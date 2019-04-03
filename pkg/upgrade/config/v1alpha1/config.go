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
)

// UpgradeConfig is wrapper over apis.UpgradeConfig which is
// upgrade config for a particular job.
type UpgradeConfig struct {
	object *apis.UpgradeConfig
}

// Builder helps to build UpgradeConfig instance
type Builder struct {
	*UpgradeConfig
	checks map[*Predicate]string
	errors []error
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
type Predicate func(*UpgradeConfig) bool

// AddCheckf adds the predicate as a condition to be validated against the
// upgrade config instance and format the message string according to format specifier.
// If only predicate and message string is provided, it will treat it as the
// value for the corresponding predicate.
func (b *Builder) AddCheckf(p Predicate, predicateMsg string, args ...interface{}) *Builder {
	b.checks[&p] = fmt.Sprintf(predicateMsg, args...)
	return b
}

// AddCheck adds the predicate as a condition to be validated against the
// upgrade config instance
func (b *Builder) AddCheck(p Predicate) *Builder {
	return b.AddCheckf(p, "")
}

// AddChecks adds the provided predicates as conditions to be validated against
// the upgrade config instance
func (b *Builder) AddChecks(predicates ...Predicate) *Builder {
	for _, check := range predicates {
		b.AddCheck(check)
	}
	return b
}

// Build returns the final instance of UpgradeConfig
func (b *Builder) Build() (*apis.UpgradeConfig, error) {
	if len(b.errors) != 0 {
		return nil, errors.Errorf("%v", b.errors)
	}
	err := b.validate()
	if err != nil {
		return nil, err
	}
	return b.UpgradeConfig.object, nil
}

// validate will run checks against UpgradeConfig instance
func (b *Builder) validate() error {
	for cond := range b.checks {
		pass := (*cond)(b.UpgradeConfig)
		if !pass {
			b.errors = append(b.errors,
				errors.Errorf("%v", b.checks[cond]))
		}
	}
	if len(b.errors) == 0 {
		return nil
	}
	return errors.Errorf("%v", b.errors)
}

// NewBuilder returns a new instance of Builder
func NewBuilder() *Builder {
	return &Builder{
		UpgradeConfig: &UpgradeConfig{
			object: &apis.UpgradeConfig{},
		},
		checks: make(map[*Predicate]string),
	}
}

// WithYamlString add object in builder struct.
func (b *Builder) WithYamlString(yamlString string) *Builder {
	err := yaml.Unmarshal([]byte(yamlString), b.object)
	if err != nil {
		b.errors = append(b.errors, err)
	}
	return b
}

// WithRawBytes add object in builder struct.
func (b *Builder) WithRawBytes(raw []byte) *Builder {
	err := yaml.Unmarshal(raw, b.object)
	if err != nil {
		b.errors = append(b.errors, err)
	}
	return b
}

// IsCASTemplateNamePresent returns predicate to check
// castemplate is present in object or not.
func IsCASTemplateNamePresent() Predicate {
	return func(uc *UpgradeConfig) bool {
		return uc.isCASTemplateNamePresent()
	}
}

// isCASTemplateNamePresent is a Predicate that checks
// castemplate is present in object or not.
func (uc *UpgradeConfig) isCASTemplateNamePresent() bool {
	return len(uc.object.CASTemplate) != 0
}

// IsResourcePresent returns predicate to check
// resources are present in object or not.
func IsResourcePresent() Predicate {
	return func(uc *UpgradeConfig) bool {
		return uc.isResourcePresent()
	}
}

// isResourcePresent is a Predicate that checks
// resources are present in object or not.
func (uc *UpgradeConfig) isResourcePresent() bool {
	return len(uc.object.Resources) != 0
}

// IsValidResource returns predicate to check present resources
// in object contains name, namespace and kind or not.
func IsValidResource() Predicate {
	return func(uc *UpgradeConfig) bool {
		return uc.isValidResource()
	}
}

// isValidResource is a Predicate that checks present resources
// in object contains name, namespace and kind or not.
func (uc *UpgradeConfig) isValidResource() bool {
	for _, resource := range uc.object.Resources {
		if resource.Kind == "" || resource.Name == "" || resource.Namespace == "" {
			return false
		}
	}
	return true
}

// IsSameKind returns predicate to check present
// resources in object are in same kind or not.
func IsSameKind() Predicate {
	return func(uc *UpgradeConfig) bool {
		return uc.isSameKind()
	}
}

// isSameKind is a Predicate that checks present
// resources in object are in same kind or not.
func (uc *UpgradeConfig) isSameKind() bool {
	var kind string
	for _, resource := range uc.object.Resources {
		if kind == "" {
			kind = resource.Kind
		}
		if resource.Kind != kind {
			return false
		}
	}
	return true
}
