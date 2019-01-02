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

// CStorBackup describes a cstor volume resource created as custom resource
type CStorBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              CStorBackupSpec   `json:"spec"`
	Status            CStorBackupStatus `json:"status"`
}

// CStorBackupSpec is the spec for a CStorBackup resource
type CStorBackupSpec struct {
	Name         string `json:"name"`
	VolumeName   string `json:"volumeName"`
	CasType      string `json:"casType"`
	SnapName     string `json:"snapName"`
	PrevSnapName string `json:"prevSnapName"`
	BackupDest   string `json:"backupDest"`
}

// CStorBackupPhase is to hold result of action.
type CStorBackupPhase string

// Status written onto CStorBackup objects.
const (
	// CSBStatusEmpty ensures the create operation is to be done, if import fails.
	CSBStatusEmpty CStorBackupPhase = ""

	// CVRStatusOnline ensures the resource is available.
	CSBStatusOnline CStorBackupPhase = "Healthy"
	// CVRStatusOffline ensures the resource is not available.
	CSBStatusOffline CStorBackupPhase = "Offline"
	// CVRStatusDegraded means that the rebuilding has not yet started.
	CSBStatusDegraded CStorBackupPhase = "Degraded"
	// CSBStatusError means that the volume status could not be found.
	CSBStatusError CStorBackupPhase = "Error"
	// CSBStatusDeletionFailed ensures the resource deletion has failed.
	CSBStatusDeletionFailed CStorBackupPhase = "Error"
	// CSBStatusInvalid ensures invalid resource.
	CSBStatusInvalid CStorBackupPhase = "Invalid"
	// CSBStatusErrorDuplicate ensures error due to duplicate resource.
	CSBStatusErrorDuplicate CStorBackupPhase = "Invalid"
	// CSBStatusPending ensures pending task of cvr resource.
	CSBStatusPending CStorBackupPhase = "Init"
)

// CStorBackupStatus is for handling status of cvr.
type CStorBackupStatus struct {
	Phase CStorBackupPhase `json:"phase"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=cstorvolumereplicas

// CStorBackupList is a list of CStorBackup resources
type CStorBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []CStorBackup `json:"items"`
}
