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
// +resource:path=cstorvolume

// CStorVolume describes a cstor volume resource created as custom resource
type CStorVolume struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              CStorVolumeSpec   `json:"spec"`
	Status            CStorVolumeStatus `json:"status"`
}

// CStorVolumeSpec is the spec for a CStorVolume resource
type CStorVolumeSpec struct {
	Capacity          string `json:"capacity"`
	TargetIP          string `json:"targetIP"`
	TargetPort        string `json:"targetPort"`
	Iqn               string `json:"iqn"`
	TargetPortal      string `json:"targetPortal"`
	Status            string `json:"status"`
	NodeBase          string `json:"nodeBase"`
	ReplicationFactor int    `json:"replicationFactor"`
	ConsistencyFactor int    `json:"consistencyFactor"`
}

// CStorVolumePhase is to hold result of action.
type CStorVolumePhase string

// CStorVolumeStatus is for handling status of cvr.
type CStorVolumeStatus struct {
	Phase           CStorVolumePhase `json:"phase"`
	ReplicaStatuses []ReplicaStatus  `json:"replicaStatuses,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=cstorvolume

// CStorVolumeList is a list of CStorVolume resources
type CStorVolumeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []CStorVolume `json:"items"`
}

// CVStatusResponse stores the reponse of istgt replica command output
// It may contain several volumes
type CVStatusResponse struct {
	CVStatuses []CVStatus `json:"volumeStatus"`
}

// CVStatus stores the status of a CstorVolume obtained from response
type CVStatus struct {
	Name            string          `json:"name"`
	Status          string          `json:"status"`
	ReplicaStatuses []ReplicaStatus `json:"replicaStatus"`
}

// ReplicaStatus stores the status of replicas
type ReplicaStatus struct {
	ID                string `json:"replicaId"`
	Status            string `json:"status"`
	CheckpointedIOSeq string `json:"checkpointedIOSeq"`
	InflightRead      string `json:"inflightRead"`
	InflightWrite     string `json:"inflightWrite"`
	InflightSync      string `json:"inflightSync"`
	UpTime            int    `json:"upTime"`
}
