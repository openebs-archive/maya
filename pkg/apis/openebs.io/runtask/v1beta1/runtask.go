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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=runtask

// RunTask forms the desired run specification as well as the result of
// trying the same
type RunTask struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RunTaskSpec   `json:"spec"`
	Status RunTaskStatus `json:"status"`
}

// RunTaskSpec is the specifications of a RunTask resource. It composes of a
// config as well as a list of run items. It specifies the desired actions
// against one or more kubernetes resources.
type RunTaskSpec struct {
	Config map[string]string `json:"config"`
	Runs   []RunItem         `json:"runs"`
}

// RunTaskStatus presents the state of runtask after execution of desired
// actions
type RunTaskStatus struct {
	Phase string `json:"phase"`
}

// RunItem composes of properties that form a single run execution
type RunItem struct {
	ID          string            `json:"id"`                    // identity of this run
	Name        string            `json:"name,omitempty"`        // description of this run
	Action      Action            `json:"action"`                // command that this run will invoke
	APIVersion  string            `json:"apiVersion,omitempty"`  // resource version
	Kind        Kind              `json:"kind"`                  // refers to resource against which action gets executed
	Options     []Option          `json:"options"`               // options while executing the action
	Conditions  []Condition       `json:"conditions,omitempty"`  // conditions to satisfy before executing
	ConditionOp ConditionOperator `json:"conditionOp,omitempty"` // operator applied against the list of conditions
}

// Option represents a logic and corresponding condition(s) to execute the
// action
type Option struct {
	Func        string            `json:"func"`
	Conditions  []Condition       `json:"conditions,omitempty"`
	ConditionOp ConditionOperator `json:"conditionOp,omitempty"`
}

// Action defines the supported commands to be executed against a supported
// resource
type Action string

const (
	Get      Action = "get"
	Template Action = "template"
	List     Action = "list"
	Create   Action = "create"
	Update   Action = "update"
	Patch    Action = "patch"
)

// Kind defines the supported resource types
type Kind string

const (
	DaemonSet Kind = "daemonset"
)

// Condition defines a condition in string format
type Condition string

// ConditionOperator defines the logical operator to be used against one or
// more conditions
type ConditionOperator string

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=runtasks

// RunTaskList is a list of RunTask resources
type RunTaskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []RunTask `json:"items"`
}
