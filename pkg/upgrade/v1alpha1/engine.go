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
	stringer "github.com/openebs/maya/pkg/apis/stringer/v1alpha1"
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
	RuntimeConfig []apis.Config
	CASTemplate   *apis.CASTemplate
	UnitOfUpgrade *upgrade.ResourceDetails
	errors        []error
}

// String implements Stringer interface
func (eb EngineBuilder) String() string {
	return stringer.Yaml("engine builder", eb)
}

// GoString implements GoStringer interface
func (eb EngineBuilder) GoString() string {
	return eb.String()
}

// WithRuntimeConfig sets runtime config in EngineBuilder.
func (eb *EngineBuilder) WithRuntimeConfig(configs []upgrade.DataItem) *EngineBuilder {
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
	eb.RuntimeConfig = runtimeConfig
	return eb
}

// WithUnitOfUpgrade sets unitOfUpgrade details in EngineBuilder.
func (eb *EngineBuilder) WithUnitOfUpgrade(item *upgrade.ResourceDetails) *EngineBuilder {
	eb.UnitOfUpgrade = item
	return eb
}

// WithCASTemplate sets casTemplate object in EngineBuilder.
func (eb *EngineBuilder) WithCASTemplate(casTemplate *apis.CASTemplate) *EngineBuilder {
	eb.CASTemplate = casTemplate
	return eb
}

// validate validates EngineBuilder struct.
func (eb *EngineBuilder) validate() error {
	if eb.CASTemplate == nil {
		return errors.New("nil castTemplate provided")
	}
	if eb.UnitOfUpgrade == nil {
		return errors.New("nil upgrade item provided")
	}
	if len(eb.errors) > 0 {
		return errors.Errorf("%v", eb.errors)
	}
	return nil
}

// Build builds a new instance of engine with the helps of EngineBuilder struct.
func (eb *EngineBuilder) Build() (e cast.Interface, err error) {
	err = eb.validate()
	if err != nil {
		err = errors.WithMessagef(err, "validation error: %+v", eb)
		return
	}

	// creating a new instance of CASTEngine
	e, err = cast.Engine(eb.CASTemplate, "", nil)
	if err != nil {
		err = errors.WithMessagef(err, "failed to create engine: %+v", eb)
		return
	}

	defaultConfig, err := cast.ConfigToMap(cast.MergeConfig([]apis.Config{}, eb.CASTemplate.Spec.Defaults))
	if err != nil {
		err = errors.WithMessagef(err, "failed to create engine: %+v", eb)
		return
	}
	runtimeConfig, err := cast.ConfigToMap(cast.MergeConfig([]apis.Config{}, eb.RuntimeConfig))
	if err != nil {
		err = errors.WithMessagef(err, "failed to create engine: %+v", eb)
		return
	}

	taskConfig := map[string]interface{}{
		nameKey:      eb.UnitOfUpgrade.Name,
		namespaceKey: eb.UnitOfUpgrade.Namespace,
		kindKey:      eb.UnitOfUpgrade.Kind,
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
