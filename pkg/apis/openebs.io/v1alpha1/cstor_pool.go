/*
Copyright 2018 The OpenEBS Authors.

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
