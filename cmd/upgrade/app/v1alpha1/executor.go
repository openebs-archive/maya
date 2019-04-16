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
	"github.com/pkg/errors"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/upgrade/v1alpha1"
	cast "github.com/openebs/maya/pkg/castemplate/v1alpha1"
	upgrade "github.com/openebs/maya/pkg/upgrade/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Executor contains list of castEngine
type Executor struct {
	engines []cast.Interface
}

// ExecutorBuilder helps to build Executor instance
type ExecutorBuilder struct {
	object *Executor
	errors []error
}

// ExecutorBuilderForConfig returns an instance of ExecutorBuilder
//It adds object in ExecutorBuilder struct with the help of config
func ExecutorBuilderForConfig(cfg *apis.UpgradeConfig) (b *ExecutorBuilder) {
	b = &ExecutorBuilder{}
	castObj, err := cast.KubeClient().
		Get(cfg.CASTemplate, metav1.GetOptions{})
	if err != nil {
		b.errors = append(b.errors,
			errors.WithMessagef(err,
				"failed to instantiate executor builder: %s", cfg))
		return
	}

	engines := []cast.Interface{}
	for _, resource := range cfg.Resources {
		resource := resource // pin it
		e, err := upgrade.NewCASTEngineBuilder().
			WithCASTemplate(castObj).
			WithUnitOfUpgrade(&resource).
			WithRuntimeConfig(cfg.Data).
			Build()
		if err != nil {
			b.errors = append(b.errors,
				errors.WithMessagef(err,
					"failed to instantiate executor builder: %s: %s", resource, cfg))
			return
		}

		engines = append(engines, e)
	}
	b.object = &Executor{engines: engines}
	return b
}

// Build builds a new instance of Executor with the help of
// ExecutorBuilder instance
func (eb *ExecutorBuilder) Build() (*Executor, error) {
	if len(eb.errors) != 0 {
		return nil, errors.Errorf("builder error: %s", eb.errors)
	}
	return eb.object, nil
}

// Execute runs list of castEngines. It returns error
// if there is any while running these engines
func (e *Executor) Execute() error {
	for _, engine := range e.engines {
		_, err := engine.Run()
		if err != nil {
			return errors.WithMessagef(err, "failed to run upgrade engine")
		}
	}
	return nil
}
