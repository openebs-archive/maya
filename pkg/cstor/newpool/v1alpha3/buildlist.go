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

package v1alpha3

import (
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
)

// ListBuilder is the builder object for CStorPoolList
type ListBuilder struct {
	CSPList *CStorPoolList
	filters PredicateList
}

// NewListBuilder returns a new instance of ListBuilder object.
func NewListBuilder() *ListBuilder {
	return &ListBuilder{
		CSPList: &CStorPoolList{
			ObjectList: &apis.NewTestCStorPoolList{},
		},
		filters: PredicateList{},
	}
}

// ListBuilderFromList builds the list based on the
// provided *CStorPoolList instances.
func ListBuilderFromList(cspl *CStorPoolList) *ListBuilder {
	lb := NewListBuilder()
	if cspl == nil {
		return lb
	}
	lb.CSPList.ObjectList.Items =
		append(lb.CSPList.ObjectList.Items,
			cspl.ObjectList.Items...)
	return lb
}

// ListBuilderFromAPIList builds the list based on the provided API CSP List
func ListBuilderFromAPIList(cspl *apis.NewTestCStorPoolList) *ListBuilder {
	lb := NewListBuilder()
	if cspl == nil {
		return lb
	}
	lb.CSPList.ObjectList.Items = append(
		lb.CSPList.ObjectList.Items,
		cspl.Items...)
	return lb
}

// List returns the list of csp
// instances that was built by this
// builder
func (lb *ListBuilder) List() *CStorPoolList {
	if lb.filters == nil || len(lb.filters) == 0 {
		return lb.CSPList
	}
	filtered := NewListBuilder().List()
	for _, cspAPI := range lb.CSPList.ObjectList.Items {
		cspAPI := cspAPI // pin it
		csp := BuilderForAPIObject(&cspAPI).CSP
		if lb.filters.all(csp) {
			filtered.ObjectList.Items = append(filtered.ObjectList.Items, *csp.Object)
		}
	}
	return filtered
}

// WithFilter adds filters on which the csp's has to be filtered
func (lb *ListBuilder) WithFilter(pred ...Predicate) *ListBuilder {
	lb.filters = append(lb.filters, pred...)
	return lb
}

// GetCStorPool returns CStorPool object from existing
// ListBuilder
func (lb *ListBuilder) GetCStorPool(cspName string) *apis.NewTestCStorPool {
	for _, cspObj := range lb.CSPList.ObjectList.Items {
		cspObj := cspObj
		if cspObj.Name == cspName {
			return &cspObj
		}
	}
	return nil
}
