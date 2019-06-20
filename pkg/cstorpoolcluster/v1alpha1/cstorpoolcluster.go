// Copyright © 2019 The OpenEBS Authors
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
	apisv1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
)

// CSPC is a wrapper over cstorpoolcluster api
// object. It provides build, validations and other common
// logic to be used by various feature specific callers.
type CSPC struct {
	object *apisv1alpha1.CStorPoolCluster
}

// CSPCList is a wrapper over cstorpoolcluster api
// object. It provides build, validations and other common
// logic to be used by various feature specific callers.
type CSPCList struct {
	items []*CSPC
}

// Len returns the number of items present
// in the CSPCList
func (c *CSPCList) Len() int {
	if c == nil {
		return 0
	}
	return len(c.items)
}

// ToAPIList converts CSPCList to API CSPCList
func (c *CSPCList) ToAPIList() *apisv1alpha1.CStorPoolClusterList {
	clist := &apisv1alpha1.CStorPoolClusterList{}
	for _, cspc := range c.items {
		clist.Items = append(clist.Items, *cspc.object)
	}
	return clist
}

type cspcBuildOption func(*CSPC)

// NewForAPIObject returns a new instance of CSPC
func NewForAPIObject(obj *apisv1alpha1.CStorPoolCluster, opts ...cspcBuildOption) *CSPC {
	c := &CSPC{object: obj}
	for _, o := range opts {
		o(c)
	}
	return c
}

// Predicate defines an abstraction
// to determine conditional checks
// against the provided cspc instance
type Predicate func(*CSPC) bool

// IsNil returns true if the CSPC instance
// is nil
func (c *CSPC) IsNil() bool {
	return c.object == nil
}

// IsNil is predicate to filter out nil CSPC
// instances
func IsNil() Predicate {
	return func(c *CSPC) bool {
		return c.IsNil()
	}
}

// PredicateList holds a list of predicate
type PredicateList []Predicate

// all returns true if all the predicates
// succeed against the provided CSPC
// instance
func (l PredicateList) all(p *CSPC) bool {
	for _, pred := range l {
		if !pred(p) {
			return false
		}
	}
	return true
}

// WithFilter adds filters on which the cspc's has to be filtered
func (b *ListBuilder) WithFilter(pred ...Predicate) *ListBuilder {
	b.filters = append(b.filters, pred...)
	return b
}
