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

// CstorReplicasGetter has a method to return a CstorReplicaInterface.
// A group's client should implement this interface.
type CstorReplicasGetter interface {
	CstorReplicas() CstorReplicaInterface
}

// CstorReplicaInterface has methods to work with CstorReplica resources.
type CstorReplicaInterface interface {
	Create(*v1alpha1.CstorReplica) (*v1alpha1.CstorReplica, error)
	Update(*v1alpha1.CstorReplica) (*v1alpha1.CstorReplica, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.CstorReplica, error)
	List(opts v1.ListOptions) (*v1alpha1.CstorReplicaList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.CstorReplica, err error)
	CstorReplicaExpansion
}

// cstorReplicas implements CstorReplicaInterface
type cstorReplicas struct {
	client rest.Interface
}

// newCstorReplicas returns a CstorReplicas
func newCstorReplicas(c *OpenebsV1alpha1Client) *cstorReplicas {
	return &cstorReplicas{
		client: c.RESTClient(),
	}
}

// Get takes name of the cstorReplica, and returns the corresponding cstorReplica object, and an error if there is any.
func (c *cstorReplicas) Get(name string, options v1.GetOptions) (result *v1alpha1.CstorReplica, err error) {
	result = &v1alpha1.CstorReplica{}
	err = c.client.Get().
		Resource("cstorreplicas").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of CstorReplicas that match those selectors.
func (c *cstorReplicas) List(opts v1.ListOptions) (result *v1alpha1.CstorReplicaList, err error) {
	result = &v1alpha1.CstorReplicaList{}
	err = c.client.Get().
		Resource("cstorreplicas").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested cstorReplicas.
func (c *cstorReplicas) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Resource("cstorreplicas").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a cstorReplica and creates it.  Returns the server's representation of the cstorReplica, and an error, if there is any.
func (c *cstorReplicas) Create(cstorReplica *v1alpha1.CstorReplica) (result *v1alpha1.CstorReplica, err error) {
	result = &v1alpha1.CstorReplica{}
	err = c.client.Post().
		Resource("cstorreplicas").
		Body(cstorReplica).
		Do().
		Into(result)
	return
}

// Update takes the representation of a cstorReplica and updates it. Returns the server's representation of the cstorReplica, and an error, if there is any.
func (c *cstorReplicas) Update(cstorReplica *v1alpha1.CstorReplica) (result *v1alpha1.CstorReplica, err error) {
	result = &v1alpha1.CstorReplica{}
	err = c.client.Put().
		Resource("cstorreplicas").
		Name(cstorReplica.Name).
		Body(cstorReplica).
		Do().
		Into(result)
	return
}

// Delete takes name of the cstorReplica and deletes it. Returns an error if one occurs.
func (c *cstorReplicas) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Resource("cstorreplicas").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *cstorReplicas) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Resource("cstorreplicas").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched cstorReplica.
func (c *cstorReplicas) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.CstorReplica, err error) {
	result = &v1alpha1.CstorReplica{}
	err = c.client.Patch(pt).
		Resource("cstorreplicas").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
