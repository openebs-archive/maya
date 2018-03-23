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

// VolumeParameterGroupsGetter has a method to return a VolumeParameterGroupInterface.
// A group's client should implement this interface.
type VolumeParameterGroupsGetter interface {
	VolumeParameterGroups() VolumeParameterGroupInterface
}

// VolumeParameterGroupInterface has methods to work with VolumeParameterGroup resources.
type VolumeParameterGroupInterface interface {
	Create(*v1alpha1.VolumeParameterGroup) (*v1alpha1.VolumeParameterGroup, error)
	Update(*v1alpha1.VolumeParameterGroup) (*v1alpha1.VolumeParameterGroup, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.VolumeParameterGroup, error)
	List(opts v1.ListOptions) (*v1alpha1.VolumeParameterGroupList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.VolumeParameterGroup, err error)
	VolumeParameterGroupExpansion
}

// volumeParameterGroups implements VolumeParameterGroupInterface
type volumeParameterGroups struct {
	client rest.Interface
}

// newVolumeParameterGroups returns a VolumeParameterGroups
func newVolumeParameterGroups(c *OpenebsV1alpha1Client) *volumeParameterGroups {
	return &volumeParameterGroups{
		client: c.RESTClient(),
	}
}

// Get takes name of the volumeParameterGroup, and returns the corresponding volumeParameterGroup object, and an error if there is any.
func (c *volumeParameterGroups) Get(name string, options v1.GetOptions) (result *v1alpha1.VolumeParameterGroup, err error) {
	result = &v1alpha1.VolumeParameterGroup{}
	err = c.client.Get().
		Resource("volumeparametergroups").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of VolumeParameterGroups that match those selectors.
func (c *volumeParameterGroups) List(opts v1.ListOptions) (result *v1alpha1.VolumeParameterGroupList, err error) {
	result = &v1alpha1.VolumeParameterGroupList{}
	err = c.client.Get().
		Resource("volumeparametergroups").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested volumeParameterGroups.
func (c *volumeParameterGroups) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Resource("volumeparametergroups").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a volumeParameterGroup and creates it.  Returns the server's representation of the volumeParameterGroup, and an error, if there is any.
func (c *volumeParameterGroups) Create(volumeParameterGroup *v1alpha1.VolumeParameterGroup) (result *v1alpha1.VolumeParameterGroup, err error) {
	result = &v1alpha1.VolumeParameterGroup{}
	err = c.client.Post().
		Resource("volumeparametergroups").
		Body(volumeParameterGroup).
		Do().
		Into(result)
	return
}

// Update takes the representation of a volumeParameterGroup and updates it. Returns the server's representation of the volumeParameterGroup, and an error, if there is any.
func (c *volumeParameterGroups) Update(volumeParameterGroup *v1alpha1.VolumeParameterGroup) (result *v1alpha1.VolumeParameterGroup, err error) {
	result = &v1alpha1.VolumeParameterGroup{}
	err = c.client.Put().
		Resource("volumeparametergroups").
		Name(volumeParameterGroup.Name).
		Body(volumeParameterGroup).
		Do().
		Into(result)
	return
}

// Delete takes name of the volumeParameterGroup and deletes it. Returns an error if one occurs.
func (c *volumeParameterGroups) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Resource("volumeparametergroups").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *volumeParameterGroups) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Resource("volumeparametergroups").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched volumeParameterGroup.
func (c *volumeParameterGroups) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.VolumeParameterGroup, err error) {
	result = &v1alpha1.VolumeParameterGroup{}
	err = c.client.Patch(pt).
		Resource("volumeparametergroups").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
