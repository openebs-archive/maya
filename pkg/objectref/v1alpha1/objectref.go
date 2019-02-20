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
	"github.com/ghodss/yaml"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/objectref/v1alpha1"
	"github.com/pkg/errors"
)

// AnnotationKey defines a custom type that
// represents a key. Constants based on this
// key are used to refer to an object.
type AnnotationKey string

const (
	// ControllerKey helps in extracting the
	// controller reference
	ControllerKey AnnotationKey = "openebs.io/controller-reference"

	// CatalogKey helps in extracting the
	// catalog reference
	CatalogKey AnnotationKey = "openebs.io/catalog-reference"

	// HashKey helps in extracting all the
	// hash references
	HashKey AnnotationKey = "openebs.io/hash-reference-list"
)

type objectref struct {
	refer string // object reference in yaml string format
}

type builder struct {
	objectref
}

// Builder returns a new instance
// of builder that can build an instance
// of objectref
func Builder() *builder {
	return &builder{}
}

// WithReference sets the provided object
// reference
//
// NOTE:
//  object reference is assumed to be in
// yaml string format
func (b *builder) WithReference(objyaml string) *builder {
	b.refer = objyaml
	return b
}

// ControllerRef builds and returns the ControllerRef
// type
func (b *builder) ControllerRef() (apis.ControllerRef, error) {
	ref := apis.ControllerRef{}
	if b.refer == "" {
		return ref, errors.New("failed to build controller reference: empty reference")
	}
	err := yaml.Unmarshal([]byte(b.refer), &ref)
	if err != nil {
		err = errors.Wrapf(err, "failed to build controller reference from: '%s'", b.refer)
	}
	return ref, err
}

// CatalogRef builds and returns the CatalogRef
// type
func (b *builder) CatalogRef() (apis.CatalogRef, error) {
	ref := apis.CatalogRef{}
	if b.refer == "" {
		return ref, errors.New("failed to build catalog reference: empty reference")
	}
	err := yaml.Unmarshal([]byte(b.refer), &ref)
	if err != nil {
		err = errors.Wrapf(err, "failed to build catalog reference from: '%s'", b.refer)
	}
	return ref, err
}

// HashRefList builds and returns the HashRefList
// type
func (b *builder) HashRefList() (apis.HashRefList, error) {
	ref := apis.HashRefList{}
	if b.refer == "" {
		return ref, errors.New("failed to build hash reference list: empty reference")
	}
	err := yaml.Unmarshal([]byte(b.refer), &ref)
	if err != nil {
		err = errors.Wrapf(err, "failed to build hash reference list from: '%s'", b.refer)
	}
	return ref, err
}
