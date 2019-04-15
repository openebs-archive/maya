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

// TypeMeta is a wrapper over metav1.TypeMeta
type TypeMeta struct {
	Object *metav1.TypeMeta
	errors []error
}

// String implements Stringer interface
func (tm TypeMeta) String() string {
	return stringer.Yaml("typemeta", tm)
}

// GoString implements GoStringer interface
func (tm TypeMeta) GoString() string {
	return tm.String()
}

// New returns a new instance of TypeMeta
func New() *TypeMeta {
	return &TypeMeta{
		Object: &metav1.TypeMeta{},
	}
}

// WithKind sets Kind property in TypeMeta instance
func (tm *TypeMeta) WithKind(kind string) *TypeMeta {
	tm.Object.Kind = kind
	return tm
}

// WithAPIVersion sets APIVersion property in TypeMeta instance
func (tm *TypeMeta) WithAPIVersion(apiVersion string) *TypeMeta {
	tm.Object.APIVersion = apiVersion
	return tm
}

// WithAPIObject sets Object property in TypeMeta instance
func (tm *TypeMeta) WithAPIObject(typeMeta *metav1.TypeMeta) *TypeMeta {
	tm.Object = typeMeta
	return tm
}

// validate validates typeMeta instance
func (tm *TypeMeta) validate() error {
	if len(tm.errors) != 0 {
		return errors.Errorf("failed to validate: build errors were found: %v", tm.errors)
	}
	validationErrs := []error{}
	if tm.Object.Kind == "" {
		validationErrs = append(validationErrs, errors.New("missing kind"))
	}
	if tm.Object.APIVersion == "" {
		validationErrs = append(validationErrs, errors.New("missing API version"))
	}
	if len(validationErrs) != 0 {
		tm.errors = append(tm.errors, validationErrs...)
		return errors.Errorf("failed to validate: %v", validationErrs)
	}
	return nil
}

// Build returns a new instance of metav1.TypeMeta
func (tm *TypeMeta) Build() (*metav1.TypeMeta, error) {
	err := tm.validate()
	if err != nil {
		return nil,
			errors.WithMessagef(err, "failed to build TypeMeta: %s", tm)
	}
	return tm.Object, nil
}
