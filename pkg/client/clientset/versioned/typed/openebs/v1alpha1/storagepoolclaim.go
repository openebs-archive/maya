/*
Copyright 2018 The Kubernetes Authors.

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

// StoragePoolClaimsGetter has a method to return a StoragePoolClaimInterface.
// A group's client should implement this interface.
type StoragePoolClaimsGetter interface {
	StoragePoolClaims() StoragePoolClaimInterface
}

// StoragePoolClaimInterface has methods to work with StoragePoolClaim resources.
type StoragePoolClaimInterface interface {
	Create(*v1alpha1.StoragePoolClaim) (*v1alpha1.StoragePoolClaim, error)
	Update(*v1alpha1.StoragePoolClaim) (*v1alpha1.StoragePoolClaim, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.StoragePoolClaim, error)
	List(opts v1.ListOptions) (*v1alpha1.StoragePoolClaimList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.StoragePoolClaim, err error)
	StoragePoolClaimExpansion
}

// storagePoolClaims implements StoragePoolClaimInterface
type storagePoolClaims struct {
	client rest.Interface
}

// newStoragePoolClaims returns a StoragePoolClaims
func newStoragePoolClaims(c *OpenebsV1alpha1Client) *storagePoolClaims {
	return &storagePoolClaims{
		client: c.RESTClient(),
	}
}

// Get takes name of the storagePoolClaim, and returns the corresponding storagePoolClaim object, and an error if there is any.
func (c *storagePoolClaims) Get(name string, options v1.GetOptions) (result *v1alpha1.StoragePoolClaim, err error) {
	result = &v1alpha1.StoragePoolClaim{}
	err = c.client.Get().
		Resource("storagepoolclaims").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of StoragePoolClaims that match those selectors.
func (c *storagePoolClaims) List(opts v1.ListOptions) (result *v1alpha1.StoragePoolClaimList, err error) {
	result = &v1alpha1.StoragePoolClaimList{}
	err = c.client.Get().
		Resource("storagepoolclaims").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested storagePoolClaims.
func (c *storagePoolClaims) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Resource("storagepoolclaims").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a storagePoolClaim and creates it.  Returns the server's representation of the storagePoolClaim, and an error, if there is any.
func (c *storagePoolClaims) Create(storagePoolClaim *v1alpha1.StoragePoolClaim) (result *v1alpha1.StoragePoolClaim, err error) {
	result = &v1alpha1.StoragePoolClaim{}
	err = c.client.Post().
		Resource("storagepoolclaims").
		Body(storagePoolClaim).
		Do().
		Into(result)
	return
}

// Update takes the representation of a storagePoolClaim and updates it. Returns the server's representation of the storagePoolClaim, and an error, if there is any.
func (c *storagePoolClaims) Update(storagePoolClaim *v1alpha1.StoragePoolClaim) (result *v1alpha1.StoragePoolClaim, err error) {
	result = &v1alpha1.StoragePoolClaim{}
	err = c.client.Put().
		Resource("storagepoolclaims").
		Name(storagePoolClaim.Name).
		Body(storagePoolClaim).
		Do().
		Into(result)
	return
}

// Delete takes name of the storagePoolClaim and deletes it. Returns an error if one occurs.
func (c *storagePoolClaims) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Resource("storagepoolclaims").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *storagePoolClaims) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Resource("storagepoolclaims").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched storagePoolClaim.
func (c *storagePoolClaims) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.StoragePoolClaim, err error) {
	result = &v1alpha1.StoragePoolClaim{}
	err = c.client.Patch(pt).
		Resource("storagepoolclaims").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
