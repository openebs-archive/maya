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

type CasKey string

const (
	// This is the cas template annotation whose value is the name of
	// cas template that will be used to provision a storagepool
	SPCreateCASTemplateCK CasKey = "cas.openebs.io/create-pool-template"

	// This is the cas template annotation whose value is the name of
	// cas template that will be used to delete a storagepool
	SPDeleteCASTemplateCK CasKey = "cas.openebs.io/delete-pool-template"
)

// TopLevelProperty represents the top level property that
// is a starting point to represent a hierarchical chain of
// properties.
//
// e.g.
// Config.prop1.subprop1 = val1
// Config.prop1.subprop2 = val2
// In above example Config is a top level object
//
// NOTE:
//  The value of any hierarchical chain of properties
// can be parsed via dot notation
type TopLevelProperty string

const (
	// ConfigTLP is a top level property supported by CAS template engine
	//
	// The policy specific properties are placed with ConfigTLP as the
	// top level property
	ConfigTLP TopLevelProperty = "Config"
	// VolumeTLP is a top level property supported by CAS template engine
	//
	// The properties provided by the caller are placed with VolumeTLP
	// as the top level property
	//
	// NOTE:
	//  CAS template engine cannot modify these properties. These are the
	// runtime properties that are provided as inputs to CAS template
	// engine.
	VolumeTLP TopLevelProperty = "Volume"
	// StoragePoolTLP is a top level property supported by CAS template engine
	//
	// The properties provided by the caller are placed with StoragePoolTLP
	// as the top level property
	//
	// NOTE:
	//  CAS template engine cannot modify these properties. These are the
	// runtime properties that are provided as inputs to CAS template
	// engine.
	StoragePoolTLP TopLevelProperty = "Storagepool"
	// TaskResultTLP is a top level property supported by CAS template engine
	//
	// The specific results after the execution of a task are placed with
	// TaskResultTLP as the top level property
	//
	// NOTE:
	//  This is typically used to feed inputs of a task's execution
	// result to **next task** before the later's execution
	TaskResultTLP TopLevelProperty = "TaskResult"
	// CurrentJsonDocTLP is a top level property supported by CAS template engine
	//
	// The result of the current task's execution is stored in this top
	// level property.
	CurrentJsonResultTLP TopLevelProperty = "JsonResult"
	// ListItemsTLP is a top level property supported by CAS template engine
	//
	// Results of one or more tasks' execution can be saved in this property.
	//
	// Example:
	//  Below shows how specific properties of a list of items can be retrieved in
	// a go template. Below dot notation is for illustration purposes and only
	// reflects the way the specific property value was set.
	//
	// {{- .ListItems.volumes.default.mypv2.ip -}}
	// {{- .ListItems.volumes.default.mypv2.status -}}
	// {{- .ListItems.volumes.openebs.mypv.ip -}}
	// {{- .ListItems.volumes.openebs.mypv.status -}}
	ListItemsTLP TopLevelProperty = "ListItems"
)

// StoragePoolTLPProperty is used to define properties that comes
// after StoragePoolTLP
type StoragePoolTLPProperty string

const (
	// OwnerCTP indicates the owner of this pool; the one who
	// is executing this policy
	//
	// NOTE:
	//  The corresponding value will be accessed as
	// {{ .Storagepool.owner }}
	OwnerCTP StoragePoolTLPProperty = "owner"
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
	// TaskResultVerifyErrTRTP is a property of TaskResultTLP
	//
	// First error found after **verification** checks done against the result of
	// the task's execution is stored in this property.
	//
	// NOTE:
	//  The corresponding value will be accessed as
	// {{ .TaskResult.<TaskIdentity>.verifyErr }}
	TaskResultVerifyErrTRTP TaskResultTLPProperty = "verifyErr"
	// TaskResultNotFoundErrTRTP is a property of TaskResultTLP
	//
	// First error found after **not found** checks done against the result of
	// the task's execution is stored in this property.
	//
	// NOTE:
	//  The corresponding value will be accessed as
	// {{ .TaskResult.<TaskIdentity>.notFoundErr }}
	TaskResultNotFoundErrTRTP TaskResultTLPProperty = "notFoundErr"
)

// ListItemsTLPProperty is the name of the property that is found
// under ListItemsTLP
type ListItemsTLPProperty string

const (
	// CurrentRepeatResourceLITP is a property of ListItemsTLP
	//
	// It is the current repeat resource due to which a task is getting
	// executed is set here
	//
	// Example:
	// {{- .ListItems.currentRepeatResource -}}
	//
	// Above templating will give the current repeat resource name
	CurrentRepeatResourceLITP ListItemsTLPProperty = "currentRepeatResource"
)
