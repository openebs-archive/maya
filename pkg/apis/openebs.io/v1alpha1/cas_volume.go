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
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/openebs/maya/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	timeout = 5 * time.Second
)

// CASKey is a typed string to represent CAS related annotations'
// or labels' keys
//
// Example 1 - Below is a sample CASVolume that makes use of some CASKey
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
//     openebs.io/storage-class: openebs-repaffinity-0.6.0
//     # this manner of setting namespace gets the second priority
//     openebs.io/namespace: default
//     openebs.io/pvc: openebs-repaffinity-0.6.0
// spec:
//   # latest way to set capacity
//   capacity: 2G
// ```
type CASKey string

const (
	// CASConfigKey is the key to fetch configurations w.r.t a CAS entity
	CASConfigKey CASKey = "cas.openebs.io/config"

	// NamespaceKey is the key to fetch cas entity's namespace
	NamespaceKey CASKey = "openebs.io/namespace"

	// PersistentVolumeClaimKey is the key to fetch name of PersistentVolumeClaim
	PersistentVolumeClaimKey CASKey = "openebs.io/persistentvolumeclaim"

	// StorageClassKey is the key to fetch name of StorageClass
	StorageClassKey CASKey = "openebs.io/storageclass"

	// StorageClassHeaderKey is the key to fetch name of StorageClass
	// This key is present only in get request headers
	StorageClassHeaderKey CASKey = "storageclass"
)

// DeprecatedKey is a typed string to represent deprecated annotations' or
// labels' key
type DeprecatedKey string

const (
	// CapacityDeprecatedKey is a label key used to set volume capacity
	//
	// NOTE:
	//  Deprecated in favour of CASVolume.Spec.Capacity
	CapacityDeprecatedKey DeprecatedKey = "volumeprovisioner.mapi.openebs.io/storage-size"

	// NamespaceDeprecatedKey is the key to fetch volume's namespace
	//
	// NOTE:
	//  Deprecated in favour of NamespaceCK
	NamespaceDeprecatedKey DeprecatedKey = "k8s.io/namespace"

	// PersistentVolumeClaimDeprecatedKey is the key to fetch volume's PVC
	//
	// NOTE:
	//  Deprecated in favour of PersistentVolumeClaimCK
	PersistentVolumeClaimDeprecatedKey DeprecatedKey = "k8s.io/pvc"

	// StorageClassDeprecatedKey is the key to fetch name of StorageClass
	//
	// NOTE:
	//  Deprecated in favour of StorageClassCK
	StorageClassDeprecatedKey DeprecatedKey = "k8s.io/storage-class"
)

// CASVolumeKey is a typed string to represent CAS Volume related annotations'
// or labels' keys
//
// Example 1 - Below is a sample StorageClass that makes use of a CASVolumeKey
// constant i.e. the cas template used to create a cas volume
//
// ```yaml
// apiVersion: storage.k8s.io/v1
// kind: StorageClass
// metadata:
//  name: openebs-standard
//  annotations:
//    cas.openebs.io/create-volume-template: cast-standard-0.6.0
// provisioner: openebs.io/provisioner-iscsi
// ```
type CASVolumeKey string

const (
	// CASTemplateKeyForVolumeCreate is the key to fetch name of CASTemplate
	// to create a CAS Volume
	CASTemplateKeyForVolumeCreate CASVolumeKey = "cas.openebs.io/create-volume-template"

	// CASTemplateKeyForVolumeRead is the key to fetch name of CASTemplate
	// to read a CAS Volume
	CASTemplateKeyForVolumeRead CASVolumeKey = "cas.openebs.io/read-volume-template"

	// CASTemplateKeyForVolumeDelete is the key to fetch name of CASTemplate
	// to delete a CAS Volume
	CASTemplateKeyForVolumeDelete CASVolumeKey = "cas.openebs.io/delete-volume-template"

	// CASTemplateKeyForVolumeList is the key to fetch name of CASTemplate
	// to list CAS Volumes
	CASTemplateKeyForVolumeList CASVolumeKey = "cas.openebs.io/list-volume-template"
)

// CASJivaVolumeDefault is a typed string to represent defaults of Jiva based
// CAS Volume properties or attributes or operations
type CASJivaVolumeDefault string

