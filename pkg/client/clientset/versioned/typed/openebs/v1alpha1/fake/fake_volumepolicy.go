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
package fake

import (
	v1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeVolumePolicies implements VolumePolicyInterface
type FakeVolumePolicies struct {
	Fake *FakeOpenebsV1alpha1
}

var volumepoliciesResource = schema.GroupVersionResource{Group: "openebs.io", Version: "v1alpha1", Resource: "volumepolicies"}

var volumepoliciesKind = schema.GroupVersionKind{Group: "openebs.io", Version: "v1alpha1", Kind: "VolumePolicy"}

// Get takes name of the volumePolicy, and returns the corresponding volumePolicy object, and an error if there is any.
func (c *FakeVolumePolicies) Get(name string, options v1.GetOptions) (result *v1alpha1.VolumePolicy, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(volumepoliciesResource, name), &v1alpha1.VolumePolicy{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VolumePolicy), err
}

// List takes label and field selectors, and returns the list of VolumePolicies that match those selectors.
func (c *FakeVolumePolicies) List(opts v1.ListOptions) (result *v1alpha1.VolumePolicyList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(volumepoliciesResource, volumepoliciesKind, opts), &v1alpha1.VolumePolicyList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.VolumePolicyList{}
	for _, item := range obj.(*v1alpha1.VolumePolicyList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested volumePolicies.
func (c *FakeVolumePolicies) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(volumepoliciesResource, opts))
}

// Create takes the representation of a volumePolicy and creates it.  Returns the server's representation of the volumePolicy, and an error, if there is any.
func (c *FakeVolumePolicies) Create(volumePolicy *v1alpha1.VolumePolicy) (result *v1alpha1.VolumePolicy, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(volumepoliciesResource, volumePolicy), &v1alpha1.VolumePolicy{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VolumePolicy), err
}

// Update takes the representation of a volumePolicy and updates it. Returns the server's representation of the volumePolicy, and an error, if there is any.
func (c *FakeVolumePolicies) Update(volumePolicy *v1alpha1.VolumePolicy) (result *v1alpha1.VolumePolicy, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(volumepoliciesResource, volumePolicy), &v1alpha1.VolumePolicy{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VolumePolicy), err
}

// Delete takes name of the volumePolicy and deletes it. Returns an error if one occurs.
func (c *FakeVolumePolicies) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteAction(volumepoliciesResource, name), &v1alpha1.VolumePolicy{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeVolumePolicies) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(volumepoliciesResource, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.VolumePolicyList{})
	return err
}

// Patch applies the patch and returns the patched volumePolicy.
func (c *FakeVolumePolicies) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.VolumePolicy, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(volumepoliciesResource, name, data, subresources...), &v1alpha1.VolumePolicy{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VolumePolicy), err
}
