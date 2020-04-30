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
// +resource:path=cstorcompletedbackup

// CStorCompletedBackup describes a cstor completed-backup resource created as custom resource
type CStorCompletedBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              CStorCompletedBackupSpec `json:"spec"`
}

// CStorCompletedBackupSpec is the spec for a CStorCompletedBackup resource
type CStorCompletedBackupSpec struct {
	// BackupName is a name of the backup or scheduled backup
	BackupName string `json:"backupName"`

	// VolumeName is a name of the volume for which this backup is destined
	VolumeName string `json:"volumeName"`

	// SnapName is a name of the current backup snapshot
	SnapName string `json:"snapName,omitempty"`

	// PrevSnapName is the last completed-backup's snapshot name
	PrevSnapName string `json:"prevSnapName,omitempty"`

	// SecondLastSnapName is a name of the second last completed-backup's snapshot
	SecondLastSnapName string `json:"secondLastSnapName"`

	// LastSnapName is the last completed-backup's snapshot name
	LastSnapName string `json:"lastSnapName"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=cstorcompletedbackup

// CStorCompletedBackupList is a list of cstorcompletedbackup resources
type CStorCompletedBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []CStorCompletedBackup `json:"items"`
}
