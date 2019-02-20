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

	objectrefapis "github.com/openebs/maya/pkg/apis/openebs.io/objectref/v1alpha1"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/openebscluster/v1alpha1"
	objectref "github.com/openebs/maya/pkg/objectref/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

// openebsCluster exposes utilities w.r.t
// OpenebsCluster API object
type openebsCluster struct {
	meta   v1.Object
	object *apis.OpenebsCluster
}

type buildOption func(*openebsCluster)

// WithObject sets the provided OpenebsCluster object
// instance against this openebs cluster utility
func WithObject(obj *apis.OpenebsCluster) buildOption {
	return func(c *openebsCluster) {
		c.object = obj
	}
}

// WithMeta sets the provided meta instance
// against this openebs cluster utility
func WithMeta(meta v1.Object) buildOption {
	return func(c *openebsCluster) {
		c.meta = meta
	}
}

// New returns a new instance of openebs cluster
func New(opts ...buildOption) *openebsCluster {
	c := &openebsCluster{}
	for _, o := range opts {
		o(c)
	}
	c.withDefaults()
	return c
}

func (c *openebsCluster) withDefaults() {
	if c.meta == nil && c.object != nil {
		c.meta = &c.object.ObjectMeta
	}
}

// GetControllerRef returns the controller referred
// to by this openebs cluster instance
func (c *openebsCluster) GetControllerRef() (objectrefapis.ControllerRef, error) {
	anns := c.meta.GetAnnotations()
	if len(anns) == 0 {
		return objectrefapis.ControllerRef{}, errors.New("failed to get openebs cluster's controller: empty annotations")
	}
	return objectref.Builder().
		WithReference(anns[string(objectref.ControllerKey)]).
		ControllerRef()
}

// IsSpecificationsChange returns true if there
// is a change in specifications between the two
// openebs cluster instances
func IsSpecificationsChange(oldc *openebsCluster, newc *openebsCluster) bool {
	return reflect.DeepEqual(oldc.object.Spec, newc.object.Spec)
}

// IsAnnotationsChange returns true if there
// is a change in annotations between the two
// openebs cluster instances
func IsAnnotationsChange(oldc *openebsCluster, newc *openebsCluster) bool {
	if len(oldc.meta.GetAnnotations()) != len(newc.meta.GetAnnotations()) {
		return true
	}
	return reflect.DeepEqual(oldc.meta.GetAnnotations(), newc.meta.GetAnnotations())
}

// IsLabelsChange returns true if there
// is a change in labels between the two
// openebs cluster instances
func IsLabelsChange(oldc *openebsCluster, newc *openebsCluster) bool {
	if len(oldc.meta.GetLabels()) != len(newc.meta.GetLabels()) {
		return true
	}
	return reflect.DeepEqual(oldc.meta.GetLabels(), newc.meta.GetLabels())
}
