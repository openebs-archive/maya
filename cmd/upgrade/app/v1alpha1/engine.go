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

// EngineList contains list of castEngine
type EngineList struct {
	engines []cast.Interface
}

// EngineListBuilder helps to build EngineList instance
type EngineListBuilder struct {
	object *EngineList
	errors []error
}

// ListEngineBuilderForResources returns an instance of EngineListBuilder
//It adds object in EngineListBuilder struct with the help of config
func ListEngineBuilderForResources(cfg *apis.UpgradeConfig) (b *EngineListBuilder) {
	b = &EngineListBuilder{}
	castObj, err := cast.KubeClient().
		Get(cfg.CASTemplate, metav1.GetOptions{})
	if err != nil {
		b.errors = append(b.errors, errors.Errorf("error getting castemplate: %v", err))
		return
	}

	engines := []cast.Interface{}
	for _, resource := range cfg.Resources {
		e, err := upgrade.NewEngine().
			WithCASTemplate(castObj).
			WithUnitOfUpgrade(&resource).
			WithRuntimeConfig(cfg.Data).
			Build()
		if err != nil {
			b.errors = append(b.errors, errors.Errorf(
				"error creating castengine for resource: %#v  error: %v", resource, err))
			continue
		}

		engines = append(engines, e)
	}
	b.object = &EngineList{engines: engines}
	return b
}

// validate will run checks errors in  EngineListBuilder instance
func (elb *EngineListBuilder) validate() error {
	if len(elb.errors) != 0 {
		return errors.Errorf("builder error: %v", elb.errors)
	}
	return nil
}

// Build builds a new instance of EngineList with the help of
// EngineListBuilder instance
func (elb *EngineListBuilder) Build() (*EngineList, error) {
	err := elb.validate()
	if err != nil {
		return nil, err
	}
	return elb.object, nil
}

// Run runs list of castEngines. It returns error if there is any
// while running these engines
func (el *EngineList) Run() error {
	for _, e := range el.engines {
		_, err := e.Run()
		if err != nil {
			return err
		}
	}
	return nil
}
