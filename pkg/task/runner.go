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
	"strconv"

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
	metaTaskYml string
	taskYml     string
}

// TaskRunner helps in running a set of Tasks in sequence
type TaskRunner struct {
	// taskSpecs is an array of task specifications
	taskSpecs []taskSpecHolder
	// rollbacks is an array of tasks that need to be run in
	// sequence in the event of any error
	rollbacks []*Task
}

func NewTaskRunner() *TaskRunner {
	return &TaskRunner{}
}

func (m *TaskRunner) AddTaskSpec(metaTaskYml, taskYml string) {
	tSpec := taskSpecHolder{
		metaTaskYml: metaTaskYml,
		taskYml:     taskYml,
	}

	m.taskSpecs = append(m.taskSpecs, tSpec)
}

// planForRollback in case of errors in executing next tasks.
// This will add to the list of rollback tasks
func (m *TaskRunner) planForRollback(t *Task, objectName string) {
	// the entire rollback policy is encapsulated
	// in the task itself; so just invoke this method
	rt, err := t.asRollback(objectName)
	if err != nil {
		glog.Warningf("Error during rollback plan: Task: '%#v' Error: '%s'", t, err.Error())
	}

	if rt == nil {
		// this task does not need a rollback or
		// can not be rollback-ed in-case of above error
		return
	}

	m.rollbacks = append(m.rollbacks, rt)
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
func (m *TaskRunner) runTasks(values map[string]interface{}, postTaskRunFn PostTaskRunFn) error {
	for idx, tSpec := range m.taskSpecs {

		// suggest task index as task's identity
		// NOTE: this may not be set if task's yaml has set its identity already
		id := v1alpha1.TaskIdentityPrefix + strconv.Itoa(idx)
		// convert the yml to task
		t, err := NewTask(id, tSpec.metaTaskYml, tSpec.taskYml, values)
		if err != nil {
			return err
		}

		// actual task execution
		result, err := t.execute()
		if err != nil {
			return err
		}

		// these are some post task execution steps
		// if execution was successful then add it to rollback set;
		// this set will be used incase of a failure while executing the
		// next or future task(s)
		// get the object that was created by this task
		taskResults := util.GetMapOfStrings(result, t.Identity)
		if taskResults == nil {
			glog.Errorf("Nil task results: Invalid task execution: Task: '%#v' Result: '%#v'", t, result)
			return fmt.Errorf("Nil task results: Invalid task execution: Task: '%s'", t.Identity)
		}

		// extract the name of this object
		objName := taskResults[string(v1alpha1.ObjectNameTRTP)]
		if len(objName) == 0 {
			glog.Errorf("Missing object name: Invalid task execution: Task: '%#v' Result: '%#v'", t, result)
			return fmt.Errorf("Missing object name: Invalid task execution: Task: '%s'", t.Identity)
		}

		m.planForRollback(t, objName)

		// this provides a way to captures the result of this task
		// Use this function & result to provide data to the next task
		// before executing the next task
		postTaskRunFn(result)
	}

	return nil
}

// Run will run all the defined tasks & will rollback in case of
// any error
func (m *TaskRunner) Run(values map[string]interface{}, postTaskRunFn PostTaskRunFn) error {
	err := m.runTasks(values, postTaskRunFn)
	if err != nil {
		glog.Errorf("Failed to run: Will rollback: Error: '%s' Values: '%#v'", err.Error(), values)
		m.rollback()
	}

	return err
}
