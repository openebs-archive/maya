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
	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
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
func (b *ListBuilder) List() (*PVCList, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("failed to list pvc: %+v", b.errs)
	}
	if b.filters == nil || len(b.filters) == 0 {
		return b.list, nil
	}
	filteredList := &PVCList{}
	for _, pvc := range b.list.items {
		if b.filters.all(pvc) {
			filteredList.items = append(filteredList.items, pvc)
		}
	}
	return filteredList, nil
}

// Len returns the number of items present
// in the PVCList of a builder
func (b *ListBuilder) Len() (int, error) {
	l, err := b.List()
	if err != nil {
		return 0, err
	}
	return l.Len(), nil
}

// APIList builds core API PVC list using listbuilder
func (b *ListBuilder) APIList() (*corev1.PersistentVolumeClaimList, error) {
	l, err := b.List()
	if err != nil {
		return nil, err
	}
	return l.ToAPIList(), nil
}

// WithFilter adds filters on which the pvc's has to be filtered
func (b *ListBuilder) WithFilter(pred ...Predicate) *ListBuilder {
	b.filters = append(b.filters, pred...)
	return b
}
