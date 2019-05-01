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
)

// CStorVolume a wrapper for CStorVolume object
type CStorVolume struct {
	// actual cstorvolume object
	object *apis.CStorVolume
}

// CStorVolumeList is a list of cstorvolume objects
type CStorVolumeList struct {
	// list of cstor volumes
	items []*CStorVolume
}

// ListBuilder enables building
// an instance of CstorVolumeList
type ListBuilder struct {
	list    *CStorVolumeList
	filters PredicateList
}

// NewListBuilder returns a new instance
// of listBuilder
func NewListBuilder() *ListBuilder {
	return &ListBuilder{list: &CStorVolumeList{}}
}

// WithAPIList builds the list of cstorvolume
// instances based on the provided
// cstorvolume api instances
func (b *ListBuilder) WithAPIList(list *apis.CStorVolumeList) *ListBuilder {
	if list == nil {
		return b
	}
	for _, c := range list.Items {
		c := c
		b.list.items = append(b.list.items, &CStorVolume{object: &c})
	}
	return b
}

// List returns the list of cstorvolume (cv)
// instances that was built by this
// builder
func (b *ListBuilder) List() *CStorVolumeList {
	if b.filters == nil || len(b.filters) == 0 {
		return b.list
	}
	filtered := &CStorVolumeList{}
	for _, cv := range b.list.items {
		if b.filters.all(cv) {
			filtered.items = append(filtered.items, cv)
		}
	}
	return filtered
}

// Len returns the number of items present
// in the CStorVolumeList
func (l *CStorVolumeList) Len() int {
	return len(l.items)
}

// Predicate defines an abstraction
// to determine conditional checks
// against the provided cstorvolume instance
type Predicate func(*CStorVolume) bool

// IsHealthy returns true if the CVR is in
// healthy state
func (p *CStorVolume) IsHealthy() bool {
	return p.object.Status.Phase == "Healthy"
}

// IsHealthy is a predicate to filter out cstorvolumes
// which is healthy
func IsHealthy() Predicate {
	return func(p *CStorVolume) bool {
		return p.IsHealthy()
	}
}

// PredicateList holds a list of cstor volume
// based predicates
type PredicateList []Predicate

// all returns true if all the predicates
// succeed against the provided cstorvolume
// instance
func (l PredicateList) all(c *CStorVolume) bool {
	for _, check := range l {
		if !check(c) {
			return false
		}
	}
	return true
}

// WithFilter adds filters on which the cstorvolume has to be filtered
func (b *ListBuilder) WithFilter(pred ...Predicate) *ListBuilder {
	b.filters = append(b.filters, pred...)
	return b
}

// NewForAPIObject returns a new instance of cstorvolume
func NewForAPIObject(obj *apis.CStorVolume) *CStorVolume {
	return &CStorVolume{
		object: obj,
	}
}
