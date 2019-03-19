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
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/template"

	m_k8s_client "github.com/openebs/maya/pkg/client/k8s"
	mach_apis_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// newK8sClient returns a new instance of K8sClient based on the provided run
// namespace.
//
// NOTE:
//  Providing a run namespace can be optional. It is optional for cluster wide
// operations.
//
// NOTE:
//  In cases where more than one namespaces are involved, **repeat**
// metatask property is used.
func newK8sClient(namespace string) (kc *m_k8s_client.K8sClient, err error) {
	ns := strings.TrimSpace(namespace)
	kc, err = m_k8s_client.NewK8sClient(ns)
	return
}

// MetaTaskAction represents the action type of RunTask
type MetaTaskAction string

const (
	// GetTA flags a action as get. Typically used to fetch
	// an object from its name.
	GetTA MetaTaskAction = "get"
	// ListTA flags a action as list. Typically used to fetch
	// a list of objects based on options.
	ListTA MetaTaskAction = "list"
	// PutTA flags a action as put. Typically used to put
	// an object.
	PutTA MetaTaskAction = "put"
	// DeleteTA flags a action as delete. Typically used to
	// delete an object.
	DeleteTA MetaTaskAction = "delete"
	// PatchTA flags a action as patch. Typically used to
	// patch an object.
	PatchTA MetaTaskAction = "patch"
	// ExecTA flags a action as exec. Typically used to
	// exec inside a container of a pod.
	ExecTA MetaTaskAction = "exec"
	// OutputTA flags the task action as output. Typically used to
	// provide a schema (i.e. a custom defined) based output after
	// running one or more tasks.
	OutputTA MetaTaskAction = "output"
)

// MetaTaskProps provides properties representing the task's meta
// information
type MetaTaskProps struct {
	// RunNamespace is the namespace where task will get
	// executed
	RunNamespace string `json:"runNamespace"`

	// Owner represents the owner of this task
	Owner string `json:"owner"`

	// ObjectName is the name of the resource that gets
	// created or will get operated by this task
	ObjectName string `json:"objectName"`

	// Options is a set of selectors that can be used for
	// tasks that are get or list based actions
	Options string `json:"options"`

	// Retry specifies the no. of times this particular task (i.e. all properties
	// remains same) can be re-tried. This is typically used along with task
	// result verify options for get or list related actions.
	//
	// A sample retry option:
	//
	// # max of 10 attempts in 20 seconds interval
	// retry: "10,20s"
	Retry string `json:"retry"`

	// Disable will disable execution of this task
	Disable bool `json:"disable"`
}

// toString returns a string representation of MetaTaskProps structure. In this
// representation, each property is separated from its value via '='. In
// addition each property=value pair is separated from other pair via '::'.
//
// Example:
//  runNamespace=default::objectName=MySvc::retry=3,20s
func (m MetaTaskProps) toString() string {
	return fmt.Sprintf("runNamespace=%s::owner=%s::objectName=%s::options=%s::retry=%s::disable=%t",
		m.RunNamespace,
		m.Owner,
		m.ObjectName,
		m.Options,
		m.Retry,
		m.Disable)
}

// selectOverride will override the current meta task properties from the given
// if the given's properties has value
func (m MetaTaskProps) selectOverride(given MetaTaskProps) MetaTaskProps {
	namespace := strings.TrimSpace(given.RunNamespace)
	if len(namespace) != 0 {
		m.RunNamespace = namespace
	}
	owner := strings.TrimSpace(given.Owner)
	if len(owner) != 0 {
		m.Owner = owner
	}
	objectname := strings.TrimSpace(given.ObjectName)
	if len(objectname) != 0 {
		m.ObjectName = objectname
	}
	options := strings.TrimSpace(given.Options)
	if len(options) != 0 {
		m.Options = options
	}
	retry := strings.TrimSpace(given.Retry)
	if len(retry) != 0 {
		m.Retry = retry
	}
	m.Disable = given.Disable

	return m
}

