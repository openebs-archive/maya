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

// CStorBackupData describes a cstor volume resource created as custom resource
type CStorBackupData struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              CStorBackupDataSpec   `json:"spec"`
	Status            CStorBackupDataStatus `json:"status"`
}

// CStorBackupDataSpec is the spec for a CStorBackupData resource
type CStorBackupDataSpec struct {
	Name                  string `json:"name"`
	VolumeName            string `json:"volumeName"`
	CasType               string `json:"casType"`
	IncrementalBackupName string `json:"incrementalBackupName"`
	LastSnapshotName      string `json:"snapshotName"`
}

// CStorBackupDataPhase is to hold result of action.
type CStorBackupDataPhase string
type CStorBackupDataProgress string

// Status written onto CStorBackupData objects.
const (
	// CSBDStatusEmpty ensures the create operation is to be done, if import fails.
	CSBDStatusEmpty CStorBackupDataPhase = ""

	// CVRStatusOnline ensures the resource is available.
	CSBDStatusOnline CStorBackupDataPhase = "Healthy"
	// CVRStatusOffline ensures the resource is not available.
	CSBDStatusOffline CStorBackupDataPhase = "Offline"
	// CVRStatusDegraded means that the rebuilding has not yet started.
	CSBDStatusDegraded CStorBackupDataPhase = "Degraded"
	// CSBDStatusError means that the volume status could not be found.
	CSBDStatusError CStorBackupDataPhase = "Error"
	// CSBDStatusDeletionFailed ensures the resource deletion has failed.
	CSBDStatusDeletionFailed CStorBackupDataPhase = "Error"
	// CSBDStatusInvalid ensures invalid resource.
	CSBDStatusInvalid CStorBackupDataPhase = "Invalid"
	// CSBDStatusErrorDuplicate ensures error due to duplicate resource.
	CSBDStatusErrorDuplicate CStorBackupDataPhase = "Invalid"
	// CSBDStatusPending ensures pending task of cvr resource.
	CSBDStatusPending CStorBackupDataPhase = "Init"
)

// CStorBackupDataStatus is for handling status of cvr.
type CStorBackupDataStatus struct {
	Phase    CStorBackupDataPhase    `json:"phase"`
	Progress CStorBackupDataProgress `json:"progress"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=cstorvolumereplicas

// CStorBackupDataList is a list of CStorBackupData resources
type CStorBackupDataList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []CStorBackupData `json:"items"`
}
