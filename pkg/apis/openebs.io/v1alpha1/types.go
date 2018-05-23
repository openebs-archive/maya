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

// StoragePoolClaimSpec is the spec for a StoragePoolClaimSpec resource
type StoragePoolClaimSpec struct {
	Name       string `json:"name"`
	Format     string `json:"format"`
	Mountpoint string `json:"mountpoint"`
	Path       string `json:"path"`
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
	Name       string `json:"name"`
	Format     string `json:"format"`
	Mountpoint string `json:"mountpoint"`
	Nodename   string `json:"nodename"`
	Message    string `json:"message"`
	Path       string `json:"path"`
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
// +resource:path=volumeparametergroup

// VolumeParameterGroup describes a VolumeParameterGroup
type VolumeParameterGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec VolumeParameterGroupSpec `json:"spec"`
}

// VolumeParameterGroupSpec is the specifications for a VolumeParameterGroup resource
type VolumeParameterGroupSpec struct {
	// Policies are a list of policies to be applied on
	// the tasks during policy execution
	Policies []Policy `json:"policies"`
	// RunTasks refers to a set of tasks to be run
	RunTasks RunTasks `json:"run"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=volumeparametergroups

// VolumeParameterGroupList is a list of VolumeParameterGroup resources
type VolumeParameterGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []VolumeParameterGroup `json:"items"`
}

// Policy is the structure that defines a volume policy
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
	// TemplateNamespace is the namespace where the tasks
	// are expected to be found
	//
	// NOTE:
	//  There are two types of namespaces possible in volume parameter group
	// i.e. TemplateNamespace & RunNamespace. A RunNamespace is the
	// namespace where tasks are expected to be run.
	TemplateNamespace string `json:"templateNamespace"`
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
	Identity string `json:"id"`
	// APIVersion is the version related to the task's actual
	// content
	APIVersion string `json:"apiVersion"`
	// Kind is the kind corresponding to the task's actual content
	Kind string `json:"kind"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced
// +resource:path=cstorpool

// CStorPool describes a cstor pool resource created as custom resource.
type CStorPool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CStorPoolSpec   `json:"spec"`
	Status CStorPoolStatus `json:"status"`
}

// CStorPoolSpec is the spec listing fields for a CStorPool resource.
type CStorPoolSpec struct {
	Disks    DiskAttr      `json:"disks"`
	PoolSpec CStorPoolAttr `json:"poolSpec"`
}

// DiskAttr stores the disk related attributes.
type DiskAttr struct {
	DiskList []string `json:"diskList"`
}

// CStorPoolAttr is to describe zpool related attributes.
type CStorPoolAttr struct {
	CacheFile        string `json:"cacheFile"`        //optional, faster if specified
	PoolType         string `json:"poolType"`         //mirror, striped
	OverProvisioning bool   `json:"overProvisioning"` //true or false
}

// CStorPoolStatus is for handling status of pool.
type CStorPoolStatus struct {
	Phase string `json:"phase"` //init/online/offline/deletion-failed
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=cstorpools

// CStorPoolList is a list of CStorPoolList resources
type CStorPoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []CStorPool `json:"items"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=cstorvolumereplica
// +genclient:nonNamespaced

// CStorVolumeReplica describes a cstor pool resource created as custom resource
type CStorVolumeReplica struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              CStorVolumeReplicaSpec `json:"spec"`
}

// CStorVolumeReplicaSpec is the spec for a CStorVolumeReplica resource
type CStorVolumeReplicaSpec struct {
	CStorControllerIP string `json:"cStorControllerIP"`
	VolName           string `json:"volName"`
	Capacity          string `json:"capacity"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=cstorvolumereplicas

// CStorVolumeReplicaList is a list of CStorVolumeReplica resources
type CStorVolumeReplicaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []CStorVolumeReplica `json:"items"`
}
