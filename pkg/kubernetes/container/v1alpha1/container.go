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
	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

type container struct {
	corev1.Container // kubernetes container type
}

// OptionFunc is a typed function that abstracts anykind of operation
// against the provided container instance
//
// This is the basic building block to create functional operations
// against the container instance
type OptionFunc func(*container)

// Predicate abstracts conditional logic w.r.t the container instance
//
// NOTE:
// Predicate is a functional approach versus traditional approach to mix
// conditions such as *if-else* within blocks of business logic
//
// NOTE:
// Predicate approach enables clear separation of conditionals from
// imperatives i.e. actions that form the business logic
type Predicate func(*container) (nameOrMsg string, ok bool)

// predicateFailedError returns the provided predicate as an error
func predicateFailedError(message string) error {
	return errors.Errorf("predicatefailed: %s", message)
}

var (
	validationFailedError = errors.New("container validation failed")
)

// asContainer transforms this container instance into corresponding kubernetes
// container type
func (c *container) asContainer() corev1.Container {
	return corev1.Container{
		Name:                     c.Name,
		Image:                    c.Image,
		Command:                  c.Command,
		Args:                     c.Args,
		WorkingDir:               c.WorkingDir,
		Ports:                    c.Ports,
		EnvFrom:                  c.EnvFrom,
		Env:                      c.Env,
		Resources:                c.Resources,
		VolumeMounts:             c.VolumeMounts,
		VolumeDevices:            c.VolumeDevices,
		LivenessProbe:            c.LivenessProbe,
		ReadinessProbe:           c.ReadinessProbe,
		Lifecycle:                c.Lifecycle,
		TerminationMessagePath:   c.TerminationMessagePath,
		TerminationMessagePolicy: c.TerminationMessagePolicy,
		ImagePullPolicy:          c.ImagePullPolicy,
		SecurityContext:          c.SecurityContext,
		Stdin:                    c.Stdin,
		StdinOnce:                c.StdinOnce,
		TTY:                      c.TTY,
	}
}

// New returns a new kubernetes container
func New(opts ...OptionFunc) corev1.Container {
	c := &container{}
	for _, o := range opts {
		o(c)
	}
	return c.asContainer()
}

// Builder provides utilities required to build a kubernetes container type
type Builder struct {
	con    *container  // container instance
	checks []Predicate // validations to be done while building the container instance
	errors []error     // errors found while building the container instance
}

// NewBuilder returns a new instance of builder
func NewBuilder() *Builder {
	return &Builder{
		con: &container{},
	}
}

// validate will run checks against container instance
func (b *Builder) validate() error {
	for _, c := range b.checks {
		if m, ok := c(b.con); !ok {
			b.errors = append(b.errors, predicateFailedError(m))
		}
	}
	if len(b.errors) == 0 {
		return nil
	}
	return validationFailedError
}

// Build returns the final kubernetes container
func (b *Builder) Build() (corev1.Container, error) {
	err := b.validate()
	if err != nil {
		return corev1.Container{}, err
	}
	return b.con.asContainer(), nil
}

// AddCheck adds the predicate as a condition to be validated against the
// container instance
func (b *Builder) AddCheck(p Predicate) *Builder {
	b.checks = append(b.checks, p)
	return b
}

// AddChecks adds the provided predicates as conditions to be validated against
// the container instance
func (b *Builder) AddChecks(p []Predicate) *Builder {
	for _, check := range p {
		b.AddCheck(check)
	}
	return b
}

// WithName sets the name of the container
func (b *Builder) WithName(name string) *Builder {
	WithName(name)(b.con)
	return b
}

// WithName sets the name of the container
func WithName(name string) OptionFunc {
	return func(c *container) {
		c.Name = name
	}
}

// WithImage sets the image of the container
func (b *Builder) WithImage(img string) *Builder {
	WithImage(img)(b.con)
	return b
}

// WithImage sets the image of the container
func WithImage(img string) OptionFunc {
	return func(c *container) {
		c.Image = img
	}
}

// WithCommand sets the command of the container
func (b *Builder) WithCommand(cmd []string) *Builder {
	WithCommand(cmd)(b.con)
	return b
}

// WithCommand sets the command of the container
func WithCommand(cmd []string) OptionFunc {
	return func(c *container) {
		c.Command = cmd
	}
}

// WithArguments sets the command arguments of the container
func (b *Builder) WithArguments(args []string) *Builder {
	WithArguments(args)(b.con)
	return b
}

// WithArguments sets the command arguments of the container
func WithArguments(args []string) OptionFunc {
	return func(c *container) {
		c.Args = args
	}
}

// WithArguments sets the command arguments of the container
func (b *Builder) WithVolumeMounts(args []corev1.VolumeMount) *Builder {
	WithVolumeMounts(args)(b.con)
	return b
}

// WithVolumeMounts sets the volume mounts of the container
func WithVolumeMounts(volumeMounts []corev1.VolumeMount) OptionFunc {
	return func(c *container) {
		c.VolumeMounts = volumeMounts
	}
}
