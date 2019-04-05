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
	"github.com/openebs/maya/pkg/task"
	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// supported represents the supported provider(s)
type supported int

// buildFn is a typed function that abstracts building of provider specific Engine
type buildFn func(casTemplate string, runtimeConfig map[string]interface{},
	runtimeConfigKey string, taskConfig []ConfigInterface) (*Engine, error)

const (
	// Kubernetes represents kubernetes as a supported provider. This is default
	// supported provider. Place other provider between kubernetes and unsupported
	kubernetes supported = iota
	// unsupported represents unsupported provider. Value of all the supported
	// providers should be less than unsupported
	unsupported supported = iota
)

// BuilderForProvider helps to build Engine for a given provider
type BuilderForProvider struct {
	provider         supported              // specific type of provider
	runtimeConfig    map[string]interface{} // runtime configuration for engine
	runtimeConfigKey string                 // key for runtimeConfig
	casTemplate      string                 // name of castemplate
	taskConfig       []ConfigInterface      // highPriority config for Engine
	build            buildFn
}

// BuildOption defines the abstraction to build a BuilderForProvider instance
type BuildOption func(*BuilderForProvider)

// New returns a new instance of BuilderForProvider meant for
// build Engine for a given provider
func New(opts ...BuildOption) *BuilderForProvider {
	b := &BuilderForProvider{}
	for _, o := range opts {
		o(b)
	}
	b.withDefault()
	return b
}

// WithRuntimeConfig sets runtimeConfig and it's key in BuilderForProvider instance
func WithRuntimeConfig(key string, config map[string]interface{}) BuildOption {
	return func(b *BuilderForProvider) {
		b.runtimeConfig = config
		b.runtimeConfigKey = key
	}
}

// WithTaskConfig sets highPriority config for Engine
func WithTaskConfig(config []ConfigInterface) BuildOption {
	return func(b *BuilderForProvider) {
		b.taskConfig = config
	}
}

// WithCasTemplate sets castemplate name in BuilderForProvider instance
func WithCasTemplate(casTemplate string) BuildOption {
	return func(b *BuilderForProvider) {
		b.casTemplate = casTemplate
	}
}

// WithProviderKubernetes marks the given provider as a Kubernetes specific provider
func WithProviderKubernetes() BuildOption {
	return func(b *BuilderForProvider) {
		b.provider = kubernetes
	}
}

// validate validates BuilderForProvider instance and returns error if there is any.
func (b *BuilderForProvider) validate() error {
	if b.casTemplate == "" {
		return errors.New("validation error : nil castTemplate provided")
	}
	if b.provider >= unsupported {
		return errors.New("validation error : no supported provider found")
	}
	if b.build == nil {
		return errors.New("validation error : nil build function provided")
	}
	return nil
}

// withDefault add default properties in BuilderForProvider instance if those
// are not present in BuilderForProvider instance
func (b *BuilderForProvider) withDefault() {
	if b.build == nil {
		if b.provider == kubernetes {
			b.build = func(casTemplate string, runtimeConfig map[string]interface{},
				runtimeConfigKey string, taskConfig []ConfigInterface) (e *Engine, err error) {
				casObj, err := KubeClient().
					Get(casTemplate, metav1.GetOptions{})
				if err != nil {
					return
				}

				f, err := task.NewK8sTaskSpecFetcher(casObj.Spec.TaskNamespace)
				if err != nil {
					return
				}

				r := task.NewTaskGroupRunner()

				ec, err := NewEngineConfig().
					WithHighPriorityConfig(taskConfig).
					WithLowPriorityConfig(casObj.Spec.Defaults).
					Build()
				if err != nil {
					return
				}

				m, err := ec.ToMap()
				if err != nil {
					return
				}

				e, err = NewEngine().
					WithCASTemplate(casObj).
					WithTaskSpecFetcher(f).
					WithTaskGroupRunner(r).
					WithCASTOptionsTLP(casObj.Labels).
					WithConfigTLP(m).
					WithRuntimeTLP(runtimeConfigKey, runtimeConfig).
					Build()
				if err != nil {
					return
				}
				return
			}
		}
	}
}

// Build builds a new instance of Engine for given provider
func (b *BuilderForProvider) Build() (e *Engine, err error) {
	err = b.validate()
	if err != nil {
		return
	}
	return b.build(b.casTemplate, b.runtimeConfig, b.runtimeConfigKey, b.taskConfig)
}
