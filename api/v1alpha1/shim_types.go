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
	// Type is the fetch strategy type.
	//
	// Deprecated: this field is ignored by the controller and exists only
	// for backward compatibility with existing manifests that specify it.
	//
	// +optional
	Type string `json:"type,omitempty"`

	// AnonHTTP fetches a binary from a public HTTP(S) URL.
	// For backward compatibility with single-architecture deployments.
	// When Platforms is also specified, Platforms takes precedence.
	// +optional
	AnonHTTP *AnonHTTPSpec `json:"anonHttp,omitempty"`

	// Platforms lists per-OS/architecture artifact sources.
	// The controller selects the matching entry for each target node.
	// When specified, this takes precedence over AnonHTTP.
	// +optional
	Platforms []PlatformArtifact `json:"platforms,omitempty"`
}

// AnonHTTPSpec defines a simple anonymous HTTP fetch (single URL, single architecture).
type AnonHTTPSpec struct {
	// Location is the direct URL to the artifact archive.
	Location string `json:"location"`
}

// PlatformArtifact maps a specific OS/Arch pair to an artifact URL.
type PlatformArtifact struct {
	// OS is the operating system. Currently only "Linux" is supported.
	// +kubebuilder:validation:Enum=linux
	OS string `json:"os"`
	// Arch is the CPU architecture.
	// Accepts Go-style ("amd64", "arm64") or uname-style ("x86_64", "aarch64").
	// +kubebuilder:validation:Enum=amd64;arm64;x86_64;aarch64
	Arch string `json:"arch"`
	// Location is the URL to the artifact archive for this platform. Must be publicly accessible.
	Location string `json:"location"`
	// SHA256 is the optional hex-encoded SHA-256 digest for verification.
	// +optional
	SHA256 string `json:"sha256,omitempty"`
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
