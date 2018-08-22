/*
Copyright 2018 The OpenEBS Authors

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

// An unstructured instance is a JSON compatible structure and is compatible
// to any kubernetes resource (native as well as custom). This instance can be
// used during following kubernetes API invocations:
//
// - Create
// - Update
// - UpdateStatus
// - Delete
// - DeleteCollection
// - Get
// - List
// - Watch
// - Patch
package v1alpha1

import (
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strings"
)

// GroupVersionResourceFromGVK returns the GroupVersionResource of the provided
// unstructured instance by making use of this instance's GroupVersionKind info
//
// NOTE:
//  Resource is assumed as plural of Kind
func GroupVersionResourceFromGVK(unstructured *unstructured.Unstructured) (gvr schema.GroupVersionResource) {
	if unstructured == nil {
		return
	}

	gvk := unstructured.GroupVersionKind()

	gvr.Group = strings.ToLower(gvk.Group)
	gvr.Version = strings.ToLower(gvk.Version)
	gvr.Resource = strings.ToLower(gvk.Kind) + "s"

	return
}

// UnstructuredCreator abstracts creation of unstructured instance from the
// provided document
type UnstructuredCreator func(document string) (*unstructured.Unstructured, error)

// CreateUnstructuredFromYaml creates an unstructured instance from the
// provided YAML document
func CreateUnstructuredFromYaml(document string) (*unstructured.Unstructured, error) {
	m := map[string]interface{}{}
	err := yaml.Unmarshal([]byte(document), &m)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create unstructured instance from yaml document")
	}

	return &unstructured.Unstructured{
		Object: m,
	}, nil
}

// CreateUnstructuredFromJson creates an unstructured instance from the
// provided JSON document
func CreateUnstructuredFromJson(document string) (*unstructured.Unstructured, error) {
	uncastObj, err := runtime.Decode(unstructured.UnstructuredJSONScheme, []byte(document))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create unstructured instance from json document")
	}

	return uncastObj.(*unstructured.Unstructured), nil
}

// UnstructuredMiddleware abstracts updating given unstructured instance
type UnstructuredMiddleware func(given *unstructured.Unstructured) (updated *unstructured.Unstructured)

// UnstructuredUpdater updates an unstructured instance by executing all the
// provided updaters
func UnstructuredUpdater(updaters []UnstructuredMiddleware) UnstructuredMiddleware {
	return func(given *unstructured.Unstructured) *unstructured.Unstructured {
		for _, u := range updaters {
			given = u(given)
		}
		return given
	}
}

// ResourceCreator abstracts creating an unstructured instance in kubernetes
// cluster
type ResourceCreator func(obj *unstructured.Unstructured, subresources ...string) (*unstructured.Unstructured, error)

// NewResourceCreator returns a new instance of ResourceCreator that is
// capable of creating a resource in kubernetes cluster
func NewResourceCreator(gvr schema.GroupVersionResource, namespace string) ResourceCreator {
	return func(obj *unstructured.Unstructured, subresources ...string) (*unstructured.Unstructured, error) {
		if obj == nil {
			return nil, fmt.Errorf("nil resource instance: failed to create '%s' at namespace '%s'", gvr, namespace)
		}

		dynamic, err := NewDynamicGetter()()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create '%s' '%s' at namespace '%s'", gvr, obj.GetName(), namespace)
		}

		unstruct, err := dynamic.Resource(gvr).Namespace(namespace).Create(obj, subresources...)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create '%s' '%s' at namespace '%s'", gvr, obj.GetName(), namespace)
		}

		return unstruct, nil
	}
}

// ResourceGetter abstracts fetching an unstructured instance from kubernetes
// cluster
type ResourceGetter func(name string, options metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error)

// NewResourceGetter returns a new instance of ResourceGetter that is capable
// of fetching an unstructured instance from kubernetes cluster
func NewResourceGetter(gvr schema.GroupVersionResource, namespace string) ResourceGetter {
	return func(name string, options metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error) {
		if len(strings.TrimSpace(name)) == 0 {
			return nil, fmt.Errorf("missing resource name: failed to get '%s' from namespace '%s'", gvr, namespace)
		}

		dynamic, err := NewDynamicGetter()()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get '%s' '%s' from namespace '%s'", gvr, name, namespace)
		}

		unstruct, err := dynamic.Resource(gvr).Namespace(namespace).Get(name, options, subresources...)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get '%s' '%s' from namespace '%s'", gvr, name, namespace)
		}

		return unstruct, nil
	}
}

// ResourceUpdater abstracts updating an unstructured instance found in
// kubernetes cluster
type ResourceUpdater func(obj *unstructured.Unstructured, subresources ...string) (*unstructured.Unstructured, error)

// NewResourceUpdater returns a new instance of ResourceUpdater that is capable
// of updating an unstructured instance found in kubernetes cluster
func NewResourceUpdater(gvr schema.GroupVersionResource, namespace string) ResourceUpdater {
	return func(obj *unstructured.Unstructured, subresources ...string) (*unstructured.Unstructured, error) {
		if obj == nil {
			return nil, fmt.Errorf("nil resource instance: failed to update '%s' at namespace '%s'", gvr, namespace)
		}

		dynamic, err := NewDynamicGetter()()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to update '%s' '%s' at namespace '%s'", gvr, obj.GetName(), namespace)
		}

		unstruct, err := dynamic.Resource(gvr).Namespace(namespace).Update(obj, subresources...)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to update '%s' '%s' at namespace '%s'", gvr, obj.GetName(), namespace)
		}

		return unstruct, nil
	}
}

// ResourceApplyOptions is used during a resource's apply operation
type ResourceApplyOptions struct {
	Getter  ResourceGetter
	Creator ResourceCreator
	Updater ResourceUpdater
}

// ResourceApplier abstracts applying an unstructured instance that may or may
// not be available in kubernetes cluster
type ResourceApplier func(obj *unstructured.Unstructured, subresources ...string) (*unstructured.Unstructured, error)

// newResourceApplier returns a new instance of ResourceApplier that is capable
// of applying an unstructured instance that may or may not be available in
// kubernetes cluster
func newResourceApplier(options ResourceApplyOptions) ResourceApplier {
	return func(obj *unstructured.Unstructured, subresources ...string) (resource *unstructured.Unstructured, err error) {
		if options.Getter == nil {
			err = fmt.Errorf("nil resource getter instance: failed to apply resource")
			return
		}

		if options.Creator == nil {
			err = fmt.Errorf("nil resource creator instance: failed to apply resource")
			return
		}

		if options.Updater == nil {
			err = fmt.Errorf("nil resource updater instance: failed to apply resource")
			return
		}

		if obj == nil {
			err = fmt.Errorf("nil resource instance: failed to apply resource")
			return
		}

		resource, err = options.Getter(obj.GetName(), metav1.GetOptions{})
		if apierrors.IsNotFound(errors.Cause(err)) {
			// create if not found
			return options.Creator(obj, subresources...)
		} else {
			// some genuine error at kubernetes server
			err = errors.Wrapf(err, "failed to apply resource '%s'", obj.GetName())
			return
		}

		return options.Updater(obj, subresources...)
	}
}

// NewResourceApplier returns a new instance of ResourceApplier that is capable
// of applying any resource into kubernetes cluster
func NewResourceApplier(gvr schema.GroupVersionResource, namespace string) ResourceApplier {
	options := ResourceApplyOptions{
		Getter:  NewResourceGetter(gvr, namespace),
		Creator: NewResourceCreator(gvr, namespace),
		Updater: NewResourceUpdater(gvr, namespace),
	}

	return newResourceApplier(options)
}
