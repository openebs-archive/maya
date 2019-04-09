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
)

const (
	// upgradeItemProperty Property used to store task configuration
	// to get these value use {{ .UpgradeItem.<key> }} in runtask
	upgradeItemProperty string = "UpgradeItem"
	// configProperty Property used to store default configuration
	// present in castemplate
	// to get these value use {{ .Config.<key>.value }} in runtask
	configProperty string = "Config"
	// runtimeProperty Property used to store runtime configuration
	// to get these value use {{ .Runtime.<key>.value }} in runtask
	runtimeProperty string = "Runtime"
	// nameKey is a property of TaskConfig this is name of the resource.
	nameKey string = "name"
	// namespaceKey is a property of TaskConfig this is namespace of the resource.
	namespaceKey string = "namespace"
	// kindKey is a property of TaskConfig this is kind of the resource.
	kindKey string = "kind"
)

// EngineBuilder helps to build a new instance of castEngine
type EngineBuilder struct {
	runtimeConfig []apis.Config
	casTemplate   *apis.CASTemplate
	unitOfUpgrade *upgrade.ResourceDetails
	errors        []error
}

// WithRuntimeConfig sets runtime config in EngineBuilder.
func (b *EngineBuilder) WithRuntimeConfig(configs []upgrade.DataItem) *EngineBuilder {
	runtimeConfig := []apis.Config{}
	for _, config := range configs {
		c := apis.Config{
			Name:    config.Name,
			Value:   config.Value,
			Data:    config.Entries,
			Enabled: "true",
		}
		runtimeConfig = append(runtimeConfig, c)
	}
	b.runtimeConfig = runtimeConfig
	return b
}

// WithUnitOfUpgrade sets unitOfUpgrade details in EngineBuilder.
func (b *EngineBuilder) WithUnitOfUpgrade(item *upgrade.ResourceDetails) *EngineBuilder {
	b.unitOfUpgrade = item
	return b
}

// WithCASTemplate sets casTemplate object in EngineBuilder.
func (b *EngineBuilder) WithCASTemplate(casTemplate *apis.CASTemplate) *EngineBuilder {
	b.casTemplate = casTemplate
	return b
}

// validate validates EngineBuilder struct.
func (b *EngineBuilder) validate() error {
	if b.casTemplate == nil {
		errors.New("failed to create cas template engine: nil castTemplate provided")
	}
	if b.unitOfUpgrade == nil {
		errors.New("failed to create cas template engine: nil upgrade item provided")
	}
	if len(b.errors) > 0 {
		errors.Errorf("validation error : %v ", b.errors)
	}
	return nil
}

// Build builds a new instance of engine with the helps of EngineBuilder struct.
func (b *EngineBuilder) Build() (e cast.Interface, err error) {
	err = b.validate()
	if err != nil {
		return
	}

	// creating a new instance of CASTEngine
	e, err = cast.Engine(b.casTemplate, "", nil)
	if err != nil {
		return
	}

	defaultConfig, err := cast.ConfigToMap(cast.MergeConfig([]apis.Config{}, b.casTemplate.Spec.Defaults))
	if err != nil {
		return
	}
	runtimeConfig, err := cast.ConfigToMap(cast.MergeConfig([]apis.Config{}, b.runtimeConfig))
	if err != nil {
		return
	}

	taskConfig := map[string]interface{}{
		nameKey:      b.unitOfUpgrade.Name,
		namespaceKey: b.unitOfUpgrade.Namespace,
		kindKey:      b.unitOfUpgrade.Kind,
	}
	e.SetValues(upgradeItemProperty, taskConfig)
	e.SetValues(configProperty, defaultConfig)
	e.SetValues(runtimeProperty, runtimeConfig)
	return
}

// NewEngine returns an empty instance of EngineBuilder
func NewEngine() *EngineBuilder {
	return &EngineBuilder{}
}
