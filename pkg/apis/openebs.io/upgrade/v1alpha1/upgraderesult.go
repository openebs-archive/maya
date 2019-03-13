/*
Copyright 2018 The OpenEBS Authors

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
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=upgrade

// UpgradeResult forms the desired specification of
// upgrade task's result that should be available for
// all the upgrade tasks
type UpgradeResult struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UpgradeResultSpec   `json:"spec"`
	Status UpgradeResultStatus `json:"status"`
}

// UpgradeResultSpec is the specifications of
// upgrade result
type UpgradeResultSpec struct {
	// TODO: Add task-executor config spec here

	// BaseVersion is the current version of openebs
	BaseVersion string `json:"baseVersion"`
	// TargetVersion is the version to which openebs
	// components needs to be upgraded
	TargetVersion string `json:"targetVersion"`
	// ResourceName contains comma separated name of
	// the resources which needs to be upgraded
	ResourceName string `json:"resourceName"`
}

// UpgradeResultStatus represents the current
// status of UpgradeTaskResult
type UpgradeResultStatus struct {
	// TotalTaskCount is the total no tasks that
	// needs to be executed to upgrade a particular
	// openebs component to desired version
	TotalTaskCount int `json:"totalTaskCount"`
	// CompletedTaskCount represents the no of tasks
	// that has successfully completed
	CompletedTaskCount int `json:"completedTaskCount"`
	// FailedTaskCount represents the no of tasks
	// that has failed
	FailedTaskCount int `json:"failedTaskCount"`
	// Tasks represents the list of tasks
	Tasks []Task `json:"tasks"`
}

// Task represents an upgrade task and its subtasks details
type Task struct {
	// Name of the task
	Name string `json:"name"`
	// TotalSubTaskCount is the no of subtasks
	// i.e. runtasks that needs to be executed
	// to complete a task
	TotalSubTaskCount int `json:"totalSubTaskCount"`
	// CompletedSubTaskCount is the no of subtasks
	// that has successfully executed
	CompletedSubTaskCount int `json:"completedSubTaskCount"`
	// FailedSubTaskCount is the no of subtasks that has
	// failed
	FailedSubTaskCount int `json:"failedSubTaskCount"`
	// LastCompletedSubTaskID is the id of the last successfully
	// completed runtask of a particular task
	LastCompletedSubTaskID string `json:"lastCompletedSubTaskID"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=upgrade

// UpgradeResultList is a list of upgrade results
type UpgradeResultList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []UpgradeResult `json:"items"`
}
