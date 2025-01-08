/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"

	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	nodev1 "k8s.io/api/node/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	rcmv1 "github.com/spinkube/runtime-class-manager/api/v1alpha1"
)

const (
	RCMOperatorFinalizer          = "rcm.spinkube.dev/finalizer"
	INSTALL                       = "install"
	UNINSTALL                     = "uninstall"
	ProvisioningStatusProvisioned = "provisioned"
	ProvisioningStatusPending     = "pending"
	K8sNameMaxLength              = 63
)

// ShimReconciler reconciles a Shim object
type ShimReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// configuration for INSTALL or UNINSTALL jobs
type opConfig struct {
	operation     string
	privileged    bool
	initContainer []corev1.Container
	args          []string
}

//+kubebuilder:rbac:groups=runtime.spinkube.dev,resources=shims,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=runtime.spinkube.dev,resources=shims/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=runtime.spinkube.dev,resources=shims/finalizers,verbs=update

// SetupWithManager sets up the controller with the Manager.
func (sr *ShimReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&rcmv1.Shim{}).
		// As we create and own the created jobs
		// Jobs are important for us to update the Shims installation status
		// on respective nodes
		Owns(&batchv1.Job{}).
		// As we don't own nodes, but need to react on node label changes,
		// we need to watch node label changes.
		// Whenever a label changes, we want to reconcile Shims, to make sure
		// that the shim is deployed on the node if it should be.
		Watches(
			&corev1.Node{},
			handler.EnqueueRequestsFromMapFunc(sr.findShimsToReconcile),
			builder.WithPredicates(predicate.LabelChangedPredicate{}),
		).
		Complete(sr)
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Shim object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (sr *ShimReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.With().Str("shim", req.Name).Logger()
	ctx = log.WithContext(ctx)

	// 1. Check if the shim resource exists
	var shimResource rcmv1.Shim
	if err := sr.Client.Get(ctx, req.NamespacedName, &shimResource); err != nil {
		log.Err(err).Msg("Unable to fetch shimResource")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Ensure the finalizer is called even if a return happens before
	defer func() {
		err := sr.ensureFinalizerForShim(ctx, &shimResource, RCMOperatorFinalizer)
		if err != nil {
			log.Error().Msgf("Failed to ensure finalizer: %s", err)
		}
	}()

	// 2. Get list of nodes where this shim is supposed to be deployed on
	nodes, err := sr.getNodeListFromShimsNodeSelector(ctx, &shimResource)
	if err != nil {
		return ctrl.Result{}, err
	}

	err = sr.updateStatus(ctx, &shimResource, nodes)
	if err != nil {
		log.Error().Msgf("Unable to update node count: %s", err)
		return ctrl.Result{}, err
	}

	// Shim has been requested for deletion, delete the child resources
	if !shimResource.DeletionTimestamp.IsZero() {
		log.Debug().Msgf("Deleting shim %s", shimResource.Name)
		err := sr.handleDeleteShim(ctx, &shimResource, nodes)
		if err != nil {
			return ctrl.Result{}, err
		}

		err = sr.removeFinalizerFromShim(ctx, &shimResource)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 3. Check if referenced runtimeClass exists in cluster
	rcExists, err := sr.runtimeClassExists(ctx, &shimResource)
	if err != nil {
		log.Error().Msgf("RuntimeClass issue: %s", err)
	}
	if !rcExists {
		log.Info().Msgf("RuntimeClass '%s' not found", shimResource.Spec.RuntimeClass.Name)
		_, err = sr.handleDeployRuntimeClass(ctx, &shimResource)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	// 4. Deploy job to each node in list
	if len(nodes.Items) > 0 {
		_, err = sr.handleInstallShim(ctx, &shimResource, nodes)
	} else {
		log.Info().Msg("No nodes found")
	}

	return ctrl.Result{}, err
}

// findShimsToReconcile finds all Shims that need to be reconciled.
// This function is required e.g. to react on node label changes.
// When the label of a node changes, we want to reconcile shims to make sure
// that the shim is deployed on the node if it should be.
func (sr *ShimReconciler) findShimsToReconcile(ctx context.Context, node client.Object) []reconcile.Request {
	_ = node
	shimList := &rcmv1.ShimList{}
	listOps := &client.ListOptions{
		Namespace: "",
	}
	err := sr.List(ctx, shimList, listOps)
	if err != nil {
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, len(shimList.Items))
	for i, item := range shimList.Items {
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
			},
		}
	}
	return requests
}

func (sr *ShimReconciler) updateStatus(ctx context.Context, shim *rcmv1.Shim, nodes *corev1.NodeList) error {
	log := log.Ctx(ctx)

	shim.Status.NodeCount = len(nodes.Items)
	shim.Status.NodeReadyCount = 0

	if len(nodes.Items) > 0 {
		for _, node := range nodes.Items {
			if node.Labels[shim.Name] == ProvisioningStatusProvisioned {
				shim.Status.NodeReadyCount++
			}
		}
	}

	// TODO: include proper status conditions to update

	if err := sr.Update(ctx, shim); err != nil {
		log.Error().Msgf("Unable to update status %s", err)
	}

	// Re-fetch shim to avoid "object has been modified" errors
	if err := sr.Client.Get(ctx, types.NamespacedName{Name: shim.Name, Namespace: shim.Namespace}, shim); err != nil {
		log.Error().Msgf("Unable to re-fetch shim: %s", err)
		return fmt.Errorf("failed to fetch shim: %w", err)
	}

	return nil
}

// handleInstallShim deploys a Job to each node in a list.
func (sr *ShimReconciler) handleInstallShim(ctx context.Context, shim *rcmv1.Shim, nodes *corev1.NodeList) (ctrl.Result, error) {
	log := log.Ctx(ctx)

	switch shim.Spec.RolloutStrategy.Type {
	case rcmv1.RolloutStrategyTypeRolling:
		{
			log.Debug().Msgf("Rolling strategy selected: maxUpdate=%d", shim.Spec.RolloutStrategy.Rolling.MaxUpdate)
			return ctrl.Result{}, errors.New("rolling strategy not implemented yet")
		}
	case rcmv1.RolloutStrategyTypeRecreate:
		{
			log.Debug().Msgf("Recreate strategy selected")
			return sr.recreateStrategyRollout(ctx, shim, nodes)
		}
	default:
		{
			log.Debug().Msgf("No rollout strategy selected; using default: recreate")
			return sr.recreateStrategyRollout(ctx, shim, nodes)
		}
	}
}

func (sr *ShimReconciler) recreateStrategyRollout(ctx context.Context, shim *rcmv1.Shim, nodes *corev1.NodeList) (ctrl.Result, error) {
	log := log.Ctx(ctx)
	shimInstallationErrors := []error{}
	for i := range nodes.Items {
		node := nodes.Items[i]

		shimProvisioned := node.Labels[shim.Name] == ProvisioningStatusProvisioned
		shimPending := node.Labels[shim.Name] == ProvisioningStatusPending
		if !shimProvisioned && !shimPending {
			err := sr.deployJobOnNode(ctx, shim, node, INSTALL)
			shimInstallationErrors = append(shimInstallationErrors, err)
		}

		if shimProvisioned {
			log.Info().Msgf("Shim %s already provisioned on Node %s", shim.Name, node.Name)
		}
	}
	return ctrl.Result{}, errors.Join(shimInstallationErrors...)
}

// deployUninstallJob deploys an uninstall Job for a Shim.
func (sr *ShimReconciler) deployJobOnNode(ctx context.Context, shim *rcmv1.Shim, node corev1.Node, jobType string) error {
	log := log.Ctx(ctx)

	if err := sr.Client.Get(ctx, types.NamespacedName{Name: node.Name}, &node); err != nil {
		log.Error().Msgf("Unable to re-fetch node: %s", err)
		return fmt.Errorf("failed to fetch node: %w", err)
	}

	log.Info().Msgf("Deploying %s-Job for Shim %s on node: %s", jobType, shim.Name, node.Name)

	var job *batchv1.Job

	switch jobType {
	case INSTALL:
		err := sr.updateNodeLabels(ctx, &node, shim, ProvisioningStatusPending)
		if err != nil {
			log.Error().Msgf("Unable to update node label %s: %s", shim.Name, err)
		}

		job, err = sr.createJobManifest(shim, &node, INSTALL)
		if err != nil {
			return err
		}
	case UNINSTALL:
		err := sr.updateNodeLabels(ctx, &node, shim, UNINSTALL)
		if err != nil {
			log.Error().Msgf("Unable to update node label %s: %s", shim.Name, err)
		}

		job, err = sr.createJobManifest(shim, &node, UNINSTALL)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid jobType: %s", jobType)
	}

	// We want to use server-side apply https://kubernetes.io/docs/reference/using-api/server-side-apply
	patchMethod := client.Apply
	patchOptions := &client.PatchOptions{
		Force:        ptr(true), // Force b/c any fields we are setting need to be owned by the spin-operator
		FieldManager: "shim-operator",
	}

	// We rely on controller-runtime to rate limit us.
	if err := sr.Client.Patch(ctx, job, patchMethod, patchOptions); err != nil {
		log.Error().Msgf("Unable to reconcile Job: %s", err)
		if err := sr.updateNodeLabels(ctx, &node, shim, "failed"); err != nil {
			log.Error().Msgf("Unable to update node label %s: %s", shim.Name, err)
		}
		return fmt.Errorf("failed to reconcile job: %w", err)
	}

	return nil
}

func (sr *ShimReconciler) updateNodeLabels(ctx context.Context, node *corev1.Node, shim *rcmv1.Shim, status string) error {
	node.Labels[shim.Name] = status

	if err := sr.Update(ctx, node); err != nil {
		return fmt.Errorf("failed to update node labels: %w", err)
	}

	return nil
}

// setOperationConfiguration sets operation specific configuration for the job manifest
func (sr *ShimReconciler) setOperationConfiguration(shim *rcmv1.Shim, opConfig *opConfig) {
	if opConfig.operation == INSTALL {
		opConfig.initContainer = []corev1.Container{{
			Image: os.Getenv("SHIM_DOWNLOADER_IMAGE"),
			Name:  "downloader",
			SecurityContext: &corev1.SecurityContext{
				Privileged: &opConfig.privileged,
			},
			Env: []corev1.EnvVar{
				{
					Name:  "SHIM_NAME",
					Value: shim.Name,
				},
				{
					Name:  "SHIM_LOCATION",
					Value: shim.Spec.FetchStrategy.AnonHTTP.Location,
				},
			},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "shim-download",
					MountPath: "/assets",
				},
			},
		}}
		opConfig.args = []string{
			"install",
			"-H",
			"/mnt/node-root",
			"-r",
			shim.Name,
		}
	}

	if opConfig.operation == UNINSTALL {
		opConfig.initContainer = nil
		opConfig.args = []string{
			"uninstall",
			"-H",
			"/mnt/node-root",
			"-r",
			shim.Name,
		}
	}
}

