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
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
)

// CStorVolumeClaim a wrapper for ume object
type CStorVolumeClaim struct {
	// actual cstorvolumeclaim object
	object *apis.CStorVolumeClaim
}

// CStorVolumeClaimList is a list of cstorvolumeclaim objects
type CStorVolumeClaimList struct {
	// list of cstor volume claims
	items []*CStorVolumeClaim
}

// ListBuilder enables building
// an instance of umeCStorVolumeClaimList
type ListBuilder struct {
	list    *CStorVolumeClaimList
	filters PredicateList
}

// NewListBuilder returns a new instance
// of listBuilder
func NewListBuilder() *ListBuilder {
	return &ListBuilder{list: &CStorVolumeClaimList{}}
}

// WithAPIList builds the list of cstorvolume claim
// instances based on the provided
// CStorVolumeClaim api instances
func (b *ListBuilder) WithAPIList(
	list *apis.CStorVolumeClaimList) *ListBuilder {
	if list == nil {
		return b
	}
	for _, c := range list.Items {
		c := c
		b.list.items = append(b.list.items, &CStorVolumeClaim{object: &c})
	}
	return b
}

// List returns the list of CStorVolumeClaims (cvcs)
// instances that was built by this
// builder
func (b *ListBuilder) List() *CStorVolumeClaimList {
	if b.filters == nil || len(b.filters) == 0 {
		return b.list
	}
	filtered := &CStorVolumeClaimList{}
	for _, cv := range b.list.items {
		if b.filters.all(cv) {
			filtered.items = append(filtered.items, cv)
		}
	}
	return filtered
}

// Len returns the number of items present
// in the CStorVolumeClaimList
func (l *CStorVolumeClaimList) Len() int {
	return len(l.items)
}

// Predicate defines an abstraction
// to determine conditional checks
// against the provided cstorvolume claim instance
type Predicate func(*CStorVolumeClaim) bool

// PredicateList holds a list of cstor volume claims
// based predicates
type PredicateList []Predicate

// all returns true if all the predicates
// succeed against the provided cstorvolumeclaim
// instance
func (l PredicateList) all(c *CStorVolumeClaim) bool {
	for _, check := range l {
		if !check(c) {
			return false
		}
	}
	return true
}

// WithFilter adds filters on which the cstorvolumeclaim has to be filtered
func (b *ListBuilder) WithFilter(pred ...Predicate) *ListBuilder {
	b.filters = append(b.filters, pred...)
	return b
}

// NewForAPIObject returns a new instance of cstorvolume
func NewForAPIObject(obj *apis.CStorVolumeClaim) *CStorVolumeClaim {
	return &CStorVolumeClaim{
		object: obj,
	}
}

// IsCVCBounded returns true only if cvc is in bound state
func (cvc *CStorVolumeClaim) IsCVCBounded() bool {
	return cvc.object.Status.Phase == apis.CStorVolumeClaimPhaseBound
}

// IsCVCBounded is a predicate to filter out cstorvolumeclaims based on bound
// state
func IsCVCBounded() Predicate {
	return func(cvc *CStorVolumeClaim) bool {
		return cvc.IsCVCBounded()
	}
}

// IsCVCPending returns true only if cvc is in pending state
func (cvc *CStorVolumeClaim) IsCVCPending() bool {
	return cvc.object.Status.Phase == apis.CStorVolumeClaimPhasePending
}

// IsCVCPending is a predicate to filter out cstorvolumeclaims based on pending state
func IsCVCPending() Predicate {
	return func(cvc *CStorVolumeClaim) bool {
		return cvc.IsCVCPending()
	}
}

// HasAnnotation returns true only if cvc annotation has volume name matching to
// provided arguments
func (cvc *CStorVolumeClaim) HasAnnotation(key, value string) bool {
	return cvc.object.GetAnnotations()[key] == value
}

// HasAnnotation is a predicate to filter out cstorvolumeclaims based on
// annotaion values
func HasAnnotation(key, value string) Predicate {
	return func(cvc *CStorVolumeClaim) bool {
		return cvc.HasAnnotation(key, value)
	}
}
