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
	corev1 "k8s.io/api/core/v1"
)

// Pod holds the api's pod objects
type Pod struct {
	object *corev1.Pod
}

// PodList holds the list of API pod instances
type PodList struct {
	items []*Pod
}

// PredicateList holds a list of predicate
type predicateList []Predicate

// Predicate defines an abstraction
// to determine conditional checks
// against the provided pod instance
type Predicate func(*Pod) bool

// ToAPIList converts PodList to API PodList
func (p *PodList) ToAPIList() *corev1.PodList {
	plist := &corev1.PodList{}
	for _, pod := range p.items {
		plist.Items = append(plist.Items, *pod.object)
	}
	return plist
}

// Len returns the number of items present in the PodList
func (p *PodList) Len() int {
	return len(p.items)
}

// all returns true if all the predicates
// succeed against the provided pod
// instance
func (l predicateList) all(p *Pod) bool {
	for _, pred := range l {
		if !pred(p) {
			return false
		}
	}
	return true
}

// IsRunning retuns true if the pod is in running
// state
func (p *Pod) IsRunning() bool {
	return p.object.Status.Phase == "Running"
}

// IsRunning is a predicate to filter out pods
// which in running state
func IsRunning() Predicate {
	return func(p *Pod) bool {
		return p.IsRunning()
	}
}

// HasLabels returns true if provided labels
// map[key]value are present in the provided PodList
// instance
func HasLabels(keyValuePair map[string]string) Predicate {
	return func(p *Pod) bool {
		//		objKeyValues := p.object.GetLabels()
		for key, value := range keyValuePair {
			if !p.HasLabel(key, value) {
				return false
			}
		}
		return true
	}
}

// HasLabel return true if provided lable
// key and value are present in the the provided PodList
// instance
func (p *Pod) HasLabel(key, value string) bool {
	val, ok := p.object.GetLabels()[key]
	if ok {
		return val == value
	}
	return false
}

// IsNil returns true if the pod instance
// is nil
func (p *Pod) IsNil() bool {
	return p.object == nil
}

// IsNil is predicate to filter out nil pod
// instances
func IsNil() Predicate {
	return func(p *Pod) bool {
		return p.IsNil()
	}
}

// GetAPIObject returns a API's Pod
func (p *Pod) GetAPIObject() *corev1.Pod {
	return p.object
}