// createJobManifest creates a Job manifest for a Shim.
func (sr *ShimReconciler) createJobManifest(shim *rcmv1.Shim, node *corev1.Node, operation string) (*batchv1.Job, error) {
	opConfig := opConfig{
		operation:  operation,
		privileged: true,
	}
	sr.setOperationConfiguration(shim, &opConfig)

	name := node.Name + "-" + shim.Name + "-" + operation
	nameMax := int(math.Min(float64(len(name)), K8sNameMaxLength))

	job := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1",
			Kind:       "Job",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name[:nameMax],
			Namespace: os.Getenv("CONTROLLER_NAMESPACE"),
			Annotations: map[string]string{
				"spinkube.dev/nodeName":  node.Name,
				"spinkube.dev/shimName":  shim.Name,
				"spinkube.dev/operation": operation,
			},
			Labels: map[string]string{
				name[:nameMax]:           "true",
				"spinkube.dev/shimName":  shim.Name,
				"spinkube.dev/operation": operation,
				"spinkube.dev/job":       "true",
			},
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					NodeName: node.Name,
					HostPID:  true,
					Volumes: []corev1.Volume{
						{
							Name: "shim-download",
						},
						{
							Name: "root-mount",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/",
								},
							},
						},
					},
					InitContainers: opConfig.initContainer,
					Containers: []corev1.Container{{
						Image: os.Getenv("SHIM_NODE_INSTALLER_IMAGE"),
						Args:  opConfig.args,
						Name:  "provisioner",
						SecurityContext: &corev1.SecurityContext{
							Privileged: &opConfig.privileged,
						},
						Env: []corev1.EnvVar{
							{
								Name:  "HOST_ROOT",
								Value: "/mnt/node-root",
							},
							{
								Name:  "SHIM_FETCH_STRATEGY",
								Value: "/mnt/node-root",
							},
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "root-mount",
								MountPath: "/mnt/node-root",
							},
							{
								Name:      "shim-download",
								MountPath: "/assets",
							},
						},
					}},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
		},
	}
	// set ttl for the installer job only if specified by the user
	if ttlStr := os.Getenv("SHIM_NODE_INSTALLER_JOB_TTL"); ttlStr != "" {
		if ttl, err := strconv.ParseInt(ttlStr, 10, 32); err == nil && ttl > 0 {
			job.Spec.TTLSecondsAfterFinished = ptr(int32(ttl))
		}
	}
	if operation == INSTALL {
		if err := ctrl.SetControllerReference(shim, job, sr.Scheme); err != nil {
			return nil, fmt.Errorf("failed to set controller reference: %w", err)
		}
	}

	return job, nil
}