// MetaTaskSpec is the specifications of a MetaTask
type MetaTaskSpec struct {
	// MetaTaskIdentity provides the identity to this task
	MetaTaskIdentity
	// MetaTaskProps provides the task's meta related properties
	MetaTaskProps
	// Action representing this task
	//
	// e.g. get based task or list based task or put based task and so on
	Action MetaTaskAction `json:"action"`
	// RepeatWith sets one or more resources for repetitive execution.
	// In other words a task template is executed multiple times based on each
	// of the item present here.
	RepeatWith RepeatWithResource `json:"repeatWith"`
}

type metaTaskExecutor struct {
	// metaTask holds the task's meta information
	metaTask MetaTaskSpec
	// identifier exposes a task's identity related operations
	identifier taskIdentifier
	// repeater exposes operations with respect to repetitive execution of this
	// task
	repeater repeatExecutor
	// k8sClient will be used to make K8s API calls
	k8sClient *m_k8s_client.K8sClient
}

// getMetaInstances is a utility function that provides required objects
// to instantiate meta task executor
func getMetaInstances(metaTaskYml string, values map[string]interface{}) (m MetaTaskSpec, i taskIdentifier, r repeatExecutor, err error) {
	// transform the yaml with provided values
	b, err := template.AsTemplatedBytes("MetaTaskSpec", metaTaskYml, values)
	if err != nil {
		return
	}

	// unmarshall the yaml bytes into m
	err = yaml.Unmarshal(b, &m)
	if err != nil {
		return
	}

	// instantiate the task identifier based out of this MetaTask
	i, err = newTaskIdentifier(m.MetaTaskIdentity)
	if err != nil {
		return
	}

	r, err = newRepeatExecutor(m.RepeatWith)
	if err != nil {
		return
	}

	return
}

// newMetaTaskExecutor provides a new instance of metaTaskExecutor
func newMetaTaskExecutor(metaTaskYml string, values map[string]interface{}) (*metaTaskExecutor, error) {

	m, i, r, err := getMetaInstances(metaTaskYml, values)
	if err != nil {
		return nil, err
	}

	k, err := newK8sClient(m.RunNamespace)
	if err != nil {
		return nil, err
	}

	return &metaTaskExecutor{
		metaTask:   m,
		identifier: i,
		repeater:   r,
		k8sClient:  k,
	}, nil
}

func (m *metaTaskExecutor) getMetaInfo() MetaTaskSpec {
	return m.metaTask
}

func (m *metaTaskExecutor) getRepeatExecutor() repeatExecutor {
	return m.repeater
}

func (m *metaTaskExecutor) isDisabled() bool {
	return m.metaTask.Disable
}

func (m *metaTaskExecutor) getIdentity() string {
	return m.metaTask.Identity
}

func (m *metaTaskExecutor) getTaskIdentity() MetaTaskIdentity {
	return m.metaTask.MetaTaskIdentity
}

func (m *metaTaskExecutor) getObjectName() string {
	return m.metaTask.ObjectName
}

func (m *metaTaskExecutor) getRunNamespace() string {
	return m.metaTask.RunNamespace
}

func (m *metaTaskExecutor) getK8sClient() *m_k8s_client.K8sClient {
	return m.k8sClient
}

func (m *metaTaskExecutor) getRetry() (attempts int, interval time.Duration) {
	retry := m.metaTask.Retry
	// "attempts,interval" format
	defRetry := "0,0s"

	// retry is a comma separated string with attempts as first element &
	// interval as second element
	retryArr := strings.Split(retry, ",")
	if len(retryArr) != 2 {
		retryArr = strings.Split(defRetry, ",")
	}

	// determine the attempts
	attempts, _ = strconv.Atoi(retryArr[0])
	if attempts < 0 {
		// no retries for negative attempt value
		attempts = 0
	}
	// determine the interval
	interval, _ = time.ParseDuration(retryArr[1])

	return
}

// getListOptions unmarshall the options in yaml doc format into meta.ListOptions
func (m *metaTaskExecutor) getListOptions() (opts mach_apis_meta_v1.ListOptions, err error) {
	err = yaml.Unmarshal([]byte(m.metaTask.Options), &opts)
	return
}

