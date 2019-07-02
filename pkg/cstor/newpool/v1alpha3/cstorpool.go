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

// CStorPool encapsulates CStorPool api object.
type CStorPool struct {
	// actual csp object
	Object *apis.NewTestCStorPool
}

// CStorPoolList encapsulates CStorPoolList api object
type CStorPoolList struct {
	// list of CSPs
	ObjectList *apis.NewTestCStorPoolList
}

// Predicate defines an abstraction to determine conditional checks against the
// provided CStorPool instance
type Predicate func(*CStorPool) bool

// PredicateList holds the list of Predicates
type PredicateList []Predicate

// all returns true if all the predicates succeed against the provided block
// device instance.
func (l PredicateList) all(c *CStorPool) bool {
	for _, pred := range l {
		if !pred(c) {
			return false
		}
	}
	return true
}

// HasAnnotation is predicate to filter out based on
// annotation in CSP instances
func HasAnnotation(key, value string) Predicate {
	return func(csp *CStorPool) bool {
		return csp.HasAnnotation(key, value)
	}
}

// HasAnnotation return true if provided annotation
// key and value are present in the the provided CSPList
// instance
func (csp *CStorPool) HasAnnotation(key, value string) bool {
	val, ok := csp.Object.GetAnnotations()[key]
	if ok {
		return val == value
	}
	return false
}

// HasLabel is predicate to filter out labeled
// CSP instances
func HasLabel(key, value string) Predicate {
	return func(csp *CStorPool) bool {
		return csp.HasLabel(key, value)
	}
}

// HasLabel returns true if provided label
// key and value are present in the provided CSP(CStorPool)
// instance
func (csp *CStorPool) HasLabel(key, value string) bool {
	val, ok := csp.Object.GetLabels()[key]
	if ok {
		return val == value
	}
	return false
}

// IsStatus is predicate to filter out CSP instances based on argument provided
func IsStatus(status string) Predicate {
	return func(csp *CStorPool) bool {
		return csp.IsStatus(status)
	}
}

// IsStatus returns true if the status on
// block device claim matches with provided status.
func (csp *CStorPool) IsStatus(status string) bool {
	return string(csp.Object.Status.Phase) == status
}

// Len returns the length og CStorPoolList.
func (cspl *CStorPoolList) Len() int {
	return len(cspl.ObjectList.Items)
}