// handleDeployRuntimeClass deploys a RuntimeClass for a Shim.
func (sr *ShimReconciler) handleDeployRuntimeClass(ctx context.Context, shim *rcmv1.Shim) (ctrl.Result, error) {
	log := log.Ctx(ctx)

	log.Info().Msgf("Deploying RuntimeClass: %s", shim.Spec.RuntimeClass.Name)
	runtimeClass, err := sr.createRuntimeClassManifest(shim)
	if err != nil {
		return ctrl.Result{}, err
	}

	// We want to use server-side apply https://kubernetes.io/docs/reference/using-api/server-side-apply
	patchMethod := client.Apply
	patchOptions := &client.PatchOptions{
		Force:        ptr(true), // Force b/c any fields we are setting need to be owned by the spin-operator
		FieldManager: "shim-operator",
	}

	// Note that we reconcile even if the deployment is in a good state. We rely on controller-runtime to rate limit us.
	if err := sr.Client.Patch(ctx, runtimeClass, patchMethod, patchOptions); err != nil {
		log.Error().Msgf("Unable to reconcile RuntimeClass %s", err)
		return ctrl.Result{}, fmt.Errorf("failed to reconcile RuntimeClass: %w", err)
	}

	return ctrl.Result{}, nil
}

// createRuntimeClassManifest creates a RuntimeClass manifest for a Shim.
func (sr *ShimReconciler) createRuntimeClassManifest(shim *rcmv1.Shim) (*nodev1.RuntimeClass, error) {
	name := shim.Spec.RuntimeClass.Name
	nameMax := int(math.Min(float64(len(name)), K8sNameMaxLength))

	nodeSelector := shim.Spec.NodeSelector
	if nodeSelector == nil {
		nodeSelector = map[string]string{}
	}

	runtimeClass := &nodev1.RuntimeClass{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "node.k8s.io/v1",
			Kind:       "RuntimeClass",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   name[:nameMax],
			Labels: map[string]string{name[:nameMax]: "true"},
		},
		Handler: shim.Spec.RuntimeClass.Handler,
		Scheduling: &nodev1.Scheduling{
			NodeSelector: nodeSelector,
		},
	}

	if err := ctrl.SetControllerReference(shim, runtimeClass, sr.Scheme); err != nil {
		return nil, fmt.Errorf("failed to set controller reference: %w", err)
	}

	return runtimeClass, nil
}

