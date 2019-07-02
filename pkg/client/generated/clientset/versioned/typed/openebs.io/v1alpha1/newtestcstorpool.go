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

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	scheme "github.com/openebs/maya/pkg/client/generated/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// NewTestCStorPoolsGetter has a method to return a NewTestCStorPoolInterface.
// A group's client should implement this interface.
type NewTestCStorPoolsGetter interface {
	NewTestCStorPools(namespace string) NewTestCStorPoolInterface
}

// NewTestCStorPoolInterface has methods to work with NewTestCStorPool resources.
type NewTestCStorPoolInterface interface {
	Create(*v1alpha1.NewTestCStorPool) (*v1alpha1.NewTestCStorPool, error)
	Update(*v1alpha1.NewTestCStorPool) (*v1alpha1.NewTestCStorPool, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.NewTestCStorPool, error)
	List(opts v1.ListOptions) (*v1alpha1.NewTestCStorPoolList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.NewTestCStorPool, err error)
	NewTestCStorPoolExpansion
}

// newTestCStorPools implements NewTestCStorPoolInterface
type newTestCStorPools struct {
	client rest.Interface
	ns     string
}

// newNewTestCStorPools returns a NewTestCStorPools
func newNewTestCStorPools(c *OpenebsV1alpha1Client, namespace string) *newTestCStorPools {
	return &newTestCStorPools{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the newTestCStorPool, and returns the corresponding newTestCStorPool object, and an error if there is any.
func (c *newTestCStorPools) Get(name string, options v1.GetOptions) (result *v1alpha1.NewTestCStorPool, err error) {
	result = &v1alpha1.NewTestCStorPool{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("newtestcstorpools").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of NewTestCStorPools that match those selectors.
func (c *newTestCStorPools) List(opts v1.ListOptions) (result *v1alpha1.NewTestCStorPoolList, err error) {
	result = &v1alpha1.NewTestCStorPoolList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("newtestcstorpools").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested newTestCStorPools.
func (c *newTestCStorPools) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("newtestcstorpools").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a newTestCStorPool and creates it.  Returns the server's representation of the newTestCStorPool, and an error, if there is any.
func (c *newTestCStorPools) Create(newTestCStorPool *v1alpha1.NewTestCStorPool) (result *v1alpha1.NewTestCStorPool, err error) {
	result = &v1alpha1.NewTestCStorPool{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("newtestcstorpools").
		Body(newTestCStorPool).
		Do().
		Into(result)
	return
}

// Update takes the representation of a newTestCStorPool and updates it. Returns the server's representation of the newTestCStorPool, and an error, if there is any.
func (c *newTestCStorPools) Update(newTestCStorPool *v1alpha1.NewTestCStorPool) (result *v1alpha1.NewTestCStorPool, err error) {
	result = &v1alpha1.NewTestCStorPool{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("newtestcstorpools").
		Name(newTestCStorPool.Name).
		Body(newTestCStorPool).
		Do().
		Into(result)
	return
}

// Delete takes name of the newTestCStorPool and deletes it. Returns an error if one occurs.
func (c *newTestCStorPools) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("newtestcstorpools").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *newTestCStorPools) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("newtestcstorpools").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched newTestCStorPool.
func (c *newTestCStorPools) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.NewTestCStorPool, err error) {
	result = &v1alpha1.NewTestCStorPool{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("newtestcstorpools").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
