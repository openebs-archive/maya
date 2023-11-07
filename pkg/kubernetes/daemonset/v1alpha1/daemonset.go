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
	"bytes"
	"text/template"

	toleration "github.com/openebs/maya/pkg/kubernetes/toleration/v1alpha1"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

type errs []error

type daemonset struct {
	daemon *appsv1.DaemonSet // kubernetes daemonset instance
	errs
}

func (d *daemonset) HasError() bool {
	return len(d.errs) > 0
}

func (d *daemonset) Errors() []error {
	return d.errs
}

// OptionFunc is a typed function that abstracts anykind of operation
// against the provided daemonset instance
//
// This is the basic building block to create functional operations
// against the daemonset instance
type OptionFunc func(*daemonset)

// Predicate abstracts conditional logic w.r.t the daemonset instance
//
// NOTE:
// Predicate is a functional approach versus traditional approach to mix
// conditions such as *if-else* within blocks of business logic
//
// NOTE:
// Predicate approach enables clear separation of conditionals from
// imperatives i.e. actions that form the business logic
type Predicate func(*daemonset) (nameOrMsg string, ok bool)

// predicateFailedError returns the provided predicate as an error
func predicateFailedError(message string) error {
	return errors.Errorf("predicatefailed: %s", message)
}

var (
	validationFailedError = errors.New("daemonset validation failed")
)

// New returns a new instance of kubernetes daemonset
func New(opts ...OptionFunc) *appsv1.DaemonSet {
	d := &daemonset{daemon: &appsv1.DaemonSet{}}
	for _, o := range opts {
		o(d)
	}
	return d.daemon
}

// builder provides utilities required to build a kubernetes daemonset instance
type builder struct {
	daemonset *daemonset  // daemonset instance
	checks    []Predicate // validations to be done against the daemonset instance
	errors    []error     // errors found while building the daemonset instance
}

// Builder returns a new instance of builder
func Builder() *builder {
	return &builder{
		daemonset: &daemonset{
			daemon: &appsv1.DaemonSet{},
		},
	}
}

// validate will run checks against daemonset instance
func (b *builder) validate() error {
	for _, c := range b.checks {
		if m, ok := c(b.daemonset); !ok {
			b.errors = append(b.errors, predicateFailedError(m))
		}
	}
	if len(b.errors) == 0 {
		return nil
	}
	return validationFailedError
}

// Build returns the final instance of kubernetes daemonset
func (b *builder) Build() (*appsv1.DaemonSet, error) {
	err := b.validate()
	if err != nil {
		return nil, err
	}
	return b.daemonset.daemon, nil
}

// AddCheck adds the predicate as a condition to be validated against the
// daemonset instance
func (b *builder) AddCheck(p Predicate) *builder {
	b.checks = append(b.checks, p)
	return b
}

// AddChecks adds the provided predicates as conditions to be validated against
// the daemonset instance
func (b *builder) AddChecks(p []Predicate) *builder {
	for _, check := range p {
		b.AddCheck(check)
	}
	return b
}

// AddNodeSelector adds the provided pair as a nodeselector to the daemonset
func (b *builder) AddNodeSelector(k, v string) *builder {
	AddNodeSelector(k, v)(b.daemonset)
	return b
}

// AddNodeSelector adds the provided pair as a nodeselector to the daemonset
func AddNodeSelector(k, v string) OptionFunc {
	return func(d *daemonset) {
		if d.daemon.Spec.Template.Spec.NodeSelector == nil {
			d.daemon.Spec.Template.Spec.NodeSelector = map[string]string{}
		}
		d.daemon.Spec.Template.Spec.NodeSelector[k] = v
	}
}

// AddInitContainer adds the provided container as an init container to the
// daemonset
func (b *builder) AddInitContainer(c corev1.Container) *builder {
	AddInitContainer(c)(b.daemonset)
	return b
}

