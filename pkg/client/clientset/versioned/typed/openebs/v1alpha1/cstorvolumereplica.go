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
package v1alpha1

import (
	v1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	scheme "github.com/openebs/maya/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// CStorVolumeReplicasGetter has a method to return a CStorVolumeReplicaInterface.
// A group's client should implement this interface.
type CStorVolumeReplicasGetter interface {
	CStorVolumeReplicas(namespace string) CStorVolumeReplicaInterface
}

// CStorVolumeReplicaInterface has methods to work with CStorVolumeReplica resources.
type CStorVolumeReplicaInterface interface {
	Create(*v1alpha1.CStorVolumeReplica) (*v1alpha1.CStorVolumeReplica, error)
	Update(*v1alpha1.CStorVolumeReplica) (*v1alpha1.CStorVolumeReplica, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.CStorVolumeReplica, error)
	List(opts v1.ListOptions) (*v1alpha1.CStorVolumeReplicaList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.CStorVolumeReplica, err error)
	CStorVolumeReplicaExpansion
}

// cStorVolumeReplicas implements CStorVolumeReplicaInterface
type cStorVolumeReplicas struct {
	client rest.Interface
	ns     string
}

// newCStorVolumeReplicas returns a CStorVolumeReplicas
func newCStorVolumeReplicas(c *OpenebsV1alpha1Client, namespace string) *cStorVolumeReplicas {
	return &cStorVolumeReplicas{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the cStorVolumeReplica, and returns the corresponding cStorVolumeReplica object, and an error if there is any.
func (c *cStorVolumeReplicas) Get(name string, options v1.GetOptions) (result *v1alpha1.CStorVolumeReplica, err error) {
	result = &v1alpha1.CStorVolumeReplica{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("cstorvolumereplicas").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of CStorVolumeReplicas that match those selectors.
func (c *cStorVolumeReplicas) List(opts v1.ListOptions) (result *v1alpha1.CStorVolumeReplicaList, err error) {
	result = &v1alpha1.CStorVolumeReplicaList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("cstorvolumereplicas").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested cStorVolumeReplicas.
func (c *cStorVolumeReplicas) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("cstorvolumereplicas").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a cStorVolumeReplica and creates it.  Returns the server's representation of the cStorVolumeReplica, and an error, if there is any.
func (c *cStorVolumeReplicas) Create(cStorVolumeReplica *v1alpha1.CStorVolumeReplica) (result *v1alpha1.CStorVolumeReplica, err error) {
	result = &v1alpha1.CStorVolumeReplica{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("cstorvolumereplicas").
		Body(cStorVolumeReplica).
		Do().
		Into(result)
	return
}

// Update takes the representation of a cStorVolumeReplica and updates it. Returns the server's representation of the cStorVolumeReplica, and an error, if there is any.
func (c *cStorVolumeReplicas) Update(cStorVolumeReplica *v1alpha1.CStorVolumeReplica) (result *v1alpha1.CStorVolumeReplica, err error) {
	result = &v1alpha1.CStorVolumeReplica{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("cstorvolumereplicas").
		Name(cStorVolumeReplica.Name).
		Body(cStorVolumeReplica).
		Do().
		Into(result)
	return
}

// Delete takes name of the cStorVolumeReplica and deletes it. Returns an error if one occurs.
func (c *cStorVolumeReplicas) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("cstorvolumereplicas").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *cStorVolumeReplicas) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("cstorvolumereplicas").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched cStorVolumeReplica.
func (c *cStorVolumeReplicas) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.CStorVolumeReplica, err error) {
	result = &v1alpha1.CStorVolumeReplica{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("cstorvolumereplicas").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
