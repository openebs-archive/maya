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
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

// ListBuilder enables building an instance of
// PVClist
type ListBuilder struct {
	list    *PVCList
	filters PredicateList
	errs    []error
}

// NewListBuilder returns an instance of ListBuilder
func NewListBuilder() *ListBuilder {
	return &ListBuilder{list: &PVCList{}}
}

// Build returns the final instance of patch
func (b *ListBuilder) Build() (*PVCList, []error) {
	if len(b.errs) > 0 {
		return nil, b.errs
	}
	return b.list, nil
}

// ListBuilderForAPIObjects builds the ListBuilder object based on PVC api list
func ListBuilderForAPIObjects(pvcs *corev1.PersistentVolumeClaimList) *ListBuilder {
	b := &ListBuilder{list: &PVCList{}}
	if pvcs == nil {
		b.errs = append(b.errs, errors.New("failed to build pvc list: missing api list"))
		return b
	}
	for _, pvc := range pvcs.Items {
		pvc := pvc
		b.list.items = append(b.list.items, &PVC{object: &pvc})
	}
	return b
}

// ListBuilderForObjects builds the ListBuilder object based on PVCList
func ListBuilderForObjects(pvcs *PVCList) *ListBuilder {
	b := &ListBuilder{}
	if pvcs == nil {
		b.errs = append(b.errs, errors.New("failed to build pvc list: missing object list"))
		return b
	}
	b.list = pvcs
	return b
}

// List returns the list of pvc
// instances that was built by this
// builder
func (b *ListBuilder) List() *PVCList {
	if b.filters == nil || len(b.filters) == 0 {
		return b.list
	}
	filtered := &PVCList{}
	for _, pvc := range b.list.items {
		if b.filters.all(pvc) {
			filtered.items = append(filtered.items, pvc)
		}
	}
	return filtered
}

// Len returns the number of items present
// in the PVCList of a builder
func (b *ListBuilder) Len() int {
	return len(b.list.items)
}

// APIList builds core API PVC list using listbuilder
func (b *ListBuilder) APIList() (*corev1.PersistentVolumeClaimList, []error) {
	l, errs := b.Build()
	if len(errs) > 0 {
		return nil, errs
	}
	if l == nil {
		errs := append(errs, errors.New("failed to build pvc list: object list nil"))
		return nil, errs
	}
	return l.ToAPIList(), nil
}
