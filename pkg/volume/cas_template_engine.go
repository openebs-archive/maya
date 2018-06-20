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

package volume

import (
	"fmt"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/task"
	"github.com/openebs/maya/pkg/util"
)

func unMarshallToConfig(config string) (configs []v1alpha1.Config, err error) {
	err = yaml.Unmarshal([]byte(config), &configs)
	return
}

// mergeConfig will merge the unique configuration elements of lowPriorityConfig
// into highPriorityConfig and return the result
func mergeConfig(highPriorityConfig, lowPriorityConfig []v1alpha1.Config) (final []v1alpha1.Config) {
	var prioritized []string

	for _, pc := range highPriorityConfig {
		final = append(final, pc)
		prioritized = append(prioritized, strings.TrimSpace(pc.Name))
	}

	for _, lc := range lowPriorityConfig {
		if !util.ContainsString(prioritized, strings.TrimSpace(lc.Name)) {
			final = append(final, lc)
		}
	}

	return
}

// CASCreator exposes method to create a cas volume
type CASCreator interface {
	create() (output []byte, err error)
}

// CASDeleter exposes method to delete a cas volume
type CASDeleter interface {
	delete() (output []byte, err error)
}

// CASLister exposes method to list one or more cas volumes
type CASLister interface {
	list() (output []byte, err error)
}

// CASReader exposes method to fetch details of a cas volume
type CASReader interface {
	read() (output []byte, err error)
}

