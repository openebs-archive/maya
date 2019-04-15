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

// ObjectMeta is a wrapper over metav1.ObjectMeta
type ObjectMeta struct {
	Object *metav1.ObjectMeta
	errors []error
}

// String implements Stringer interface
func (om ObjectMeta) String() string {
	return stringer.Yaml("objectmeta", om)
}

// GoString implements GoStringer interface
func (om ObjectMeta) GoString() string {
	return om.String()
}

// New returns a new instance of ObjectMeta
func New() *ObjectMeta {
	return &ObjectMeta{
		Object: &metav1.ObjectMeta{},
	}
}

// WithName adds name in ObjectMeta instance
func (om *ObjectMeta) WithName(name string) *ObjectMeta {
	om.Object.Name = name
	return om
}

// WithNamespace adds namespace in ObjectMeta instance
func (om *ObjectMeta) WithNamespace(namespace string) *ObjectMeta {
	om.Object.Namespace = namespace
	return om
}

// WithLabels adds labels in ObjectMeta instance
func (om *ObjectMeta) WithLabels(labels map[string]string) *ObjectMeta {
	om.Object.Labels = labels
	return om
}

// WithAnnotations adds annotations in ObjectMeta instance
func (om *ObjectMeta) WithAnnotations(annotations map[string]string) *ObjectMeta {
	om.Object.Annotations = annotations
	return om
}

// WithOwnerReferences owner references in ObjectMeta instance
func (om *ObjectMeta) WithOwnerReferences(ownerReferences ...metav1.OwnerReference) *ObjectMeta {
	om.Object.OwnerReferences = append(om.Object.OwnerReferences, ownerReferences...)
	return om
}

// WithAPIObject sets Object property in ObjectMeta instance
func (om *ObjectMeta) WithAPIObject(objectMeta *metav1.ObjectMeta) *ObjectMeta {
	om.Object = objectMeta
	return om
}

// validate validates ObjectMeta instance
func (om *ObjectMeta) validate() error {
	if len(om.errors) != 0 {
		return errors.Errorf("failed to validate: build errors were found: %v", om.errors)
	}
	validationErrs := []error{}
	if om.Object.Name == "" {
		validationErrs = append(validationErrs, errors.New("missing name"))
	}
	if len(validationErrs) != 0 {
		om.errors = append(om.errors, validationErrs...)
		return errors.Errorf("failed to validate: %v", validationErrs)
	}
	return nil
}

// Build builds a new instance of metav1.ObjectMeta
func (om *ObjectMeta) Build() (*metav1.ObjectMeta, error) {
	err := om.validate()
	if err != nil {
		return nil,
			errors.WithMessagef(err, "failed to build ObjectMeta: %s", om)
	}
	return om.Object, nil
}
