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

// CSPI encapsulates CStorPoolInstance api object.
type CSPI struct {
	// actual CSPI object
	Object *apis.CStorPoolInstance
}

// CSPIList encapsulates CStorPoolList api object
type CSPIList struct {
	// list of CSPIs
	ObjectList *apis.CStorPoolInstanceList
}

// Predicate defines an abstraction to determine conditional checks against the
// provided CStorPoolInstance
type Predicate func(*CSPI) bool

// PredicateList holds the list of Predicates
type PredicateList []Predicate

// all returns true if all the predicates succeed against the provided block
// device instance.
func (l PredicateList) all(c *CSPI) bool {
	for _, pred := range l {
		if !pred(c) {
			return false
		}
	}
	return true
}

// HasAnnotation is predicate to filter out based on
// annotation in CSPI instances
func HasAnnotation(key, value string) Predicate {
	return func(c *CSPI) bool {
		return c.HasAnnotation(key, value)
	}
}

// HasAnnotation return true if provided annotation
// key and value are present in the the provided CSPIList
// instance
func (c *CSPI) HasAnnotation(key, value string) bool {
	val, ok := c.Object.GetAnnotations()[key]
	if ok {
		return val == value
	}
	return false
}

// HasNodeName is predicate to filter out based on
// node name of CSPI instances.
func HasNodeName(nodeName string) Predicate {
	return func(c *CSPI) bool {
		return c.HasNodeName(nodeName)
	}
}

// HasNodeName returns true if the CSPI belongs
// to the provided node name.
func (c *CSPI) HasNodeName(nodeName string) bool {
	return c.Object.Spec.HostName == nodeName
}

// HasLabel is predicate to filter out labeled
// CSPI instances
func HasLabel(key, value string) Predicate {
	return func(c *CSPI) bool {
		return c.HasLabel(key, value)
	}
}

// HasLabel returns true if provided label
// key and value are present in the provided
// CSPI
func (c *CSPI) HasLabel(key, value string) bool {
	val, ok := c.Object.GetLabels()[key]
	if ok {
		return val == value
	}
	return false
}

// IsStatus is predicate to filter out CSP instances based on argument provided
func IsStatus(status string) Predicate {
	return func(c *CSPI) bool {
		return c.IsStatus(status)
	}
}

// IsStatus returns true if the status on
// block device claim matches with provided status.
func (c *CSPI) IsStatus(status string) bool {
	return string(c.Object.Status.Phase) == status
}

// Len returns the length of CStorPoolInstanceList.
func (c *CSPIList) Len() int {
	return len(c.ObjectList.Items)
}
