/*
Copyright 2017 The OpenEBS Authors.

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

// TopLevelProperty represents the top level property that
// is a starting point to represent a hierarchical chain of
// properties.
//
// e.g.
// Policy.prop1.subprop1 = val1
// Policy.prop1.subprop2 = val2
// In above example Policy is a top level object
//
// NOTE:
//  The value of any hierarchical chain of properties
// can be parsed via dot notation
type TopLevelProperty string

const (
	// PolicyTLP is a top level property supported by volume
	// policy engine
	//
	// The policy specific properties are placed with
	// PolicyTLP as the top level property
	PolicyTLP TopLevelProperty = "Policy"
	// VolumeTLP is a top level property supported by volume
	// policy engine
	//
	// The properties provided by the caller are placed
	// with VolumeTLP as the top level property
	//
	// NOTE:
	//  Policy engine cannot modify these properties.
	// These are the runtime properties that are provided
	// as inputs to policy engine
	VolumeTLP TopLevelProperty = "Volume"
	// TaskTLP is a top level property supported by
	// volume policy engine
	//
	// The task related properties as placed with TaskTLP
	// as the top level property
	TaskTLP TopLevelProperty = "Task"
	// TaskResultTLP is a top level property supported by
	// volume policy engine
	//
	// The specific results after the execution of a task
	// as placed with TaskResultTLP as the top level property
	//
	// NOTE:
	//  This is typically used to feed inputs of a task's execution
	// result to next task before the latter's execution
	TaskResultTLP TopLevelProperty = "TaskResult"
)

// VolumeTLPProperty is used to define properties that comes
// after VolumeTLP
type VolumeTLPProperty string

const (
	// OwnerVTP indicates the owner of this volume; the one who
	// is executing this policy
	//
	// NOTE:
	//  The corresponding value will be accessed as
	// {{ .Volume.owner }}
	OwnerVTP VolumeTLPProperty = "owner"
	// RunNamespaceVTP is the namespace where this policy is
	// supposed to run
	// NOTE:
	//  The corresponding value will be accessed as
	// {{ .Volume.runNamespace }}
	RunNamespaceVTP VolumeTLPProperty = "runNamespace"
	// CapacityVTP is the capacity of the volume
	// NOTE:
	//  The corresponding value will be accessed as
	// {{ .Volume.capacity }}
	CapacityVTP VolumeTLPProperty = "capacity"
	// PersistentVolumeClaimVTP is the PVC of the volume
	// NOTE:
	//  The corresponding value will be accessed as
	// {{ .Volume.pvc }}
	PersistentVolumeClaimVTP VolumeTLPProperty = "pvc"
)

// PolicyTLPProperty is the name of the property that is found
// under PolicyTLP
type PolicyTLPProperty string

const (
	// EnabledPTP is the enabled property of the policy
	// NOTE:
	//  The corresponding value will be accessed as
	// {{ .Policy.<PolicyName>.enabled }}
	EnabledPTP PolicyTLPProperty = "enabled"
	// ValuePTP is the value property of the policy
	// NOTE:
	//  The corresponding value will be accessed as
	// {{ .Policy.<PolicyName>.value }}
	ValuePTP PolicyTLPProperty = "value"
	// DataPTP is the data property of the policy
	// NOTE:
	//  The corresponding value will be accessed as
	// {{ .Policy.<PolicyName>.data }}
	DataPTP PolicyTLPProperty = "data"
)

const (
	TaskIdentityPrefix string = "key"
)

// TaskTLPProperty is the name of the property that is found
// under TaskTLP
type TaskTLPProperty string

const (
	// APIVersionTTP is the apiVersion property of the task
	// NOTE:
	//  The corresponding value will be accessed as
	// {{ .Task.<TaskIdentity>.apiVersion }}
	APIVersionTTP TaskTLPProperty = "apiVersion"
	// KindTTP is the kind property of the task
	// NOTE:
	//  The corresponding value will be accessed as
	// {{ .Task.<TaskIdentity>.kind }}
	KindTTP TaskTLPProperty = "kind"
)

// TaskResultTLPProperty is the name of the property that is found
// under TaskResultTLP
type TaskResultTLPProperty string

const (
	// ObjectNameTRTP is the objectName property of the
	// TaskResultTLP
	//
	// NOTE:
	//  The corresponding value will be accessed as
	// {{ .TaskResult.<TaskIdentity>.objectName }}
	ObjectNameTRTP TaskResultTLPProperty = "objectName"
	// AnnotationsTRTP is the annotations property of the
	// TaskResultTLP
	//
	// NOTE:
	//  The corresponding value will be accessed as
	// {{ .TaskResult.<TaskIdentity>.annotations }}
	AnnotationsTRTP TaskResultTLPProperty = "annotations"
)
