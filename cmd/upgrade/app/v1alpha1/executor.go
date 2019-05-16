/*
Copyright 2019 The OpenEBS Authors.

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
	"os"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/upgrade/v1alpha1"
	cast "github.com/openebs/maya/pkg/castemplate/v1alpha1"
	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	pod "github.com/openebs/maya/pkg/kubernetes/pod/v1alpha1"
	upgrade "github.com/openebs/maya/pkg/upgrade/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// envPodName represent name of pod in which upgrade process is running.
	// This is required to get owner reference for upgrade result cr.
	envPodName = "OPENEBS_IO_POD_NAME"
	// envPodNamespace represent namespace of pod in which upgrade process is running.
	// This is required to get owner reference for upgrade result cr.
	envPodNamespace = "OPENEBS_IO_POD_NAMESPACE"
)

// Executor contains list of castEngine
type Executor struct {
	engines []cast.Interface
}

// ExecutorBuilder helps to build Executor instance
type ExecutorBuilder struct {
	*errors.ErrorList
	object *Executor
}

// ExecutorBuilderForConfig returns an instance of ExecutorBuilder
//It adds object in ExecutorBuilder struct with the help of config
func ExecutorBuilderForConfig(cfg *apis.UpgradeConfig) *ExecutorBuilder {
	executorBuilder := &ExecutorBuilder{
		ErrorList: &errors.ErrorList{},
	}

	podName := os.Getenv(envPodName)
	if podName == "" {
		executorBuilder.Errors = append(executorBuilder.Errors,
			errors.Errorf("failed to instantiate executor builder: ENV {%s} not present", envPodName))
		return executorBuilder
	}
	podNamespace := os.Getenv(envPodNamespace)
	if podNamespace == "" {
		executorBuilder.Errors = append(executorBuilder.Errors,
			errors.Errorf("failed to instantiate executor builder: ENV {%s} not present", envPodNamespace))
		return executorBuilder

	}

	p, err := pod.NewKubeClient().WithNamespace(podNamespace).
		Get(podName, metav1.GetOptions{})
	if err != nil {
		executorBuilder.Errors = append(executorBuilder.Errors,
			errors.Wrap(err, "failed to instantiate executor builder"))
		return executorBuilder
	}

	owners := p.ObjectMeta.OwnerReferences
	if len(owners) == 0 {
		executorBuilder.Errors = append(executorBuilder.Errors,
			errors.New("failed to instantiate executor builder: pod does not have a owner"))
		return executorBuilder
	}

	castObj, err := cast.KubeClient().
		Get(cfg.CASTemplate, metav1.GetOptions{})
	if err != nil {
		executorBuilder.Errors = append(executorBuilder.Errors,
			errors.Wrapf(err, "failed to instantiate executor builder: %s", cfg))
		return executorBuilder
	}

	// tasks represents list of runtask present in castemplate
	// These entries are present in upgrade result cr.
	tasks := []apis.UpgradeResultTask{}
	for _, taskName := range castObj.Spec.RunTasks.Tasks {
		task := apis.UpgradeResultTask{
			Name: taskName,
		}
		tasks = append(tasks, task)
	}

	engines := []cast.Interface{}
	for _, resource := range cfg.Resources {
		resource := resource // pin it
		upgradeResult, err := NewUpgradeResultGetOrCreateBuilder().
			WithSelfNamespace(podNamespace).
			WithOwner(&owners[0]).
			WithUpgradeConfig(cfg).
			WithResourceDetails(&resource).
			WithTasks(tasks).
			GetOrCreate()
		if err != nil {
			executorBuilder.Errors = append(executorBuilder.Errors,
				errors.Wrapf(err,
					"failed to instantiate executor builder: %s: %s", resource, cfg))
			return executorBuilder
		}
		e, err := upgrade.NewCASTEngineBuilder().
			WithCASTemplate(castObj).
			WithUnitOfUpgrade(&resource).
			WithRuntimeConfig(cfg.Data).
			WithUpgradeResult(upgradeResult).
			Build()
		if err != nil {
			executorBuilder.Errors = append(executorBuilder.Errors,
				errors.Wrapf(err,
					"failed to instantiate executor builder: %s: %s", resource, cfg))
			return executorBuilder
		}

		engines = append(engines, e)
	}
	executorBuilder.object = &Executor{engines: engines}
	return executorBuilder
}

// Build builds a new instance of Executor with the help of
// ExecutorBuilder instance
func (eb *ExecutorBuilder) Build() (*Executor, error) {
	if len(eb.Errors) != 0 {
		return nil, eb.ErrorList.WithStack("failed to build executor")
	}
	return eb.object, nil
}

// Execute runs list of castEngines. It returns error
// if there is any while running these engines
func (e *Executor) Execute() error {
	for _, engine := range e.engines {
		_, err := engine.Run()
		if err != nil {
			return errors.Wrap(err, "failed to run upgrade engine")
		}
	}
	return nil
}
