// Copyright © 2018-2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Builder is the builder object for CStorVolume
type Builder struct {
	Cvr  *CVR
	errs []error
}

// NewBuilder returns new instance of Builder
func NewBuilder() *Builder {
	return &Builder{Cvr: &CVR{object: &apis.CStorVolumeReplica{}}}
}

// WithName sets the Name field of CStorVolumeReplica with provided value.
func (b *Builder) WithName(name string) *Builder {
	if len(name) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build cvr object: missing name"),
		)
		return b
	}
	b.Cvr.object.Name = name
	return b
}

// WithGenerateName sets the GenerateName field of
// CStorVolumeReplica with provided value
func (b *Builder) WithGenerateName(name string) *Builder {
	if len(name) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build cvr object: missing generateName"),
		)
		return b
	}

	b.Cvr.object.GenerateName = name
	return b
}

// WithNamespace sets the Namespace field of
// CStorVolumeReplica provided arguments
func (b *Builder) WithNamespace(namespace string) *Builder {
	if len(namespace) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build cvr object: missing namespace"),
		)
		return b
	}
	b.Cvr.object.Namespace = namespace
	return b
}

// WithAnnotations merges existing annotations if any
// with the ones that are provided here
func (b *Builder) WithAnnotations(annotations map[string]string) *Builder {
	if len(annotations) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build cvr object: no new annotations"),
		)
		return b
	}

	if b.Cvr.object.Labels == nil {
		return b.WithAnnotationsNew(annotations)
	}

	for key, value := range annotations {
		b.Cvr.object.Annotations[key] = value
	}
	return b
}

// WithAnnotationsNew resets existing annotations if any with
// ones that are provided here
func (b *Builder) WithAnnotationsNew(annotations map[string]string) *Builder {
	if len(annotations) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build cvr object: missing annotations"),
		)
		return b
	}

	// copy of original map
	newannotations := map[string]string{}
	for key, value := range annotations {
		newannotations[key] = value
	}

	// override
	b.Cvr.object.Annotations = newannotations
	return b
}

// WithOwnerRefernceNew sets ownerrefernce if any with
// ones that are provided here
func (b *Builder) WithOwnerRefernceNew(ownerRefernce []metav1.OwnerReference) *Builder {
	if len(ownerRefernce) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build cvr object: no new ownerRefernce"),
		)
		return b
	}

	b.Cvr.object.OwnerReferences = ownerRefernce
	return b
}

// WithLabels merges existing labels if any
// with the ones that are provided here
func (b *Builder) WithLabels(labels map[string]string) *Builder {
	if len(labels) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build cvr object: missing labels"),
		)
		return b
	}

	if b.Cvr.object.Labels == nil {
		return b.WithLabelsNew(labels)
	}

	for key, value := range labels {
		b.Cvr.object.Labels[key] = value
	}
	return b
}

// WithLabelsNew resets existing labels if any with
// ones that are provided here
func (b *Builder) WithLabelsNew(labels map[string]string) *Builder {
	if len(labels) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build cvr object: missing labels"),
		)
		return b
	}

	// copy of original map
	newlbls := map[string]string{}
	for key, value := range labels {
		newlbls[key] = value
	}

	// override
	b.Cvr.object.Labels = newlbls
	return b
}

// WithTargetIP sets the target IP address field of
// CStorVolumeReplica with provided arguments
func (b *Builder) WithTargetIP(targetip string) *Builder {
	if len(targetip) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build cvr object: missing targetip"),
		)
		return b
	}
	b.Cvr.object.Spec.TargetIP = targetip
	return b
}

// WithCapacity sets the Capacity field of
// CStorVolumeReplica with provided arguments
func (b *Builder) WithCapacity(capacity string) *Builder {
	if len(capacity) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build cvr object: missing capacity"),
		)
		return b
	}
	b.Cvr.object.Spec.Capacity = capacity
	return b
}

// WithFinalizers merges the existing finalizers if any
// with the ones provided arguments
func (b *Builder) WithFinalizers(finalizers []string) *Builder {
	if len(finalizers) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build cvr object: missing finalizers"),
		)
		return b
	}

	if b.Cvr.object.Finalizers == nil {
		return b.WithFinalizersNew(finalizers)
	}
	// copy of original slice
	newfinalizers := []string{}
	newfinalizers = append(newfinalizers, finalizers...)

	b.Cvr.object.Finalizers = append(
		b.Cvr.object.Finalizers,
		newfinalizers...,
	)
	return b
}

// WithFinalizersNew resets any existing finalizers
// and overrides them with provided arguments
func (b *Builder) WithFinalizersNew(finalizers []string) *Builder {
	if len(finalizers) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build cvr object: no new finalizers"),
		)
		return b
	}

	// copy of original slice
	newfinalizers := []string{}
	newfinalizers = append(newfinalizers, finalizers...)

	// override
	b.Cvr.object.Finalizers = newfinalizers
	return b
}

// Build returns the CStorVolumeReplica API instance
func (b *Builder) Build() (*apis.CStorVolumeReplica, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("%+v", b.errs)
	}
	return b.Cvr.object, nil
}
