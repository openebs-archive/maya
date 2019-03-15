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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	// Register adds UpgradeResult and UpgradeResultList objects to
	// SchemeBuilder so they can be added to a Scheme
	SchemeBuilder.Register(&UpgradeResult{}, &UpgradeResultList{})
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=upgrade

// UpgradeResult contains the desired specifications of an
// upgrade result
type UpgradeResult struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Config UpgradeResultConfig `json:"config"`
	Status UpgradeResultStatus `json:"status"`
}

// UpgradeResultStatus represents the current state of UpgradeResult
type UpgradeResultStatus struct {
	// Resources is the total no of resources that
	// needs to be upgraded to a desired version
	Resources int `json:"resources"`
	// UpgradedResources represents the no of resources
	// that has been successfully upgraded
	UpgradedResources int `json:"upgradedResources"`
	// FailedResources represents the no of resources
	// that has failed to upgrade
	FailedResources int `json:"failedResources"`
	// ResourceList is the list of resources that needs to
	// be upgraded
	ResourceList []UpgradeResource `json:"resourceList"`
}

// UpgradeResource represents a resource that needs to
// be upgraded to a desired version
type UpgradeResource struct {
	// Name of the resource to be upgraded
	Name string `json:"name"`
	// Kind is the type of resource i.e.
	// PVC, SPC, ..
	Kind string `json:"kind"`
	// APIVersion of the resource
	APIVersion string `json:"apiVersion"`
	// Namespace of the resource
	Namespace string `json:"namespace"`
	// Status is the status of the resource
	Status string `json:"status"`
	// Message is a human readable message
	// indicating details about the resource state
	Message string `json:"message"`
	// BeforeUpgrade represents the state of the
	// related resources before upgrade i.e. for a
	// PV, related resources could be cvr, target
	// deployment, target svc, etc.
	BeforeUpgrade []ResourceState `json:"beforeUpgrade"`
	// AfterUpgrade represents the state of the
	// related resources after upgrade
	AfterUpgrade []ResourceState `json:"afterUpgrade"`
	// Tasks are the runtasks that needs to be
	// executed to perform this upgrade
	Tasks []Task `json:"tasks"`
}

// Task represents details of a task(runtask) required
// to be executed for upgrading a particular resource
type Task struct {
	// Name of the task
	Name string `json:"name"`
	// Status is the status of the task which
	// could be successful or failed
	Status string `json:"status"`
	// LastTransitionTime is the last time the status
	// transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// Message is a human readable message
	// indicating details about the task
	Message string `json:"message"`
	// LastError is the last error occurred
	// while executing this task
	LastError string `json:"lastError"`
	// StartTime of the task
	StartTime *metav1.Time `json:"startTime"`
	// EndTime of the task
	EndTime *metav1.Time `json:"endTime"`
	// Retries is the no of times this task
	// has tried to execute
	Retries int `json:"retries"`
}

// ResourceState represents the state of a resource
type ResourceState struct {
	// Name of the resource
	Name string `json:"name"`
	// Kind is the type of resource i.e.
	// cvr, deployment, ..
	Kind string `json:"kind"`
	// APIVersion of the resource
	APIVersion string `json:"apiVersion"`
	// Namespace of the resource
	Namespace string `json:"namespace"`
	// Status of the resource
	Status string `json:"status"`
	// LastTransitionTime is the last time the status
	// transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// Message is a human readable message indicating details about the transition.
	Message string `json:"message"`
}

// UpgradeResultConfig represents the entire config
// of UpgradeResult i.e. same as task-executor job config
type UpgradeResultConfig struct {
	// TODO: Add Task-executor job config here
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=upgrades

// UpgradeResultList is a list of UpgradeResults
type UpgradeResultList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []UpgradeResult `json:"items"`
}
