/*
Copyright 2017 The OpenEBS Authors

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
	v1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	scheme "github.com/openebs/maya/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// CstorPoolsGetter has a method to return a CstorPoolInterface.
// A group's client should implement this interface.
type CstorPoolsGetter interface {
	CstorPools() CstorPoolInterface
}

// CstorPoolInterface has methods to work with CstorPool resources.
type CstorPoolInterface interface {
	Create(*v1alpha1.CstorPool) (*v1alpha1.CstorPool, error)
	Update(*v1alpha1.CstorPool) (*v1alpha1.CstorPool, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.CstorPool, error)
	List(opts v1.ListOptions) (*v1alpha1.CstorPoolList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.CstorPool, err error)
	CstorPoolExpansion
}

// cstorPools implements CstorPoolInterface
type cstorPools struct {
	client rest.Interface
}

// newCstorPools returns a CstorPools
func newCstorPools(c *OpenebsV1alpha1Client) *cstorPools {
	return &cstorPools{
		client: c.RESTClient(),
	}
}

// Get takes name of the cstorPool, and returns the corresponding cstorPool object, and an error if there is any.
func (c *cstorPools) Get(name string, options v1.GetOptions) (result *v1alpha1.CstorPool, err error) {
	result = &v1alpha1.CstorPool{}
	err = c.client.Get().
		Resource("cstorpools").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of CstorPools that match those selectors.
func (c *cstorPools) List(opts v1.ListOptions) (result *v1alpha1.CstorPoolList, err error) {
	result = &v1alpha1.CstorPoolList{}
	err = c.client.Get().
		Resource("cstorpools").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested cstorPools.
func (c *cstorPools) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Resource("cstorpools").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a cstorPool and creates it.  Returns the server's representation of the cstorPool, and an error, if there is any.
func (c *cstorPools) Create(cstorPool *v1alpha1.CstorPool) (result *v1alpha1.CstorPool, err error) {
	result = &v1alpha1.CstorPool{}
	err = c.client.Post().
		Resource("cstorpools").
		Body(cstorPool).
		Do().
		Into(result)
	return
}

// Update takes the representation of a cstorPool and updates it. Returns the server's representation of the cstorPool, and an error, if there is any.
func (c *cstorPools) Update(cstorPool *v1alpha1.CstorPool) (result *v1alpha1.CstorPool, err error) {
	result = &v1alpha1.CstorPool{}
	err = c.client.Put().
		Resource("cstorpools").
		Name(cstorPool.Name).
		Body(cstorPool).
		Do().
		Into(result)
	return
}

// Delete takes name of the cstorPool and deletes it. Returns an error if one occurs.
func (c *cstorPools) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Resource("cstorpools").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *cstorPools) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Resource("cstorpools").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched cstorPool.
func (c *cstorPools) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.CstorPool, err error) {
	result = &v1alpha1.CstorPool{}
	err = c.client.Patch(pt).
		Resource("cstorpools").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
