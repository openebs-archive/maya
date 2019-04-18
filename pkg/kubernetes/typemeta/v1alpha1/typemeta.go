/*
Copyright 2019 The OpenEBS Authors.

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
	"github.com/pkg/errors"

	stringer "github.com/openebs/maya/pkg/apis/stringer/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Builder helps to build TypeMeta
type Builder struct {
	meta   *TypeMeta
	errors []error
}

// TypeMeta is a wrapper over metav1.TypeMeta
type TypeMeta struct {
	object *metav1.TypeMeta
}

// String implements Stringer interface
func (tm TypeMeta) String() string {
	return stringer.Yaml("typemeta", tm.object)
}

// GoString implements GoStringer interface
func (tm TypeMeta) GoString() string {
	return tm.String()
}

// NewBuilder returns a new instance of Builder
func NewBuilder() *Builder {
	return &Builder{
		meta: &TypeMeta{
			object: &metav1.TypeMeta{},
		},
	}
}

// NewBuilderForAPIObject returns a new instance of Builder
// for given metav1.TypeMeta
func NewBuilderForAPIObject(meta *metav1.TypeMeta) *Builder {
	b := &Builder{}
	if meta == nil {
		b.errors = append(b.errors,
			errors.New("failed to init builder: nil TypeMeta provided"))
		b.meta = &TypeMeta{object: &metav1.TypeMeta{}}
		return b
	}
	b.meta = &TypeMeta{object: meta}
	return b
}

// WithKind sets Kind property in TypeMeta instance
func (b *Builder) WithKind(kind string) *Builder {
	b.meta.object.Kind = kind
	return b
}

// WithAPIVersion sets APIVersion property in TypeMeta instance
func (b *Builder) WithAPIVersion(apiVersion string) *Builder {
	b.meta.object.APIVersion = apiVersion
	return b
}

// validate validates Builder instance
func (b *Builder) validate() error {
	if len(b.errors) != 0 {
		return errors.Errorf("failed to validate: build errors were found: %v", b.errors)
	}
	validationErrs := []error{}
	if b.meta.object.Kind == "" {
		validationErrs = append(validationErrs, errors.New("missing kind"))
	}
	if b.meta.object.APIVersion == "" {
		validationErrs = append(validationErrs, errors.New("missing API version"))
	}
	if len(validationErrs) != 0 {
		b.errors = append(b.errors, validationErrs...)
		return errors.Errorf("validation error(s) found: %v", validationErrs)
	}
	return nil
}

// Build returns a new instance of metav1.TypeMeta
func (b *Builder) Build() (*metav1.TypeMeta, error) {
	err := b.validate()
	if err != nil {
		return nil,
			errors.Wrapf(err, "failed to build TypeMeta: %s", b.meta.object)
	}
	return b.meta.object, nil
}
