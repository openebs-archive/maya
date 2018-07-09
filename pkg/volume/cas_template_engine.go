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

// addTaskResultsToTaskResultTLP will add a run task's results to TaskResultTLP
//
// NOTE:
//  This is a concrete implementation of task.PostTaskRunFn type.
// Since task package does the low level execution, it has the results of
// the execution as well the properties of resulting objects. This
// function will be passed as a **closure** till task execution.

// In other words it gets executed lazily post the run task's execution.
//
// NOTE:
//  This will enable templating a run task template as follows:
//
// {{ .TaskResult.<Identity>.<prop1> }}
// {{ .TaskResult.<Identity>.<prop2> }}
//
// NOTE:
//  Above parsing scheme is translated by running `go template` against the run
// task template
func (c *casCommon) addTaskResultsToTaskResultTLP(taskResultsMap map[string]interface{}) {
	if taskResultsMap == nil {
		// nothing to do
		return
	}

	// task results are mapped against the task's identity
	for tID, tResults := range taskResultsMap {
		// task result is added or updated to template values
		// task result is set against TaskResultTLP.tID path of template values
		util.SetNestedField(c.templateValues, tResults, string(v1alpha1.TaskResultTLP), tID)
	}
}

// prepareTasksForExec prepares the taskGroupRunner instance with the
// info needed to run the tasks
func (c *casCommon) prepareTasksForExec() error {
	// prepare the tasks mentioned in cas template
	for _, t := range c.casTemplate.Spec.RunTasks.Tasks {
		// fetch task & meta task specifications from task name
		metaTaskYml, taskYml, err := c.taskSpecFetcher.Fetch(t.TaskName)
		if err != nil {
			return err
		}

		// prepare the task group runner by adding the run task's specifications
		//
		// NOTE: a run task has two specs:
		//  1/ meta task specifications &
		//  2/ task specifications
		err = c.taskGroupRunner.AddTaskSpec(t.Identity, metaTaskYml, taskYml)
		if err != nil {
			return err
		}
	}

	return nil
}

// prepareOutputTask prepares the taskGroupRunner instance with the
// info needed to output after successful execution of the tasks
func (c *casCommon) prepareOutputTask() error {
	opTask := c.casTemplate.Spec.RunTasks.Output
	if len(opTask.TaskName) == 0 {
		// no output task was specified; nothing to do
		return nil
	}

	metaOpYml, taskOpYml, err := c.taskSpecFetcher.Fetch(opTask.TaskName)
	if err != nil {
		return err
	}

	c.taskGroupRunner.SetOutputTaskSpec(metaOpYml, taskOpYml)

	return nil
}

// run executes the cas engine based on the tasks set in the cas template
func (c *casCommon) run() (output []byte, err error) {
	err = c.prepareTasksForExec()
	if err != nil {
		return nil, err
	}

	err = c.prepareOutputTask()
	if err != nil {
		return nil, err
	}

	return c.taskGroupRunner.Run(c.templateValues, c.addTaskResultsToTaskResultTLP)
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

	f, err := task.NewK8sTaskSpecFetcher(casTemplate.Spec.RunTasks.TaskNamespace)
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
		casCommon: casCommon{
			casTemplate: casTemplate,
			templateValues: map[string]interface{}{
				// runtime volume info is set against volume top level property
				string(v1alpha1.VolumeTLP): runtimeVolumeVals,
			},
			taskSpecFetcher: f,
			taskGroupRunner: r,
		},
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

// getAnnotations will fetch the annotations from all tasks
func (c *casCreate) getAnnotations() (map[string]string, error) {
	// extract results of all tasks
	allTasksResults := c.templateValues[string(v1alpha1.TaskResultTLP)]
	if allTasksResults == nil {
		return nil, nil
	}

	annotations := map[string]string{}

	if allTasksResultsMap, ok := allTasksResults.(map[string]interface{}); ok {
		// iterate through each task & capture its annotation from task results
		for tID, _ := range allTasksResultsMap {
			if strings.Contains(tID, string(v1alpha1.AnnotationsTRTP)) {
				isMerged := util.MergeMapOfStrings(annotations, util.GetMapOfStrings(allTasksResultsMap, tID))
				if !isMerged {
					return nil, fmt.Errorf("failed to add annotations from task having task result id '%s'", tID)
				}
			}
		}
	}

	return annotations, nil
}

// create creates a cas volume
func (c *casCreate) create() (output []byte, err error) {
	// add config
	err = c.addConfigToConfigTLP()
	if err != nil {
		return nil, err
	}

	return c.run()
	//if err != nil {
	//	return nil, err
	//}

	//return c.getAnnotations()
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

	fr, err := task.NewK8sTaskSpecFetcher(casTemplate.Spec.RunTasks.TaskNamespace)
	if err != nil {
		return nil, err
	}

	gr := task.NewTaskGroupRunner()
	if gr == nil {
		return nil, fmt.Errorf("failed to create cas template engine: nil task group runner")
	}

	return &casEngine{
		casCommon: casCommon{
			casTemplate: casTemplate,
			templateValues: map[string]interface{}{
				// runtime volume info is set against volume top level property
				string(v1alpha1.VolumeTLP): runtimeVolumeVals,
			},
			taskSpecFetcher: fr,
			taskGroupRunner: gr,
		},
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
