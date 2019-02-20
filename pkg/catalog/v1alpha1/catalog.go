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
	"reflect"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/catalog/v1alpha1"
	objectrefapis "github.com/openebs/maya/pkg/apis/openebs.io/objectref/v1alpha1"
	objectref "github.com/openebs/maya/pkg/objectref/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

// catalog exposes utilities w.r.t
// api catalog object
type catalog struct {
	meta   v1.Object
	object *apis.Catalog
}

type buildOption func(*catalog)

// WithObject sets the provided catalog object
// instance against this catalog utility
func WithObject(obj *apis.Catalog) buildOption {
	return func(c *catalog) {
		c.object = obj
	}
}

// WithMeta sets the provided meta instance
// against this catalog utility
func WithMeta(meta v1.Object) buildOption {
	return func(c *catalog) {
		c.meta = meta
	}
}

// New returns a new instance of catalog
func New(opts ...buildOption) *catalog {
	c := &catalog{}
	for _, o := range opts {
		o(c)
	}
	c.withDefaults()
	return c
}

func (c *catalog) withDefaults() {
	if c.meta == nil && c.object != nil {
		c.meta = &c.object.ObjectMeta
	}
}

// GetControllerRef returns the controller that is
// refered to by this catalog instance
func (c *catalog) GetControllerRef() (objectrefapis.ControllerRef, error) {
	anns := c.meta.GetAnnotations()
	if len(anns) == 0 {
		return objectrefapis.ControllerRef{}, errors.New("failed to get catalog's controller reference: empty annotations")
	}
	return objectref.Builder().
		WithReference(anns[string(objectref.ControllerKey)]).
		ControllerRef()
}

// IsSpecificationsChange returns true if there
// is a change in specifications between the two
// catalog instances
func IsSpecificationsChange(oldc *catalog, newc *catalog) bool {
	return reflect.DeepEqual(oldc.object.Spec, newc.object.Spec)
}

// IsAnnotationsChange returns true if there
// is a change in annotations between the two
// catalog instances
func IsAnnotationsChange(oldc *catalog, newc *catalog) bool {
	if len(oldc.meta.GetAnnotations()) != len(newc.meta.GetAnnotations()) {
		return true
	}
	return reflect.DeepEqual(oldc.meta.GetAnnotations(), newc.meta.GetAnnotations())
}

// IsLabelsChange returns true if there
// is a change in labels between the two
// catalog instances
func IsLabelsChange(oldc *catalog, newc *catalog) bool {
	if len(oldc.meta.GetLabels()) != len(newc.meta.GetLabels()) {
		return true
	}
	return reflect.DeepEqual(oldc.meta.GetLabels(), newc.meta.GetLabels())
}