// handleDeleteShim deletes all possible child resources of a Shim. It will ignore NotFound errors.
func (sr *ShimReconciler) handleDeleteShim(ctx context.Context, shim *rcmv1.Shim, nodes *corev1.NodeList) error {
	// deploy uninstall job on every node in node list
	for i := range nodes.Items {
		node := nodes.Items[i]

		if _, exists := node.Labels[shim.Name]; exists {
			err := sr.deployJobOnNode(ctx, shim, node, UNINSTALL)
			if client.IgnoreNotFound(err) != nil {
				return err
			}
		} else {
			log.Info().Msgf("Shim %s has no label on Node %s", shim.Name, node.Name)
		}
	}
	return nil
}

func (sr *ShimReconciler) getNodeListFromShimsNodeSelector(ctx context.Context, shim *rcmv1.Shim) (*corev1.NodeList, error) {
	nodes := &corev1.NodeList{}
	if shim.Spec.NodeSelector != nil {
		err := sr.List(ctx, nodes, client.MatchingLabels(shim.Spec.NodeSelector))
		if err != nil {
			return &corev1.NodeList{}, fmt.Errorf("failed to get node list: %w", err)
		}
	} else {
		err := sr.List(ctx, nodes)
		if err != nil {
			return &corev1.NodeList{}, fmt.Errorf("failed to get node list: %w", err)
		}
	}

	return nodes, nil
}

