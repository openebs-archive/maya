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
	"k8s.io/apimachinery/pkg/types"

	stringer "github.com/openebs/maya/pkg/apis/stringer/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// OwnerReference is a wrapper over metav1.OwnerReference
type OwnerReference struct {
	Object *metav1.OwnerReference
	errors []error
}

// String implements Stringer interface
func (or OwnerReference) String() string {
	return stringer.Yaml("ownerreference", or)
}

// GoString implements GoStringer interface
func (or OwnerReference) GoString() string {
	return or.String()
}

// New returns a new instance of OwnerReference
func New() *OwnerReference {
	return &OwnerReference{
		Object: &metav1.OwnerReference{},
	}
}

// WithName sets name in OwnerReference instance
func (or *OwnerReference) WithName(name string) *OwnerReference {
	or.Object.Name = name
	return or
}

// WithKind sets kind in OwnerReference instance
func (or *OwnerReference) WithKind(kind string) *OwnerReference {
	or.Object.Kind = kind
	return or
}

// WithAPIVersion sets api version in OwnerReference instance
func (or *OwnerReference) WithAPIVersion(apiVersion string) *OwnerReference {
	or.Object.APIVersion = apiVersion
	return or
}

// WithUID sets uid in OwnerReference instance
func (or *OwnerReference) WithUID(uid types.UID) *OwnerReference {
	or.Object.UID = uid
	return or
}

// WithAPIObject sets Object property in OwnerReference instance
func (or *OwnerReference) WithAPIObject(ownerReference *metav1.OwnerReference) *OwnerReference {
	or.Object = ownerReference
	return or
}

// validate validates OwnerReference instance
func (or *OwnerReference) validate() error {
	if len(or.errors) != 0 {
		return errors.Errorf("failed to validate: build errors were found: %v", or.errors)
	}
	validationErrs := []error{}
	if or.Object.Name == "" {
		validationErrs = append(validationErrs, errors.New("missing name"))
	}
	if or.Object.Kind == "" {
		validationErrs = append(validationErrs, errors.New("missing kind"))
	}
	if or.Object.APIVersion == "" {
		validationErrs = append(validationErrs, errors.New("missing api version"))
	}
	if or.Object.UID == "" {
		validationErrs = append(validationErrs, errors.New("missing uid"))
	}
	if len(validationErrs) != 0 {
		or.errors = append(or.errors, validationErrs...)
		return errors.Errorf("failed to validate: %v", validationErrs)
	}
	return nil
}

// Build builds a new instance of matav1.OwnerReference
func (or *OwnerReference) Build() (*metav1.OwnerReference, error) {
	err := or.validate()
	if err != nil {
		return nil,
			errors.WithMessagef(err, "failed to build OwnerReference: %s", or)
	}
	return or.Object, nil
}
