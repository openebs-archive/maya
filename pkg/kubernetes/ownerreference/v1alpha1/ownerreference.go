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
	ownerRef *OwnerReference
	errors   []error
}

// OwnerReference is a wrapper over metav1.OwnerReference
type OwnerReference struct {
	object *metav1.OwnerReference
}

// String implements Stringer interface
func (or OwnerReference) String() string {
	return stringer.Yaml("ownerreference", or.object)
}

// GoString implements GoStringer interface
func (or OwnerReference) GoString() string {
	return or.String()
}

// NewBuilder returns a new instance of Builder
func NewBuilder() *Builder {
	return &Builder{
		ownerRef: &OwnerReference{
			object: &metav1.OwnerReference{},
		},
	}
}

// NewBuilderForAPIObject returns a new instance of Builder
// for given metav1.OwnerReference
func NewBuilderForAPIObject(meta *metav1.OwnerReference) *Builder {
	b := &Builder{}
	if meta == nil {
		b.errors = append(b.errors,
			errors.New("failed to init builder: nil OwnerReference provided"))
		b.ownerRef = &OwnerReference{object: &metav1.OwnerReference{}}
		return b
	}
	b.ownerRef = &OwnerReference{object: meta}
	return b
}

// WithName sets name in OwnerReference instance
func (b *Builder) WithName(name string) *Builder {
	b.ownerRef.object.Name = name
	return b
}

// WithKind sets kind in OwnerReference instance
func (b *Builder) WithKind(kind string) *Builder {
	b.ownerRef.object.Kind = kind
	return b
}

// WithAPIVersion sets api version in OwnerReference instance
func (b *Builder) WithAPIVersion(apiVersion string) *Builder {
	b.ownerRef.object.APIVersion = apiVersion
	return b
}

// WithUID sets uid in OwnerReference instance
func (b *Builder) WithUID(uid types.UID) *Builder {
	b.ownerRef.object.UID = uid
	return b
}

// WithControllerOption sets Controller property in OwnerReference instance
func (b *Builder) WithControllerOption(controller *bool) *Builder {
	b.ownerRef.object.Controller = controller
	return b
}

// WithBlockOwnerDeletionOption sets BlockOwnerDeletion property in OwnerReference instance
func (b *Builder) WithBlockOwnerDeletionOption(blockOwnerDeletion *bool) *Builder {
	b.ownerRef.object.BlockOwnerDeletion = blockOwnerDeletion
	return b
}

// validate validates OwnerReference instance
func (b *Builder) validate() error {
	if len(b.errors) != 0 {
		return errors.Errorf("failed to validate: build errors were found: %v", b.errors)
	}
	validationErrs := []error{}
	if b.ownerRef.object.Name == "" {
		validationErrs = append(validationErrs, errors.New("missing name"))
	}
	if b.ownerRef.object.Kind == "" {
		validationErrs = append(validationErrs, errors.New("missing kind"))
	}
	if b.ownerRef.object.APIVersion == "" {
		validationErrs = append(validationErrs, errors.New("missing api version"))
	}
	if b.ownerRef.object.UID == "" {
		validationErrs = append(validationErrs, errors.New("missing uid"))
	}
	if len(validationErrs) != 0 {
		b.errors = append(b.errors, validationErrs...)
		return errors.Errorf("validation error(s) found: %v", validationErrs)
	}
	return nil
}

// Build builds a new instance of matav1.OwnerReference
func (b *Builder) Build() (*metav1.OwnerReference, error) {
	err := b.validate()
	if err != nil {
		return nil,
			errors.Wrapf(err, "failed to build OwnerReference: %s", b.ownerRef.object)
	}
	return b.ownerRef.object, nil
}
