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

package v1alpha2

import (
	"text/template"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
)

type annotationKey string

const scheduleOnHostAnnotation annotationKey = "volume.kubernetes.io/selected-node"

type csp struct {
	// actual cstor pool object
	object *apis.CStorPool
}

// predicate defines an abstraction
// to determine conditional checks
// against the provided csp instance
type predicate func(*csp) bool

type predicateList []predicate

// all returns true if all the predicates
// succeed against the provided csp
// instance
func (l predicateList) all(c *csp) bool {
	for _, pred := range l {
		if !pred(c) {
			return false
		}
	}
	return true
}

// IsNotUID returns true if provided csp
// instance's UID does not match with any
// of the provided UIDs
func IsNotUID(uids ...string) predicate {
	return func(c *csp) bool {
		for _, uid := range uids {
			if uid == string(c.object.GetUID()) {
				return false
			}
		}
		return true
	}
}

// HasAnnotation returns true if provided annotation
// key and value are present in the provided CSP
// instance
func HasAnnotation(key, value string) predicate {
	return func(c *csp) bool {
		val, ok := c.object.GetAnnotations()[key]
		if ok {
			return val == value
		}
		return false
	}
}

// CSPList holds the list of cstorpools
type CSPList struct {
	// list of cstor pools
	items []*csp
}

// Filter will filter the csp instances
// if all the predicates succeed against that
// csp.
func (l *CSPList) Filter(p ...predicate) *CSPList {
	var plist predicateList
	plist = append(plist, p...)
	if len(plist) == 0 {
		return l
	}

	filtered := ListBuilder().List()
	for _, csp := range l.items {
		if plist.all(csp) {
			filtered.items = append(filtered.items, csp)
		}
	}
	return filtered
}

// listBuilder enables building a
// list of csp instances
type listBuilder struct {
	list *CSPList
}

// ListBuilder returns a new instance of
// listBuilder object
func ListBuilder() *listBuilder {
	return &listBuilder{list: &CSPList{}}
}

// WithUIDs builds a list of cstor pools
// based on the provided pool UIDs
func (b *listBuilder) WithUIDs(poolUIDs ...string) *listBuilder {
	for _, uid := range poolUIDs {
		obj := &csp{&apis.CStorPool{}}
		obj.object.SetUID(types.UID(uid))
		b.list.items = append(b.list.items, obj)
	}
	return b
}

// WithUIDNodeMap builds a cspList based on the provided
// map of uid and nodename
func (b *listBuilder) WithUIDNode(UIDNode map[string]string) *listBuilder {
	for k, v := range UIDNode {
		obj := &csp{&apis.CStorPool{}}
		obj.object.SetUID(types.UID(k))
		obj.object.SetAnnotations(map[string]string{string(scheduleOnHostAnnotation): v})
		b.list.items = append(b.list.items, obj)
	}
	return b
}

// WithList builds the list based on the provided
// *apis.CStorPool instances
func (b *listBuilder) WithList(pools *CSPList) *listBuilder {
	if pools == nil {
		return b
	}
	b.list.items = append(b.list.items, pools.items...)
	return b
}

// WithAPIList builds the list based on the provided
// *apis.CStorPoolList
func (b *listBuilder) WithAPIList(pools *apis.CStorPoolList) *listBuilder {
	if pools == nil {
		return b
	}
	for _, pool := range pools.Items {
		pool := pool //pin it
		b.list.items = append(b.list.items, &csp{&pool})
	}

	return b
}

// List returns the list of csp
// instances that were built by
// this builder
func (b *listBuilder) List() *CSPList {
	return b.list
}

// GetPoolUIDs retuns the UIDs of the pools
// available in the list
func (l *CSPList) GetPoolUIDs() []string {
	uids := []string{}
	for _, pool := range l.items {
		uids = append(uids, string(pool.object.GetUID()))
	}
	return uids
}

// newListFromUIDNode exposes WithUIDNodeMap
// to CAS Templates
func newListFromUIDNode(UIDNodeMap map[string]string) *CSPList {
	return ListBuilder().WithUIDNode(UIDNodeMap).List()
}

// newListFromUIDs exposes WithUIDs to CASTemplates
func newListFromUIDs(uids []string) *CSPList {
	return ListBuilder().WithUIDs(uids...).List()
}

// TemplateFunctions exposes a few functions as go
// template functions
func TemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"createCSPListFromUIDs":       newListFromUIDs,
		"createCSPListFromUIDNodeMap": newListFromUIDNode,
	}
}
