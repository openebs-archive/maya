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

package v1beta1

import (
	apis "github.com/openebs/maya/pkg/apis/openebs.io/runtask/v1beta1"
	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"
)

type errs []error

type runtask struct {
	task *apis.RunTask
	errs
}

func (d *runtask) HasError() bool {
	return len(d.errs) > 0
}

func (d *runtask) Errors() []error {
	return d.errs
}

// OptionFunc is a typed function that abstracts anykind of operation
// against the provided runtask instance
//
// This is the basic building block to create functional operations
// against the runtask instance
type OptionFunc func(*runtask)

// Predicate abstracts conditional logic w.r.t the runtask instance
//
// NOTE:
// Predicate is a functional approach versus traditional approach to mix
// conditions such as *if-else* within blocks of business logic
//
// NOTE:
// Predicate approach enables clear separation of conditionals from
// imperatives i.e. actions that form the business logic
type Predicate func(*runtask) (nameOrMsg string, ok bool)

// predicateFailedError returns the provided predicate as an error
func predicateFailedError(message string) error {
	return errors.Errorf("predicatefailed: %s", message)
}

var (
	validationFailedError = errors.New("runtask validation failed")
)

// New returns a new instance of runtask
func New(opts ...OptionFunc) *apis.RunTask {
	r := &runtask{task: &apis.RunTask{}}
	for _, o := range opts {
		o(r)
	}
	return r.task
}

// Update returns the updated instance of runtask
func Update(r *apis.RunTask, opts ...OptionFunc) *apis.RunTask {
	rt := &runtask{task: r}
	for _, o := range opts {
		o(rt)
	}
	return rt.task
}

// builder provides utilities required to build a runtask instance
type builder struct {
	runtask *runtask    // runtask instance
	checks  []Predicate // validations to be done against the runtask instance
	errors  []error     // errors found while building the runtask instance
}

// Builder returns a new instance of runtask builder
func Builder() *builder {
	return &builder{
		runtask: &runtask{
			task: &apis.RunTask{},
		},
	}
}

// validate will run checks against runtask instance
func (b *builder) validate() error {
	for _, c := range b.checks {
		if m, ok := c(b.runtask); !ok {
			b.errors = append(b.errors, predicateFailedError(m))
		}
	}
	if len(b.errors) == 0 {
		return nil
	}
	return validationFailedError
}

// Build returns the final instance of RunTask
func (b *builder) Build() (*apis.RunTask, error) {
	err := b.validate()
	if err != nil {
		return nil, err
	}
	return b.runtask.task, nil
}

// AddCheck adds the predicate as a condition to be validated against the
// runtask instance
func (b *builder) AddCheck(p Predicate) *builder {
	b.checks = append(b.checks, p)
	return b
}

// AddChecks adds the provided predicates as conditions to be validated against
// the runtask instance
func (b *builder) AddChecks(p []Predicate) *builder {
	for _, check := range p {
		b.AddCheck(check)
	}
	return b
}

// WithConfig sets the config to the runtask instance
func (b *builder) WithConfig(c map[string]string) *builder {
	WithConfig(c)(b.runtask)
	return b
}

// WithConfig sets the config to the runtask instance
func WithConfig(c map[string]string) OptionFunc {
	return func(r *runtask) {
		r.task.Spec.Config = c
	}
}

// AddRunItem adds a run item to the runtask instance
func AddRunItem(i apis.RunItem) OptionFunc {
	return func(r *runtask) {
		r.task.Spec.Runs = append(r.task.Spec.Runs, i)
	}
}

// AddRunItem adds a run item to the runtask instance
func (b *builder) AddRunItem(i apis.RunItem) *builder {
	AddRunItem(i)(b.runtask)
	return b
}

// AddRunItems adds a list of run items to the runtask instance
func (b *builder) AddRunItems(i []apis.RunItem) *builder {
	for _, run := range i {
		b.AddRunItem(run)
	}
	return b
}

// WithSpec sets the runtask specifications
func (b *builder) WithSpec(spec *apis.RunTask) *builder {
	WithSpec(spec)(b.runtask)
	return b
}

// WithSpec sets the runtask specifications
func WithSpec(spec *apis.RunTask) OptionFunc {
	return func(r *runtask) {
		r.task = spec
	}
}

// WithUnmarshal unmarshals the provided yaml into corresponding runtask
// instance
func WithUnmarshal(yml string) OptionFunc {
	return func(r *runtask) {
		u := apis.RunTask{}
		err := yaml.Unmarshal([]byte(yml), &u)
		if err != nil {
			r.errs = append(r.errs, err)
			return
		}
		r.task = &u
	}
}

// WithUnmarshal unmarshals the provided yaml into corresponding runtask
// instance
func (b *builder) WithUnmarshal(yaml string) *builder {
	WithUnmarshal(yaml)(b.runtask)
	return b
}

// WithStatus sets the runtask status
//
// NOTE: It is typically invoked post runtask execution
func (b *builder) WithStatus(s apis.RunTaskStatus) *builder {
	WithStatus(s)(b.runtask)
	return b
}

// WithStatus sets the runtask status
//
// NOTE: It is typically invoked post runtask execution
func WithStatus(s apis.RunTaskStatus) OptionFunc {
	return func(r *runtask) {
		r.task.Status = s
	}
}
