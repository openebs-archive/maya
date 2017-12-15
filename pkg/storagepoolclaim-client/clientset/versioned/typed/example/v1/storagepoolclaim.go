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
package v1

import (
	v1 "github.com/openebs/maya/pkg/storagepoolclaim-apis/openebs.io/v1"
	scheme "github.com/openebs/maya/pkg/storagepoolclaim-client/clientset/versioned/scheme"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// StoragepoolclaimsGetter has a method to return a StoragepoolclaimInterface.
// A group's client should implement this interface.
type StoragepoolclaimsGetter interface {
	Storagepoolclaims(namespace string) StoragepoolclaimInterface
}

// StoragepoolclaimInterface has methods to work with Storagepoolclaim resources.
type StoragepoolclaimInterface interface {
	Create(*v1.Storagepoolclaim) (*v1.Storagepoolclaim, error)
	Update(*v1.Storagepoolclaim) (*v1.Storagepoolclaim, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error
	Get(name string, options meta_v1.GetOptions) (*v1.Storagepoolclaim, error)
	List(opts meta_v1.ListOptions) (*v1.StoragepoolclaimList, error)
	Watch(opts meta_v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Storagepoolclaim, err error)
	StoragepoolclaimExpansion
}

// storagepoolclaims implements StoragepoolclaimInterface
type storagepoolclaims struct {
	client rest.Interface
	ns     string
}

// newStoragepoolclaims returns a Storagepoolclaims
func newStoragepoolclaims(c *ExampleV1Client, namespace string) *storagepoolclaims {
	return &storagepoolclaims{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the storagepoolclaim, and returns the corresponding storagepoolclaim object, and an error if there is any.
func (c *storagepoolclaims) Get(name string, options meta_v1.GetOptions) (result *v1.Storagepoolclaim, err error) {
	result = &v1.Storagepoolclaim{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("storagepoolclaims").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Storagepoolclaims that match those selectors.
func (c *storagepoolclaims) List(opts meta_v1.ListOptions) (result *v1.StoragepoolclaimList, err error) {
	result = &v1.StoragepoolclaimList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("storagepoolclaims").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested storagepoolclaims.
func (c *storagepoolclaims) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("storagepoolclaims").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a storagepoolclaim and creates it.  Returns the server's representation of the storagepoolclaim, and an error, if there is any.
func (c *storagepoolclaims) Create(storagepoolclaim *v1.Storagepoolclaim) (result *v1.Storagepoolclaim, err error) {
	result = &v1.Storagepoolclaim{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("storagepoolclaims").
		Body(storagepoolclaim).
		Do().
		Into(result)
	return
}

// Update takes the representation of a storagepoolclaim and updates it. Returns the server's representation of the storagepoolclaim, and an error, if there is any.
func (c *storagepoolclaims) Update(storagepoolclaim *v1.Storagepoolclaim) (result *v1.Storagepoolclaim, err error) {
	result = &v1.Storagepoolclaim{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("storagepoolclaims").
		Name(storagepoolclaim.Name).
		Body(storagepoolclaim).
		Do().
		Into(result)
	return
}

// Delete takes name of the storagepoolclaim and deletes it. Returns an error if one occurs.
func (c *storagepoolclaims) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("storagepoolclaims").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *storagepoolclaims) DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("storagepoolclaims").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched storagepoolclaim.
func (c *storagepoolclaims) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Storagepoolclaim, err error) {
	result = &v1.Storagepoolclaim{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("storagepoolclaims").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
