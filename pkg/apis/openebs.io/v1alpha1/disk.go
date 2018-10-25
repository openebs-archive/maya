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

// DiskStatus provides current state of the disk (Active/Inactive)
type DiskStatus struct {
	State string `json:"state"`
}

// DiskCapacity provides disk size in byte
type DiskCapacity struct {
	Storage uint64 `json:"storage"`
}

// DiskDetails contains basic and static info of a disk
type DiskDetails struct {
	Model  string `json:"model"`  // Model is model of disk
	Serial string `json:"serial"` // Serial is serial no of disk
	Vendor string `json:"vendor"` // Vendor is vendor of disk
}

// DiskDevlink holds the mapping between type and links like by-id type or by-path type link
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