func (m *metaTaskExecutor) isCommand() bool {
	return m.identifier.isCommand()
}

func (m *metaTaskExecutor) isList() bool {
	return m.metaTask.Action == ListTA
}

func (m *metaTaskExecutor) isGet() bool {
	return m.metaTask.Action == GetTA
}

func (m *metaTaskExecutor) isPut() bool {
	return m.metaTask.Action == PutTA
}

func (m *metaTaskExecutor) isDelete() bool {
	return m.metaTask.Action == DeleteTA
}

func (m *metaTaskExecutor) isPatch() bool {
	return m.metaTask.Action == PatchTA
}

func (m *metaTaskExecutor) isExec() bool {
	return m.metaTask.Action == ExecTA
}

func (m *metaTaskExecutor) isPutExtnV1B1Deploy() bool {
	return m.identifier.isExtnV1B1Deploy() && m.isPut()
}

func (m *metaTaskExecutor) isPutBatchV1Job() bool {
	return m.identifier.isBatchV1Job() && m.isPut()
}

func (m *metaTaskExecutor) isPutAppsV1STS() bool {
	return m.identifier.isAppsV1STS() && m.isPut()
}

func (m *metaTaskExecutor) isPatchExtnV1B1Deploy() bool {
	return m.identifier.isExtnV1B1Deploy() && m.isPatch()
}

func (m *metaTaskExecutor) isPutAppsV1B1Deploy() bool {
	return m.identifier.isAppsV1B1Deploy() && m.isPut()
}

func (m *metaTaskExecutor) isPatchOEV1alpha1SPC() bool {
	return m.identifier.isStoragePoolClaim() && m.isPatch()
}

func (m *metaTaskExecutor) isPatchAppsV1B1Deploy() bool {
	return m.identifier.isAppsV1B1Deploy() && m.isPatch()
}

func (m *metaTaskExecutor) isPutCoreV1Service() bool {
	return m.identifier.isCoreV1Service() && m.isPut()
}

func (m *metaTaskExecutor) isPatchCoreV1Service() bool {
	return m.identifier.isCoreV1Service() && m.isPatch()
}

func (m *metaTaskExecutor) isDeleteExtnV1B1Deploy() bool {
	return m.identifier.isExtnV1B1Deploy() && m.isDelete()
}

func (m *metaTaskExecutor) isDeleteExtnV1B1ReplicaSet() bool {
	return m.identifier.isExtnV1B1ReplicaSet() && m.isDelete()
}

func (m *metaTaskExecutor) isDeleteAppsV1B1Deploy() bool {
	return m.identifier.isAppsV1B1Deploy() && m.isDelete()
}

func (m *metaTaskExecutor) isDeleteCoreV1Service() bool {
	return m.identifier.isCoreV1Service() && m.isDelete()
}

func (m *metaTaskExecutor) isListCoreV1PVC() bool {
	return m.identifier.isCoreV1PVC() && m.isList()
}

func (m *metaTaskExecutor) isListCoreV1PV() bool {
	return m.identifier.isCoreV1PV() && m.isList()
}

func (m *metaTaskExecutor) isListCoreV1Pod() bool {
	return m.identifier.isCoreV1Pod() && m.isList()
}

func (m *metaTaskExecutor) isListCoreV1Service() bool {
	return m.identifier.isCoreV1Service() && m.isList()
}

func (m *metaTaskExecutor) isListExtnV1B1Deploy() bool {
	return m.identifier.isExtnV1B1Deploy() && m.isList()
}

func (m *metaTaskExecutor) isListAppsV1B1Deploy() bool {
	return m.identifier.isAppsV1B1Deploy() && m.isList()
}

func (m *metaTaskExecutor) isGetStorageV1SC() bool {
	return m.identifier.isStorageV1SC() && m.isGet()
}

func (m *metaTaskExecutor) isGetBatchV1Job() bool {
	return m.identifier.isBatchV1Job() && m.isGet()
}

