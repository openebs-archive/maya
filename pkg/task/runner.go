/*
Copyright 2017 The OpenEBS Authors

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
	"strings"

	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/util"
)

// PostTaskRunFn is a closure definition that provides option
// to act on individual task's result
type PostTaskRunFn func(taskResult map[string]interface{})

// taskSpecHolder is a utility structure that composes specifications
// of task as well as metatask
type taskSpecHolder struct {
	identity    string
	metaTaskYml string
	taskYml     string
}

// TaskRunner helps in running a set of Tasks in sequence
type TaskRunner struct {
	// taskSpecs is an array of task specifications
	taskSpecs []taskSpecHolder
	// rollbacks is an array of tasks that need to be run in
	// sequence in the event of any error
	rollbacks []*taskExecutor
}

func NewTaskRunner() *TaskRunner {
	return &TaskRunner{}
}

func (m *TaskRunner) AddTaskSpec(identity, metaTaskYml, taskYml string) error {
	if len(strings.TrimSpace(identity)) == 0 {
		fmt.Errorf("Missing task identity: Failed to add task spec")
	}

	tSpec := taskSpecHolder{
		identity:    strings.TrimSpace(identity),
		metaTaskYml: metaTaskYml,
		taskYml:     taskYml,
	}
	m.taskSpecs = append(m.taskSpecs, tSpec)

	return nil
}

// planForRollback in case of errors in executing next tasks.
// This will add to the list of rollback tasks
//
// NOTE:
//  This is just the planning for rollback & not actual rollback.
// In the events of issues this planning will be useful.
func (m *TaskRunner) planForRollback(tSpec taskSpecHolder, te *taskExecutor, objectName string) {
	// the entire rollback policy is encapsulated
	// in the task itself; so just invoke this method
	rte, err := te.asRollbackInstance(objectName)
	if err != nil {
		glog.Warningf("Error during rollback plan: Error: '%s' Identity: '%s' Meta YAML '%s' YAML '%s'", err.Error(), tSpec.identity, tSpec.metaTaskYml, tSpec.taskYml)
	}

	if rte == nil {
		// this task does not need a rollback or
		// can not be rollback-ed in-case of above error
		return
	}

	m.rollbacks = append(m.rollbacks, rte)
}

// rollback will rollback the run operation
func (m *TaskRunner) rollback() {
	count := len(m.rollbacks)
	if count == 0 {
		glog.Infof("Nothing to rollback")
		return
	}

	// execute the rollback tasks in reverse order
	for i := count - 1; i >= 0; i-- {
		_, err := m.rollbacks[i].execute()
		if err != nil {
			glog.Warningf("Error during rollback: Task: '%#v' Error: '%s'", m.rollbacks[i], err.Error())
		}
	}
}

// Run will run all tasks in the sequence of provided array
//
// NOTE:
//  The error is logged with verbose details before being returned
func (m *TaskRunner) runTasks(values map[string]interface{}, postTaskRunFn PostTaskRunFn) error {
	for _, tSpec := range m.taskSpecs {
		// build a task executor
		// this is all about utilizing the meta task information
		te, err := newTaskExecutor(tSpec.identity, tSpec.metaTaskYml, tSpec.taskYml, values)
		if err != nil {
			// log with verbose details
			glog.Errorf("Failed to execute task: Identity: '%s' Meta YAML: '%s' Values: '%#v'", tSpec.identity, tSpec.metaTaskYml, values)
			return err
		}

		// actual task execution begins here
		result, err := te.execute()
		if err != nil {
			// log with verbose details
			glog.Errorf("Failed to execute task: Identity: '%s' Meta Task: '%#v' YAML: '%s' Values: '%#v'", tSpec.identity, te.getMetaTaskExecutor(), tSpec.taskYml, values)
			return err
		}

		// these are some post task execution steps
		// if execution was successful then add it to rollback set;
		// this set will be used incase of a failure while executing the
		// next or future task(s)
		// get the object that was created by this task
		taskResults := util.GetMapOfStrings(result, tSpec.identity)
		if taskResults == nil {
			// log with verbose details
			glog.Errorf("Nil task results: could not execute task: Identity: '%s' Meta Task: '%#v' YAML: '%s' Values: '%#v' Result: '%#v'", tSpec.identity, te.getMetaTaskExecutor(), tSpec.taskYml, values, result)
			// return error minus verbosity
			return fmt.Errorf("Nil task results: could not execute task: Identity: '%s'", tSpec.identity)
		}

		// extract the name of this object
		objName := taskResults[string(v1alpha1.ObjectNameTRTP)]
		if len(objName) == 0 {
			// log with verbose details
			glog.Errorf("Missing object name: Invalid task execution: Identity: '%s' Meta Task: '%#v' YAML: '%s' Values: '%#v' Result: '%#v'", tSpec.identity, te.getMetaTaskExecutor(), tSpec.taskYml, values, result)
			// return error minus verbosity
			return fmt.Errorf("Missing object name: Invalid task execution: Identity: '%s'", tSpec.identity)
		}

		// this is planning & not the actual rollback
		m.planForRollback(tSpec, te, objName)

		// executing this closure provides a way to capture the result of this task;
		// this is used to provide data to the next task before the latter's
		// execution
		postTaskRunFn(result)
	}

	return nil
}

// Run will run all the defined tasks & will rollback in case of
// any error
//
// NOTE: values will be modified to include the results from execution of
// each task
func (m *TaskRunner) Run(values map[string]interface{}, postTaskRunFn PostTaskRunFn) error {
	err := m.runTasks(values, postTaskRunFn)
	if err != nil {
		glog.Errorf("Failed to execute task: will rollback previous task(s): Error: '%s'", err.Error())
		m.rollback()
	}

	return err
}
