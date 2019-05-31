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
	stringer "github.com/openebs/maya/pkg/apis/stringer/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
	Group    []DiskGroup   `json:"group"`
	PoolSpec CStorPoolAttr `json:"poolSpec"`
}

// DiskGroup contains a collection of disk for a given pool topology in CSP.
type DiskGroup struct {
	// Item contains a list of CspDisks.
	Item []CspDisk `json:"disk"`
}

// CspDisk contains the details of disk present on CSP.
type CspDisk struct {
	// Name is the name of the disk resource.
	Name string `json:"name"`
	// DeviceID is the device id of the disk resource. In case of sparse disks, it contains the device path.
	DeviceID string `json:"deviceID"`
	// InUseByPool tells whether the disk is present on spc. If disk is present on SPC, it is true else false.
	InUseByPool bool `json:"inUseByPool"`
}

// BlockDeviceAttr stores the BD related attributes.
type BlockDeviceAttr struct {
	BlockDeviceList []string `json:"blockDeviceList"`
}

// CStorPoolAttr is to describe zpool related attributes.
type CStorPoolAttr struct {
	CacheFile        string `json:"cacheFile"`        //optional, faster if specified
	PoolType         string `json:"poolType"`         //mirrored, striped
	OverProvisioning bool   `json:"overProvisioning"` //true or false
}

// CStorPoolPhase is a typed string for phase field of CStorPool.
type CStorPoolPhase string

// Status written onto CStorPool and CStorVolumeReplica objects.
// Resetting state to either Empty or Pending need to be done with care,
// as, label clear and pool creation depends on this state.
const (
	// CStorPoolStatusEmpty ensures the create operation is to be done, if import fails.
	// Check comment
	CStorPoolStatusEmpty CStorPoolPhase = ""
	// CStorPoolStatusOnline signifies that the pool is online.
	CStorPoolStatusOnline CStorPoolPhase = "Healthy"
	// CStorPoolStatusOffline signifies that the pool is offline.
	CStorPoolStatusOffline CStorPoolPhase = "Offline"
	// CStorPoolStatusDegraded signifies that the pool is degraded.
	CStorPoolStatusDegraded CStorPoolPhase = "Degraded"
	// CStorPoolStatusFaulted signifies that the pool is faulted.
	CStorPoolStatusFaulted CStorPoolPhase = "Faulted"
	// CStorPoolStatusRemoved signifies that the pool is removed.
	CStorPoolStatusRemoved CStorPoolPhase = "Removed"
	// CStorPoolStatusUnavail signifies that the pool is not available.
	CStorPoolStatusUnavail CStorPoolPhase = "Unavail"
	// CStorPoolStatusDeletionFailed signifies that the pool status could not be fetched.
	CStorPoolStatusError CStorPoolPhase = "Error"
	// CStorPoolStatusDeletionFailed ensures the resource deletion has failed.
	CStorPoolStatusDeletionFailed CStorPoolPhase = "DeletionFailed"
	// CStorPoolStatusInvalid ensures invalid resource.
	CStorPoolStatusInvalid CStorPoolPhase = "Invalid"
	// CStorPoolStatusErrorDuplicate ensures error due to duplicate resource.
	CStorPoolStatusErrorDuplicate CStorPoolPhase = "ErrorDuplicate"
	// CStorPoolStatusPending ensures pending task for cstorpool.
	// Check comment
	CStorPoolStatusPending CStorPoolPhase = "Pending"
)

// CStorPoolStatus is for handling status of pool.
type CStorPoolStatus struct {
	Phase    CStorPoolPhase        `json:"phase"`
	Capacity CStorPoolCapacityAttr `json:"capacity"`
}

type CStorPoolCapacityAttr struct {
	Total string `json:"total"`
	Free  string `json:"free"`
	Used  string `json:"used"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=cstorpools

// CStorPoolList is a list of CStorPoolList resources
type CStorPoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []CStorPool `json:"items"`
}

// String implements Stringer interface
func (c *CStorPool) String() string {
	return stringer.Yaml("cstorpool", c)
}

// GoString implements GoStringer interface
func (c *CStorPool) GoString() string {
	return c.String()
}