func (m *metaTaskExecutor) isGetExtnV1B1Deploy() bool {
	return m.identifier.isExtnV1B1Deploy() && m.isGet()
}

func (m *metaTaskExecutor) isDeleteBatchV1Job() bool {
	return m.identifier.isBatchV1Job() && m.isDelete()
}

func (m *metaTaskExecutor) isDeleteAppsV1STS() bool {
	return m.identifier.isAppsV1STS() && m.isDelete()
}

func (m *metaTaskExecutor) isGetOEV1alpha1Disk() bool {
	return m.identifier.isOEV1alpha1Disk() && m.isGet()
}

func (m *metaTaskExecutor) isGetOEV1alpha1SPC() bool {
	return m.identifier.isOEV1alpha1SPC() && m.isGet()
}
func (m *metaTaskExecutor) isGetOEV1alpha1SP() bool {
	return m.identifier.isOEV1alpha1SP() && m.isGet()
}

func (m *metaTaskExecutor) isGetOEV1alpha1UR() bool {
	return m.identifier.isOEV1alpha1UR() && m.isGet()
}

func (m *metaTaskExecutor) isGetCoreV1PVC() bool {
	return m.identifier.isCoreV1PVC() && m.isGet()
}

func (m *metaTaskExecutor) isGetCoreV1PV() bool {
	return m.identifier.isCoreV1PV() && m.isGet()
}

func (m *metaTaskExecutor) isPutOEV1alpha1SP() bool {
	return m.identifier.isOEV1alpha1SP() && m.isPut()
}

func (m *metaTaskExecutor) isPutOEV1alpha1CSP() bool {
	return m.identifier.isOEV1alpha1CSP() && m.isPut()
}

func (m *metaTaskExecutor) isPutOEV1alpha1CSV() bool {
	return m.identifier.isOEV1alpha1CV() && m.isPut()
}

func (m *metaTaskExecutor) isPutOEV1alpha1CVR() bool {
	return m.identifier.isOEV1alpha1CVR() && m.isPut()
}

func (m *metaTaskExecutor) isDeleteOEV1alpha1SP() bool {
	return m.identifier.isOEV1alpha1SP() && m.isDelete()
}

func (m *metaTaskExecutor) isDeleteOEV1alpha1CSP() bool {
	return m.identifier.isOEV1alpha1CSP() && m.isDelete()
}

func (m *metaTaskExecutor) isDeleteOEV1alpha1CSV() bool {
	return m.identifier.isOEV1alpha1CV() && m.isDelete()
}

func (m *metaTaskExecutor) isDeleteOEV1alpha1CVR() bool {
	return m.identifier.isOEV1alpha1CVR() && m.isDelete()
}

func (m *metaTaskExecutor) isPatchOEV1alpha1CSV() bool {
	return m.identifier.isOEV1alpha1CV() && m.isPatch()
}

func (m *metaTaskExecutor) isPatchOEV1alpha1CVR() bool {
	return m.identifier.isOEV1alpha1CVR() && m.isPatch()
}

func (m *metaTaskExecutor) isListOEV1alpha1Disk() bool {
	return m.identifier.isOEV1alpha1Disk() && m.isList()
}

func (m *metaTaskExecutor) isListOEV1alpha1SP() bool {
	return m.identifier.isOEV1alpha1SP() && m.isList()
}

func (m *metaTaskExecutor) isListOEV1alpha1CSP() bool {
	return m.identifier.isOEV1alpha1CSP() && m.isList()
}

func (m *metaTaskExecutor) isListOEV1alpha1CVR() bool {
	return m.identifier.isOEV1alpha1CVR() && m.isList()
}

func (m *metaTaskExecutor) isListOEV1alpha1UR() bool {
	return m.identifier.isOEV1alpha1UR() && m.isList()
}

func (m *metaTaskExecutor) isListOEV1alpha1CV() bool {
	return m.identifier.isOEV1alpha1CV() && m.isList()
}

