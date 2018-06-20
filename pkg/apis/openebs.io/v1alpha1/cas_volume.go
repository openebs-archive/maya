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

// CASVolumeKey is a typed string to represent cas volume related annotations'
// or labels' keys
//
// Example 1 - Below is a sample CASVolume that makes use of some CASVolumeKey
// constants.
//
// NOTE:
//  This specification is sent by openebs provisioner as http create request in
// its payload.
//
// ```yaml
// kind: CASVolume
// apiVersion: v1alpha1
// metadata:
//   name: jiva-cas-vol
//   # this way of setting namespace gets the first priority
//   namespace: default
//   labels:
//     # deprecated way to set capacity
//     volumeprovisioner.mapi.openebs.io/storage-size: 2G
//     k8s.io/storage-class: openebs-repaffinity-0.6.0
//     # this manner of setting namespace gets the second priority
//     k8s.io/namespace: default
//     k8s.io/pvc: openebs-repaffinity-0.6.0
// # latest way to set capacity
// capacity: 2G
// ```
//
// Example 2 - Below is a sample StorageClass that makes use of a CASVolumeKey
// constant i.e. the cas template used to create a cas volume
//
// ```yaml
// apiVersion: storage.k8s.io/v1
// kind: StorageClass
// metadata:
//  name: openebs-standard
//  annotations:
//    cas.openebs.io/create-template: cast-standard-0.6.0
// provisioner: openebs.io/provisioner-iscsi
// ```
type CASVolumeKey string

const (
	// CASTemplateCVK is the key to fetch name of CASTemplate custom resource
	// to create a cas volume
	CASTemplateCVK CASVolumeKey = "cas.openebs.io/create-template"

	// CASTemplateForReadCVK is the key to fetch name of CASTemplate custom
	// resource to read a cas volume
	CASTemplateForReadCVK CASVolumeKey = "cas.openebs.io/read-template"

	// CASTemplateForDeleteCVK is the key to fetch name of CASTemplate custom
	// resource to delete a cas volume
	CASTemplateForDeleteCVK CASVolumeKey = "cas.openebs.io/delete-template"

	// CASTemplateForListCVK is the key to fetch name of CASTemplate custom
	// resource to list cas volumes
	CASTemplateForListCVK CASVolumeKey = "cas.openebs.io/list-template"

	// CASConfigCVK is the key to fetch configurations w.r.t a CAS volume
	CASConfigCVK CASVolumeKey = "cas.openebs.io/config"

	// NamespaceCVK is the key to fetch volume's namespace
	NamespaceCVK CASVolumeKey = "openebs.io/namespace"

	// PersistentVolumeClaimCVK is the key to fetch volume's PVC
	PersistentVolumeClaimCVK CASVolumeKey = "openebs.io/pvc"

	// StorageClassCVK is the key to fetch volume's SC
	StorageClassCVK CASVolumeKey = "openebs.io/storage-class"
)

// CASVolumeDeprecatedKey is a typed string to represent cas volume related
// annotations' or labels' key
type CASVolumeDeprecatedKey string

const (
	// Deprecated in favour of CASVolume.Spec.Capacity
	//
	// CapacityCVDK is a label key used to set volume capacity
	CapacityCVDK CASVolumeDeprecatedKey = "volumeprovisioner.mapi.openebs.io/storage-size"

	// NamespaceCVK is the key to fetch volume's namespace
	NamespaceCVDK CASVolumeDeprecatedKey = "k8s.io/namespace"

	// PersistentVolumeClaimCVK is the key to fetch volume's PVC
	PersistentVolumeClaimCVDK CASVolumeDeprecatedKey = "k8s.io/pvc"

	// StorageClassCVK is the key to fetch volume's SC
	StorageClassCVDK CASVolumeDeprecatedKey = "k8s.io/storage-class"
)

// CASVolumeDefault is a typed string to represent default cas volume related
// properties or related attributes or related operations
type CASVolumeDefault string

const (
	// NOTE:
	//  As per the current design there is no default cas template to create a
	// cas volume. It is expected that the StorageClass will explicitly set the
	// cas template name required to create a cas volume. However reading,
	// deleting & listing of cas volume(s) have corresponding cas templates that
	// are used implicitly i.e. each of them have their own default cas template.

	// CASTemplateForReadCVD is the default cas template to read a cas volume
	CASTemplateForReadCVD CASVolumeDefault = "read-cas-tpl"
	// CASTemplateForListCVD is the default cas template to list cas volumes
	CASTemplateForListCVD CASVolumeDefault = "list-cas-tpl"
	// CASTemplateForDeleteCVD is the default cas template to delete a cas volume
	CASTemplateForDeleteCVD CASVolumeDefault = "delete-cas-tpl"
)

// CASVolume represents a cas volume
type CASVolume struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Spec i.e. specifications of this cas volume
	Spec CASVolumeSpec `json:"spec"`
	// Status of this cas volume
	Status CASVolumeStatus `json:"status"`
}

// CASVolumeSpec has the properties of a cas volume
type CASVolumeSpec struct {
	// Capacity will hold the capacity of this Volume
	Capacity string `json:"capacity,omitempty" protobuf:"bytes,1,opt,name=capacity"`
}

// CASVolumeStatus provides status of a cas volume
type CASVolumeStatus struct {
	// Phase indicates if a volume is available, pending or failed
	Phase VolumePhase
	// A human-readable message indicating details about why the volume
	// is in this state
	Message string
	// Reason is a brief CamelCase string that describes any failure and is meant
	// for machine parsing and tidy display in the CLI
	Reason string
}

// VolumePhase defines phase of a volume
type VolumePhase string

const (
	// VolumePending - used for Volumes that are not available
	VolumePending VolumePhase = "Pending"
	// VolumeAvailable - used for Volumes that are available
	VolumeAvailable VolumePhase = "Available"
	// VolumeFailed - used for Volumes that failed for some reason
	VolumeFailed VolumePhase = "Failed"
)

// CASVolumeList is a list of CASVolume resources
type CASVolumeList struct {
	metav1.ListOptions `json:",inline"`
	metav1.ObjectMeta  `json:"metadata,omitempty"`
	metav1.ListMeta    `json:"metalist"`

	// Items are the list of volumes
	Items []CASVolume `json:"items"`
}
