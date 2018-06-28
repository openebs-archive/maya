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

package engine

import (
	"fmt"

	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/task"
	"github.com/openebs/maya/pkg/util"
)

// CASCreator exposes method to create a cas entity
type CASCreator interface {
	Create() (output []byte, err error)
}

// CASDeleter exposes method to delete a cas entity
type CASDeleter interface {
	Delete() (output []byte, err error)
}

// CASLister exposes method to list one or more cas entities
type CASLister interface {
	List() (output []byte, err error)
}

// CASReader exposes method to fetch details of a cas entity
type CASReader interface {
	Read() (output []byte, err error)
}

// casCommon consists of common properties required by various
// CAS engines
type casCommon struct {
	// casTemplate refers to a CASTemplate
	casTemplate *v1alpha1.CASTemplate
	// templateValues is the data consisting of cas template default config, or
	// updated config (due to custom cas template engine(s), runtime values,
	// specific run task's results, all run tasks' list results in hierarchical
	// format.
	//
	// This is used while running `go template` against the run tasks (which are
	// yaml document templates)
	templateValues map[string]interface{}
	// runtimeKey represents the top level property to be set in the template
	// values
	runtimeKey string
	// runtimeValues represent the top level property value to be set against the
	// runtimeKey in the template values
	//
	// NOTE:
	//  runtimeKey along with runtimeValues represents those values that are not
	// set in the CASTemplate's config
	runtimeValues map[string]interface{}
	// taskSpecFetcher will fetch a task specification
	taskSpecFetcher task.TaskSpecFetcher
	// taskGroupRunner will run the tasks in the sequence specified in the cas
	// template
	taskGroupRunner *task.TaskGroupRunner
}

// buildCASCommon constructs and returns a casCommon structure based on the
// given arguments
func buildCASCommon(
	casTemplate *v1alpha1.CASTemplate,
	runtimeKey string,
	runtimeValues map[string]interface{},
	fetcher task.TaskSpecFetcher,
	grpRunner *task.TaskGroupRunner) casCommon {

	return casCommon{
		casTemplate: casTemplate,
		templateValues: map[string]interface{}{
			// cas template's default config is set as a top level property
			string(v1alpha1.ConfigTLP): casTemplate.Spec.Defaults,
			// list items is set as a top level property
			string(v1alpha1.ListItemsTLP): map[string]interface{}{},
			// task result is set as a top level property
			string(v1alpha1.TaskResultTLP): nil,
		},
		runtimeKey:      runtimeKey,
		runtimeValues:   runtimeValues,
		taskSpecFetcher: fetcher,
		taskGroupRunner: grpRunner,
	}
}

// SetConfig sets (or resets if already existing) the CAS template related
// config elements
//
// NOTE:
//  This provides an opportunity to override the default config of the CAS
// template
func (c *casCommon) SetConfig(config map[string]interface{}) {
	c.templateValues[string(v1alpha1.ConfigTLP)] = config
}

// setRuntimeValues sets (or resets if already existing) the runtime elements
// into CAS template as template values
//
// NOTE:
//  Runtime values are different from Config elements even though both of these
// are treated as template values
func (c *casCommon) setRuntimeValues() {
	util.SetNestedField(c.templateValues, c.runtimeValues, c.runtimeKey)
}

// prepareTasksForExec prepares the taskGroupRunner instance with the
// info needed to run the tasks
func (c *casCommon) prepareTasksForExec() error {
	// prepare the tasks mentioned in cas template
	for _, taskName := range c.casTemplate.Spec.RunTasks.Tasks {
		// fetch runtask from task name
		runtask, err := c.taskSpecFetcher.Fetch(taskName)
		if err != nil {
			return err
		}

		// prepare the task group runner by adding the runtask
		err = c.taskGroupRunner.AddRunTask(runtask)
		if err != nil {
			return err
		}
	}

	return nil
}

// prepareOutputTask prepares the taskGroupRunner instance with the
// yaml template which becomes the output of the runner. This is
// invoked after successful execution of all the runtasks.
func (c *casCommon) prepareOutputTask() (err error) {
	opTaskName := c.casTemplate.Spec.OutputTask
	if len(opTaskName) == 0 {
		// no output task was specified; nothing to do
		return
	}

	runtask, err := c.taskSpecFetcher.Fetch(opTaskName)
	if err != nil {
		return
	}

	err = c.taskGroupRunner.SetOutputTask(runtask)
	return
}

// Run executes the cas engine based on the tasks set in the cas template
func (c *casCommon) Run() (output []byte, err error) {
	// set the runtime template values before the actual run
	c.setRuntimeValues()

	err = c.prepareTasksForExec()
	if err != nil {
		return
	}

	err = c.prepareOutputTask()
	if err != nil {
		return
	}

	return c.taskGroupRunner.Run(c.templateValues)
}

// CASEngine supports various operations w.r.t a CAS entity
// This is a common engine that supports operations related to a CAS entity
// via CASTemplate.
//
// It implements these interfaces:
// - CASCreator
// - CASReader
// - CASLister
// - CASDeleter
type CASEngine struct {
	// casCommon has common properties required by this engine
	casCommon
}

// NewCASEngine returns a new instance of casEngine based on
// the provided cas configs & runtime volume values
//
// NOTE:
//  runtime volume values set at **runtime** by openebs storage provisioner
// (a kubernetes dynamic storage provisioner)
func NewCASEngine(casTemplate *v1alpha1.CASTemplate, runtimeKey string, runtimeValues map[string]interface{}) (*CASEngine, error) {
	if casTemplate == nil {
		return nil, fmt.Errorf("failed to create cas template engine: nil cas template")
	}

	fr, err := task.NewK8sTaskSpecFetcher(casTemplate.Spec.TaskNamespace)
	if err != nil {
		return nil, err
	}

	gr := task.NewTaskGroupRunner()
	if gr == nil {
		return nil, fmt.Errorf("failed to create cas template engine: nil task group runner")
	}

	return &CASEngine{
		casCommon: buildCASCommon(casTemplate, runtimeKey, runtimeValues, fr, gr),
	}, nil
}

// create creates a CAS entity
func (c *CASEngine) Create() (output []byte, err error) {
	return c.Run()
}

// read gets the details of a CAS entity
func (c *CASEngine) Read() (output []byte, err error) {
	return c.Run()
}

// delete deletes a CAS entity
func (c *CASEngine) Delete() (output []byte, err error) {
	return c.Run()
}

// list gets the details of one or more CAS entities
func (c *CASEngine) List() (output []byte, err error) {
	return c.Run()
}
