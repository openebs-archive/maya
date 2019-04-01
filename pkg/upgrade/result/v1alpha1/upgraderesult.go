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
	apis "github.com/openebs/maya/pkg/apis/openebs.io/upgrade/v1alpha1"
	"github.com/openebs/maya/pkg/template"
	"github.com/pkg/errors"
)

type upgradeResult struct {
	// upgrade result object
	object *apis.UpgradeResult
}

type upgradeResultList struct {
	// list of upgrade results
	items []*upgradeResult
}

// builder enables building an instance of
// upgradeResult
type builder struct {
	*upgradeResult
	checks []Predicate
	errors []error
}

// Predicate abstracts conditional logic w.r.t the upgradeResult instance
//
// NOTE:
// Predicate is a functional approach versus traditional approach to mix
// conditions such as *if-else* within blocks of business logic
//
// NOTE:
// Predicate approach enables clear separation of conditionals from
// imperatives i.e. actions that form the business logic
type Predicate func(*upgradeResult) (message string, ok bool)

// predicateFailedError returns the provided predicate as an error
func predicateFailedError(message string) error {
	return errors.Errorf("predicatefailed: %s", message)
}

// Builder returns a new instance of builder
func Builder() *builder {
	return &builder{upgradeResult: &upgradeResult{
		object: &apis.UpgradeResult{},
	}}
}

// BuilderForRuntask returns a new instance
// of builder for runtasks
func BuilderForRuntask(context, yml string, values map[string]interface{}) *builder {
	b := &builder{}
	uresult := &apis.UpgradeResult{}

	raw, err := template.AsTemplatedBytes(context, yml, values)
	if err != nil {
		b.errors = append(b.errors, err)
		return b
	}
	err = yaml.Unmarshal(raw, uresult)
	if err != nil {
		b.errors = append(b.errors, err)
		return b
	}
	b.upgradeResult = &upgradeResult{
		object: uresult,
	}
	return b
}

// Build returns the final instance of upgradeResult
func (b *builder) Build() (*apis.UpgradeResult, error) {
	err := b.validate()
	if err != nil {
		return nil, err
	}
	return b.upgradeResult.object, nil
}

// validate will run checks against upgrade
// result instance
func (b *builder) validate() error {
	for _, c := range b.checks {
		if m, ok := c(b.upgradeResult); !ok {
			b.errors = append(b.errors, predicateFailedError(m))
		}
	}
	if len(b.errors) == 0 {
		return nil
	}
	return errors.New("upgrade result build validation failed")
}

// AddCheck adds the predicate as a condition to be validated against the
// upgradeResult instance
func (b *builder) AddCheck(p Predicate) *builder {
	b.checks = append(b.checks, p)
	return b
}

// AddChecks adds the provided predicates as conditions to be validated against
// the upgradeResult instance
func (b *builder) AddChecks(p []Predicate) *builder {
	for _, check := range p {
		b.AddCheck(check)
	}
	return b
}

// listBuilder enables building
// an instance of upgradeResultList
type listBuilder struct {
	list *upgradeResultList
}

// ListBuilder returns a new instance
// of listBuilder
func ListBuilder() *listBuilder {
	return &listBuilder{list: &upgradeResultList{}}
}

// WithAPIList builds the list of ur
// instances based on the provided
// ur api instances
func (b *listBuilder) WithAPIList(list *apis.UpgradeResultList) *listBuilder {
	if list == nil {
		return b
	}
	for i := range list.Items {
		b.list.items = append(b.list.items, &upgradeResult{object: &list.Items[i]})
	}
	return b
}

// List returns the list of ur
// instances that was built by this
// builder
func (b *listBuilder) List() *upgradeResultList {
	return b.list
}
