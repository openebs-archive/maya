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

package task

import (
	"strings"
)

// RepeatWithKind is a typed string that determines the kind of a repeat
// resource
type RepeatWithKind string

const (
	// NamespaceRWK indicates that the repeat resource is of kind Kubernetes
	// namespace
	NamespaceRWK RepeatWithKind = "namespace"
)

// RepeatWithResource enables repetitive execution of a task based on this
// resource. There can be one or more resources defined in this object.
//
// The task that has its property set with this object automatically sets itself
// to be executed multiple times where the number of times depends on the number
// of resources set in this object.
type RepeatWithResource struct {
	// Kind represent the type of resource that determines the repetition of a
	// task execution
	Kind RepeatWithKind `json:"kind"`
	// Resources lists the names of resources that determine the repetition of
	// a task execution
	//
	// NOTE:
	//  All resources should belong to one kind
	Resources []string `json:"resources"`
}

// repeatWithResourceExecutor exposes operations with respect to
// RepeatWithResource
type repeatWithResourceExecutor struct {
	repeatWith RepeatWithResource
}

// newRepeatWithResourceExecutor returns a new instance of
// repeatWithResourceExecutor
func newRepeatWithResourceExecutor(repeatWith RepeatWithResource) repeatWithResourceExecutor {
	return repeatWithResourceExecutor{
		repeatWith: repeatWith,
	}
}

// newRepeatWithResourceExecByObjectNames returns a new instance of
// newRepeatWithResourceExecutor based on the object names
//
// NOTE:
//  The repeat resource can be different than object resource. In this function
// a repeat instance is being created out of the name of objects due to the
// result of repetitive task executions.
//
// NOTE:
//  ObjectName refers to the name of the object i.e. result of a task execution
func newRepeatWithResourceExecByObjectNames(objectNames string) repeatWithResourceExecutor {
	return repeatWithResourceExecutor{
		repeatWith: RepeatWithResource{
			Resources: strings.Split(strings.TrimSpace(objectNames), ","),
		},
	}
}

// isNamespaceRepeat flags if the repeat resource is of kind kubernetes
// namespace
func (r repeatWithResourceExecutor) isNamespaceRepeat() bool {
	return r.repeatWith.Kind == NamespaceRWK
}

// isRepeat flags if there is any requirement to repeat the task execution
func (r repeatWithResourceExecutor) isRepeat() bool {
	return len(r.repeatWith.Resources) > 0
}

// getResources returns the list of resources based on which a task will get
// executed multiple times, each time depending on exactly one resource.
func (r repeatWithResourceExecutor) getResources() []string {
	return r.repeatWith.Resources
}
