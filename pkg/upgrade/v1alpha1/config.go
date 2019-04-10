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
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/upgrade/v1alpha1"
	stringer "github.com/openebs/maya/pkg/apis/stringer/v1alpha1"
)

// Config is wrapper over apis.UpgradeConfig which is
// upgrade config for a particular job.
type Config struct {
	Object *apis.UpgradeConfig
}

// String implements Stringer interface
func (c Config) String() string {
	return stringer.Yaml("upgrade config", c)
}

// GoString implements GoStringer interface
func (c Config) GoString() string {
	return c.String()
}

// ConfigBuilder helps to build UpgradeConfig instance
type ConfigBuilder struct {
	Config *Config
	checks map[*Predicate]string
	Errors []error
}

// String implements Stringer interface
func (cb ConfigBuilder) String() string {
	return stringer.Yaml("upgrade config builder", cb)
}

// GoString implements GoStringer interface
func (cb ConfigBuilder) GoString() string {
	return cb.String()
}

// Predicate abstracts conditional logic w.r.t the upgradeConfig instance
//
// NOTE:
// Predicate is a functional approach versus traditional approach to mix
// conditions such as *if-else* within blocks of business logic
//
// NOTE:
// Predicate approach enables clear separation of conditionals from
// imperatives i.e. actions that form the business logic
type Predicate func(*Config) bool

// AddCheckf adds the predicate as a condition to be validated against the
// upgrade config instance and format the message string according to format specifier.
// If only predicate and message string is provided, it will treat it as the
// value for the corresponding predicate.
func (cb *ConfigBuilder) AddCheckf(p Predicate, predicateMsg string, args ...interface{}) *ConfigBuilder {
	cb.checks[&p] = fmt.Sprintf(predicateMsg, args...)
	return cb
}

// AddCheck adds the predicate as a condition to be validated against the
// upgrade config instance
func (cb *ConfigBuilder) AddCheck(p Predicate) *ConfigBuilder {
	return cb.AddCheckf(p, "")
}

// AddChecks adds the provided predicates as conditions to be validated against
// the upgrade config instance
func (cb *ConfigBuilder) AddChecks(predicates ...Predicate) *ConfigBuilder {
	for _, check := range predicates {
		cb.AddCheck(check)
	}
	return cb
}

// Build returns the final instance of UpgradeConfig
func (cb *ConfigBuilder) Build() (*apis.UpgradeConfig, error) {
	err := cb.validate()
	if err != nil {
		return nil, err
	}
	return cb.Config.Object, nil
}

// validate will run checks against UpgradeConfig instance
func (cb *ConfigBuilder) validate() error {
	if len(cb.Errors) != 0 {
		return errors.Errorf("builder error: %v %+v", cb.Errors, cb)
	}
	validationError := []error{}
	for cond := range cb.checks {
		pass := (*cond)(cb.Config)
		if !pass {
			validationError = append(validationError,
				errors.Errorf("failed to validate: %s ", cb.checks[cond]))
		}
	}
	if len(validationError) == 0 {
		return nil
	}
	cb.Errors = append(cb.Errors, validationError...)
	return errors.Errorf("validation error: %s %+v", validationError, cb.Config)
}

// NewConfigBuilder returns a new instance of ConfigBuilder
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		Config: &Config{
			Object: &apis.UpgradeConfig{},
		},
		checks: make(map[*Predicate]string),
	}
}

// ConfigBuilderForYaml add object in ConfigBuilder struct.
// with the help of yaml string
func ConfigBuilderForYaml(yamlString string) *ConfigBuilder {
	cb := &ConfigBuilder{
		Config: &Config{
			Object: &apis.UpgradeConfig{},
		},
		checks: make(map[*Predicate]string),
	}
	err := yaml.Unmarshal([]byte(yamlString), cb.Config.Object)
	if err != nil {
		cb.Errors = append(cb.Errors, errors.Wrapf(err, "file content: %s", yamlString))
	}
	return cb
}

// ConfigBuilderForRaw add object in ConfigBuilder struct.
// With the help of raw bytes
func ConfigBuilderForRaw(raw []byte) *ConfigBuilder {
	cb := &ConfigBuilder{
		Config: &Config{
			Object: &apis.UpgradeConfig{},
		},
		checks: make(map[*Predicate]string),
	}
	err := yaml.Unmarshal(raw, cb.Config.Object)
	if err != nil {
		cb.Errors = append(cb.Errors, errors.Wrapf(err, "file content: %s", string(raw)))
	}
	return cb
}

// IsCASTemplateName returns predicate to check
// castemplate is present in object or not.
func IsCASTemplateName() Predicate {
	return func(uc *Config) bool {
		return uc.isCASTemplateName()
	}
}

// isCASTemplateName is a Predicate that checks
// castemplate is present in object or not.
func (c *Config) isCASTemplateName() bool {
	return len(c.Object.CASTemplate) != 0
}

// IsResource returns predicate to check
// resources are present in object or not.
func IsResource() Predicate {
	return func(c *Config) bool {
		return c.isResource()
	}
}

// isResource is a Predicate that checks
// resources are present in object or not.
func (c *Config) isResource() bool {
	return len(c.Object.Resources) != 0
}

// IsValidResource returns predicate to check present resources
// in object contains name, namespace and kind or not.
func IsValidResource() Predicate {
	return func(c *Config) bool {
		return c.isValidResource()
	}
}

// isValidResource is a Predicate that checks present resources
// in object contains name, namespace and kind or not.
func (c *Config) isValidResource() bool {
	for _, resource := range c.Object.Resources {
		if resource.Kind == "" || resource.Name == "" || resource.Namespace == "" {
			return false
		}
	}
	return true
}

// IsSameKind returns predicate to check present
// resources in object are in same kind or not.
func IsSameKind() Predicate {
	return func(c *Config) bool {
		return c.isSameKind()
	}
}

// isSameKind is a Predicate that checks present
// resources in object are in same kind or not.
func (c *Config) isSameKind() bool {
	var kind string
	for _, resource := range c.Object.Resources {
		if kind == "" {
			kind = resource.Kind
		}
		if resource.Kind != kind {
			return false
		}
	}
	return true
}
