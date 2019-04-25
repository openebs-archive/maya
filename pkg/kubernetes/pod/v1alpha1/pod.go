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
	v1 "k8s.io/api/core/v1"
)

// pod holds the api's pod objects
type pod struct {
	object *v1.Pod
}

// podList holds the list of pod instances
type podList struct {
	items []*pod
}

// listBuilder enables building an instance of
// podlist
type listBuilder struct {
	list    *podList
	filters predicateList
}

// WithAPIList builds the list of pod
// instances based on the provided
// pod list api instance
func (b *listBuilder) WithAPIList(pods *v1.PodList) *listBuilder {
	if pods == nil {
		return b
	}
	b.WithAPIObject(pods.Items...)
	return b
}

// WithObjects builds the list of pod
// instances based on the provided
// pod list instance
func (b *listBuilder) WithObject(pods ...*pod) *listBuilder {
	b.list.items = append(b.list.items, pods...)
	return b
}

// WithAPIList builds the list of pod
// instances based on the provided
// pod api instances
func (b *listBuilder) WithAPIObject(pods ...v1.Pod) *listBuilder {
	for _, p := range pods {
		p := p //pin it
		b.list.items = append(b.list.items, &pod{&p})
	}
	return b
}

// List returns the list of pod
// instances that was built by this
// builder
func (b *listBuilder) List() *podList {
	if b.filters == nil && len(b.filters) == 0 {
		return b.list
	}
	filtered := &podList{}
	for _, pod := range b.list.items {
		if b.filters.all(pod) {
			filtered.items = append(filtered.items, pod)
		}
	}
	return filtered
}

// Len returns the number of items present in the podList
func (p *podList) Len() int {
	return len(p.items)
}

// ToAPIList converts podList to API podList
func (p *podList) ToAPIList() *v1.PodList {
	plist := &v1.PodList{}
	for _, pod := range p.items {
		plist.Items = append(plist.Items, *pod.object)
	}
	return plist
}

// ListBuilder returns a instance of listBuilder
func ListBuilder() *listBuilder {
	return &listBuilder{list: &podList{items: []*pod{}}}
}

// predicate defines an abstraction
// to determine conditional checks
// against the provided pod instance
type predicate func(*pod) bool

// IsRunning retuns true if the pod is in running
// state
func (p *pod) IsRunning() bool {
	return p.object.Status.Phase == "Running"
}

// IsRunning is a predicate to filter out pods
// which in running state
func IsRunning() predicate {
	return func(p *pod) bool {
		return p.IsRunning()
	}
}

// IsNil returns true if the pod instance
// is nil
func (p *pod) IsNil() bool {
	return p.object == nil
}

// IsNil is predicate to filter out nil pod
// instances
func IsNil() predicate {
	return func(p *pod) bool {
		return p.IsNil()
	}
}

// predicateList holds a list of predicate
type predicateList []predicate

// all returns true if all the predicates
// succeed against the provided pod
// instance
func (l predicateList) all(p *pod) bool {
	for _, pred := range l {
		if !pred(p) {
			return false
		}
	}
	return true
}

// WithFilter add filters on which the pod
// has to be filtered
func (b *listBuilder) WithFilter(pred ...predicate) *listBuilder {
	b.filters = append(b.filters, pred...)
	return b
}
