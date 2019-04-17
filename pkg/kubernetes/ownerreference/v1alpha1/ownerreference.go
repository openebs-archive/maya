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

// Builder helps to build OwnerReference
type Builder struct {
	errors []error
	Object *OwnerReference
}

// OwnerReference is a wrapper over metav1.OwnerReference
type OwnerReference struct {
	OwnerReference *metav1.OwnerReference
}

// String implements Stringer interface
func (or OwnerReference) String() string {
	return stringer.Yaml("ownerreference", or)
}

// GoString implements GoStringer interface
func (or OwnerReference) GoString() string {
	return or.String()
}

// New returns a new instance of Builder
func New() *Builder {
	return &Builder{
		Object: &OwnerReference{
			OwnerReference: &metav1.OwnerReference{},
		},
	}
}

// WithName sets name in OwnerReference instance
func (b *Builder) WithName(name string) *Builder {
	b.Object.OwnerReference.Name = name
	return b
}

// WithKind sets kind in OwnerReference instance
func (b *Builder) WithKind(kind string) *Builder {
	b.Object.OwnerReference.Kind = kind
	return b
}

// WithAPIVersion sets api version in OwnerReference instance
func (b *Builder) WithAPIVersion(apiVersion string) *Builder {
	b.Object.OwnerReference.APIVersion = apiVersion
	return b
}

// WithUID sets uid in OwnerReference instance
func (b *Builder) WithUID(uid types.UID) *Builder {
	b.Object.OwnerReference.UID = uid
	return b
}

// WithControllerOption sets Controller property in OwnerReference instance
func (b *Builder) WithControllerOption(controller *bool) *Builder {
	b.Object.OwnerReference.Controller = controller
	return b
}

// WithBlockOwnerDeletionOption sets BlockOwnerDeletion property in OwnerReference instance
func (b *Builder) WithBlockOwnerDeletionOption(blockOwnerDeletion *bool) *Builder {
	b.Object.OwnerReference.BlockOwnerDeletion = blockOwnerDeletion
	return b
}

// WithAPIObject sets OwnerReference property in OwnerReference instance
func (b *Builder) WithAPIObject(ownerReference *metav1.OwnerReference) *Builder {
	b.Object.OwnerReference = ownerReference
	return b
}

// validate validates OwnerReference instance
func (b *Builder) validate() error {
	if len(b.errors) != 0 {
		return errors.Errorf("failed to validate: build errors were found: %v", b.errors)
	}
	validationErrs := []error{}
	if b.Object.OwnerReference.Name == "" {
		validationErrs = append(validationErrs, errors.New("missing name"))
	}
	if b.Object.OwnerReference.Kind == "" {
		validationErrs = append(validationErrs, errors.New("missing kind"))
	}
	if b.Object.OwnerReference.APIVersion == "" {
		validationErrs = append(validationErrs, errors.New("missing api version"))
	}
	if b.Object.OwnerReference.UID == "" {
		validationErrs = append(validationErrs, errors.New("missing uid"))
	}
	if len(validationErrs) != 0 {
		b.errors = append(b.errors, validationErrs...)
		return errors.Errorf("failed to validate: %v", validationErrs)
	}
	return nil
}

// Build builds a new instance of matav1.OwnerReference
func (b *Builder) Build() (*metav1.OwnerReference, error) {
	err := b.validate()
	if err != nil {
		return nil,
			errors.WithMessagef(err, "failed to build OwnerReference: %s", b.Object)
	}
	return b.Object.OwnerReference, nil
}
