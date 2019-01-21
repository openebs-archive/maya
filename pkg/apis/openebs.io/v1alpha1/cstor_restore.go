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

// CStorRestore describes a cstor volume resource created as custom resource
type CStorRestore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              CStorRestoreSpec   `json:"spec"`
	Status            CStorRestoreStatus `json:"status"`
}

// CStorRestoreSpec is the spec for a CStorRestore resource
type CStorRestoreSpec struct {
	Name       string `json:"name"`
	VolumeName string `json:"volumeName"`
	CasType    string `json:"casType"`
	RestoreSrc string `json:"backupSrc"`
	TargetIP   string `json:"targetIP"`
}

// CStorRestorePhase is to hold result of action.
type CStorRestorePhase string

// Status written onto CStorRestore objects.
const (
	// RSTStatusEmpty ensures the create operation is to be done, if import fails.
	RSTStatusEmpty CStorRestorePhase = ""

	// CVRStatusOnline ensures the resource is available.
	RSTStatusOnline CStorRestorePhase = "Healthy"
	// CVRStatusOffline ensures the resource is not available.
	RSTStatusOffline CStorRestorePhase = "Offline"
	// CVRStatusDegraded means that the rebuilding has not yet started.
	RSTStatusDegraded CStorRestorePhase = "Degraded"
	// RSTStatusError means that the volume status could not be found.
	RSTStatusError CStorRestorePhase = "Error"
	// RSTStatusDeletionFailed ensures the resource deletion has failed.
	RSTStatusDeletionFailed CStorRestorePhase = "Error"
	// RSTStatusInvalid ensures invalid resource.
	RSTStatusInvalid CStorRestorePhase = "Invalid"
	// RSTStatusErrorDuplicate ensures error due to duplicate resource.
	RSTStatusErrorDuplicate CStorRestorePhase = "Duplicate"
	// RSTStatusPending ensures pending task of cvr resource.
	RSTStatusPending CStorRestorePhase = "Init"
)

// CStorRestoreStatus is for handling status of cvr.
type CStorRestoreStatus struct {
	Phase CStorRestorePhase `json:"phase"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=cstorvolumereplicas

// CStorRestoreList is a list of CStorRestore resources
type CStorRestoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []CStorRestore `json:"items"`
}