// AddInitContainers adds the provided containers as init containers to the
// daemonset
func (b *builder) AddInitContainers(c []corev1.Container) *builder {
	for _, container := range c {
		b.AddInitContainer(container)
	}
	return b
}

// AddInitContainer adds the provided container as an init container to the
// daemonset
func AddInitContainer(c corev1.Container) OptionFunc {
	return func(d *daemonset) {
		d.daemon.Spec.Template.Spec.InitContainers = append(d.daemon.Spec.Template.Spec.InitContainers, c)
	}
}

// AddContainer adds the provided container as a container to the daemonset
func (b *builder) AddContainer(c corev1.Container) *builder {
	AddContainer(c)(b.daemonset)
	return b
}

// AddContainers adds the provided containers to the daemonset
func (b *builder) AddContainers(c []corev1.Container) *builder {
	for _, container := range c {
		b.AddContainer(container)
	}
	return b
}

// AddContainer adds the provided container options as a container to the
// daemonset
func AddContainer(c corev1.Container) OptionFunc {
	return func(d *daemonset) {
		d.daemon.Spec.Template.Spec.Containers = append(d.daemon.Spec.Template.Spec.Containers, c)
	}
}

// AddToleration adds the provided toleration to the daemonset
func (b *builder) AddToleration(t corev1.Toleration) *builder {
	AddToleration(t)(b.daemonset)
	return b
}

// AddTolerations adds the provided tolerations to the daemonset
func (b *builder) AddTolerations(t []corev1.Toleration) *builder {
	for _, toleration := range t {
		b.AddToleration(toleration)
	}
	return b
}

// AddToleration adds the provided toleration to the daemonset
func AddToleration(t corev1.Toleration) OptionFunc {
	return func(d *daemonset) {
		d.daemon.Spec.Template.Spec.Tolerations = append(d.daemon.Spec.Template.Spec.Tolerations, t)
	}
}

// NoScheduleOnMaster disables scheduling of daemonset pods on kubernetes
// master node
func (b *builder) NoScheduleOnMaster() *builder {
	NoScheduleOnMaster()(b.daemonset)
	return b
}

// NoScheduleOnMaster disables scheduling of daemonset pods on kubernetes
// master node
func NoScheduleOnMaster() OptionFunc {
	return func(d *daemonset) {
		t := toleration.NoScheduleOnMaster()
		d.daemon.Spec.Template.Spec.Tolerations = append(d.daemon.Spec.Template.Spec.Tolerations, t)
	}
}

// WithSpec sets the daemonset specifications
func (b *builder) WithSpec(spec *appsv1.DaemonSet) *builder {
	WithSpec(spec)(b.daemonset)
	return b
}

// WithSpec sets the daemonset specifications
func WithSpec(spec *appsv1.DaemonSet) OptionFunc {
	return func(d *daemonset) {
		d.daemon = spec
	}
}

// WithTemplate executes the provided template in yaml format and unmarshalls
// it into daemonset
func (b *builder) WithTemplate(tpl string, data interface{}) *builder {
	WithTemplate(tpl, data)(b.daemonset)
	return b
}

// WithTemplate executes the provided template in yaml format and unmarshalls
// it into daemonset
func WithTemplate(tpl string, data interface{}) OptionFunc {
	return func(d *daemonset) {
		t := template.New("daemonset")
		t, err := t.Parse(tpl)
		if err != nil {
			d.errs = append(d.errs, err)
			return
		}
		// buf is an io.Writer impl as required by template
		var buf bytes.Buffer
		// execute into the buffer
		err = t.Execute(&buf, data)
		if err != nil {
			d.errs = append(d.errs, err)
			return
		}
		newd := &appsv1.DaemonSet{}
		err = yaml.Unmarshal(buf.Bytes(), newd)
		if err != nil {
			d.errs = append(d.errs, err)
			return
		}
		d.daemon = newd
	}
}
