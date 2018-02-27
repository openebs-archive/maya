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

// CstorCrdsGetter has a method to return a CstorCrdInterface.
// A group's client should implement this interface.
type CstorCrdsGetter interface {
	CstorCrds(namespace string) CstorCrdInterface
}

// CstorCrdInterface has methods to work with CstorCrd resources.
type CstorCrdInterface interface {
	Create(*v1alpha1.CstorCrd) (*v1alpha1.CstorCrd, error)
	Update(*v1alpha1.CstorCrd) (*v1alpha1.CstorCrd, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.CstorCrd, error)
	List(opts v1.ListOptions) (*v1alpha1.CstorCrdList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.CstorCrd, err error)
	CstorCrdExpansion
}

// cstorCrds implements CstorCrdInterface
type cstorCrds struct {
	client rest.Interface
	ns     string
}

// newCstorCrds returns a CstorCrds
func newCstorCrds(c *OpenebsV1alpha1Client, namespace string) *cstorCrds {
	return &cstorCrds{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the cstorCrd, and returns the corresponding cstorCrd object, and an error if there is any.
func (c *cstorCrds) Get(name string, options v1.GetOptions) (result *v1alpha1.CstorCrd, err error) {
	result = &v1alpha1.CstorCrd{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("cstorcrds").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of CstorCrds that match those selectors.
func (c *cstorCrds) List(opts v1.ListOptions) (result *v1alpha1.CstorCrdList, err error) {
	result = &v1alpha1.CstorCrdList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("cstorcrds").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested cstorCrds.
func (c *cstorCrds) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("cstorcrds").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a cstorCrd and creates it.  Returns the server's representation of the cstorCrd, and an error, if there is any.
func (c *cstorCrds) Create(cstorCrd *v1alpha1.CstorCrd) (result *v1alpha1.CstorCrd, err error) {
	result = &v1alpha1.CstorCrd{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("cstorcrds").
		Body(cstorCrd).
		Do().
		Into(result)
	return
}

// Update takes the representation of a cstorCrd and updates it. Returns the server's representation of the cstorCrd, and an error, if there is any.
func (c *cstorCrds) Update(cstorCrd *v1alpha1.CstorCrd) (result *v1alpha1.CstorCrd, err error) {
	result = &v1alpha1.CstorCrd{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("cstorcrds").
		Name(cstorCrd.Name).
		Body(cstorCrd).
		Do().
		Into(result)
	return
}

// Delete takes name of the cstorCrd and deletes it. Returns an error if one occurs.
func (c *cstorCrds) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("cstorcrds").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *cstorCrds) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("cstorcrds").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched cstorCrd.
func (c *cstorCrds) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.CstorCrd, err error) {
	result = &v1alpha1.CstorCrd{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("cstorcrds").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
