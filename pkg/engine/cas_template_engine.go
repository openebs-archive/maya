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

	"github.com/ghodss/yaml"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/task"
	"github.com/openebs/maya/pkg/util"
)

// UnMarshallToConfig un-marshalls the given cas template config in a yaml
// string format to a typed list of cas template config
func UnMarshallToConfig(config string) (configs []v1alpha1.Config, err error) {
	err = yaml.Unmarshal([]byte(config), &configs)
	return
}

// MergeConfig will merge the unique configuration elements of lowPriority
// into highPriority and return the result
func MergeConfig(highPriority, lowPriority []v1alpha1.Config) (final []v1alpha1.Config) {
	var book []string
	for _, h := range highPriority {
		final = append(final, h)
		book = append(book, strings.TrimSpace(h.Name))
	}
	for _, l := range lowPriority {
		// include only if the config was not present earlier in high priority
		// configuration
		if !util.ContainsString(book, strings.TrimSpace(l.Name)) {
			final = append(final, l)
		}
	}
	return
}

// ConfigToMap transforms CAS template's config type to a nested map
func ConfigToMap(all []v1alpha1.Config) (m map[string]interface{}, err error) {
	var configName string
	m = map[string]interface{}{}
	for _, config := range all {
		configName = strings.TrimSpace(config.Name)
		if len(configName) == 0 {
			err = fmt.Errorf("failed to map config: missing config name: config '%+v'", config)
			return nil, err
		}
		confHierarchy := map[string]interface{}{
			configName: map[string]string{
				string(v1alpha1.EnabledPTP): config.Enabled,
				string(v1alpha1.ValuePTP):   config.Value,
			},
		}
		isMerged := util.MergeMapOfObjects(m, confHierarchy)
		if !isMerged {
			err = fmt.Errorf("failed to map config: unable to merge config '%s' to config hierarchy", configName)
			return nil, err
		}
	}
	return
}

// Interface abstracts various CRUD operations exposed by engine
type Interface interface {
	Configurer
	Runner
}

// TODO
// Check if pkg should be engine/cast vs. executor/cast vs. operator/cast
// vs. controller/cast

// Configurer abstracts configuring cast template engine
type Configurer interface {
	SetConfig(values map[string]interface{})
	SetValues(key string, values map[string]interface{})
}

// Runner abstracts execution of engine
type Runner interface {
	Run() (output []byte, err error)
}

// engine supports various operations w.r.t a CAS entity. This is a generic
// engine that supports operations related to one or more CAS entities via
// CASTemplate
type engine struct {
	cast            *v1alpha1.CASTemplate  // refers to a CASTemplate
	values          map[string]interface{} // values used during go templating
	key             string                 // a path to store template values
	taskSpecFetcher task.TaskSpecFetcher   // to fetch runtask specification
	taskGroupRunner *task.TaskGroupRunner  // runs the tasks in the sequence specified in cas template
}

// initValues provides engine specific initialization values to be used during
// template execution
func initValues() (v map[string]interface{}) {
	return map[string]interface{}{
		// configTLP is set to nil, this would be reset to cas template defaults
		// if the action methods i.e. create/read/list/delete do not set it
		string(v1alpha1.ConfigTLP): nil,
		// list items is set as a top level property
		string(v1alpha1.ListItemsTLP): map[string]interface{}{},
		// task result is set as a top level property
		string(v1alpha1.TaskResultTLP): map[string]interface{}{},
	}
}

// New returns a new instance of engine
func New(cast *v1alpha1.CASTemplate, key string, values map[string]interface{}) (e *engine, err error) {
	if cast == nil {
		err = fmt.Errorf("failed to create cas template engine: nil cas template")
		return
	}
	f, err := task.NewK8sTaskSpecFetcher(cast.Spec.TaskNamespace)
	if err != nil {
		return
	}
	r := task.NewTaskGroupRunner()
	if r == nil {
		err = fmt.Errorf("failed to create cas template engine: nil task group runner")
		return
	}

	ev := initValues()

	// add/override with provided values iif key is not empty
	key = strings.TrimSpace(key)
	if len(key) != 0 {
		ev[key] = values
	}

	e = &engine{
		cast:            cast,
		key:             key,
		values:          ev,
		taskSpecFetcher: f,
		taskGroupRunner: r,
	}
	return
}

// setDefaults sets cas template default config values
func (c *engine) setDefaults() (err error) {
	m, err := ConfigToMap(c.cast.Spec.Defaults)
	if err != nil {
		return
	}
	c.values[string(v1alpha1.ConfigTLP)] = m
	return
}

// setDefaultsIfEmptyConfig sets cas template default config values if config
// property is not set
func (c *engine) setDefaultsIfEmptyConfig() (err error) {
	if c.values[string(v1alpha1.ConfigTLP)] != nil {
		return
	}
	return c.setDefaults()
}

// setLabels sets cas template labels as template values
func (c *engine) setLabels() (err error) {
	c.values[string(v1alpha1.CASTOptionsTLP)] = c.cast.Labels
	return
}

// SetConfig sets cas template config as template values
func (c *engine) SetConfig(v map[string]interface{}) {
	c.values[string(v1alpha1.ConfigTLP)] = v
}

// SetValues sets template values to be used during cas template execution
func (c *engine) SetValues(k string, v map[string]interface{}) {
	if len(k) == 0 {
		return
	}
	c.key = k
	util.SetNestedField(c.values, v, c.key)
}

// prepareTasksForExec prepares the taskGroupRunner instance with the
// info needed to run the tasks
func (c *engine) prepareTasksForExec() error {
	// prepare the tasks mentioned in cas template
	for _, taskName := range c.cast.Spec.RunTasks.Tasks {
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
func (c *engine) prepareOutputTask() (err error) {
	opTaskName := c.cast.Spec.OutputTask
	if len(opTaskName) == 0 {
		// no output task was specified; nothing to do
		return
	}
	runtask, err := c.taskSpecFetcher.Fetch(opTaskName)
	if err != nil {
		return
	}
	return c.taskGroupRunner.SetOutputTask(runtask)
}

// prepareFallback prepares the taskGroupRunner instance with the
// fallback template which is used in case of specific errors e.g. version
// mismatch error
func (c *engine) prepareFallback() {
	f := c.cast.Spec.Fallback
	c.taskGroupRunner.SetFallback(f)
}

// Run executes the cas engine based on the tasks set in the cas template
func (c *engine) Run() (output []byte, err error) {
	c.setLabels()
	err = c.setDefaultsIfEmptyConfig()
	if err != nil {
		return
	}
	err = c.prepareTasksForExec()
	if err != nil {
		return
	}
	err = c.prepareOutputTask()
	if err != nil {
		return
	}
	c.prepareFallback()
	return c.taskGroupRunner.Run(c.values)
}