// runtimeClassExists checks whether a RuntimeClass for a Shim exists.
func (sr *ShimReconciler) runtimeClassExists(ctx context.Context, shim *rcmv1.Shim) (bool, error) {
	log := log.Ctx(ctx)

	if shim.Spec.RuntimeClass.Name != "" {
		runtimeClass, err := sr.getRuntimeClass(ctx, shim)
		if err != nil {
			log.Debug().Msgf("No RuntimeClass '%s' found", shim.Spec.RuntimeClass.Name)

			return false, err
		}
		log.Debug().Msgf("RuntimeClass found: %s", runtimeClass.Name)
		return true, nil
	}
	log.Debug().Msg("Shim.Spec.RuntimeClass not defined")
	return false, nil
}

// getRuntimeClass finds a RuntimeClass.
func (sr *ShimReconciler) getRuntimeClass(ctx context.Context, shim *rcmv1.Shim) (*nodev1.RuntimeClass, error) {
	rc := nodev1.RuntimeClass{}
	err := sr.Client.Get(ctx, types.NamespacedName{Name: shim.Spec.RuntimeClass.Name, Namespace: shim.Namespace}, &rc)
	if err != nil {
		return nil, fmt.Errorf("failed to get runtimeClass: %w", err)
	}
	return &rc, nil
}

// removeFinalizerFromShim removes the finalizer from a Shim.
func (sr *ShimReconciler) removeFinalizerFromShim(ctx context.Context, shim *rcmv1.Shim) error {
	if controllerutil.ContainsFinalizer(shim, RCMOperatorFinalizer) {
		controllerutil.RemoveFinalizer(shim, RCMOperatorFinalizer)
		if err := sr.Client.Update(ctx, shim); err != nil {
			return fmt.Errorf("failed to remove finalizer: %w", err)
		}
	}
	return nil
}

// ensureFinalizerForShim ensures the finalizer is present on a Shim resource.
func (sr *ShimReconciler) ensureFinalizerForShim(ctx context.Context, shim *rcmv1.Shim, finalizer string) error {
	if !controllerutil.ContainsFinalizer(shim, finalizer) {
		controllerutil.AddFinalizer(shim, finalizer)
		if err := sr.Client.Update(ctx, shim); err != nil {
			return fmt.Errorf("failed to set finalizer: %w", err)
		}
	}
	return nil
}

func ptr[T any](v T) *T {
	return &v
}
