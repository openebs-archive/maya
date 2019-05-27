/*
Copyright 2019 The OpenEBS Authors

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
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeviceClaimSpec defines the desired state of BlockDeviceClaim
type DeviceClaimSpec struct {
	Requirements    DeviceClaimRequirements `json:"requirements"`                 // the requirements in the claim like Capacity, IOPS
	DeviceType      string                  `json:"deviceType"`                   // DeviceType represents the type of drive like SSD, HDD etc.,
	HostName        string                  `json:"hostName"`                     // Node name from where blockdevice has to be claimed.
	Details         DeviceClaimDetails      `json:"deviceClaimDetails,omitempty"` // Details of the device to be claimed
	BlockDeviceName string                  `json:"blockDeviceName,omitempty"`    // BlockDeviceName is the reference to the block-device backing this claim
}

// DeviceClaimStatus defines the observed state of BlockDeviceClaim
type DeviceClaimStatus struct {
	Phase DeviceClaimPhase `json:"phase"`
}

// DeviceClaimPhase is a typed string for phase field of BlockDeviceClaim.
type DeviceClaimPhase string

// BlockDeviceClaim CR, when created pass through phases before it got some Devices Assigned.
// Given below table, have all phases which BlockDeviceClaim CR can go before it is marked done.
const (
	// BlockDeviceClaimStatusEmpty represents that the BlockDeviceClaim was just created.
	BlockDeviceClaimStatusEmpty DeviceClaimPhase = ""

	// BlockDeviceClaimStatusPending represents BlockDeviceClaim has not been assigned devices yet. Rather
	// search is going on for matching devices.
	BlockDeviceClaimStatusPending DeviceClaimPhase = "Pending"

	// BlockDeviceClaimStatusInvalidCapacity represents BlockDeviceClaim has invalid capacity request i.e. 0/-1
	BlockDeviceClaimStatusInvalidCapacity DeviceClaimPhase = "Invalid Capacity Request"

	// BlockDeviceClaimStatusDone represents BlockDeviceClaim has been assigned backing blockdevice and ready for use.
	BlockDeviceClaimStatusDone DeviceClaimPhase = "Bound"
)

// DeviceClaimRequirements defines the request by the claim, eg, Capacity, IOPS
type DeviceClaimRequirements struct {
	// Requests describes the minimum resources required. eg: if storage resource of 10G is
	// requested minimum capacity of 10G should be available
	Requests v1.ResourceList `json:"requests"`
}

const (
	// ResourceCapacity defines the capacity required as v1.Quantity
	ResourceCapacity v1.ResourceName = "capacity"
)

// DeviceClaimDetails defines the details of the block device that should be claimed
type DeviceClaimDetails struct {
	DeviceFormat   string `json:"formatType,omitempty"`     //Format of the device required, eg:ext4, xfs
	MountPoint     string `json:"mountPoint,omitempty"`     //MountPoint of the device required. Claim device from the specified mountpoint.
	AllowPartition bool   `json:"allowPartition,omitempty"` //AllowPartition represents whether to claim a full block device or a device that is a partition
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=blockDeviceClaim
// +k8s:openapi-gen=true

// BlockDeviceClaim is the Schema for the block device claim API
type BlockDeviceClaim struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeviceClaimSpec   `json:"spec,omitempty"`
	Status DeviceClaimStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=blockDeviceClaimList
// +k8s:openapi-gen=true

// BlockDeviceClaimList contains a list of BlockDeviceClaim
type BlockDeviceClaimList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BlockDeviceClaim `json:"items"`
}
