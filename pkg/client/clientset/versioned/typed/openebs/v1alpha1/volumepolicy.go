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

// VolumePoliciesGetter has a method to return a VolumePolicyInterface.
// A group's client should implement this interface.
type VolumePoliciesGetter interface {
	VolumePolicies() VolumePolicyInterface
}

// VolumePolicyInterface has methods to work with VolumePolicy resources.
type VolumePolicyInterface interface {
	Create(*v1alpha1.VolumePolicy) (*v1alpha1.VolumePolicy, error)
	Update(*v1alpha1.VolumePolicy) (*v1alpha1.VolumePolicy, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.VolumePolicy, error)
	List(opts v1.ListOptions) (*v1alpha1.VolumePolicyList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.VolumePolicy, err error)
	VolumePolicyExpansion
}

// volumePolicies implements VolumePolicyInterface
type volumePolicies struct {
	client rest.Interface
}

// newVolumePolicies returns a VolumePolicies
func newVolumePolicies(c *OpenebsV1alpha1Client) *volumePolicies {
	return &volumePolicies{
		client: c.RESTClient(),
	}
}

// Get takes name of the volumePolicy, and returns the corresponding volumePolicy object, and an error if there is any.
func (c *volumePolicies) Get(name string, options v1.GetOptions) (result *v1alpha1.VolumePolicy, err error) {
	result = &v1alpha1.VolumePolicy{}
	err = c.client.Get().
		Resource("volumepolicies").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of VolumePolicies that match those selectors.
func (c *volumePolicies) List(opts v1.ListOptions) (result *v1alpha1.VolumePolicyList, err error) {
	result = &v1alpha1.VolumePolicyList{}
	err = c.client.Get().
		Resource("volumepolicies").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested volumePolicies.
func (c *volumePolicies) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Resource("volumepolicies").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a volumePolicy and creates it.  Returns the server's representation of the volumePolicy, and an error, if there is any.
func (c *volumePolicies) Create(volumePolicy *v1alpha1.VolumePolicy) (result *v1alpha1.VolumePolicy, err error) {
	result = &v1alpha1.VolumePolicy{}
	err = c.client.Post().
		Resource("volumepolicies").
		Body(volumePolicy).
		Do().
		Into(result)
	return
}

// Update takes the representation of a volumePolicy and updates it. Returns the server's representation of the volumePolicy, and an error, if there is any.
func (c *volumePolicies) Update(volumePolicy *v1alpha1.VolumePolicy) (result *v1alpha1.VolumePolicy, err error) {
	result = &v1alpha1.VolumePolicy{}
	err = c.client.Put().
		Resource("volumepolicies").
		Name(volumePolicy.Name).
		Body(volumePolicy).
		Do().
		Into(result)
	return
}

// Delete takes name of the volumePolicy and deletes it. Returns an error if one occurs.
func (c *volumePolicies) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Resource("volumepolicies").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *volumePolicies) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Resource("volumepolicies").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched volumePolicy.
func (c *volumePolicies) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.VolumePolicy, err error) {
	result = &v1alpha1.VolumePolicy{}
	err = c.client.Patch(pt).
		Resource("volumepolicies").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
