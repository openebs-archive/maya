package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=storagepoolclaim

// StoragePoolClaim describes a StoragePoolClaim.
type StoragePoolClaim struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec StoragePoolClaimSpec `json:"spec"`
}

// StoragePoolClaimSpec is the spec for a StoragePoolClaim resource
type StoragePoolClaimSpec struct {
	Path string `json:"path"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=storagepoolclaims

// StoragePoolClaimList is a list of StoragePoolClaim resources
type StoragePoolClaimList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []StoragePoolClaim `json:"items"`
}

// +genclient
// +genclient:noStatus
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=storagepool

// StoragePool describes a StoragePool.
type StoragePool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec StoragePoolSpec `json:"spec"`
}

// StoragePoolSpec is the spec for a StoragePool resource
type StoragePoolSpec struct {
	Path string `json:"path"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=storagepools

// StoragePoolList is a list of StoragePool resources
type StoragePoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []StoragePool `json:"items"`
}

// +genclient
// +genclient:noStatus
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=volumepolicy

// VolumePolicy describes a VolumePolicy
type VolumePolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec VolumePolicySpec `json:"spec"`
}

// VolumePolicySpec is the specifications for a VolumePolicy resource
type VolumePolicySpec struct {
	// Policies are a list of policies to be applied on
	// the tasks during policy execution
	Policies []Policy `json:"policies"`
	// RunTasks refers to a set of tasks to be run
	RunTasks RunTasks `json:"run"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=volumepolicies

// VolumePolicyList is a list of VolumePolicy resources
type VolumePolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []VolumePolicy `json:"items"`
}

// Policy defines various policy based properties
type Policy struct {
	// Name of the policy
	Name string `json:"name"`
	// Enabled flags if this policy is enabled or disabled
	// true indicates enabled while false indicates disabled
	Enabled string `json:"enabled"`
	// Value represents any specific value that is relevant
	// to this policy
	Value string `json:"value"`
	// Data represents an arbitrary map of key value pairs
	Data map[string]string `json:"data"`
}

// RunTasks contains fields to run a set of
// tasks
type RunTasks struct {
	// SearchNamespace is the namespace where the tasks
	// are expected to be found
	//
	// NOTE:
	//  There are two types of namespaces possible in volume policy
	// i.e. SearchNamespace & RunNamespace. A RunNamespace is the
	// namespace where tasks are expected to be run.
	SearchNamespace string `json:"searchNamespace"`
	// Items is a set of order-ed tasks
	Tasks []Task `json:"tasks"`
}

// Task has information about a task
type Task struct {
	// TemplateName is the name of the template. A template
	// represents a task in yaml format. A template refers to
	// a K8s ConfigMap.
	TemplateName string `json:"template"`
	// Identity is the unique identity that can differentiate
	// two tasks even when using the same template
	Identity string `json:"identity"`
	// APIVersion is the version related to the task's actual
	// content
	APIVersion string `json:"apiVersion"`
	// Kind is the kind corresponding to the task's actual content
	Kind string `json:"kind"`
}