const (
// NOTE:
//  As per the current design there is no default CAS template to create a
// CAS Volume. It is expected that the StorageClass will explicitly set the
// cas template name required to create a CAS Volume. However reading,
// deleting & listing of cas volume(s) have corresponding cas templates that
// are used implicitly i.e. read, delete & list have their own default cas
// templates.

// DefaultCASTemplateForJivaVolumeRead is the default cas template to read
// a Jiva based CAS Volume
//DefaultCASTemplateForJivaVolumeRead CASJivaVolumeDefault = "read-cstor-cas-volume-tpl"
// DefaultCASTemplateForJivaVolumeList is the default cas template to list
// Jiva based CAS Volumes
//DefaultCASTemplateForJivaVolumeList CASJivaVolumeDefault = "list-cstor-cas-volume-tpl"
// DefaultCASTemplateForJivaVolumeDelete is the default cas template to
// delete a Jiva based CAS Volume
//DefaultCASTemplateForJivaVolumeDelete CASJivaVolumeDefault = "delete-cstor-cas-volume-tpl"
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
	Capacity string `json:"capacity"`
	// Iqn will hold the iqn value of this Volume
	Iqn string `json:"iqn"`
	// TargetPortal will hold the target portal for this volume
	TargetPortal string `json:"targetPortal"`
	// TargetIP will hold the targetIP for this Volume
	TargetIP string `json:"targetIP"`
	// TargetPort will hold the targetIP for this Volume
	TargetPort string `json:"targetPort"`
	// Replicas will hold the replica count for this volume
	Replicas string `json:"replicas"`
	// CasType will hold the storage engine used to provision this volume
	CasType string `json:"casType"`
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

// FetchVolumeInfo fetches and fills CASVolume structure from URL given to it
func (volInfo *CASVolume) FetchVolumeInfo(URL string, volname string, namespace string) error {
	url := URL
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("namespace", namespace)

	c := &http.Client{
		Timeout: timeout,
	}
	resp, err := c.Do(req)
	if err != nil {
		fmt.Printf("Can't get a response, error found: %v", err)
		return err
	}
	if resp != nil {
		if resp.StatusCode == 500 {
			fmt.Printf("Volume: %s not found at M_API server\n", volname)
			return util.InternalServerError
		} else if resp.StatusCode == 503 {
			fmt.Println("M_API server not reachable")
			return util.ServerUnavailable
		} else if resp.StatusCode == 404 {
			fmt.Printf("Volume: %s not found at M_API server\n", volname)
			return util.PageNotFound
		}
	} else {
		fmt.Println("M_API server not reachable")
		return err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(volInfo)
	if err != nil {
		fmt.Println("Response decode failed")
		return err
	}
	if volInfo.Status.Reason == "pending" {
		fmt.Println("VOLUME status Unknown to M_API server")
	}
	return err
}

// GetCASType returns the CASType of the volume in lowercase
func (volInfo *CASVolume) GetCASType() string {
	if len(volInfo.Spec.CasType) == 0 {
		return "jiva"
	}
	return strings.ToLower(volInfo.Spec.CasType)
}

// GetClusterIP returns the ClusterIP of the cluster
func (volInfo *CASVolume) GetClusterIP() string {
	if val, ok := volInfo.ObjectMeta.Annotations["openebs.io/cluster-ips"]; ok {
		return val
	} else if val, ok := volInfo.ObjectMeta.Annotations["vsm.openebs.io/cluster-ips"]; ok {
		return val
	}
	return ""
}

// GetControllerStatus returns the status of the volume controller
func (volInfo *CASVolume) GetControllerStatus() string {
	if val, ok := volInfo.ObjectMeta.Annotations["openebs.io/controller-status"]; ok {
		return val
	} else if val, ok := volInfo.ObjectMeta.Annotations["vsm.openebs.io/controller-status"]; ok {
		return val
	}
	return ""
}

// GetIQN returns the IQN of the volume
func (volInfo *CASVolume) GetIQN() string {
	return volInfo.Spec.Iqn
}

// GetVolumeName returns the volume name
func (volInfo *CASVolume) GetVolumeName() string {
	return volInfo.ObjectMeta.Name
}

// GetTargetPortal returns the TargetPortal of the volume
func (volInfo *CASVolume) GetTargetPortal() string {
	return volInfo.Spec.TargetPortal
}

// GetVolumeSize returns the capacity of the volume
func (volInfo *CASVolume) GetVolumeSize() string {
	return volInfo.Spec.Capacity
}

// GetReplicaCount returns the volume replica count
func (volInfo *CASVolume) GetReplicaCount() string {
	return volInfo.Spec.Replicas
}

// GetReplicaStatus returns the replica status of the volume replica
func (volInfo *CASVolume) GetReplicaStatus() string {
	if val, ok := volInfo.ObjectMeta.Annotations["openebs.io/replica-status"]; ok {
		return val
	} else if val, ok := volInfo.ObjectMeta.Annotations["vsm.openebs.io/replica-status"]; ok {
		return val
	}
	return ""
}

// GetReplicaIP returns the IP of volume replica
func (volInfo *CASVolume) GetReplicaIP() string {
	if val, ok := volInfo.ObjectMeta.Annotations["openebs.io/replica-ips"]; ok {
		return val
	} else if val, ok := volInfo.ObjectMeta.Annotations["vsm.openebs.io/replica-ips"]; ok {
		return val
	}
	return ""
}