func (m *metaTaskExecutor) isExecCoreV1Pod() bool {
	return m.identifier.isCoreV1Pod() && m.isExec()
}

// getRollbackMetaInstances is a utility function that provides objects
// required to build a rollback based meta task executor
func getRollbackMetaInstances(given MetaTaskSpec, objectName string) (m MetaTaskSpec, i taskIdentifier, err error) {
	m = MetaTaskSpec{
		// rollback currently understands only Delete action
		Action: DeleteTA,
		MetaTaskProps: MetaTaskProps{
			ObjectName:   objectName,
			RunNamespace: given.RunNamespace,
			Owner:        given.Owner,
		},
		MetaTaskIdentity: given.MetaTaskIdentity,
	}

	// instantiate the task identifier based out of this MetaTaskSpec
	i, err = newTaskIdentifier(m.MetaTaskIdentity)
	return
}

// asRollbackInstance defines a metaTaskExecutor suitable for
// rollback operation.
//
// It translates a `put` action into a `delete` action
// keeping the objectName & other properties
// of the rollback task same as the original task
//
// NOTE:
//  The bool return with value as `false` implies there is no
// need for a rollback
func (m *metaTaskExecutor) asRollbackInstance(objectName string) (*metaTaskExecutor, bool, error) {
	// there is no rollback when task is disabled
	// there is no rollback when original action is not put
	if m.isDisabled() || !m.isPut() {
		return nil, false, nil
	}

	if len(objectName) == 0 {
		errMsg := fmt.Sprintf("failed to build rollback instance for task '%s': object name is missing", m.getIdentity())
		glog.Errorf(fmt.Sprintf("%s: meta task '%+v'", errMsg, m.getMetaInfo()))
		return nil, true, fmt.Errorf(errMsg)
	}

	rbSpec, i, err := getRollbackMetaInstances(m.metaTask, objectName)
	if err != nil {
		return nil, true, err
	}

	k, err := newK8sClient(rbSpec.RunNamespace)
	if err != nil {
		return nil, true, err
	}

	return &metaTaskExecutor{
		metaTask:   rbSpec,
		identifier: i,
		k8sClient:  k,
	}, true, nil
}

// getRepeatMetaInstances is a utility function that provides various objects
// required to build a repeat meta task executor
func getRepeatMetaInstances(given MetaTaskSpec, repeatIndex int) (m MetaTaskSpec, i taskIdentifier, r repeatExecutor, err error) {
	// build a meta task spec from the given one
	m = MetaTaskSpec{
		Action:           given.Action,
		MetaTaskProps:    given.MetaTaskProps,
		MetaTaskIdentity: given.MetaTaskIdentity,
		RepeatWith:       given.RepeatWith,
	}

	// instantiate the task identifier based out of this MetaTask
	i, err = newTaskIdentifier(m.MetaTaskIdentity)
	if err != nil {
		return
	}

	r, err = newRepeatExecutor(m.RepeatWith)
	if err != nil {
		return
	}

	// get repeat meta props based on index
	rptMetaProps, err := r.getMeta(repeatIndex)
	if err != nil {
		return
	}

	// final meta task spec corresponding the to repeat index
	m.MetaTaskProps = m.MetaTaskProps.selectOverride(rptMetaProps)
	return
}

// asRepeatInstance returns a new instance of metaTaskExecutor
// based on the provided meta task properties
func (m *metaTaskExecutor) asRepeatInstance(repeatIndex int) (*metaTaskExecutor, error) {
	if !m.repeater.isMetaRepeat() {
		// old executor will suffice if repeater is not based on meta task
		return m, nil
	}

	rSpec, i, r, err := getRepeatMetaInstances(m.metaTask, repeatIndex)
	if err != nil {
		return nil, err
	}

	k, err := newK8sClient(rSpec.RunNamespace)
	if err != nil {
		return nil, err
	}

	return &metaTaskExecutor{
		metaTask:   rSpec,
		identifier: i,
		repeater:   r,
		k8sClient:  k,
	}, nil
}
