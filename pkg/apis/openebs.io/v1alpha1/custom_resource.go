/*
Copyright 2017 The OpenEBS Authors.

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

// +genclient
// +genclient:noStatus
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=storagepoolclaim

// StoragePoolClaim describes a StoragePoolClaim custom resource.
type StoragePoolClaim struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec StoragePoolClaimSpec `json:"spec"`
	Status StoragePoolClaimStatus `json:"status"`

}
// StoragePoolClaimSpec is the spec for a StoragePoolClaimSpec resource
type StoragePoolClaimSpec struct {
	Name         string `json:"name"`
	Format       string `json:"format"`
	Mountpoint	 string `json:"mountpoint"`
	Path       	 string `json:"path"`
	Type       	 string `json:"type"`
	NodeSelector []string `json:"nodeSelector"`
	Capacity     string `json:"capacity"`
	MaxPools     int16 `json:"maxPools"`
	Disks DiskAttr `json:"disks"`
	PoolSpec CStorPoolAttr `json:"poolSpec"`
}
// StoragePoolClaim is for handling status of pool.
type StoragePoolClaimStatus struct {
	Phase string `json:"phase"`
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
	Disks    DiskAttr      `json:"disks"`
	PoolSpec CStorPoolAttr `json:"poolSpec"`

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
// +resource:path=disk

// Disk describes disk resource.
type Disk struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata, omitempty"`

	Spec   DiskSpec   `json:"spec"`
	Status DiskStatus `json:"status"`
}

// DiskSpec is the specification for the disk stored as CRD
type DiskSpec struct {
	Path     string        `json:"path"`               //Path contain devpath (e.g. /dev/sdb)
	Capacity DiskCapacity  `json:"capacity"`           //Capacity
	Details  DiskDetails   `json:"details"`            //Details contains static attributes (model, serial ..)
	DevLinks []DiskDevLink `json:"devlinks,omitempty"` //DevLinks contains soft links of one disk
}

type DiskStatus struct {
	State string `json:"state"` //current state of the disk (Active/Inactive)
}

type DiskCapacity struct {
	Storage uint64 `json:"storage"` //disk size in byte
}

// DiskDetails contains basic and static info of a disk
type DiskDetails struct {
	Model  string `json:"model"`  // Model is model of disk
	Serial string `json:"serial"` // Serial is serial no of disk
	Vendor string `json:"vendor"` // Vendor is vendor of disk
}

// DiskDevlink holds the maping between type and links like by-id type or by-path type link
type DiskDevLink struct {
	Kind  string   `json:"kind,omitempty"`  // Kind is the type of link like by-id or by-path.
	Links []string `json:"links,omitempty"` // Links are the soft links of Type type
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=disks

// DiskList is a list of Disk object resources
type DiskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Disk `json:"items"`
}

// +genclient
// +genclient:noStatus
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=castemplate

// CASTemplate describes a Container Attached Storage template that is used
// to provision a CAS volume
type CASTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec CASTemplateSpec `json:"spec"`
}

// CASTemplateSpec is the specifications for a CASTemplate resource
type CASTemplateSpec struct {
	// Update specifications to update a CAS volume
	Update CASUpdateSpec `json:"update"`
	// Defaults are a list of default configurations that may be applied
	// during provisioning of a CAS volume
	Defaults []Config `json:"defaultConfig"`
	// TaskNamespace is the namespace where the tasks
	// are expected to be found
	TaskNamespace string `json:"taskNamespace"`
	// RunTasks refers to a set of tasks to be run
	RunTasks RunTasks `json:"run"`
	// OutputTask is the task that has the CAS template result's output
	// format
	OutputTask string `json:"output"`
}

// CASUpdateSpec is the specification to update a CAS volume
// One or more CAS volumes may be updated at a time based on PVC or SC
// respectively.
type CASUpdateSpec struct {
	// Kind refers to a Kubernetes kind. In this case it can be a
	// StorageClass or a PVC
	Kind string `json:"kind"`
	// Name refers to the Kubernetes resource. In this case it can
	// be the name of StorageClass or PVC.
	Name string `json:"name"`
	// Selector filters the relevant CAS volumes to be updated
	Selector string `json:"selector"`
	// CurrentVersion is the expected current version of a CAS volume
	// that is eligible for update
	CurrentVersion string `json:"currentVersion"`
	// DesiredVersion is the desired version of a CAS volume after a
	// successful update
	DesiredVersion string `json:"desiredVersion"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=castemplates

// CASTemplateList is a list of CASTemplate resources
type CASTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []CASTemplate `json:"items"`
}

// Config holds a configuration element
//
// For example, it can represent a config property of a CAS volume
type Config struct {
	// Name of the config
	Name string `json:"name"`
	// Enabled flags if this config is enabled or disabled;
	// true indicates enabled while false indicates disabled
	Enabled string `json:"enabled"`
	// Value represents any specific value that is applicable
	// to this config
	Value string `json:"value"`
	// Data represents an arbitrary map of key value pairs
	Data map[string]string `json:"data"`
}

// RunTasks contains fields to run a set of
// tasks
type RunTasks struct {
	// Items is a set of order-ed tasks
	Tasks []string `json:"tasks"`
}

// Task has information about an action and a resource where the action
// is performed against the resource.
//
// For example a resource can be a kubernetes resource and the corresponding
// action can be to apply this resource to kubernetes cluster.
//type Task struct {
// TaskName is the name of the task.
//
// NOTE: A task refers to a K8s ConfigMap.
//TaskName string `json:"task"`
// Identity is the unique identity that can differentiate
// two tasks even when using the same template
//Identity string `json:"id"`
// APIVersion is the version related to the task's resource
//APIVersion string `json:"apiVersion"`
// Kind is the kind corresponding to the task's resource
//Kind string `json:"kind"`
//}

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

type CStorPoolPhase string

// Status written onto CStorPool and CStorVolumeReplica objects.
const (
	// CStorPoolStatusEmpty ensures the create operation is to be done, if import fails.
	CStorPoolStatusEmpty CStorPoolPhase = ""
	// CStorPoolStatusOnline ensures the resource is available.
	CStorPoolStatusOnline CStorPoolPhase = "Online"
	// CStorPoolStatusOffline ensures the resource is not available.
	CStorPoolStatusOffline CStorPoolPhase = "Offline"
	// CStorPoolStatusDeletionFailed ensures the resource deletion has failed.
	CStorPoolStatusDeletionFailed CStorPoolPhase = "DeletionFailed"
	// CStorPoolStatusInvalid ensures invalid resource.
	CStorPoolStatusInvalid CStorPoolPhase = "Invalid"
	// CStorPoolStatusErrorDuplicate ensures error due to duplicate resource.
	CStorPoolStatusErrorDuplicate CStorPoolPhase = "ErrorDuplicate"
	// CStorPoolStatusPending ensures pending task for cstorpool.
	CStorPoolStatusPending CStorPoolPhase = "Pending"
)

// CStorPoolStatus is for handling status of pool.
type CStorPoolStatus struct {
	Phase CStorPoolPhase `json:"phase"`
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

// CStorVolumeReplica describes a cstor volume resource created as custom resource
type CStorVolumeReplica struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              CStorVolumeReplicaSpec   `json:"spec"`
	Status            CStorVolumeReplicaStatus `json:"status"`
}

// CStorVolumeReplicaSpec is the spec for a CStorVolumeReplica resource
type CStorVolumeReplicaSpec struct {
	TargetIP string `json:"targetIP"`
	Capacity string `json:"capacity"`
}

// CStorVolumeReplicaPhase is to hold result of action.
type CStorVolumeReplicaPhase string

// Status written onto CStorVolumeReplica objects.
const (
	// CVRStatusEmpty ensures the create operation is to be done, if import fails.
	CVRStatusEmpty CStorVolumeReplicaPhase = ""
	// CVRStatusOnline ensures the resource is available.
	CVRStatusOnline CStorVolumeReplicaPhase = "Online"
	// CVRStatusOffline ensures the resource is not available.
	CVRStatusOffline CStorVolumeReplicaPhase = "Offline"
	// CVRStatusDeletionFailed ensures the resource deletion has failed.
	CVRStatusDeletionFailed CStorVolumeReplicaPhase = "DeletionFailed"
	// CVRStatusInvalid ensures invalid resource.
	CVRStatusInvalid CStorVolumeReplicaPhase = "Invalid"
	// CVRStatusErrorDuplicate ensures error due to duplicate resource.
	CVRStatusErrorDuplicate CStorVolumeReplicaPhase = "ErrorDuplicate"
	// CVRStatusPending ensures pending task of cvr resource.
	CVRStatusPending CStorVolumeReplicaPhase = "Pending"
)

// CStorVolumeReplicaStatus is for handling status of cvr.
type CStorVolumeReplicaStatus struct {
	Phase CStorVolumeReplicaPhase `json:"phase"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=cstorvolumereplicas

// CStorVolumeReplicaList is a list of CStorVolumeReplica resources
type CStorVolumeReplicaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []CStorVolumeReplica `json:"items"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=cstorvolume

// CStorVolume describes a cstor volume resource created as custom resource
type CStorVolume struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              CStorVolumeSpec `json:"spec"`
}

// CStorVolumeSpec is the spec for a CStorVolume resource
type CStorVolumeSpec struct {
	Capacity     string `json:"capacity"`
	TargetIP     string `json:"targetIP"`
	TargetPort   string `json:"targetPort"`
	Iqn          string `json:"iqn"`
	TargetPortal string `json:"targetPortal"`
	Status       string `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=cstorvolume

// CStorVolumeReplicaList is a list of CStorVolume resources
type CStorVolumeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []CStorVolumeReplica `json:"items"`
}
