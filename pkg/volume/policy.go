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

	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/task"
	"github.com/openebs/maya/pkg/util"
)

type policyEngine struct {
	// policy is the openebs volume policy specs
	policy *v1alpha1.VolumePolicy
	// values is the data in hierarchical format
	// that is fed to each policy task
	//
	// NOTE:
	// A Task is a templated yaml which needs specific
	// value to be set against corresponding placeholder
	values map[string]interface{}
	// taskSpecFetcher will fetch a task specification
	taskSpecFetcher task.TaskSpecFetcher
	// taskRunner will run the tasks
	taskRunner *task.TaskRunner
}

// PolicyEngine returns a new instance of policyEngine based on
// the provided volume policy & volume property values
//
// NOTE:
//  volumeVals are the properties set against Volume as top level property.
// These volume values are set at runtime by the clients and provided to this
// engine
func PolicyEngine(policy *v1alpha1.VolumePolicy, volumeVals map[string]string) (*policyEngine, error) {
	if policy == nil {
		return nil, fmt.Errorf("Nil policy")
	}

	if len(volumeVals) == 0 {
		return nil, fmt.Errorf("Nil volume values")
	}

	f, err := task.NewK8sTaskSpecFetcher(policy.Spec.RunTasks.SearchNamespace)
	if err != nil {
		return nil, err
	}

	r := task.NewTaskRunner()
	if r == nil {
		return nil, fmt.Errorf("Nil task runner")
	}

	return &policyEngine{
		policy: policy,
		values: map[string]interface{}{
			string(v1alpha1.VolumeTLP): volumeVals,
		},
		taskSpecFetcher: f,
		taskRunner:      r,
	}, nil
}

// addPolicyValuesToPolicyTLP will add a policy's values to
// PolicyTLP.
//
// NOTE:
//  This will enable parsing of a particular
// policy's specific property as follows:
//
// {{ .Policy.<PolicyName>.enabled }}
// {{ .Policy.<PolicyName>.value }}
// {{ .Policy.<PolicyName>.data }}
//
// NOTE:
//  Above parsing scheme is made possible due to execution
// of each policy task by translating this task via
// text/template library
func (p *policyEngine) addPolicyValuesToPolicyTLP() error {
	allPoliciesValues := map[string]interface{}{}

	for _, p := range p.policy.Spec.Policies {
		p.Name = strings.TrimSpace(p.Name)
		if len(p.Name) == 0 {
			return fmt.Errorf("Missing name in policy '%#v'", p)
		}

		pValues := map[string]interface{}{
			p.Name: map[string]string{
				string(v1alpha1.EnabledPTP): p.Enabled,
				string(v1alpha1.ValuePTP):   p.Value,
				//string(v1alpha1.DataPTP):    p.Data,
			},
		}

		isMerged := util.MergeMapOfObjects(allPoliciesValues, pValues)
		if !isMerged {
			return fmt.Errorf("Failed to add policy values: '%s'", p.Name)
		}
	}

	// this is set to policy as the top level property
	p.values[string(v1alpha1.PolicyTLP)] = allPoliciesValues

	return nil
}

// addTaskResultsToTaskResultTLP will add a task's results to TaskResultTLP
//
// NOTE:
//  This is a concrete implementation of task.PostTaskRunFn type.
// Since task package does the low level execution, it has the results of
// the execution as well the properties of resulting objects. This
// function will be used as a closure that will be passed all the
// way till task execution & will be executed lazily.
//
// NOTE:
//  This will enable parsing of a particular
// policy's specific property as follows:
//
// {{ .TaskResult.<Identity>.<key1> }}
// {{ .TaskResult.<Identity>.<key2> }}
//
// NOTE:
//  Above parsing scheme is made possible due to execution
// of each policy task by translating this task via
// text/template library
func (p *policyEngine) addTaskResultsToTaskResultTLP(taskResultsMap map[string]interface{}) {
	if taskResultsMap == nil {
		// nothing to do
		return
	}

	for tID, tResults := range taskResultsMap {
		util.SetNestedField(p.values, tResults, string(v1alpha1.TaskResultTLP), tID)
	}
}

// prepareTasksForExec prepares the taskrunner with the
// info needed to run the tasks
func (p *policyEngine) prepareTasksForExec() error {
	// prepare the tasks mentioned in this policy
	for _, t := range p.policy.Spec.RunTasks.Tasks {
		// fetch the task & metatask's specifications from the task's
		// template name
		metaTaskYml, taskYml, err := p.taskSpecFetcher.Fetch(t.TemplateName)
		if err != nil {
			return err
		}

		// prepare the task runner by adding this task's details
		err = p.taskRunner.AddTaskSpec(t.Identity, metaTaskYml, taskYml)
		if err != nil {
			return err
		}
	}

	return nil
}

// getAnnotations will return the annotations from all tasks
func (p *policyEngine) getAnnotations() (map[string]string, error) {
	// extract results of all tasks
	allTasksResults := p.values[string(v1alpha1.TaskResultTLP)]
	if allTasksResults == nil {
		return nil, nil
	}

	annotations := map[string]string{}

	if allTasksResultsMap, ok := allTasksResults.(map[string]interface{}); ok {
		// iterate through each task & capture its annotation based results
		for tID, _ := range allTasksResultsMap {
			if strings.Contains(tID, string(v1alpha1.AnnotationsTRTP)) {
				isMerged := util.MergeMapOfStrings(annotations, util.GetMapOfStrings(allTasksResultsMap, tID))
				if !isMerged {
					return nil, fmt.Errorf("Failed to add annotations: '%s'", tID)
				}
			}
		}
	}

	return annotations, nil
}

// Execute this volume policy
func (p *policyEngine) execute() (annotations map[string]string, err error) {
	// set Policy TLP
	err = p.addPolicyValuesToPolicyTLP()
	if err != nil {
		return nil, err
	}

	err = p.prepareTasksForExec()
	if err != nil {
		return nil, err
	}

	err = p.taskRunner.Run(p.values, p.addTaskResultsToTaskResultTLP)
	if err != nil {
		return nil, err
	}

	return p.getAnnotations()
}
