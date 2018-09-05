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
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strings"
)

// kind represents the name of the kubernetes kind
type kind string

// resource converts the name of the kubernetes kind to corresponding kubernetes
// resource's name
//
// NOTE:
//  This may not be the best of approaches to get name of a resource. However,
// this fits the current requirement. This might need a revisit depending on
// future requirements.
func (k kind) resource() (resource string) {
	resource = strings.ToLower(string(k))
	switch resource {
	case "":
		return
	case "storageclass":
		return "storageclasses"
	default:
		return resource + "s"
	}
	return
}

// isNamespaced flags if the kind is namespaced or not
//
// NOTE:
//  This may not be the best of approaches to flag a resource as namespaced or
// not. However, this fits the current requirement. This might need a revisit
// depending on future requirements.
func (k kind) isNamespaced() (no bool) {
	ks := strings.ToLower(string(k))
	switch ks {
	case "customresourcedefinition":
		return no
	case "storageclass":
		return no
	case "persistentvolume":
		return no
	case "castemplate":
		return no
	case "storagepoolclaim":
		return no
	case "cstorpool":
		return no
	case "storagepool":
		return no
	default:
		return !no
	}
	return !no
}

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
	gvr.Resource = kind(gvk.Kind).resource()

	return
}

// WithBytesUnstructuredCreator abstracts creation of unstructured instance
// from the provided bytes
type WithBytesUnstructuredCreator func(raw []byte) (*unstructured.Unstructured, error)

// CreateUnstructuredFromYamlBytes creates an unstructured instance from the
// provided YAML document in bytes
//
// NOTE:
//  This is an implementation of WithBytesUnstructuredCreator
func CreateUnstructuredFromYamlBytes(raw []byte) (*unstructured.Unstructured, error) {
	m := map[string]interface{}{}
	err := yaml.Unmarshal(raw, &m)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create unstructured instance from yaml bytes")
	}

	return &unstructured.Unstructured{
		Object: m,
	}, nil
}

// UnstructuredCreator abstracts creation of unstructured instance from the
// provided document
type UnstructuredCreator func(document string) (*unstructured.Unstructured, error)

// CreateUnstructuredFromYaml creates an unstructured instance from the
// provided YAML document
func CreateUnstructuredFromYaml(document string) (*unstructured.Unstructured, error) {
	return CreateUnstructuredFromYamlBytes([]byte(document))
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

// UnstructuredPredicate abstracts evaluating a condition against the provided
// unstructured instance
type UnstructuredPredicate func(given *unstructured.Unstructured) bool

// IsNamespaceScoped flags if the given unstructured instance is namespace
// scoped
//
// NOTE:
//  This is a UnstructuredPredicate implementation
func IsNamespaceScoped(given *unstructured.Unstructured) bool {
	return kind(given.GetKind()).isNamespaced()
}

// UnstructuredOptions provides a set of properties that can be used as a
// utility for various operations related to unstructured instance
type UnstructuredOptions struct {
	Namespace string
	Labels    map[string]string
}

// UpdateNamespaceP updates the unstructured's namespace conditionally
func UpdateNamespaceP(o UnstructuredOptions, p UnstructuredPredicate) UnstructuredMiddleware {
	return func(given *unstructured.Unstructured) (updated *unstructured.Unstructured) {
		if p(given) {
			return UpdateNamespace(o)(given)
		}
		return given
	}
}

// UpdateNamespace updates the unstructured's namespace
func UpdateNamespace(o UnstructuredOptions) UnstructuredMiddleware {
	return func(given *unstructured.Unstructured) (updated *unstructured.Unstructured) {
		if given == nil {
			return given
		}
		namespace := strings.TrimSpace(o.Namespace)
		if len(namespace) == 0 {
			return given
		}
		given.SetNamespace(namespace)
		return given
	}
}

// UpdateLabels updates the unstructured's labels
func UpdateLabels(o UnstructuredOptions) UnstructuredMiddleware {
	return func(given *unstructured.Unstructured) (updated *unstructured.Unstructured) {
		if given == nil {
			return given
		}
		if len(o.Labels) == 0 {
			return given
		}
		orig := given.GetLabels()
		if orig == nil {
			orig = map[string]string{}
		}
		for k, v := range o.Labels {
			orig[k] = v
		}
		given.SetLabels(orig)
		return given
	}
}

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

// UnstructList represents a list of unstructured instances
type UnstructList struct {
	Items []*unstructured.Unstructured
}

// MapAll will execute the all UnstructuredMiddlewares on each unstructured
// instance
func (u UnstructList) MapAll(ml []UnstructuredMiddleware) (ul UnstructList) {
	for _, unstruct := range u.Items {
		for _, m := range ml {
			unstruct = m(unstruct)
		}
		if unstruct != nil {
			ul.Items = append(ul.Items, unstruct)
		}
	}
	return
}

// MapIf will execute the UnstructuredMiddleware conditionally based on
// UnstructuredPredicate
func (u UnstructList) MapIf(m UnstructuredMiddleware, p UnstructuredPredicate) (ul UnstructList) {
	for _, unstruct := range u.Items {
		if p(unstruct) {
			unstruct = m(unstruct)
		}
		if unstruct != nil {
			ul.Items = append(ul.Items, unstruct)
		}
	}
	return
}
