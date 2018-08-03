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
	"strings"

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

// CASEngine supports various operations w.r.t a CAS entity
// This is a common / generic engine that supports operations related to a
// CAS entity via CASTemplate.
//
// It implements these interfaces:
// - CASCreator
// - CASReader
// - CASLister
// - CASDeleter
type CASEngine struct {
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
	// taskSpecFetcher will fetch a task specification
	taskSpecFetcher task.TaskSpecFetcher
	// taskGroupRunner will run the tasks in the sequence specified in the cas
	// template
	taskGroupRunner *task.TaskGroupRunner
}

// buildCASEngine constructs and returns a CASEngine object based on the
// given arguments
func buildCASEngine(
	casTemplate *v1alpha1.CASTemplate,
	runtimeKey string,
	runtimeValues map[string]interface{},
	fetcher task.TaskSpecFetcher,
	grpRunner *task.TaskGroupRunner) *CASEngine {

	templateValues := map[string]interface{}{
		// cas template's default config is set as a top level property
		string(v1alpha1.ConfigTLP): casTemplate.Spec.Defaults,
		// list items is set as a top level property
		string(v1alpha1.ListItemsTLP): map[string]interface{}{},
		// task result is set as a top level property
		string(v1alpha1.TaskResultTLP): nil,
	}

	rk := strings.TrimSpace(runtimeKey)
	if len(rk) != 0 {
		// runtime values is set as a top level property in template values
		//
		// NOTE:
		//  runtimeKey is as per the caller/client code of this module
		templateValues[rk] = runtimeValues
	}

	return &CASEngine{
		casTemplate:     casTemplate,
		templateValues:  templateValues,
		runtimeKey:      rk,
		taskSpecFetcher: fetcher,
		taskGroupRunner: grpRunner,
	}
}

// NewCASEngine returns a new instance of CASEngine based on
// the provided cas configs & runtime volume values
//
// NOTE:
//  runtime volume values set at **runtime** by openebs storage provisioner
// (a kubernetes dynamic storage provisioner)
func NewCASEngine(casTemplate *v1alpha1.CASTemplate, runtimeKey string, runtimeValues map[string]interface{}) (engine *CASEngine, err error) {
	if casTemplate == nil {
		err = fmt.Errorf("failed to create cas template engine: nil cas template")
		return
	}

	fr, err := task.NewK8sTaskSpecFetcher(casTemplate.Spec.TaskNamespace)
	if err != nil {
		return
	}

	gr := task.NewTaskGroupRunner()
	if gr == nil {
		err = fmt.Errorf("failed to create cas template engine: nil task group runner")
		return
	}

	engine = buildCASEngine(casTemplate, runtimeKey, runtimeValues, fr, gr)
	return
}

// AddConfigToConfigTLP will add final cas volume configurations to ConfigTLP.
//
// NOTE:
//  This will enable templating a run task template as follows:
//
// {{ .Config.<ConfigName>.enabled }}
// {{ .Config.<ConfigName>.value }}
//
// NOTE:
//  Above parsing scheme is translated by running `go template` against the run
// task template
func (c *CASEngine) AddConfigToConfigTLP(allConfigs []v1alpha1.Config) error {
	var configName string
	allConfigsHierarchy := map[string]interface{}{}

	for _, config := range allConfigs {
		configName = strings.TrimSpace(config.Name)
		if len(configName) == 0 {
			return fmt.Errorf("failed to add config as a top level property: missing config name: config '%+v'", config)
		}

		configHierarchy := map[string]interface{}{
			configName: map[string]string{
				string(v1alpha1.EnabledPTP): config.Enabled,
				string(v1alpha1.ValuePTP):   config.Value,
			},
		}

		isMerged := util.MergeMapOfObjects(allConfigsHierarchy, configHierarchy)
		if !isMerged {
			return fmt.Errorf("failed to merge config: unable to add config '%s' to config hierarchy", configName)
		}
	}

	// update merged config as the top level property
	c.SetConfig(allConfigsHierarchy)
	return nil
}

// SetConfig sets (or resets if already existing) the CAS template related
// config elements
//
// NOTE:
//  This provides an opportunity to override the default config of the CAS
// template
func (c *CASEngine) SetConfig(config map[string]interface{}) {
	c.templateValues[string(v1alpha1.ConfigTLP)] = config
}

// setRuntimeValues sets (or resets if already existing) the runtime elements
// into CAS template as template values
//
// NOTE:
//  Runtime values are different from Config elements even though both of these
// are treated as template values
func (c *CASEngine) setRuntimeValues(runtimeValues map[string]interface{}) (err error) {
	if len(c.runtimeKey) == 0 {
		err = fmt.Errorf("failed to set template runtime values: runtime key is not set")
		return
	}

	util.SetNestedField(c.templateValues, runtimeValues, c.runtimeKey)
	return
}

// prepareTasksForExec prepares the taskGroupRunner instance with the
// info needed to run the tasks
func (c *CASEngine) prepareTasksForExec() error {
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
func (c *CASEngine) prepareOutputTask() (err error) {
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
func (c *CASEngine) Run() (output []byte, err error) {
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

// create creates a CAS entity
func (c *CASEngine) Create() (output []byte, err error) {
	err = c.AddConfigToConfigTLP(c.casTemplate.Spec.Defaults)
	if err != nil {
		return nil, err
	}
	return c.Run()
}

// read gets the details of a CAS entity
func (c *CASEngine) Read() (output []byte, err error) {
	err = c.AddConfigToConfigTLP(c.casTemplate.Spec.Defaults)
	if err != nil {
		return nil, err
	}
	return c.Run()
}

// delete deletes a CAS entity
func (c *CASEngine) Delete() (output []byte, err error) {
	err = c.AddConfigToConfigTLP(c.casTemplate.Spec.Defaults)
	if err != nil {
		return nil, err
	}
	return c.Run()
}

// list gets the details of one or more CAS entities
func (c *CASEngine) List() (output []byte, err error) {
	err = c.AddConfigToConfigTLP(c.casTemplate.Spec.Defaults)
	if err != nil {
		return nil, err
	}
	return c.Run()
}
