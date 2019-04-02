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
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/openebs/maya/pkg/template"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
)

// PatchType is a custom type that holds the patch type
type PatchType string

const (
	// PatchTypeJSON refers to a generic json patch type that is understood
	// by Kubernetes API as well
	PatchTypeJSON PatchType = "json"
	// PatchTypeMerge refers to a generic json merge patch type that is
	// understood by Kubernetes API as well
	PatchTypeMerge PatchType = "merge"
	// PatchTypeStrategic refers to a patch type that is understood
	// by Kubernetes API only
	PatchTypeStrategic PatchType = "strategic"
)

// kubePatchTypes map task patch types
// to k8s patch types
var kubePatchTypes = map[PatchType]types.PatchType{
	PatchTypeJSON:      types.JSONPatchType,
	PatchTypeMerge:     types.MergePatchType,
	PatchTypeStrategic: types.StrategicMergePatchType,
}

var predicatesInfo = map[*Predicate]string{}

// Patch will consist of patch that gets applied
// against a particular resource
type Patch struct {
	// Type determines the type of patch to be applied
	Type types.PatchType `json:"type"`
	// object determines the actual patch object
	Object []byte `json:"object"`
}

// GoString provides the essential Patch struct details
func (p *Patch) GoString() string {
	return fmt.Sprintf("Patch{Type: %s}", p.Type)
}

// String provides the essential Patch details
func (p *Patch) String() string {
	return fmt.Sprintf("patch with type '%s'", p.Type)
}

// Builder returns a new instance of builder
type Builder struct {
	patch  *Patch
	checks []*Predicate
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
type Predicate func(*Patch) bool

// predicateFailedError returns the provided predicate as an error
func predicateFailedError(message string) error {
	return errors.Errorf("predicatefailed: %s", message)
}

// NewBuilder returns a new instance of builder
func NewBuilder() *Builder {
	return &Builder{
		patch: &Patch{},
	}
}

// BuilderForObject returns a new instance of builder
// when a patch obj and patch type is given
func BuilderForObject(Type types.PatchType, obj []byte) *Builder {
	return &Builder{
		patch: &Patch{
			Type:   Type,
			Object: obj,
		},
	}
}

// BuilderForRuntask returns a new instance of builder
// for runtask when a runtask patch yaml is given
func BuilderForRuntask(context, templateYaml string,
	templateValues map[string]interface{}) *Builder {
	// runTaskPatch reflects a runtask having
	// patch action
	type runTaskPatch struct {
		// Type determines the type of patch to be applied
		Type PatchType `json:"type"`
		// Spec determines the actual patch object
		Spec string `json:"pspec"`
	}
	t := &runTaskPatch{}
	b := &Builder{}
	p, err := template.AsTemplatedBytes(context, templateYaml, templateValues)
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
		Type:   kubePatchTypes[t.Type],
		Object: []byte(t.Spec),
	}
	return b
}

// validate will run checks against patch instance
func (b *Builder) validate() error {
	for _, c := range b.checks {
		if ok := (*c)(b.patch); !ok {
			b.errors = append(b.errors,
				errors.Errorf("predicatefailed: %s", predicatesInfo[c]))
		}
	}
	if len(b.errors) == 0 {
		return nil
	}
	return errors.Errorf("patch validation failed: %v", b.errors)
}

// Build returns the final instance of patch
func (b *Builder) Build() (*Patch, error) {
	if len(b.errors) != 0 {
		return nil, errors.Errorf("%v", b.errors)
	}
	err := b.validate()
	if err != nil {
		return nil, err
	}
	return b.patch, nil
}

// AddCheck adds the predicate as a condition to be validated against the
// patch instance
func (b *Builder) AddCheck(p Predicate, predicateInfo string) *Builder {
	predicatesInfo[&p] = predicateInfo
	b.checks = append(b.checks, &p)
	return b
}

// AddChecks adds the provided predicates as conditions to be validated against
// the patch instance
func (b *Builder) AddChecks(predicates ...Predicate) *Builder {
	for _, check := range predicates {
		b.AddCheck(check, "")
	}
	return b
}

// ToJSON converts the patch to json format
func (p *Patch) ToJSON() ([]byte, error) {
	m := map[string]interface{}{}
	err := yaml.Unmarshal(p.Object, &m)
	if err != nil {
		return nil, err
	}
	return json.Marshal(m)
}

// JSON builds and returns JSON format of the patch object
func (b *Builder) JSON() ([]byte, error) {
	p, err := b.Build()
	if err != nil {
		return nil, err
	}
	return p.ToJSON()
}

// IsValidType returns a predicate for
// checking valid patch type
func IsValidType() Predicate {
	return func(p *Patch) bool {
		return p.IsValidType()
	}
}

// IsValidType returns true if provided patch
// type is one of the valid patch types
func (p *Patch) IsValidType() bool {
	return p.Type == types.JSONPatchType || p.Type == types.MergePatchType ||
		p.Type == types.StrategicMergePatchType
}
