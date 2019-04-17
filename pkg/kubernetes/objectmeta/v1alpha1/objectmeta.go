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

// Builder helps to build ObjectMeta
type Builder struct {
	errors []error
	Object *ObjectMeta
}

// ObjectMeta is a wrapper over metav1.ObjectMeta
type ObjectMeta struct {
	ObjectMeta *metav1.ObjectMeta
}

// String implements Stringer interface
func (om ObjectMeta) String() string {
	return stringer.Yaml("objectmeta", om)
}

// GoString implements GoStringer interface
func (om ObjectMeta) GoString() string {
	return om.String()
}

// New returns a new instance of Builder
func New() *Builder {
	return &Builder{
		Object: &ObjectMeta{
			ObjectMeta: &metav1.ObjectMeta{},
		},
	}
}

// WithName adds name in ObjectMeta instance
func (b *Builder) WithName(name string) *Builder {
	b.Object.ObjectMeta.Name = name
	return b
}

// WithNamespace adds namespace in ObjectMeta instance
func (b *Builder) WithNamespace(namespace string) *Builder {
	b.Object.ObjectMeta.Namespace = namespace
	return b
}

// WithLabels adds labels in ObjectMeta instance
func (b *Builder) WithLabels(labels map[string]string) *Builder {
	b.Object.ObjectMeta.Labels = labels
	return b
}

// WithAnnotations adds annotations in ObjectMeta instance
func (b *Builder) WithAnnotations(annotations map[string]string) *Builder {
	b.Object.ObjectMeta.Annotations = annotations
	return b
}

// WithOwnerReferences owner references in ObjectMeta instance
func (b *Builder) WithOwnerReferences(ownerReferences ...metav1.OwnerReference) *Builder {
	b.Object.ObjectMeta.OwnerReferences = append(b.Object.ObjectMeta.OwnerReferences, ownerReferences...)
	return b
}

// WithAPIObject sets Object property in ObjectMeta instance
func (b *Builder) WithAPIObject(objectMeta *metav1.ObjectMeta) *Builder {
	b.Object.ObjectMeta = objectMeta
	return b
}

// validate validates ObjectMeta instance
func (b *Builder) validate() error {
	if len(b.errors) != 0 {
		return errors.Errorf("failed to validate: build errors were found: %v", b.errors)
	}
	validationErrs := []error{}
	if b.Object.ObjectMeta.Name == "" {
		validationErrs = append(validationErrs, errors.New("missing name"))
	}
	if len(validationErrs) != 0 {
		b.errors = append(b.errors, validationErrs...)
		return errors.Errorf("failed to validate: %v", validationErrs)
	}
	return nil
}

// Build builds a new instance of metav1.ObjectMeta
func (b *Builder) Build() (*metav1.ObjectMeta, error) {
	err := b.validate()
	if err != nil {
		return nil,
			errors.WithMessagef(err, "failed to build ObjectMeta: %s", b.Object)
	}
	return b.Object.ObjectMeta, nil
}
