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

package v1alpha2

import (
	"strings"

	clientset "github.com/openebs/maya/pkg/client/k8s/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8sdynamic "k8s.io/client-go/dynamic"
)

// getClientsetFn is a typed function that
// abstracts fetching of clientset
type getClientsetFn func() (clientset k8sdynamic.Interface, err error)

// CreateFn is a typed function that abstracts
// creating of unstructured object
type CreateFn func(cli k8sdynamic.Interface, obj *unstructured.Unstructured, opts metav1.CreateOptions, subresources ...string) (*unstructured.Unstructured, error)

// GetFn is a typed function that abstracts
// fetching of unstructured object
type GetFn func(cli k8sdynamic.Interface, name string, namespace string, gvr schema.GroupVersionResource, opts metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error)

// DeleteFn is a typed function that abstract  deletion
// of unstructured object
type DeleteFn func(cli k8sdynamic.Interface, obj *unstructured.Unstructured, opts *metav1.DeleteOptions, subresources ...string) error

// Kubeclient enables kubernetes API operations on catalog
// instance
type Kubeclient struct {
	// clientset refers to clientset
	// that will be responsible to
	// make kubernetes API calls
	clientset k8sdynamic.Interface

	// functions useful during mocking
	getClientset getClientsetFn
	create       CreateFn
	get          GetFn
	delete       DeleteFn
}

// KubeclientBuildOption defines the abstraction to build
// a Kubeclient instance
type KubeclientBuildOption func(*Kubeclient)

// withDefaults sets default options for Kubeclient
func withDefaults(k *Kubeclient) error {
	if k.clientset == nil {
		cli, err := clientset.Dynamic().Provide()
		if err != nil {
			return err
		}
		k.clientset = cli
	}
	if k.get == nil {
		k.get = func(cli k8sdynamic.Interface, name string, namespace string, gvr schema.GroupVersionResource, opts metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error) {
			u, err := cli.Resource(gvr).Namespace(namespace).Get(name, opts, subresources...)
			if err != nil {
				return nil, err
			}
			return u, nil
		}
	}
	if k.create == nil {
		k.create = func(cli k8sdynamic.Interface, obj *unstructured.Unstructured, opts metav1.CreateOptions, subresources ...string) (*unstructured.Unstructured, error) {
			return cli.Resource(clientset.GroupVersionResourceFromGVK(obj)).Namespace(obj.GetNamespace()).Create(obj, opts, subresources...)
		}
	}
	if k.delete == nil {
		k.delete = func(cli k8sdynamic.Interface, obj *unstructured.Unstructured, opts *metav1.DeleteOptions, subresources ...string) error {
			return cli.Resource(clientset.GroupVersionResourceFromGVK(obj)).Namespace(obj.GetNamespace()).Delete(obj.GetName(), opts, subresources...)
		}
	}
	return nil
}

// WithClient sets the kubernetes client against
// the Kubeclient instance
func WithClient(c k8sdynamic.Interface) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.clientset = c
	}
}

// KubeClient returns a new instance of Kubeclient meant for
// catalog operations
func KubeClient(opts ...KubeclientBuildOption) (*Kubeclient, error) {
	k := &Kubeclient{}
	for _, o := range opts {
		o(k)
	}
	err := withDefaults(k)
	if err != nil {
		return nil, err
	}
	return k, nil
}

// getClientOrCached returns either a new instance
// of kubernetes client or its cached copy
func (k *Kubeclient) getClientOrCached() (k8sdynamic.Interface, error) {
	if k.clientset != nil {
		return k.clientset, nil
	}
	cli, err := k.getClientset()
	if err != nil {
		return nil, err
	}
	k.clientset = cli
	return k.clientset, nil
}

// Get returns an unstructured instance from kubernetes
// cluster
func (k *Kubeclient) Get(name string, gOpts ...GetOptionFn) (*unstructured.Unstructured, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("failed to get unstructured instance: missing name")
	}
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	getOptions := NewGetOption(gOpts...)
	return k.get(cli, name, getOptions.namespace, getOptions.grv, getOptions.getOption, getOptions.subresources...)
}

// CreateAllOrNone creates all the provided
// unstructured instances at kubernetes cluster
// or none in case of any error
func (k *Kubeclient) CreateAllOrNone(u ...*unstructured.Unstructured) []error {
	var (
		errs    []error
		created []*unstructured.Unstructured
	)
	for _, ustruct := range u {
		err := k.Create(ustruct)
		if err != nil {
			errs = append(errs, err)
			break
		}
		created = append(created, ustruct)
	}
	if len(errs) > 0 {
		k.DeleteAll(created...)
	}
	return errs
}

