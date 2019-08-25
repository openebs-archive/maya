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

// TODO
// Move this file to pkg/k8sresource/v1alpha1

package v1alpha1

import (
	"strings"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ResourceCreator abstracts creating an unstructured instance in kubernetes
// cluster
type ResourceCreator interface {
	Create(obj *unstructured.Unstructured, subresources ...string) (*unstructured.Unstructured, error)
}

// ResourceGetter abstracts fetching an unstructured instance from kubernetes
// cluster
type ResourceGetter interface {
	Get(name string, options metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error)
}

// ResourceLister abstracts fetching an unstructured list of instance from kubernetes
// cluster
type ResourceLister interface {
	List(options metav1.ListOptions) (*unstructured.UnstructuredList, error)
}

// ResourceUpdater abstracts updating an unstructured instance found in
// kubernetes cluster
type ResourceUpdater interface {
	Update(oldobj, newobj *unstructured.Unstructured, subresources ...string) (u *unstructured.Unstructured, err error)
}

// ResourceApplier abstracts applying an unstructured instance that may or may
// not be available in kubernetes cluster
type ResourceApplier interface {
	Apply(obj *unstructured.Unstructured, subresources ...string) (*unstructured.Unstructured, error)
}

// ResourceDeleter abstracts deletes an unstructured instance that is available in kubernetes cluster
type ResourceDeleter interface {
	Delete(obj *unstructured.Unstructured, subresources ...string) error
}

type resource struct {
	gvr       schema.GroupVersionResource // identify a resource
	namespace string                      // namespace where this resource is to be operated at
}

// Resource returns a new resource instance
func Resource(gvr schema.GroupVersionResource, namespace string) *resource {
	return &resource{gvr: gvr, namespace: namespace}
}

// Create creates a new resource in kubernetes cluster
func (r *resource) Create(obj *unstructured.Unstructured, subresources ...string) (u *unstructured.Unstructured, err error) {
	if obj == nil {
		err = errors.Errorf("nil resource instance: failed to create resource '%s' at '%s'", r.gvr, r.namespace)
		return
	}
	dynamic, err := Dynamic().Provide()
	if err != nil {
		err = errors.Wrapf(err, "failed to create resource '%s' '%s' at '%s'", r.gvr, obj.GetName(), r.namespace)
		return
	}
	u, err = dynamic.Resource(r.gvr).Namespace(r.namespace).Create(obj, metav1.CreateOptions{}, subresources...)
	if err != nil {
		err = errors.Wrapf(err, "failed to create resource '%s' '%s' at '%s'", r.gvr, obj.GetName(), r.namespace)
		return
	}
	return
}

// Delete deletes a existing resource in kubernetes cluster
func (r *resource) Delete(obj *unstructured.Unstructured, subresources ...string) error {
	if obj == nil {
		return errors.Errorf("nil resource instance: failed to delete resource '%s' at '%s'", r.gvr, r.namespace)
	}
	dynamic, err := Dynamic().Provide()
	if err != nil {
		return errors.Wrapf(err, "failed to delete resource '%s' '%s' at '%s'", r.gvr, obj.GetName(), r.namespace)
	}
	err = dynamic.Resource(r.gvr).Namespace(r.namespace).Delete(obj.GetName(), &metav1.DeleteOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to delete resource '%s' '%s' at '%s'", r.gvr, obj.GetName(), r.namespace)
	}
	return nil
}

// Get returns a specific resource from kubernetes cluster
func (r *resource) Get(name string, opts metav1.GetOptions, subresources ...string) (u *unstructured.Unstructured, err error) {
	if len(strings.TrimSpace(name)) == 0 {
		err = errors.Errorf("missing resource name: failed to get resource '%s' at '%s'", r.gvr, r.namespace)
		return
	}
	dynamic, err := Dynamic().Provide()
	if err != nil {
		err = errors.Wrapf(err, "failed to get resource '%s' '%s' at '%s'", r.gvr, name, r.namespace)
		return
	}
	u, err = dynamic.Resource(r.gvr).Namespace(r.namespace).Get(name, opts, subresources...)
	if err != nil {
		err = errors.Wrapf(err, "failed to get resource '%s' '%s' at '%s'", r.gvr, name, r.namespace)
		return
	}
	return
}

// Update updates the resource at kubernetes cluster
func (r *resource) Update(oldobj, newobj *unstructured.Unstructured, subresources ...string) (u *unstructured.Unstructured, err error) {
	if oldobj == nil {
		err = errors.Errorf("nil old resource instance: failed to update resource '%s' at '%s'", r.gvr, r.namespace)
		return
	}
	if newobj == nil {
		err = errors.Errorf("nil new resource instance: failed to update resource '%s' at '%s'", r.gvr, r.namespace)
		return
	}
	dynamic, err := Dynamic().Provide()
	if err != nil {
		err = errors.Wrapf(err, "failed to update resource '%s' '%s' at '%s'", r.gvr, oldobj.GetName(), r.namespace)
		return
	}

	resourceVersion := oldobj.GetResourceVersion()
	newobj.SetResourceVersion(resourceVersion)

	u, err = dynamic.Resource(r.gvr).Namespace(r.namespace).Update(newobj, metav1.UpdateOptions{}, subresources...)
	if err != nil {
		err = errors.Wrapf(err, "failed to update resource '%s' '%s' at '%s'", r.gvr, oldobj.GetName(), r.namespace)
		return
	}
	return
}

// List returns a list of specific resource at kubernetes cluster
func (r *resource) List(opts metav1.ListOptions) (u *unstructured.UnstructuredList, err error) {
	dynamic, err := Dynamic().Provide()
	if err != nil {
		err = errors.Wrapf(err, "failed to list resource '%s'  at '%s'", r.gvr, r.namespace)
		return
	}
	u, err = dynamic.Resource(r.gvr).Namespace(r.namespace).List(opts)
	if err != nil {
		err = errors.Wrapf(err, "failed to list resource '%s'  at '%s'", r.gvr, r.namespace)
		return
	}
	return
}

// ResourceApplyOptions is a utility instance used during the resource's apply
// operation
type ResourceApplyOptions struct {
	Getter  ResourceGetter
	Creator ResourceCreator
	Updater ResourceUpdater
}

// createOrUpdate is a resource that is suitable to be executed as an apply
// operation
type createOrUpdate struct {
	*resource
	options ResourceApplyOptions // options used during resource's apply operation
}

// CreateOrUpdate returns a new instance of createOrUpdate resource
func CreateOrUpdate(gvr schema.GroupVersionResource, namespace string) *createOrUpdate {
	resource := Resource(gvr, namespace)
	options := ResourceApplyOptions{Getter: resource, Creator: resource, Updater: resource}
	return &createOrUpdate{resource: resource, options: options}
}

// Apply applies a resource to the kubernetes cluster. In other words, it
// creates a new resource if it does not exist or updates the existing resource.
func (r *createOrUpdate) Apply(obj *unstructured.Unstructured, subresources ...string) (resource *unstructured.Unstructured, err error) {
	if r.options.Getter == nil {
		err = errors.New("nil resource getter instance: failed to apply resource")
		return
	}
	if r.options.Creator == nil {
		err = errors.New("nil resource creator instance: failed to apply resource")
		return
	}
	if r.options.Updater == nil {
		err = errors.New("nil resource updater instance: failed to apply resource")
		return
	}
	if obj == nil {
		err = errors.New("nil resource instance: failed to apply resource")
		return
	}
	resource, err = r.options.Getter.Get(obj.GetName(), metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(errors.Cause(err)) {
			return r.options.Creator.Create(obj, subresources...)
		}
		return nil, err
	}
	// NOTE: We are returning same unstructured object in case if spc and sc are
	// already installed
	if obj.GetKind() == "StoragePoolClaim" || obj.GetKind() == "StorageClass" {
		glog.Infof(
			"{%s/%s} already exist in cluster",
			obj.GroupVersionKind(),
			obj.GetName(),
		)
		resource, err = obj, nil
		return
	}
	return r.options.Updater.Update(resource, obj, subresources...)
}

// ResourceDeleteOptions is a utility instance used during the resource's delete operations
type ResourceDeleteOptions struct {
	Deleter ResourceDeleter
}

// Delete is a resource that is suitable to be executed as a Delete operation
type Delete struct {
	*resource
	options ResourceDeleteOptions
}

// DeleteResource returns a new instance of delete resource
func DeleteResource(gvr schema.GroupVersionResource, namespace string) *Delete {
	resource := Resource(gvr, namespace)
	options := ResourceDeleteOptions{Deleter: resource}
	return &Delete{resource: resource, options: options}
}

// Delete deletes a resource from a kubernetes cluster
func (d *Delete) Delete(obj *unstructured.Unstructured, subresources ...string) error {
	if d.options.Deleter == nil {
		return errors.New("nil resource deleter instance: failed to delete resource")
	} else if obj == nil {
		return errors.New("nil resource instance: failed to delete resource")
	}
	return d.options.Deleter.Delete(obj, subresources...)
}

// ResourceListOptions is a utility instance used during the resource's list operations
type ResourceListOptions struct {
	Lister ResourceLister
}

// List is a resource resource that is suitable to be executed as a List operation
type List struct {
	*resource
	options ResourceListOptions
}

// ListResource returns a new instance of list resource
func ListResource(gvr schema.GroupVersionResource, namespace string) *List {
	resource := Resource(gvr, namespace)
	options := ResourceListOptions{Lister: resource}
	return &List{resource: resource, options: options}
}

// List lists a resource from a kubernetes cluster
func (l *List) List(options metav1.ListOptions) (u *unstructured.UnstructuredList, err error) {
	if l.options.Lister == nil {
		err = errors.New("nil resource lister instance: failed to list resource")
		return
	}
	return l.options.Lister.List(options)
}

// ResourceGetOptions is a utility instance used during the resource's get operations
type ResourceGetOptions struct {
	Getter ResourceGetter
}

// Get is resource that is suitable to be executed as Get operation
type Get struct {
	*resource
	options ResourceGetOptions
}

// GetResource returns a new instance of get resource
func GetResource(gvr schema.GroupVersionResource, namespace string) *Get {
	resource := Resource(gvr, namespace)
	options := ResourceGetOptions{Getter: resource}
	return &Get{resource: resource, options: options}
}

// Get gets a resource from a kubernetes cluster
func (g *Get) Get(name string, opts metav1.GetOptions, subresources ...string) (u *unstructured.Unstructured, err error) {
	if g.options.Getter == nil {
		err = errors.New("nil resource getter instance: failed to get resource")
		return
	}
	return g.options.Getter.Get(name, opts, subresources...)
}
