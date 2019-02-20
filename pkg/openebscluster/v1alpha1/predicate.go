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
	apis "github.com/openebs/maya/pkg/apis/openebs.io/openebscluster/v1alpha1"
	objectref "github.com/openebs/maya/pkg/objectref/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type predicate struct {
	funcs []predicateFunc
}

type predicateFunc func(*openebsCluster) bool

// IsAnnotationsEmpty returns true if openebs
// cluster annotations is empty
func IsAnnotationsEmpty() predicateFunc {
	return func(c *openebsCluster) bool {
		if c.meta == nil || len(c.meta.GetAnnotations()) == 0 {
			return true
		}
		return false
	}
}

// IsControlled returns true if openebs cluster
// is controlled by any object
func IsControlled() predicateFunc {
	return func(c *openebsCluster) bool {
		if IsAnnotationsEmpty()(c) {
			return false
		}
		if c.meta.GetAnnotations()[string(objectref.ControllerKey)] == "" {
			return false
		}
		return true
	}
}

// IsNotControlled returns true if openebs cluster
// is not controlled by any object
func IsNotControlled() predicateFunc {
	return func(c *openebsCluster) bool {
		return !IsControlled()(c)
	}
}

// IsControlledByType returns true if openebs
// cluster is controlled by the provided kind
func IsControlledByType(kind string) predicateFunc {
	return func(c *openebsCluster) bool {
		if IsNotControlled()(c) {
			return false
		}
		ctl, _ := c.GetControllerRef()
		if ctl.Kind == kind {
			return true
		}
		return false
	}
}

// IsControlledByName returns true if openebs
// cluster is controlled by the provided
// controller name
func IsControlledByName(name string) predicateFunc {
	return func(c *openebsCluster) bool {
		if IsNotControlled()(c) {
			return false
		}
		ctl, _ := c.GetControllerRef()
		if ctl.Name == name {
			return true
		}
		return false
	}
}

// IsControlledByNameIfSet returns true if openebs
// cluster is controlled by the provided controller
// name.
//
// NOTE:
//  In case, if openebs cluster is not set with any
// controller name, then it returns true.
func IsControlledByNameIfSet(name string) predicateFunc {
	return func(c *openebsCluster) bool {
		if IsNotControlled()(c) {
			return false
		}
		ctl, _ := c.GetControllerRef()
		if ctl.Name == "" || ctl.Name == name {
			return true
		}
		return false
	}
}

// IsControlledByNamespace returns true if
// openebs cluster is controlled by any controller
// running in the provided namespace
func IsControlledByNamespace(namespace string) predicateFunc {
	return func(c *openebsCluster) bool {
		if IsNotControlled()(c) {
			return false
		}
		ctl, _ := c.GetControllerRef()
		if ctl.Namespace == namespace {
			return true
		}
		return false
	}
}

// IsControlledByNamespaceIfSet returns true if
// openebs cluster is controlled by any controller
// running in the provided namespace
//
// NOTE:
//  In case, if openebs cluster is not set with any
// namespace, then it returns true.
func IsControlledByNamespaceIfSet(namespace string) predicateFunc {
	return func(c *openebsCluster) bool {
		if IsNotControlled()(c) {
			return false
		}
		ctl, _ := c.GetControllerRef()
		if ctl.Namespace == "" || ctl.Namespace == namespace {
			return true
		}
		return false
	}
}

// Predicate returns a new instance of
// openebs cluster based predicate instance
func Predicate(p ...predicateFunc) *predicate {
	return &predicate{funcs: p}
}

// all checks the registered predicates
// and returns true if all of them
// succeed
func (p *predicate) all(c *openebsCluster) bool {
	for _, pred := range p.funcs {
		if !pred(c) {
			return false
		}
	}
	return true
}

// Create returns true if all openebs cluster
// related checks succeed
//
// NOTE:
//  Create implements controller runtime's Predicate
func (p *predicate) Create(e event.CreateEvent) bool {
	c := New(
		WithMeta(e.Meta),
		WithObject((e.Object).(*apis.OpenebsCluster)),
	)
	return p.all(c)
}

// Delete returns true if all openebs cluster
// related checks succeed
//
// NOTE:
//  Delete implements controller runtime's Predicate
func (p *predicate) Delete(e event.DeleteEvent) bool {
	c := New(
		WithMeta(e.Meta),
		WithObject((e.Object).(*apis.OpenebsCluster)),
	)
	return p.all(c)
}

// Update returns true if there is some valid
// changes to openebs cluster resource and the
// new change is as per the pre-defined predicates
//
// NOTE:
//  Update implements controller runtime's Predicate
func (p predicate) Update(e event.UpdateEvent) bool {
	oldc := New(
		WithMeta(e.MetaOld),
		WithObject((e.ObjectOld).(*apis.OpenebsCluster)),
	)
	newc := New(
		WithMeta(e.MetaNew),
		WithObject((e.ObjectNew).(*apis.OpenebsCluster)),
	)
	// was there any change?
	isValidChange := IsAnnotationsChange(oldc, newc) ||
		IsLabelsChange(oldc, newc) ||
		IsSpecificationsChange(oldc, newc)
	// return based on validity in change(s) & as per
	// pre-defined predicates against the new instance
	return isValidChange && p.all(newc)
}

// Generic returns true if all openebs cluster
// related checks succeed
//
// NOTE:
//  Generic implements controller runtime's Predicate
func (p predicate) Generic(e event.GenericEvent) bool {
	c := New(
		WithMeta(e.Meta),
		WithObject((e.Object).(*apis.OpenebsCluster)),
	)
	return p.all(c)
}
