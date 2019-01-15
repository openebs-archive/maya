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
// +resource:path=cstorbackups

// CStorRestoreData describes a cstor volume resource created as custom resource
type CStorRestoreData struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              CStorRestoreDataSpec   `json:"spec"`
	Status            CStorRestoreDataStatus `json:"status"`
}

// CStorRestoreDataSpec is the spec for a CStorRestoreData resource
type CStorRestoreDataSpec struct {
	Name         string `json:"name"`
	VolumeName   string `json:"volumeName"`
	CasType      string `json:"casType"`
	SnapName     string `json:"snapName"`
	PrevSnapName string `json:"prevSnapName"`
}

// CStorRestoreDataPhase is to hold result of action.
type CStorRestoreDataPhase string
type CStorRestoreDataProgress string

// Status written onto CStorRestoreData objects.
const (
	// CSRDStatusEmpty ensures the create operation is to be done, if import fails.
	CSRDStatusEmpty CStorRestoreDataPhase = ""

	// CVRStatusOnline ensures the resource is available.
	CSRDStatusOnline CStorRestoreDataPhase = "Healthy"
	// CVRStatusOffline ensures the resource is not available.
	CSRDStatusOffline CStorRestoreDataPhase = "Offline"
	// CVRStatusDegraded means that the rebuilding has not yet started.
	CSRDStatusDegraded CStorRestoreDataPhase = "Degraded"
	// CSRDStatusError means that the volume status could not be found.
	CSRDStatusError CStorRestoreDataPhase = "Error"
	// CSRDStatusDeletionFailed ensures the resource deletion has failed.
	CSRDStatusDeletionFailed CStorRestoreDataPhase = "Error"
	// CSRDStatusInvalid ensures invalid resource.
	CSRDStatusInvalid CStorRestoreDataPhase = "Invalid"
	// CSRDStatusErrorDuplicate ensures error due to duplicate resource.
	CSRDStatusErrorDuplicate CStorRestoreDataPhase = "Invalid"
	// CSRDStatusPending ensures pending task of cvr resource.
	CSRDStatusPending CStorRestoreDataPhase = "Init"
)

// CStorRestoreDataStatus is for handling status of cvr.
type CStorRestoreDataStatus struct {
	Phase    CStorRestoreDataPhase    `json:"phase"`
	Progress CStorRestoreDataProgress `json:"progress"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=cstorvolumereplicas

// CStorRestoreDataList is a list of CStorRestoreData resources
type CStorRestoreDataList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []CStorRestoreData `json:"items"`
}
