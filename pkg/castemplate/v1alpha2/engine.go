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

package v1alpha1

import (
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/task"
	"github.com/pkg/errors"
)

// Engine supports various operations w.r.t a CAS entity. This is a generic engine
// that supports operations related to one or more CAS entities via CASTemplate
type Engine struct {
	cast            *v1alpha1.CASTemplate  // refers to a CASTemplate
	values          map[string]interface{} // values used during go templating
	taskSpecFetcher task.TaskSpecFetcher   // to fetch runtask specification
	taskGroupRunner *task.TaskGroupRunner  // runs the tasks in the sequence specified in cas template
	errors          []error
}

// NewEngine returns an empty instance of Engine
func NewEngine() *Engine {
	return &Engine{
		values: map[string]interface{}{
			// configTLP is set to nil, this would be reset to cas template defaults
			// if the action methods i.e. create/read/list/delete do not set it
			string(v1alpha1.ConfigTLP): nil,
			// list items is set as a top level property
			string(v1alpha1.ListItemsTLP): map[string]interface{}{},
			// task result is set as a top level property
			string(v1alpha1.TaskResultTLP): map[string]interface{}{},
		}}
}

// WithCASTemplate sets casTemplate object in Engine
func (e *Engine) WithCASTemplate(casTemplate *v1alpha1.CASTemplate) *Engine {
	e.cast = casTemplate
	return e
}

// WithTaskSpecFetcher sets taskSpecFetcher object in Engine
func (e *Engine) WithTaskSpecFetcher(taskSpecFetcher task.TaskSpecFetcher) *Engine {
	e.taskSpecFetcher = taskSpecFetcher
	return e
}

// WithTaskGroupRunner sets taskGroupRunner object in Engine
func (e *Engine) WithTaskGroupRunner(taskGroupRunner *task.TaskGroupRunner) *Engine {
	e.taskGroupRunner = taskGroupRunner
	return e
}

// WithCASTOptionsTLP sets CASTOptionsTLP object in Engine
func (e *Engine) WithCASTOptionsTLP(labels map[string]string) *Engine {
	e.values[string(v1alpha1.CASTOptionsTLP)] = labels
	return e
}

// WithConfigTLP sets ConfigTLP object in Engine
func (e *Engine) WithConfigTLP(config map[string]interface{}) *Engine {
	e.values[string(v1alpha1.ConfigTLP)] = config
	return e
}

// WithRuntimeTLP sets RuntimeTLP object in Engine
func (e *Engine) WithRuntimeTLP(key string, runtimeConfig map[string]interface{}) *Engine {
	e.values[key] = runtimeConfig
	return e
}

// validate validates Engine instance and returns error if there is any.
func (e *Engine) validate() error {
	if e.cast == nil {
		return errors.New("validation error : castemplate not present")
	}
	if e.taskSpecFetcher == nil {
		return errors.New("validation error : nil taskSpecFetcher present")
	}
	if e.taskGroupRunner == nil {
		return errors.New("validation error : nil taskGroupRunner present")
	}
	if e.values == nil {
		return errors.New("validation error : nil values present")
	}
	if len(e.errors) != 0 {
		return errors.Errorf("validation error : %v ", e.errors)
	}
	return nil
}

// Build builds Engine instance
func (e *Engine) Build() (*Engine, error) {
	err := e.validate()
	if err != nil {
		return nil, err
	}
	err = e.prepareTasksForExec()
	if err != nil {
		return nil, err
	}
	err = e.prepareOutputTask()
	if err != nil {
		return nil, err
	}
	e.prepareFallback()
	return e, nil
}

// prepareTasksForExec prepares the taskGroupRunner instance with the
// info needed to run the tasks
func (e *Engine) prepareTasksForExec() error {
	// prepare the tasks mentioned in cas template
	for _, taskName := range e.cast.Spec.RunTasks.Tasks {
		// fetch runtask from task name
		runtask, err := e.taskSpecFetcher.Fetch(taskName)
		if err != nil {
			return err
		}
		// prepare the task group runner by adding the runtask
		err = e.taskGroupRunner.AddRunTask(runtask)
		if err != nil {
			return err
		}
	}
	return nil
}

// prepareOutputTask prepares the taskGroupRunner instance with the
// yaml template which becomes the output of the runner. This is
// invoked after successful execution of all the runtasks.
func (e *Engine) prepareOutputTask() (err error) {
	opTaskName := e.cast.Spec.OutputTask
	if len(opTaskName) == 0 {
		// no output task was specified; nothing to do
		return
	}
	runtask, err := e.taskSpecFetcher.Fetch(opTaskName)
	if err != nil {
		return
	}
	return e.taskGroupRunner.SetOutputTask(runtask)
}

// prepareFallback prepares the taskGroupRunner instance with the fallback
// template which is used in case of specific errors - version mismatch error
func (e *Engine) prepareFallback() {
	f := e.cast.Spec.Fallback
	e.taskGroupRunner.SetFallback(f)
}

// Run executes the cas engine based on the tasks set in the cas template
func (e *Engine) Run() (output []byte, err error) {
	return e.taskGroupRunner.Run(e.values)
}
