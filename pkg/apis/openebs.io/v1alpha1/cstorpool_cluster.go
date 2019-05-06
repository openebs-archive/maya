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

// +genclient
// +genclient:noStatus
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=cstorpoolcluster

// CStorPoolCluster describes a CStorPoolCluster custom resource.
type CStorPoolCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              CStorPoolClusterSpec   `json:"spec"`
	Status            CStorPoolClusterStatus `json:"status"`
}

// CStorPoolClusterSpec is the spec for a CStorPoolClusterSpec resource
type CStorPoolClusterSpec struct {
	Name     string                     `json:"name"`
	Type     string                     `json:"type"`
	MaxPools *int                       `json:"maxPools"`
	MinPools int                        `json:"minPools"`
	Nodes    []CStorPoolClusterNodeSpec `json:"nodes"`
	PoolSpec CStorPoolAttr              `json:"poolSpec"`
}

// CStorPoolClusterNodeSpec is the spec for node where pool should be created.
type CStorPoolClusterNodeSpec struct {
	// Name is the name of the node.
	Name string `json:"name"`
	// PoolSpec is the pool related specification that is used to provision cstor pool on the node.
	PoolSpec CStorPoolAttr `json:"poolSpec"`
	// DiskGroups contains the list of disk groups that should be used for pool provisioning on that node.
	DiskGroups []CStorPoolClusterDiskGroups `json:"groups"`
}

// CStorPoolClusterDiskGroups contains details of a disk group.
type CStorPoolClusterDiskGroups struct {
	// Name is the name of the disk group.
	Name string `json:"name"`
	// Disks is the disks present in the group.
	Disks []CStorPoolClusterDisk `json:"disks"`
}

// CStorPoolClusterDisk contains the name details of a disk CR.
type CStorPoolClusterDisk struct {
	// Name is the name of the disk CR.
	Name string `json:"name"`
	// ID is a unique id associated to this disk either by user or system.
	// This ID does not change even if the disk is replaced so one can think it as a slot for this disk.
	ID string `json:"id"`
}

// CStorPoolClusterStatus is for handling status of pool.
type CStorPoolClusterStatus struct {
	Phase string `json:"phase"`
}

// CStorPoolClusterSpecAttr is to describe zpool related attributes.
type CStorPoolClusterSpecAttr struct {
	CacheFile        string `json:"cacheFile"`        //optional, faster if specified
	PoolType         string `json:"poolType"`         //mirrored, striped
	OverProvisioning bool   `json:"overProvisioning"` //true or false
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=cstorpoolclusters

// CStorPoolClusterList is a list of CStorPoolCluster resources
type CStorPoolClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []CStorPoolCluster `json:"items"`
}
