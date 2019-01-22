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
)

type runitem struct {
	apis.RunItem
}

// asRunItem transforms this runitem instance into corresponding
// RunItem type
func (t *runitem) asRunItem() apis.RunItem {
	return apis.RunItem{
		ID:          t.ID,
		Name:        t.Name,
		Action:      t.Action,
		APIVersion:  t.APIVersion,
		Kind:        t.Kind,
		Options:     t.Options,
		Conditions:  t.Conditions,
		ConditionOp: t.ConditionOp,
	}
}

// OptionFunc is a typed function that abstracts any kind of operation
// against the provided runitem instance
//
// This is the basic building block to create functional operations
// against the runitem instance
type OptionFunc func(*runitem)

// Predicate abstracts conditional logic w.r.t runitem instance
//
// NOTE:
// Predicate is a functional approach versus traditional approach to mix
// conditions such as *if-else* within blocks of business logic
//
// NOTE:
// Predicate approach enables clear separation of conditionals from
// imperatives i.e. actions that form the business logic
type Predicate func(*runitem) (message string, ok bool)

// New returns a new RunItem
func New(opts ...OptionFunc) apis.RunItem {
	var t runitem
	for _, o := range opts {
		o(&t)
	}
	return t.asRunItem()
}

// builder provides utilities required to build a runitem instance
type builder struct {
	runitem *runitem
}

// Builder returns a new instance of builder
func Builder() *builder {
	return &builder{runitem: &runitem{}}
}

// Build returns the final kubernetes toleration
func (b *builder) Build() apis.RunItem {
	return b.runitem.asRunItem()
}

// WithID sets the runitem identity
func WithID(id string) OptionFunc {
	return func(r *runitem) {
		r.ID = id
	}
}

// WithID sets the runitem identity
func (b *builder) WithID(id string) *builder {
	WithID(id)(b.runitem)
	return b
}

// WithName sets the name of runitem
func WithName(name string) OptionFunc {
	return func(r *runitem) {
		r.Name = name
	}
}

// WithName sets the name of runitem
func (b *builder) WithName(name string) *builder {
	WithName(name)(b.runitem)
	return b
}

// WithAction sets the runitem action
func WithAction(a apis.Action) OptionFunc {
	return func(r *runitem) {
		r.Action = a
	}
}

// WithAction sets the runitem action
func (b *builder) WithAction(a apis.Action) *builder {
	WithAction(a)(b.runitem)
	return b
}

// WithAPIVersion sets the runitem api version
func WithAPIVersion(version string) OptionFunc {
	return func(r *runitem) {
		r.APIVersion = version
	}
}

// WithAPIVersion sets the runitem api version
func (b *builder) WithAPIVersion(version string) *builder {
	WithAPIVersion(version)(b.runitem)
	return b
}

// WithKind sets the runitem kind
func WithKind(k apis.Kind) OptionFunc {
	return func(r *runitem) {
		r.Kind = k
	}
}

// WithKind sets the runitem kind
func (b *builder) WithKind(k apis.Kind) *builder {
	WithKind(k)(b.runitem)
	return b
}

// WithOptions sets the runitem options
func WithOptions(o []apis.Option) OptionFunc {
	return func(r *runitem) {
		r.Options = o
	}
}

// WithOptions sets the runitem options
func (b *builder) WithOptions(o []apis.Option) *builder {
	WithOptions(o)(b.runitem)
	return b
}

// AddOption adds a runitem option
func AddOption(o apis.Option) OptionFunc {
	return func(r *runitem) {
		r.Options = append(r.Options, o)
	}
}

// AddOption adds a runitem option
func (b *builder) AddOption(o apis.Option) *builder {
	AddOption(o)(b.runitem)
	return b
}

// WithConditions sets the runitem conditions
func WithConditions(c []apis.Condition) OptionFunc {
	return func(r *runitem) {
		r.Conditions = c
	}
}

// WithConditions sets the runitem conditions
func (b *builder) WithConditions(c []apis.Condition) *builder {
	WithConditions(c)(b.runitem)
	return b
}

// AddCondition adds a runitem condition
func AddCondition(c apis.Condition) OptionFunc {
	return func(r *runitem) {
		r.Conditions = append(r.Conditions, c)
	}
}

// AddCondition adds a runitem condition
func (b *builder) AddCondition(c apis.Condition) *builder {
	AddCondition(c)(b.runitem)
	return b
}

// WithConditionOperator sets the runitem condition operator
func WithConditionOperator(o apis.ConditionOperator) OptionFunc {
	return func(r *runitem) {
		r.ConditionOp = o
	}
}

// WithConditionOperator sets the runitem condition operator
func (b *builder) WithConditionOperator(o apis.ConditionOperator) *builder {
	WithConditionOperator(o)(b.runitem)
	return b
}