// DeleteAll deletes all the provided unstructured
// instances at kubernetes cluster
func (k *Kubeclient) DeleteAll(u ...*unstructured.Unstructured) []error {
	var errs []error
	for _, ustruct := range u {
		err := k.Delete(ustruct)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// Create creates an unstructured instance at
// kubernetes cluster
func (k *Kubeclient) Create(u *unstructured.Unstructured, cOpts ...CreateOptionFn) error {
	if u == nil {
		return errors.Errorf("create failed: nil unstruct instance was provided")
	}
	cli, err := k.getClientOrCached()
	if err != nil {
		return err
	}
	opt := NewCreateOption(cOpts...)
	_, err = k.create(cli, u, opt.createOptions, opt.subresources...)
	return err
}

// Delete deletes the unstructured instance from
// kubernetes cluster
func (k *Kubeclient) Delete(u *unstructured.Unstructured, dOpts ...DeleteOptionFn) error {
	cli, err := k.getClientOrCached()
	if err != nil {
		return err
	}
	opts := NewDeleteOption(dOpts...)
	return k.delete(cli, u, opts.deleteOptions)
}

// GetOption holds the kubernetes options
// to get a resource
type GetOption struct {
	namespace    string
	getOption    metav1.GetOptions
	grv          schema.GroupVersionResource
	subresources []string
}

// GetOptionFn abstracts the construction of GetOption
type GetOptionFn func(*GetOption)

// WithGetNamespace is a GetOptionFn to provide
// namespace
func WithGetNamespace(namespace string) GetOptionFn {
	return func(opt *GetOption) {
		opt.namespace = namespace
	}
}

// WithGetOption is a GetOptionsFn to provide
// kubernetes getoption
func WithGetOption(getOption metav1.GetOptions) GetOptionFn {
	return func(opt *GetOption) {
		opt.getOption = getOption
	}
}

// WithGroupVersionResource is a GetOptionFn to provide
// GroupResourceVersion
func WithGroupVersionResource(r schema.GroupVersionResource) GetOptionFn {
	return func(opt *GetOption) {
		opt.grv = r
	}
}

// WithGetSubResources is a GetOptionFn to provide
// subresources
func WithGetSubResources(r ...string) GetOptionFn {
	return func(opt *GetOption) {
		opt.subresources = r
	}
}

// NewGetOption returns a new instance of GetOption
func NewGetOption(gOpts ...GetOptionFn) *GetOption {
	opts := &GetOption{getOption: metav1.GetOptions{}, grv: schema.GroupVersionResource{}}
	for _, o := range gOpts {
		o(opts)
	}
	return opts
}

// DeleteOption holds kubernetes options to delete a
// resource
type DeleteOption struct {
	deleteOptions *metav1.DeleteOptions
	subresources  []string
}

// NewDeleteOption returns a new instance of
// DeleteOption
func NewDeleteOption(dOpts ...DeleteOptionFn) *DeleteOption {
	opts := &DeleteOption{deleteOptions: &metav1.DeleteOptions{}}
	for _, o := range dOpts {
		o(opts)
	}
	return opts
}

// DeleteOptionFn abstracts the construvtion of delete
// option
type DeleteOptionFn func(*DeleteOption)

// WithDeleteOption is a DeleteOptionFn to provide
// kubernetes delete option
func WithDeleteOption(deleteOpt *metav1.DeleteOptions) DeleteOptionFn {
	return func(opt *DeleteOption) {
		opt.deleteOptions = deleteOpt
	}
}

// WithDeleteSubResources is a DeleteOptionFn to provide
// subresources during delete
func WithDeleteSubResources(r ...string) DeleteOptionFn {
	return func(opt *DeleteOption) {
		opt.subresources = r
	}
}

// CreateOption holds the kubernetes option to create a resource
type CreateOption struct {
	createOptions metav1.CreateOptions
	subresources  []string
}

// NewCreateOption returns a new instance of CreateOption
func NewCreateOption(cOpts ...CreateOptionFn) *CreateOption {
	opts := &CreateOption{metav1.CreateOptions{}, []string{}}
	for _, o := range cOpts {
		o(opts)
	}
	return opts
}

// CreateOptionFn abstracts the construction of CreateOption
type CreateOptionFn func(*CreateOption)

// WithCreateOption is CreateOptionFn to provide kubernetes
// createOption for creating a resource
func WithCreateOption(opt metav1.CreateOptions) CreateOptionFn {
	return func(createOpt *CreateOption) {
		createOpt.createOptions = opt
	}
}

// WithCreateSubResources is CreateOptionFn to kubernetes
// subresources during resource creation
func WithCreateSubResources(r ...string) CreateOptionFn {
	return func(createOpt *CreateOption) {
		createOpt.subresources = r
	}
}
