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
	"github.com/pkg/errors"

	upgrade "github.com/openebs/maya/pkg/apis/openebs.io/upgrade/v1alpha1"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	cast "github.com/openebs/maya/pkg/castemplate/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/generated/clientset/internalclientset"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ExecutorEngine struct {
	engine        cast.Interface         // generic CAS template engine
	defaultConfig []apis.Config          // default cas storagepool config found in CASTemplate
	runtimeConfig []cast.ConfigInterface // runtimeConfig is the config that is provided
}

type builder struct {
	runtimeConfig []cast.ConfigInterface
	taskConfig    map[string]interface{}
	casTemplate   string
	clientset     *clientset.Clientset
	taskConfigKey string
	errors        []error
}

type buildOption func(*builder)

func WithRuntimeConfig(configs []upgrade.DataItem) buildOption {
	return func(b *builder) {
		runtimeConfig := []cast.ConfigInterface{}
		for _, config := range configs {
			c := cast.ConfigInterface(config)
			runtimeConfig = append(runtimeConfig, c)
		}
		b.runtimeConfig = runtimeConfig
	}
}

func WithTaskConfig(config map[string]interface{}) buildOption {
	return func(b *builder) {
		b.taskConfig = config
	}
}

func WithClientset(cs *clientset.Clientset) buildOption {
	return func(b *builder) {
		b.clientset = cs
	}
}

func WithCASTemplate(casTemplate string) buildOption {
	return func(b *builder) {
		b.casTemplate = casTemplate
	}
}

func WithTaskConfigKey(taskConfigKey string) buildOption {
	return func(b *builder) {
		b.taskConfigKey = taskConfigKey
	}
}

func (b *builder) validate() error {
	if b.clientset == nil {
		errors.New("failed to create cas template engine: nil clientset provided")
	}
	if len(b.taskConfig) == 0 {
		errors.New("failed to create cas template engine: nil TaskConfig provided")
	}
	if b.casTemplate == "" {
		errors.New("failed to create cas template engine: nil castTemplate provided")
	}
	if len(b.errors) > 0 {
		errors.Errorf("validation error : %v ", b.errors)
	}
	return nil
}

func (b *builder) GetExecutorEngine() (e *ExecutorEngine, err error) {
	castObj, err := cast.KubeClient(cast.WithClientset(b.clientset)).
		Get(b.casTemplate, meta_v1.GetOptions{})
	if err != nil {
		return
	}
	eEngine, err := cast.Engine(castObj, b.taskConfigKey, b.taskConfig)
	if err != nil {
		return
	}
	e = &ExecutorEngine{
		engine:        eEngine,
		defaultConfig: castObj.Spec.Defaults,
		runtimeConfig: b.runtimeConfig,
	}
	return
}

func New(opts ...buildOption) *builder {
	b := &builder{}
	for _, o := range opts {
		o(b)
	}
	return b
}

// Run runs one instance of executorEngine
func (ee *ExecutorEngine) Run() (op []byte, err error) {
	m, err := cast.ConfigToMap(cast.MergeConfigf(ee.runtimeConfig, ee.defaultConfig))
	if err != nil {
		return
	}
	ee.engine.SetConfig(m)
	return ee.engine.Run()
}
