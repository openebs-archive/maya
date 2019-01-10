/*
Copyright 2018 The OpenEBS Authors

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
	corev1 "k8s.io/api/core/v1"
)

type TolerationKey string

const (
	// MasterNodeTolerationKey determines the master node via this toleration key
	MasterNodeTolerationKey TolerationKey = "node-role.kubernetes.io/master"
)

type toleration struct {
	corev1.Toleration // kubernetes toleration
}

// asToleration transforms this toleration instance into corresponding
// kubernetes toleration type
func (t *toleration) asToleration() corev1.Toleration {
	return corev1.Toleration{
		Key:               t.Key,
		Operator:          t.Operator,
		Value:             t.Value,
		Effect:            t.Effect,
		TolerationSeconds: t.TolerationSeconds,
	}
}

// OptionFunc is a typed function that abstracts anykind of operation
// against the provided toleration instance
//
// This is the basic building block to create functional operations
// against the toleration instance
type OptionFunc func(*toleration)

// Predicate abstracts conditional logic w.r.t toleration instance
//
// NOTE:
// Predicate is a functional approach versus traditional approach to mix
// conditions such as *if-else* within blocks of business logic
//
// NOTE:
// Predicate approach enables clear separation of conditionals from
// imperatives i.e. actions that form the business logic
type Predicate func(*toleration) (message string, ok bool)

// New returns a new kubernetes toleration
func New(opts ...OptionFunc) corev1.Toleration {
	var t toleration
	for _, o := range opts {
		o(&t)
	}
	return t.asToleration()
}

// NoScheduleOnMaster returns a kubernetes toleration that disables scheduling
// of Pods on master kubernetes nodes
func NoScheduleOnMaster() corev1.Toleration {
	return Builder().
		NoSchedule().
		WithKey(string(MasterNodeTolerationKey)).
		Build()
}

// builder provides utilities required to build a kubernetes daemonset instance
type builder struct {
	toleration *toleration // toleration instance
}

// Builder returns a new instance of builder
func Builder() *builder {
	return &builder{toleration: &toleration{}}
}

// Build returns the final kubernetes toleration
func (b *builder) Build() corev1.Toleration {
	return b.toleration.asToleration()
}

// NoSchedule sets NoSchedule effect to the toleration
func (b *builder) NoSchedule() *builder {
	NoSchedule()(b.toleration)
	return b
}

// NoSchedule sets NoSchedule effect to the toleration
func NoSchedule() OptionFunc {
	return func(t *toleration) {
		t.Effect = corev1.TaintEffectNoSchedule
	}
}

// WithKey sets the key of the toleration
func (b *builder) WithKey(k string) *builder {
	WithKey(k)(b.toleration)
	return b
}

// WithKey sets the key of the toleration
func WithKey(k string) OptionFunc {
	return func(t *toleration) {
		t.Key = k
	}
}

// WithEffect sets the effect of the toleration
func (b *builder) WithEffect(e corev1.TaintEffect) *builder {
	WithEffect(e)(b.toleration)
	return b
}

// WithEffect sets the effect of the toleration
func WithEffect(e corev1.TaintEffect) OptionFunc {
	return func(t *toleration) {
		t.Effect = e
	}
}

// WithOperator sets the operator of the toleration
func (b *builder) WithOperator(o corev1.TolerationOperator) *builder {
	WithOperator(o)(b.toleration)
	return b
}

// WithOperator sets the operator of the toleration
func WithOperator(o corev1.TolerationOperator) OptionFunc {
	return func(t *toleration) {
		t.Operator = o
	}
}

// WithValue sets the value of the toleration
func (b *builder) WithValue(val string) *builder {
	WithValue(val)(b.toleration)
	return b
}

// WithValue sets the value of the toleration
func WithValue(val string) OptionFunc {
	return func(t *toleration) {
		t.Value = val
	}
}

// WithTolerationSeconds sets the toleration seconds of the toleration
func (b *builder) WithTolerationSeconds(sec *int64) *builder {
	WithTolerationSeconds(sec)(b.toleration)
	return b
}

// WithTolerationSeconds sets the toleration seconds of the toleration
func WithTolerationSeconds(sec *int64) OptionFunc {
	return func(t *toleration) {
		t.TolerationSeconds = sec
	}
}
