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
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=csivolumeinfo

// CSIVolume describes a csi volume resource created as custom resource
type CSIVolume struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              CSIVolumeSpec `json:"spec"`
}

// CSIVolumeSpec is the spec for a CStorVolume resource
type CSIVolumeSpec struct {
	// Volume part of CSIVolume contains info specific to CSIVolumes in general
	Volume VolumeInfo
	// ISCSI part of CSIVolume contains info specific to ISCSI protocol,
	// this is filled only if the volume type is iSCSI
	ISCSI ISCSIInfo
}

// VolumeInfo contains the volume related info
// for all types of volumes in CSIVolumeSpec
type VolumeInfo struct {
	// Volname of a volume will hold the name of the CSI Volume
	Volname string `json:"volname"`
	// CASType of a volume is the backend type for OpenEBS volumes
	CASType string `json:"castype"`
	// OwnerNodeID of a volume will hold the ownerNodeID of the Volume
	OwnerNodeID string `json:"ownernodeID"`
	// Capacity of a volume will hold the capacity of the Volume
	Capacity string `json:"capacity"`
	// FSType of a volume will specify the format type - ext4(default), xfs of PV
	FSType string `json:"fsType"`
	// AccessMode of a volume will hold the access mode of the volume
	AccessModes []string `json:"accessMode"`
	// MountPath of the volume will hold the path on which the volume is mounted
	// on that node
	MountPath string `json:"mountPath"`
	// ReadOnly specifies if the volume needs to be mounted in ReadOnly mode
	ReadOnly bool `json:"readOnly"`
	// MountOptions specifies the options with which mount needs to be attempted
	MountOptions []string `json:"mountOptions"`
	// Device Path specifies the device path which is returned when the iSCSI
	// login is successful
	DevicePath string `json:"devicePath"`
}

// ISCSIInfo contains info specific to ISCSI protocol,
// this can be used only if only if the volume type exposed
// by the vendor is iSCSI
type ISCSIInfo struct {
	// Iqn of a volume will hold the iqn value of the Volume
	Iqn string `json:"iqn"`
	// TargetPortal of a volume will hold the target portal of the volume
	TargetPortal string `json:"targetPortal"`
	// IscsiInterface
	IscsiInterface string `json:"iscsiInterface"`
	// Lun of volume will specify the lun number 0, 1.. on iSCSI Volume. (default: 0)
	Lun string `json:"lun"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=csivolumeinfo

// CSIVolumeList is a list of CSIVolume resources
type CSIVolumeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []CSIVolume `json:"items"`
}
