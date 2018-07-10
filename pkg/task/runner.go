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
// to act on an individual task's result
type PostTaskRunFn func(taskResult map[string]interface{})

// taskSpecHolder is a utility structure that composes specifications
// of task & its metatask
type taskSpecHolder struct {
	identity    string
	metaTaskYml string
	taskYml     string
}

// TaskGroupRunner helps in running a set of Tasks in sequence
type TaskGroupRunner struct {
	// taskSpecs is an array of task specifications
	taskSpecs []taskSpecHolder
	// outputTaskSpec holds the specifications to return this runner's
	// output in the format defined in the output task spec
	outputTaskSpec taskSpecHolder
	// rollbacks is an array of tasks that need to be run in
	// sequence in the event of any error
	rollbacks []*taskExecutor
}

func NewTaskGroupRunner() *TaskGroupRunner {
	return &TaskGroupRunner{}
}

func (m *TaskGroupRunner) AddTaskSpec(identity, metaTaskYml, taskYml string) error {
	if len(strings.TrimSpace(identity)) == 0 {
		fmt.Errorf("failed to add task spec: missing task identity")
	}

	tSpec := taskSpecHolder{
		identity:    strings.TrimSpace(identity),
		metaTaskYml: metaTaskYml,
		taskYml:     taskYml,
	}
	m.taskSpecs = append(m.taskSpecs, tSpec)

	return nil
}

// SetOutputTaskSpec sets this runner with a format that will be used
// to return the output after successful execution of this runner. This output
// format is specified in a task.
func (m *TaskGroupRunner) SetOutputTaskSpec(metaTaskYml, taskYml string) {
	m.outputTaskSpec = taskSpecHolder{
		// it is assumed that there will be only one task that
		// determines the output of this runner; hence the identity
		// is hardcoded here
		identity:    "output",
		metaTaskYml: metaTaskYml,
		taskYml:     taskYml,
	}
	return
}

// planForRollback in case of errors while executing next set of tasks.
// This will add to the list of rollback tasks
//
// NOTE:
//  This is just the planning for rollback & not actual rollback.
// In the events of issues this planning will be useful.
func (m *TaskGroupRunner) planForRollback(tSpec taskSpecHolder, te *taskExecutor, objectName string) {
	// the entire rollback policy is encapsulated
	// in the task itself; so just invoke this method
	rte, err := te.asRollbackInstance(objectName)
	if err != nil {
		glog.Warningf("failed to plan for rollback: error '%s': identity '%s': meta yaml '%s': task yaml '%s'", err.Error(), tSpec.identity, tSpec.metaTaskYml, tSpec.taskYml)
	}

	if rte == nil {
		// this task does not need a rollback or
		// can not be rollback-ed in-case of above error
		return
	}

	m.rollbacks = append(m.rollbacks, rte)
}

// rollback will rollback the run operation
func (m *TaskGroupRunner) rollback() {
	count := len(m.rollbacks)
	if count == 0 {
		glog.Infof("nothing to rollback")
		return
	}

	// execute the rollback tasks in **reverse order**
	for i := count - 1; i >= 0; i-- {
		_, err := m.rollbacks[i].Execute()
		if err != nil {
			// warn this rollback error & continue with the next rollbacks
			glog.Warningf("failed to rollback: task '%#v': error '%s'", m.rollbacks[i], err.Error())
		}
	}
}

// runAllTasks will run all tasks in the sequence as defined in the array
//
// NOTE:
//  The error is logged with verbose details before being returned
func (m *TaskGroupRunner) runAllTasks(values map[string]interface{}, postTaskRunFn PostTaskRunFn) error {
	for _, tSpec := range m.taskSpecs {
		// build a task executor
		// this is all about utilizing the meta task information
		te, err := newTaskExecutor(tSpec.identity, tSpec.metaTaskYml, tSpec.taskYml, values)
		if err != nil {
			// log with verbose details
			glog.Errorf("failed to execute task: identity '%s': meta yaml '%s': template values '%#v'", tSpec.identity, tSpec.metaTaskYml, values)
			return err
		}

		// actual task execution is done here
		result, err := te.Execute()
		if err != nil {
			// log with verbose details
			// TODO
			// log with pretty print !!! Currently task yaml is logged in a readable
			// manner
			glog.Errorf("failed to execute task: identity '%s': meta task '%#v': task yaml '%s': template values '%#v'", tSpec.identity, te.getMetaTaskExecutor(), tSpec.taskYml, values)
			return err
		}

		// below are some of the housekeeping activities post the task execution
		//
		// if execution was successful then add it to rollback plans; these plans
		// will be used incase of a failure while executing the next or future
		// task(s)
		//
		// get the object that was created by this task
		taskResults := util.GetMapOfStrings(result, tSpec.identity)
		if taskResults == nil {
			// log with verbose details
			glog.Errorf("failed to execute task: nil task results: identity '%s': meta task '%#v': task yaml '%s': template values: '%#v': result '%#v'", tSpec.identity, te.getMetaTaskExecutor(), tSpec.taskYml, values, result)
			// return error minus verbosity
			return fmt.Errorf("failed to execute task: nil task results: identity '%s'", tSpec.identity)
		}

		// extract the name of this object
		objName := taskResults[string(v1alpha1.ObjectNameTRTP)]
		if len(objName) == 0 {
			// log with verbose details
			glog.Errorf("failed to execute task: missing object name: identity '%s': meta task '%#v': task yaml '%s': template values '%#v': result: '%#v'", tSpec.identity, te.getMetaTaskExecutor(), tSpec.taskYml, values, result)
			// return error minus verbosity
			return fmt.Errorf("failed to execute task: missing object name: identity '%s'", tSpec.identity)
		}

		// this is planning & not the actual rollback
		m.planForRollback(tSpec, te, objName)

		// executing this closure provides a way to capture the result of this task
		// into the values. This is useful to provide data to the next task before
		// this next task's execution
		postTaskRunFn(result)
	}

	return nil
}

// runOutput gets the output of this runner once all the tasks were executed
// successfully
func (m *TaskGroupRunner) runOutput(values map[string]interface{}) (output []byte, err error) {
	if len(m.outputTaskSpec.taskYml) == 0 {
		glog.Warningf("did not run output: empty output task was provided")
		return
	}

	te, err := newTaskExecutor(m.outputTaskSpec.identity, m.outputTaskSpec.metaTaskYml, m.outputTaskSpec.taskYml, values)
	if err != nil {
		return
	}

	output, err = te.Output()
	return
}

// Run will run all the defined tasks & will rollback in case of any error
//
// NOTE: values is mutated (i.e. gets modified after each task execution) to
// let the task execution result be made available to the next task before execution
// of this next task
func (m *TaskGroupRunner) Run(values map[string]interface{}, postTaskRunFn PostTaskRunFn) (output []byte, err error) {
	err = m.runAllTasks(values, postTaskRunFn)
	if err != nil {
		glog.Errorf("failed to run tasks: will rollback previously run task(s): error '%s'", err.Error())
		m.rollback()
	} else {
		// return this runner's output if there were no errors
		return m.runOutput(values)
	}

	return
}
