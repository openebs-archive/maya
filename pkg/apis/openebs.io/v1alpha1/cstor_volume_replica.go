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

// CVRKey represents the properties of a cstorvolumereplica
type CVRKey string

const (
	// CloneEnableKEY is used to enable/disable cloning for a cstorvolumereplica
	CloneEnableKEY CVRKey = "openebs.io/cloned"

	// SourceVolumeKey stores the name of source volume whose snapshot is used to
	// create this cvr
	SourceVolumeKey CVRKey = "openebs.io/source-volume"

	// SnapshotNameKey stores the name of the snapshot being used to restore this replica
	SnapshotNameKey CVRKey = "openebs.io/snapshot"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=cstorvolumereplica

// CStorVolumeReplica describes a cstor volume resource created as custom resource
type CStorVolumeReplica struct {
	metav1.TypeMeta                 `json:",inline"`
	metav1.ObjectMeta               `json:"metadata,omitempty"`
	Spec   CStorVolumeReplicaSpec   `json:"spec"`
	Status CStorVolumeReplicaStatus `json:"status"`
}

// CStorVolumeReplicaSpec is the spec for a CStorVolumeReplica resource
type CStorVolumeReplicaSpec struct {
	TargetIP string `json:"targetIP"`
	Capacity string `json:"capacity"`
}

// CStorVolumeReplicaPhase is to hold result of action.
type CStorVolumeReplicaPhase string

// Status written onto CStorVolumeReplica objects.
const (
	// CVRStatusEmpty ensures the create operation is to be done, if import fails.
	CVRStatusEmpty CStorVolumeReplicaPhase = ""
	// CVRStatusOnline ensures the resource is available.
	CVRStatusOnline CStorVolumeReplicaPhase = "Healthy"
	// CVRStatusOffline ensures the resource is not available.
	CVRStatusOffline CStorVolumeReplicaPhase = "Offline"
	// CVRStatusDegraded means that the rebuilding has not yet started.
	CVRStatusDegraded CStorVolumeReplicaPhase = "Degraded"
	// CVRStatusRebuilding means that the volume is in re-building phase.
	CVRStatusRebuilding CStorVolumeReplicaPhase = "Rebuilding"
	// CVRStatusRebuilding means that the volume status could not be found.
	CVRStatusError CStorVolumeReplicaPhase = "Error"
	// CVRStatusDeletionFailed ensures the resource deletion has failed.
	CVRStatusDeletionFailed CStorVolumeReplicaPhase = "Error"
	// CVRStatusInvalid ensures invalid resource.
	CVRStatusInvalid CStorVolumeReplicaPhase = "Invalid"
	// CVRStatusErrorDuplicate ensures error due to duplicate resource.
	CVRStatusErrorDuplicate CStorVolumeReplicaPhase = "Invalid"
	// CVRStatusPending ensures pending task of cvr resource.
	CVRStatusPending CStorVolumeReplicaPhase = "Init"
)

// CStorVolumeReplicaStatus is for handling status of cvr.
type CStorVolumeReplicaStatus struct {
	Phase    CStorVolumeReplicaPhase `json:"phase"`
	Capacity CStorVolumeCapacityAttr `json:"capacity"`
}

// CStorVolumeCapacityAttr is for storing the volume capacity.
type CStorVolumeCapacityAttr struct {
	TotalAllocated string `json:"totalAllocated"`
	Used           string `json:"used"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=cstorvolumereplicas

// CStorVolumeReplicaList is a list of CStorVolumeReplica resources
type CStorVolumeReplicaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []CStorVolumeReplica `json:"items"`
}
