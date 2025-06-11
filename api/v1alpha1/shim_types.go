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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ShimSpec defines the desired state of Shim
type ShimSpec struct {
	NodeSelector    map[string]string `json:"nodeSelector,omitempty"`
	FetchStrategy   FetchStrategy     `json:"fetchStrategy"`
	RuntimeClass    RuntimeClassSpec  `json:"runtimeClass"`
	RolloutStrategy RolloutStrategy   `json:"rolloutStrategy"`
	// ContainerdRuntimeOptions is a map of containerd runtime options for the shim plugin.
	// See an example of configuring cgroup driver via runtime options: https://github.com/containerd/containerd/blob/main/docs/cri/config.md#cgroup-driver
	ContainerdRuntimeOptions map[string]string `json:"containerdRuntimeOptions,omitempty"`
}

type FetchStrategy struct {
	Type     string       `json:"type"`
	AnonHTTP AnonHTTPSpec `json:"anonHttp"`
}

type AnonHTTPSpec struct {
	Location string `json:"location"`
}

type RuntimeClassSpec struct {
	Name    string `json:"name"`
	Handler string `json:"handler"`
}

// +kubebuilder:validation:Enum=rolling;recreate
type RolloutStrategyType string

const (
	RolloutStrategyTypeRolling  RolloutStrategyType = "rolling"
	RolloutStrategyTypeRecreate RolloutStrategyType = "recreate"
)

type RolloutStrategy struct {
	Type    RolloutStrategyType `json:"type"`
	Rolling RollingSpec         `json:"rolling,omitempty"`
}

type RollingSpec struct {
	MaxUpdate int `json:"maxUpdate"`
}

// ShimStatus defines the observed state of Shim
// +operator-sdk:csv:customresourcedefinitions:type=status
type ShimStatus struct {
	Conditions     []metav1.Condition `json:"conditions,omitempty"`
	NodeCount      int                `json:"nodes"`
	NodeReadyCount int                `json:"nodesReady"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=shims,scope=Cluster
// +kubebuilder:printcolumn:JSONPath=".spec.runtimeClass.name",name=RuntimeClass,type=string
// +kubebuilder:printcolumn:JSONPath=".status.nodesReady",name=Ready,type=integer
// +kubebuilder:printcolumn:JSONPath=".status.nodes",name=Nodes,type=integer
// Shim is the Schema for the shims API
type Shim struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ShimSpec   `json:"spec,omitempty"`
	Status ShimStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ShimList contains a list of Shim
type ShimList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Shim `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Shim{}, &ShimList{})
}
