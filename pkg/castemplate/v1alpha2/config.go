/*
Copyright 2017 The OpenEBS Authors

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
	"fmt"
	"strings"

	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/util"
)

// ConfigInterface abstracts get of various fields of config
type ConfigInterface interface {
	GetName() string
	GetValue() string
	IsEnabled() string
	GetData() map[string]string
}

// EngineConfig contains highPriority and lowPriority config
type EngineConfig struct {
	highPriority []ConfigInterface
	lowPriority  []v1alpha1.Config
}

// NewEngineConfig returns new instance of EngineConfig
func NewEngineConfig() *EngineConfig {
	return &EngineConfig{}
}

// WithHighPriorityConfig sets highPriority config in EngineConfig instance
func (ec *EngineConfig) WithHighPriorityConfig(highPriority []ConfigInterface) *EngineConfig {
	ec.highPriority = highPriority
	return ec
}

// WithLowPriorityConfig sets lowPriority config in EngineConfig instance
func (ec *EngineConfig) WithLowPriorityConfig(lowPriority []v1alpha1.Config) *EngineConfig {
	ec.lowPriority = lowPriority
	return ec
}

// validate validates EngineConfig struct and returns if there is any error
func (ec *EngineConfig) validate() error {
	return nil
}

// Build builds EngineConfig with highPriority and lowPriority config
func (ec *EngineConfig) Build() (*EngineConfig, error) {
	err := ec.validate()
	if err != nil {
		return nil, err
	}
	return ec, nil
}

// merge will merge the unique configuration elements of lowPriority
// into highPriority and return the result
func (ec *EngineConfig) merge() (final []v1alpha1.Config) {
	var book []string
	for _, h := range ec.highPriority {
		c := v1alpha1.Config{
			Name:    h.GetName(),
			Enabled: h.IsEnabled(),
			Value:   h.GetValue(),
			Data:    h.GetData(),
		}
		final = append(final, c)
		book = append(book, strings.TrimSpace(h.GetName()))
	}
	for _, l := range ec.lowPriority {
		// include only if the config was not present earlier in high priority configuration
		if !util.ContainsString(book, strings.TrimSpace(l.Name)) {
			final = append(final, l)
		}
	}
	return
}

// ToMap transforms CAS template's config type to a nested map
func (ec *EngineConfig) ToMap() (m map[string]interface{}, err error) {
	final := ec.merge()
	var configName string
	m = map[string]interface{}{}
	for _, config := range final {
		configName = strings.TrimSpace(config.Name)
		if len(configName) == 0 {
			err = fmt.Errorf("failed to map config: missing config name: config '%+v'", config)
			return nil, err
		}
		confHierarchy := map[string]interface{}{
			configName: map[string]string{
				string(v1alpha1.EnabledPTP): config.Enabled,
				string(v1alpha1.ValuePTP):   config.Value,
			},
		}
		isMerged := util.MergeMapOfObjects(m, confHierarchy)
		if !isMerged {
			err = fmt.Errorf("failed to map config: unable to merge config '%s' to config hierarchy", configName)
			return nil, err
		}
	}
	return
}
