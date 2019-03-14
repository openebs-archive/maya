/*
Copyright 2019 The OpenEBS Authors

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
	"github.com/ghodss/yaml"
	"github.com/openebs/maya/pkg/template"
	"github.com/pkg/errors"
	api_core_v1 "k8s.io/api/core/v1"
)

type podexec struct {
	object       *api_core_v1.PodExecOptions
	ignoreErrors bool
	errs         []error
}

// AsAPIPodExec validate and returns PodExecOptions object pointer and error
// depending on ignoreErrors opt and errors
func (p *podexec) AsAPIPodExec() (podExecOptions *api_core_v1.PodExecOptions, err error) {
	err = p.Validate()
	if err != nil && !p.ignoreErrors {
		return nil, err
	}
	return p.object, nil
}

// WithTemplate takes Yaml values which is given in runtask and key in which configuration
// is present and unmarshal it with PodExecOptions.
func WithTemplate(context, yamlString string, values map[string]interface{}) (p *podexec, err error) {
	p = &podexec{}
	b, err := template.AsTemplatedBytes(context, yamlString, values)
	if err != nil {
		return nil, err
	}
	exec := &api_core_v1.PodExecOptions{}
	err = yaml.Unmarshal(b, exec)
	if err != nil {
		return nil, err
	}
	p.object = exec
	return p, nil
}

// Validate validates PodExecOptions it mainly checks for container name is present or not and
// commands are present or not.
func (p *podexec) Validate() error {
	if len(p.errs) != 0 {
		return errors.Errorf("validation failed: %v", p.errs)
	}
	if len(p.object.Command) == 0 {
		return errors.New("validation failed: command not provided")
	}
	if p.object.Container == "" {
		return errors.New("validation failed: container name not provided")
	}
	return nil
}

type buildOption func(*podexec)

// IgnoreErrors is a buildOption that is used ignore errors
func IgnoreErrors() buildOption {
	return func(p *podexec) {
		p.ignoreErrors = true
	}
}

// Apply applies all build options in podexec
func (p *podexec) Apply(opts ...buildOption) (*podexec, error) {
	for _, o := range opts {
		o(p)
	}
	return p, nil
}
