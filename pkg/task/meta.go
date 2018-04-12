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
	"strconv"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/openebs/maya/pkg/template"

	mach_apis_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TaskAction signifies the action to be taken
// against a Task
type TaskAction string

const (
	// GetTA flags a action as get. Typically used to fetch
	// an object from its name.
	GetTA TaskAction = "get"
	// ListTA flags a action as list. Typically used to fetch
	// a list of objects based on options.
	ListTA TaskAction = "list"
	// PutTA flags a action as put. Typically used to put
	// an object.
	PutTA TaskAction = "put"
	// DeleteTA flags a action as delete. Typically used to
	// delete an object.
	DeleteTA TaskAction = "delete"
	// PatchTA flags a action as patch. Typically used to
	// patch an object.
	PatchTA TaskAction = "patch"
)

// MetaTask contains information about a Task
type MetaTask struct {
	// TaskIdentity provides the required identity
	// about the task
	TaskIdentity
	// RunNamespace is the namespace where task will get
	// executed
	RunNamespace string `json:"runNamespace"`
	// Action to be invoked on the task
	Action TaskAction `json:"action"`
	// Owner represents the owner of this task
	Owner string `json:"owner"`
	// ObjectName is the name of the target that is
	// created or operated by the task
	ObjectName string `json:"objectName"`
	// Options is a set of selectors that can be used for
	// get or list actions
	Options string `json:"options"`
	// TaskResultQueries will consist of the queries to be run against the
	// task's result
	TaskResultQueries []TaskResultQuery `json:"queries"`
	// Retry specifies the no. of times this task can be tried
	// This is typically used along with task result verify options
	// for get or list related actions
	//
	// A sample retry option:
	//
	// # max of 10 attempts in 20 seconds interval
	// retry: "10,20s"
	Retry string `json:"retry"`
	// TaskPatch will consist of patches that gets applied
	// against the task object
	TaskPatch `json:"patch"`
}

type metaTaskExecutor struct {
	// metaTask holds various metadata related info w.r.t the task
	metaTask MetaTask
	// identifier is a utility struct that enables a task's identity
	// related operations
	identifier taskIdentifier
}

// newMetaTaskExecutor provides a new instance of metaTaskExecutor
func newMetaTaskExecutor(identity, yml string, values map[string]interface{}) (*metaTaskExecutor, error) {
	// transform the yaml with provided values
	b, err := template.AsTemplatedBytes("MetaTask", yml, values)
	if err != nil {
		return nil, err
	}

	// unmarshall the yaml bytes into this instance
	var m MetaTask
	err = yaml.Unmarshal(b, &m)
	if err != nil {
		return nil, err
	}

	// set the task identity
	m.Identity = identity

	// instantiate the task identifier based out of this MetaTask
	i, err := newTaskIdentifier(m.TaskIdentity)
	if err != nil {
		return nil, err
	}

	return &metaTaskExecutor{
		metaTask:   m,
		identifier: i,
	}, nil
}

func (m *metaTaskExecutor) getMetaInfo() MetaTask {
	return m.metaTask
}

func (m *metaTaskExecutor) getTaskResultQueries() []TaskResultQuery {
	return m.metaTask.TaskResultQueries
}

func (m *metaTaskExecutor) getTaskPatch() TaskPatch {
	return m.metaTask.TaskPatch
}

func (m *metaTaskExecutor) getObjectName() string {
	return m.metaTask.ObjectName
}

func (m *metaTaskExecutor) getRunNamespace() string {
	return m.metaTask.RunNamespace
}

func (m *metaTaskExecutor) getRetry() (attempts int, interval time.Duration) {
	retry := m.metaTask.Retry
	// "attempts,interval" format
	defRetry := "1,0s"

	// retry is a comma separated string with attempts as first element &
	// interval as second element
	retryArr := strings.Split(retry, ",")
	if len(retryArr) != 2 {
		retryArr = strings.Split(defRetry, ",")
	}

	// determine the attempts
	attempts, _ = strconv.Atoi(retryArr[0])
	if attempts == 0 {
		attempts = 1
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

func (m *metaTaskExecutor) isPutExtnV1B1Deploy() bool {
	return m.identifier.isExtnV1B1Deploy() && m.isPut()
}

func (m *metaTaskExecutor) isPatchExtnV1B1Deploy() bool {
	return m.identifier.isExtnV1B1Deploy() && m.isPatch()
}

func (m *metaTaskExecutor) isPutAppsV1B1Deploy() bool {
	return m.identifier.isAppsV1B1Deploy() && m.isPut()
}

func (m *metaTaskExecutor) isPatchAppsV1B1Deploy() bool {
	return m.identifier.isAppsV1B1Deploy() && m.isPatch()
}

func (m *metaTaskExecutor) isPutCoreV1Service() bool {
	return m.identifier.isCoreV1Service() && m.isPut()
}

func (m *metaTaskExecutor) isDeleteExtnV1B1Deploy() bool {
	return m.identifier.isExtnV1B1Deploy() && m.isDelete()
}

func (m *metaTaskExecutor) isDeleteAppsV1B1Deploy() bool {
	return m.identifier.isAppsV1B1Deploy() && m.isDelete()
}

func (m *metaTaskExecutor) isDeleteCoreV1Service() bool {
	return m.identifier.isCoreV1Service() && m.isDelete()
}

func (m *metaTaskExecutor) isListCoreV1Pod() bool {
	return m.identifier.isCoreV1Pod() && m.isList()
}

func (m *metaTaskExecutor) isGetOEV1alpha1SP() bool {
	return m.identifier.isOEV1alpha1SP() && m.isGet()
}

func (m *metaTaskExecutor) isGetCoreV1PVC() bool {
	return m.identifier.isCoreV1PVC() && m.isGet()
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
	// there is no rollback when original action is not put
	if !m.isPut() {
		return nil, false, nil
	}

	// build the rollback version of this MetaTask
	rbMT := MetaTask{
		Action:     DeleteTA,
		ObjectName: objectName,
		TaskIdentity: TaskIdentity{
			Identity:   m.metaTask.Identity,
			Kind:       m.metaTask.Kind,
			APIVersion: m.metaTask.APIVersion,
		},
		RunNamespace: m.metaTask.RunNamespace,
		Owner:        m.metaTask.Owner,
	}

	// instantiate the task identifier based out of this MetaTask
	i, err := newTaskIdentifier(m.metaTask.TaskIdentity)
	if err != nil {
		return nil, true, err
	}

	return &metaTaskExecutor{
		metaTask:   rbMT,
		identifier: i,
	}, true, nil
}