// casCommon consists of common properties required by various
// cas engines
type casCommon struct {
	// casTemplate refers to a CASTemplate
	casTemplate *v1alpha1.CASTemplate
	// templateValues is the data (consisting of final config, runtime volume info,
	// and run task's results) in hierarchical format. This is used while running
	// `go template` against the run tasks (which are yaml document templates)
	templateValues map[string]interface{}
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
	runtimeVals map[string]string,
	fetcher task.TaskSpecFetcher,
	grpRunner *task.TaskGroupRunner) casCommon {

	return casCommon{
		casTemplate: casTemplate,
		templateValues: map[string]interface{}{
			// runtime volume info is set against volume top level property
			string(v1alpha1.VolumeTLP): runtimeVals,
			// list items is set as a top level property
			string(v1alpha1.ListItemsTLP): map[string]interface{}{},
		},
		taskSpecFetcher: fetcher,
		taskGroupRunner: grpRunner,
	}
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

// initTaskResultAsTLP adds task result as a top level property i.e. becomes
// part of values to be used for templating
//
// NOTE:
//  This will enable templating a run task template as follows:
//
// {{ .TaskResult }}
// {{ .TaskResult.<nestedProperty1> }}
func (c *casCommon) initTaskResultAsTLP() {
	c.templateValues[string(v1alpha1.TaskResultTLP)] = nil
}

// run executes the cas engine based on the tasks set in the cas template
func (c *casCommon) run() (output []byte, err error) {
	c.initTaskResultAsTLP()

	err = c.prepareTasksForExec()
	if err != nil {
		return
	}

	err = c.prepareOutputTask()
	if err != nil {
		return
	}

	//return c.taskGroupRunner.Run(c.templateValues, c.addTaskResultsToTaskResultTLP)
	return c.taskGroupRunner.Run(c.templateValues)
}

// casCreate is capable of creating a CAS volume
//
// It implements following interfaces:
// - CASCreator
type casCreate struct {
	// casCommon has the common properties required by this engine
	casCommon
	// defaultConfig is the default cas volume configurations found
	// in the CASTemplate
	defaultConfig []v1alpha1.Config
	// casConfigSC is the cas volume config found in the StorageClass
	casConfigSC []v1alpha1.Config
	// casConfigPVC is the cas volume config found in the PersistentVolumeClaim
	casConfigPVC []v1alpha1.Config
}

// NewCASCreate returns a new instance of casCreate based on
// the provided cas configs & runtime volume values
//
// NOTE:
//  runtime volume values set at **runtime** by openebs storage provisioner
// (a kubernetes dynamic storage provisioner)
func NewCASCreate(casConfigPVC string, casConfigSC string, casTemplate *v1alpha1.CASTemplate, runtimeVolumeVals map[string]string) (*casCreate, error) {
	if casTemplate == nil {
		return nil, fmt.Errorf("failed to create cas template engine: nil cas template")
	}

	if len(runtimeVolumeVals) == 0 {
		return nil, fmt.Errorf("failed to create cas template engine: nil runtime volume values")
	}

	f, err := task.NewK8sTaskSpecFetcher(casTemplate.Spec.TaskNamespace)
	if err != nil {
		return nil, err
	}

	r := task.NewTaskGroupRunner()
	if r == nil {
		return nil, fmt.Errorf("failed to create cas template engine: nil task group runner")
	}

	casConfPVC, err := unMarshallToConfig(casConfigPVC)
	if err != nil {
		return nil, err
	}

	casConfSC, err := unMarshallToConfig(casConfigSC)
	if err != nil {
		return nil, err
	}

	return &casCreate{
		casCommon:     buildCASCommon(casTemplate, runtimeVolumeVals, f, r),
		defaultConfig: casTemplate.Spec.Defaults,
		casConfigSC:   casConfSC,
		casConfigPVC:  casConfPVC,
	}, nil
}

func (c *casCreate) prepareFinalConfig() (final []v1alpha1.Config) {
	// merge unique config elements from SC with config from PVC
	mc := mergeConfig(c.casConfigPVC, c.casConfigSC)
	// merge unique config from above result with default config from CASTemplate
	final = mergeConfig(mc, c.defaultConfig)
	return
}

// addConfigToConfigTLP will add final cas volume configurations to ConfigTLP.
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
func (c *casCreate) addConfigToConfigTLP() error {
	var configName string
	allConfigsHierarchy := map[string]interface{}{}
	allConfigs := c.prepareFinalConfig()

	for _, config := range allConfigs {
		configName = strings.TrimSpace(config.Name)
		if len(configName) == 0 {
			return fmt.Errorf("failed to merge config '%#v': missing config name", config)
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

	// config top level property is added as part of values to be used for
	// templating
	c.templateValues[string(v1alpha1.ConfigTLP)] = allConfigsHierarchy
	return nil
}

// create creates a cas volume
func (c *casCreate) create() (output []byte, err error) {
	// add config
	err = c.addConfigToConfigTLP()
	if err != nil {
		return nil, err
	}

	return c.run()
}

// casEngine supports various cas volume related operations
//
// It implements these interfaces:
// - CASReader
// - CASLister
// - CASDeleter
type casEngine struct {
	// casCommon has common properties required by this engine
	casCommon
}

// NewCASEngine returns a new instance of casCreate based on
// the provided cas configs & runtime volume values
//
// NOTE:
//  runtime volume values set at **runtime** by openebs storage provisioner
// (a kubernetes dynamic storage provisioner)
func NewCASEngine(casTemplate *v1alpha1.CASTemplate, runtimeVolumeVals map[string]string) (*casEngine, error) {
	if casTemplate == nil {
		return nil, fmt.Errorf("failed to create cas template engine: nil cas template")
	}

	if len(runtimeVolumeVals) == 0 {
		return nil, fmt.Errorf("failed to create cas template engine: nil runtime volume values")
	}

	fr, err := task.NewK8sTaskSpecFetcher(casTemplate.Spec.TaskNamespace)
	if err != nil {
		return nil, err
	}

	gr := task.NewTaskGroupRunner()
	if gr == nil {
		return nil, fmt.Errorf("failed to create cas template engine: nil task group runner")
	}

	return &casEngine{
		casCommon: buildCASCommon(casTemplate, runtimeVolumeVals, fr, gr),
	}, nil
}

// read gets the details of a cas volume
func (c *casEngine) read() (output []byte, err error) {
	return c.run()
}

// delete deletes a cas volume
func (c *casEngine) delete() (output []byte, err error) {
	return c.run()
}

// list gets the details of one or more cas volumes
func (c *casEngine) list() (output []byte, err error) {
	return c.run()
}
