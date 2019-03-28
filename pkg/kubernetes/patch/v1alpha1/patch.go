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
	"encoding/json"

	"github.com/ghodss/yaml"
	"github.com/openebs/maya/pkg/template"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
)

// TaskPatchType is a custom type that holds the patch type
type TaskPatchType string

const (
	// JsonTPT refers to a generic json patch type that is understood
	// by Kubernetes API as well
	JsonTPT TaskPatchType = "json"
	// MergeTPT refers to a generic json merge patch type that is
	// understood by Kubernetes API as well
	MergeTPT TaskPatchType = "merge"
	// StrategicTPT refers to a patch type that is understood
	// by Kubernetes API only
	StrategicTPT TaskPatchType = "strategic"
)

var taskPatchTypes = map[TaskPatchType]types.PatchType{
	JsonTPT:      types.JSONPatchType,
	MergeTPT:     types.MergePatchType,
	StrategicTPT: types.StrategicMergePatchType,
}

// Patch will consist of patch that gets applied
// against a particular resource
type Patch struct {
	// Type determines the type of patch to be applied
	Type types.PatchType `json:"type"`
	// object determines the actual patch object
	Object []byte `json:"object"`
}

type builder struct {
	patch  *Patch
	checks []Predicate
	errors []error
}

// Predicate abstracts conditional logic w.r.t the patch instance
//
// NOTE:
// Predicate is a functional approach versus traditional approach to mix
// conditions such as *if-else* within blocks of business logic
//
// NOTE:
// Predicate approach enables clear separation of conditionals from
// imperatives i.e. actions that form the business logic
type Predicate func(*Patch) (message string, ok bool)

// predicateFailedError returns the provided predicate as an error
func predicateFailedError(message string) error {
	return errors.Errorf("predicatefailed: %s", message)
}

// Builder returns a new instance of builder
func Builder() *builder {
	return &builder{
		patch: &Patch{},
	}
}

// BuilderForObject returns a new instance of builder
// when a patch obj and patch type is given
func BuilderForObject(Type types.PatchType, obj []byte) *builder {
	return &builder{
		patch: &Patch{
			Type:   Type,
			Object: obj,
		},
	}
}

// BuilderForRuntask returns a new instance of builder
// for runtask when a runtask patch yaml is given
func BuilderForRuntask(context, yml string, values map[string]interface{}) *builder {
	type taskPatch struct {
		// Type determines the type of patch to be applied
		Type TaskPatchType `json:"type"`
		// Specs determines the actual patch object
		Specs string `json:"pspec"`
	}
	t := &taskPatch{}
	b := &builder{}
	p, err := template.AsTemplatedBytes(context, yml, values)
	if err != nil {
		b.errors = append(b.errors, err)
		return b
	}
	// unmarshall into taskPatch
	err = yaml.Unmarshal(p, t)
	if err != nil {
		b.errors = append(b.errors, err)
		return b
	}
	b.patch = &Patch{
		Type:   (taskPatchType(t.Type)),
		Object: []byte(t.Specs),
	}
	return b
}

// validate will run checks against patch instance
func (b *builder) validate() error {
	for _, c := range b.checks {
		if m, ok := c(b.patch); !ok {
			b.errors = append(b.errors, predicateFailedError(m))
		}
	}
	if len(b.errors) == 0 {
		return nil
	}
	return errors.New("patch validation failed")
}

// Build returns the final instance of patch
func (b *builder) Build() (*Patch, error) {
	err := b.validate()
	if err != nil {
		return nil, err
	}
	return b.patch, nil
}

// AddCheck adds the predicate as a condition to be validated against the
// patch instance
func (b *builder) AddCheck(p Predicate) *builder {
	b.checks = append(b.checks, p)
	return b
}

// AddChecks adds the provided predicates as conditions to be validated against
// the patch instance
func (b *builder) AddChecks(p []Predicate) *builder {
	for _, check := range p {
		b.AddCheck(check)
	}
	return b
}

// ToJSON converts the patch in yaml document format to corresponding
// json document
func (p *Patch) ToJSON() ([]byte, error) {
	m := map[string]interface{}{}
	err := yaml.Unmarshal(p.Object, &m)
	if err != nil {
		return nil, err
	}
	return json.Marshal(m)
}

// IsValidPatchType returns true if provided patch
// type is one of the valid patch types
func IsValidPatchType() Predicate {
	return func(p *Patch) (string, bool) {
		if p.Type == types.JSONPatchType || p.Type == types.MergePatchType ||
			p.Type == types.StrategicMergePatchType {
			return "Valid patch Type", true
		}
		return "Invalid patch type", false
	}
}

// taskPatchType maps the runtask patch type to
// kubernetes patch type
func taskPatchType(tpt TaskPatchType) types.PatchType {
	return taskPatchTypes[tpt]
}
