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

// FakeVolumeParameterGroups implements VolumeParameterGroupInterface
type FakeVolumeParameterGroups struct {
	Fake *FakeOpenebsV1alpha1
}

var volumeparametergroupsResource = schema.GroupVersionResource{Group: "openebs.io", Version: "v1alpha1", Resource: "volumeparametergroups"}

var volumeparametergroupsKind = schema.GroupVersionKind{Group: "openebs.io", Version: "v1alpha1", Kind: "VolumeParameterGroup"}

// Get takes name of the volumeParameterGroup, and returns the corresponding volumeParameterGroup object, and an error if there is any.
func (c *FakeVolumeParameterGroups) Get(name string, options v1.GetOptions) (result *v1alpha1.VolumeParameterGroup, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(volumeparametergroupsResource, name), &v1alpha1.VolumeParameterGroup{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VolumeParameterGroup), err
}

// List takes label and field selectors, and returns the list of VolumeParameterGroups that match those selectors.
func (c *FakeVolumeParameterGroups) List(opts v1.ListOptions) (result *v1alpha1.VolumeParameterGroupList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(volumeparametergroupsResource, volumeparametergroupsKind, opts), &v1alpha1.VolumeParameterGroupList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.VolumeParameterGroupList{}
	for _, item := range obj.(*v1alpha1.VolumeParameterGroupList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested volumeParameterGroups.
func (c *FakeVolumeParameterGroups) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(volumeparametergroupsResource, opts))
}

// Create takes the representation of a volumeParameterGroup and creates it.  Returns the server's representation of the volumeParameterGroup, and an error, if there is any.
func (c *FakeVolumeParameterGroups) Create(volumeParameterGroup *v1alpha1.VolumeParameterGroup) (result *v1alpha1.VolumeParameterGroup, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(volumeparametergroupsResource, volumeParameterGroup), &v1alpha1.VolumeParameterGroup{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VolumeParameterGroup), err
}

// Update takes the representation of a volumeParameterGroup and updates it. Returns the server's representation of the volumeParameterGroup, and an error, if there is any.
func (c *FakeVolumeParameterGroups) Update(volumeParameterGroup *v1alpha1.VolumeParameterGroup) (result *v1alpha1.VolumeParameterGroup, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(volumeparametergroupsResource, volumeParameterGroup), &v1alpha1.VolumeParameterGroup{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VolumeParameterGroup), err
}

// Delete takes name of the volumeParameterGroup and deletes it. Returns an error if one occurs.
func (c *FakeVolumeParameterGroups) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteAction(volumeparametergroupsResource, name), &v1alpha1.VolumeParameterGroup{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeVolumeParameterGroups) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(volumeparametergroupsResource, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.VolumeParameterGroupList{})
	return err
}

// Patch applies the patch and returns the patched volumeParameterGroup.
func (c *FakeVolumeParameterGroups) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.VolumeParameterGroup, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(volumeparametergroupsResource, name, data, subresources...), &v1alpha1.VolumeParameterGroup{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VolumeParameterGroup), err
}
