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

// CStorVolumesGetter has a method to return a CStorVolumeInterface.
// A group's client should implement this interface.
type CStorVolumesGetter interface {
	CStorVolumes(namespace string) CStorVolumeInterface
}

// CStorVolumeInterface has methods to work with CStorVolume resources.
type CStorVolumeInterface interface {
	Create(*v1alpha1.CStorVolume) (*v1alpha1.CStorVolume, error)
	Update(*v1alpha1.CStorVolume) (*v1alpha1.CStorVolume, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.CStorVolume, error)
	List(opts v1.ListOptions) (*v1alpha1.CStorVolumeList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.CStorVolume, err error)
	CStorVolumeExpansion
}

// cStorVolumes implements CStorVolumeInterface
type cStorVolumes struct {
	client rest.Interface
	ns     string
}

// newCStorVolumes returns a CStorVolumes
func newCStorVolumes(c *OpenebsV1alpha1Client, namespace string) *cStorVolumes {
	return &cStorVolumes{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the cStorVolume, and returns the corresponding cStorVolume object, and an error if there is any.
func (c *cStorVolumes) Get(name string, options v1.GetOptions) (result *v1alpha1.CStorVolume, err error) {
	result = &v1alpha1.CStorVolume{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("cstorvolumes").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of CStorVolumes that match those selectors.
func (c *cStorVolumes) List(opts v1.ListOptions) (result *v1alpha1.CStorVolumeList, err error) {
	result = &v1alpha1.CStorVolumeList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("cstorvolumes").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested cStorVolumes.
func (c *cStorVolumes) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("cstorvolumes").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a cStorVolume and creates it.  Returns the server's representation of the cStorVolume, and an error, if there is any.
func (c *cStorVolumes) Create(cStorVolume *v1alpha1.CStorVolume) (result *v1alpha1.CStorVolume, err error) {
	result = &v1alpha1.CStorVolume{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("cstorvolumes").
		Body(cStorVolume).
		Do().
		Into(result)
	return
}

// Update takes the representation of a cStorVolume and updates it. Returns the server's representation of the cStorVolume, and an error, if there is any.
func (c *cStorVolumes) Update(cStorVolume *v1alpha1.CStorVolume) (result *v1alpha1.CStorVolume, err error) {
	result = &v1alpha1.CStorVolume{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("cstorvolumes").
		Name(cStorVolume.Name).
		Body(cStorVolume).
		Do().
		Into(result)
	return
}

// Delete takes name of the cStorVolume and deletes it. Returns an error if one occurs.
func (c *cStorVolumes) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("cstorvolumes").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *cStorVolumes) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("cstorvolumes").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched cStorVolume.
func (c *cStorVolumes) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.CStorVolume, err error) {
	result = &v1alpha1.CStorVolume{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("cstorvolumes").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
