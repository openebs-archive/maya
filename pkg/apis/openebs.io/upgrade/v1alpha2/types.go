/*
Copyright 2019 The OpenEBS Authors

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

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=upgradetask
// +k8s:openapi-gen=true

// UpgradeTask represents an upgrade task
type UpgradeTask struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// Spec i.e. specifications of the upgradeTask
	Spec UpgradeTaskSpec `json:"spec" protobuf:"bytes,2,opt,name=spec"`
	// Status of upgradeTask
	Status UpgradeTaskStatus `json:"status" protobuf:"bytes,3,opt,name=status"`
}

// UpgradeTaskSpec is the properties of an upgrade task
type UpgradeTaskSpec struct {
	FromVersion  string       `json:"fromVersion" protobuf:"bytes,1,name=fromVersion"`
	ToVersion    string       `json:"toVersion" protobuf:"bytes,2,name=toVersion"`
	Flags        Flags        `json:"flags,omitempty" protobuf:"bytes,3,name=flags"`
	ResourceType ResourceType `json:"resourceType" protobuf:"bytes,4,name=resourceType"`
	ImagePrefix  string       `json:"imagePrefix" protobuf:"bytes,5,name=imagePrefix"`
	ImageTag     string       `json:"imageTag" protobuf:"bytes,6,name=imageTag"`
}

// Flags provides additional optional arguments
type Flags struct {
	Timeout int `json:"timeout,omitempty" protobuf:"varint,1,opt,name=resourceType"`
}

// ResourceType is the type of resource which is to be upgraded.
// Exactly one of its members must be set.
type ResourceType struct {
	JivaVolume       *JivaVolume       `json:"jivaVolume,omitempty" protobuf:"bytes,1,opt,name=jivaVolume"`
	CStorVolume      *CStorVolume      `json:"cstorVolume,omitempty" protobuf:"bytes,2,opt,name=cstorVolume"`
	CStorPool        *CStorPool        `json:"cstorPool,omitempty" protobuf:"bytes,3,opt,name=cstorPool"`
	StoragePoolClaim *StoragePoolClaim `json:"storagePoolClaim,omitempty" protobuf:"bytes,3,opt,name=storagePoolClaim"`
}

// JivaVolume is the ResourceType for jiva volume
type JivaVolume struct {
	PVName string         `json:"pvName,omitempty" protobuf:"bytes,1,name=pvName"`
	Flags  *ResourceFlags `json:"flags,omitempty" protobuf:"bytes,2,opt,name=flags"`
}

// CStorVolume is the ResourceType for cstor volume
type CStorVolume struct {
	PVName string         `json:"pvName,omitempty" protobuf:"bytes,1,name=pvName"`
	Flags  *ResourceFlags `json:"flags,omitempty" protobuf:"bytes,2,opt,name=flags"`
}

// CStorPool is the ResourceType for cstor pool
type CStorPool struct {
	PoolName string         `json:"poolName,omitempty" protobuf:"bytes,1,name=poolName"`
	Flags    *ResourceFlags `json:"flags,omitempty" protobuf:"bytes,2,opt,name=flags"`
}

// StoragePoolClaim is the ResourceType for storage pool claim
type StoragePoolClaim struct {
	PoolName string         `json:"poolName,omitempty" protobuf:"bytes,1,name=poolName"`
	Flags    *ResourceFlags `json:"flags,omitempty" protobuf:"bytes,2,opt,name=flags"`
}

// ResourceFlags provides additional options for a particular resource
type ResourceFlags struct {
	IgnoreStepsOnError []string `json:"ignoreStepsOnError,omitempty" protobuf:"bytes,1,opt,name=ignoreStepsOnError"`
}

// UpgradeTaskStatus provides status of a cas volume
type UpgradeTaskStatus struct {
	// Phase indicates if a volume is available, pending or failed
	Phase                   UpgradePhase              `json:"phase,omitempty" protobuf:"bytes,1,opt,name=phase,casttype=UpgradePhase"`
	StartTime               metav1.Time               `json:"startTime,omitempty" protobuf:"bytes,2,opt,name=startTime"`
	CompletedTime           metav1.Time               `json:"completedTime,omitempty" protobuf:"bytes,3,opt,name=completedTime"`
	UpgradeDetailedStatuses []UpgradeDetailedStatuses `json:"upgradeDetailedStatuses,omitempty" protobuf:"bytes,4,rep,name=upgradeDetailedStatuses"`
}

// UpgradeDetailedStatuses represents the latest available observations
// of a UpgradeTask current state.
type UpgradeDetailedStatuses struct {
	Step          UpgradeStep `json:"step,omitempty" protobuf:"bytes,1,opt,name=step"`
	StartTime     metav1.Time `json:"startTime,omitempty" protobuf:"bytes,2,opt,name=startTime"`
	LastUpdatedAt metav1.Time `json:"lastUpdatedAt,omitempty" protobuf:"bytes,3,opt,name=lastUpdatedAt"`
	State         State       `json:"state" protobuf:"bytes,4,opt,name=state"`
}

// State represents the state of the step performed during the upgrade.
type State struct {
	Phase StatePhase `json:"phase,omitempty" protobuf:"bytes,1,opt,name=phase"`
	// A human-readable message indicating details about why the volume
	// is in this state
	Message string `json:"message,omitempty" protobuf:"bytes,2,opt,name=message"`
	// Reason is a brief CamelCase string that describes any failure and is meant
	// for machine parsing and tidy display in the CLI
	Reason string `json:"reason,omitempty" protobuf:"bytes,3,opt,name=reason"`
}

// UpgradeStep is the current step being performed for a particular resource upgrade
type UpgradeStep string

const (
	// PreUpgrade ...
	PreUpgrade UpgradeStep = "PRE_UPGRADE"
	// TargetUpgrade ...
	TargetUpgrade UpgradeStep = "TARGET_UPGRADE"
	// ReplicaUpgrade ...
	ReplicaUpgrade UpgradeStep = "REPLICA_UPGRADE"
	// Verify ...
	Verify UpgradeStep = "VERIFY"
	// Rollback ...
	Rollback UpgradeStep = "ROLLBACK"
)

// StatePhase ...
type StatePhase string

const (
	// WaitingState ...
	WaitingState StatePhase = "Waiting"
	// ErroredState ...
	ErroredState StatePhase = "Errored"
	// CompletedState ...
	CompletedState StatePhase = "Completed"
)

// UpgradePhase defines phase of a volume
type UpgradePhase string

const (
	// UpgradeStarted - used for Upgrades that are Started
	UpgradeStarted UpgradePhase = "Started"
	// UpgradeSuccess - used for Upgrades that are not available
	UpgradeSuccess UpgradePhase = "Success"
	// UpgradeError - used for Upgrades that Error for some reason
	UpgradeError UpgradePhase = "Error"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=upgradetasks
// +k8s:openapi-gen=true

// UpgradeTaskList is a list of UpgradeTask resources
type UpgradeTaskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items are the list of upgrade task items
	Items []UpgradeTask `json:"items" protobuf:"bytes,2,rep,name=items"`
}
