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
	apisv1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
)

// ListBuilder enables building an instance of
// CSPCList
type ListBuilder struct {
	list    *CSPCList
	filters PredicateList
	errs    []error
}

// NewListBuilder returns an instance of ListBuilder
func NewListBuilder() *ListBuilder {
	return &ListBuilder{list: &CSPCList{}}
}

// ListBuilderForAPIObjects builds the ListBuilder object based on CSPC api list
func ListBuilderForAPIObjects(cspcs *apisv1alpha1.CStorPoolClusterList) *ListBuilder {
	b := NewListBuilder()
	if cspcs == nil {
		b.errs = append(b.errs, errors.New("failed to build cspc list: missing api list"))
		return b
	}
	for _, cspc := range cspcs.Items {
		cspc := cspc
		b.list.items = append(b.list.items, &CSPC{object: &cspc})
	}
	return b
}

// ListBuilderForObjects builds the ListBuilder object based on CSPCList
func ListBuilderForObjects(cspcs *CSPCList) *ListBuilder {
	b := NewListBuilder()
	if cspcs == nil {
		b.errs = append(b.errs, errors.New("failed to build cspc list: missing object list"))
		return b
	}
	b.list = cspcs
	return b
}

// List returns the list of cspc
// instances that was built by this
// builder
func (b *ListBuilder) List() (*CSPCList, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("failed to list cspc: %+v", b.errs)
	}
	if b.filters == nil || len(b.filters) == 0 {
		return b.list, nil
	}
	filteredList := &CSPCList{}
	for _, cspc := range b.list.items {
		if b.filters.all(cspc) {
			filteredList.items = append(filteredList.items, cspc)
		}
	}
	return filteredList, nil
}

// Len returns the number of items present
// in the CSPCList of a builder
func (b *ListBuilder) Len() (int, error) {
	l, err := b.List()
	if err != nil {
		return 0, err
	}
	return l.Len(), nil
}

// APIList builds core API CSPC list using listbuilder
func (b *ListBuilder) APIList() (*apisv1alpha1.CStorPoolClusterList, error) {
	l, err := b.List()
	if err != nil {
		return nil, err
	}
	return l.ToAPIList(), nil
}
