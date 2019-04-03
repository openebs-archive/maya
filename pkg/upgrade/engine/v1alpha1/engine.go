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
	castkey "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	cast "github.com/openebs/maya/pkg/castemplate/v1alpha1"
)

// Builder helps to build a new instance of castEngine
type Builder struct {
	runtimeConfig []apis.Config
	casTemplate   *apis.CASTemplate
	unitOfUpgrade *upgrade.ResourceDetails
	errors        []error
}

// WithRuntimeConfig sets runtime config in builder.
func (b *Builder) WithRuntimeConfig(configs []upgrade.DataItem) *Builder {
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

// WithUnitOfUpgrade sets unitOfUpgrade details in builder.
func (b *Builder) WithUnitOfUpgrade(item *upgrade.ResourceDetails) *Builder {
	b.unitOfUpgrade = item
	return b
}

// WithCASTemplate sets casTemplate object in builder.
func (b *Builder) WithCASTemplate(casTemplate *apis.CASTemplate) *Builder {
	b.casTemplate = casTemplate
	return b
}

// validate validates builder struct.
func (b *Builder) validate() error {
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

// Build builds a new instance of engine with the helps of builder struct.
func (b *Builder) Build() (e cast.Interface, err error) {
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
		string(castkey.NameTCTP):      b.unitOfUpgrade.Name,
		string(castkey.NamespaceTCTP): b.unitOfUpgrade.Namespace,
		string(castkey.KindTCTP):      b.unitOfUpgrade.Kind,
	}
	// set taskconfig against UpgradeItem key
	// to get these value use {{ .UpgradeItem.<key> }} in runtask
	e.SetValues(string(castkey.UpgradeItemTLP), taskConfig)
	// set defaultconfig against Config key
	// to get these value use {{ .Config.<key>.value }} in runtask
	e.SetValues(string(castkey.ConfigTLP), defaultConfig)
	// set runtimeconfig against Runtime key
	// to get these value use {{ .Runtime.<key>.value }} in runtask
	e.SetValues(string(castkey.RuntimeTLP), runtimeConfig)
	return
}

// New returns an empty instance of builder
func New() *Builder {
	return &Builder{}
}
