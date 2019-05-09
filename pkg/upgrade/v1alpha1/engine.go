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
	upgrade "github.com/openebs/maya/pkg/apis/openebs.io/upgrade/v1alpha1"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	stringer "github.com/openebs/maya/pkg/apis/stringer/v1alpha1"
	cast "github.com/openebs/maya/pkg/castemplate/v1alpha1"
	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
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
	// upgradeResultNameKey is a property of upgrade result cr's name for the resource.
	upgradeResultNameKey string = "upgradeResultName"
	// upgradeResultNamespaceKey is a property of upgrade result cr's namespace for the resource.
	upgradeResultNamespaceKey string = "upgradeResultNamespace"
)

// CASTEngineBuilder helps to build a new instance of castEngine
type CASTEngineBuilder struct {
	*errors.ErrorList
	RuntimeConfig []apis.Config
	CASTemplate   *apis.CASTemplate
	UnitOfUpgrade *upgrade.ResourceDetails
	UpgradeResult *upgrade.UpgradeResult
}

// String implements Stringer interface
func (ceb CASTEngineBuilder) String() string {
	return stringer.Yaml("cast engine builder", ceb)
}

// GoString implements GoStringer interface
func (ceb CASTEngineBuilder) GoString() string {
	return ceb.String()
}

// WithRuntimeConfig sets runtime config in CASTEngineBuilder.
func (ceb *CASTEngineBuilder) WithRuntimeConfig(configs []upgrade.DataItem) *CASTEngineBuilder {
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
	ceb.RuntimeConfig = runtimeConfig
	return ceb
}

// WithUnitOfUpgrade sets unitOfUpgrade details in CASTEngineBuilder.
func (ceb *CASTEngineBuilder) WithUnitOfUpgrade(item *upgrade.ResourceDetails) *CASTEngineBuilder {
	ceb.UnitOfUpgrade = item
	return ceb
}

// WithCASTemplate sets casTemplate object in CASTEngineBuilder.
func (ceb *CASTEngineBuilder) WithCASTemplate(casTemplate *apis.CASTemplate) *CASTEngineBuilder {
	ceb.CASTemplate = casTemplate
	return ceb
}

// WithUpgradeResult sets upgradeResult in CASTEngineBuilder.
func (ceb *CASTEngineBuilder) WithUpgradeResult(obj *upgrade.UpgradeResult) *CASTEngineBuilder {
	ceb.UpgradeResult = obj
	return ceb
}

// validate validates CASTEngineBuilder struct.
func (ceb *CASTEngineBuilder) validate() error {
	if ceb.CASTemplate == nil {
		return errors.New("missing castemplate")
	}
	if ceb.UnitOfUpgrade == nil {
		return errors.New("missing upgrade item")
	}
	if ceb.UpgradeResult == nil {
		return errors.New("missing upgrade result")
	}
	if len(ceb.Errors) > 0 {
		return ceb.ErrorList.WithStack("failed to build cast engine")
	}
	return nil
}

// Build builds a new instance of engine with the helps of CASTEngineBuilder struct.
func (ceb *CASTEngineBuilder) Build() (e cast.Interface, err error) {
	err = ceb.validate()
	if err != nil {
		err = errors.Wrapf(err, "failed to build cast engine: failed to validate: %s", ceb)
		return
	}

	// creating a new instance of CASTEngine
	e, err = cast.Engine(ceb.CASTemplate, "", nil)
	if err != nil {
		err = errors.Wrapf(err, "failed to build cast engine: %s", ceb)
		return
	}

	defaultConfig, err := cast.ConfigToMap(ceb.CASTemplate.Spec.Defaults)
	if err != nil {
		err = errors.Wrapf(err, "failed to build cast engine: %s", ceb)
		return
	}
	runtimeConfig, err := cast.ConfigToMap(ceb.RuntimeConfig)
	if err != nil {
		err = errors.Wrapf(err, "failed to build cast engine: %s", ceb)
		return
	}

	taskConfig := map[string]interface{}{
		nameKey:                   ceb.UnitOfUpgrade.Name,
		namespaceKey:              ceb.UnitOfUpgrade.Namespace,
		kindKey:                   ceb.UnitOfUpgrade.Kind,
		upgradeResultNameKey:      ceb.UpgradeResult.Name,
		upgradeResultNamespaceKey: ceb.UpgradeResult.Namespace,
	}
	e.SetValues(upgradeItemProperty, taskConfig)
	e.SetValues(configProperty, defaultConfig)
	e.SetValues(runtimeProperty, runtimeConfig)
	return
}

// NewCASTEngineBuilder returns an empty instance of CASTEngineBuilder
func NewCASTEngineBuilder() *CASTEngineBuilder {
	return &CASTEngineBuilder{
		ErrorList: &errors.ErrorList{},
	}
}
