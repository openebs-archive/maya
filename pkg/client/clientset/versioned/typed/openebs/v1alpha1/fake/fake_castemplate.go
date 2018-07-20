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

// FakeCASTemplates implements CASTemplateInterface
type FakeCASTemplates struct {
	Fake *FakeOpenebsV1alpha1
}

var castemplatesResource = schema.GroupVersionResource{Group: "openebs.io", Version: "v1alpha1", Resource: "castemplates"}

var castemplatesKind = schema.GroupVersionKind{Group: "openebs.io", Version: "v1alpha1", Kind: "CASTemplate"}

// Get takes name of the cASTemplate, and returns the corresponding cASTemplate object, and an error if there is any.
func (c *FakeCASTemplates) Get(name string, options v1.GetOptions) (result *v1alpha1.CASTemplate, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(castemplatesResource, name), &v1alpha1.CASTemplate{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CASTemplate), err
}

// List takes label and field selectors, and returns the list of CASTemplates that match those selectors.
func (c *FakeCASTemplates) List(opts v1.ListOptions) (result *v1alpha1.CASTemplateList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(castemplatesResource, castemplatesKind, opts), &v1alpha1.CASTemplateList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.CASTemplateList{}
	for _, item := range obj.(*v1alpha1.CASTemplateList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested cASTemplates.
func (c *FakeCASTemplates) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(castemplatesResource, opts))
}

// Create takes the representation of a cASTemplate and creates it.  Returns the server's representation of the cASTemplate, and an error, if there is any.
func (c *FakeCASTemplates) Create(cASTemplate *v1alpha1.CASTemplate) (result *v1alpha1.CASTemplate, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(castemplatesResource, cASTemplate), &v1alpha1.CASTemplate{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CASTemplate), err
}

// Update takes the representation of a cASTemplate and updates it. Returns the server's representation of the cASTemplate, and an error, if there is any.
func (c *FakeCASTemplates) Update(cASTemplate *v1alpha1.CASTemplate) (result *v1alpha1.CASTemplate, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(castemplatesResource, cASTemplate), &v1alpha1.CASTemplate{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CASTemplate), err
}

// Delete takes name of the cASTemplate and deletes it. Returns an error if one occurs.
func (c *FakeCASTemplates) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteAction(castemplatesResource, name), &v1alpha1.CASTemplate{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeCASTemplates) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(castemplatesResource, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.CASTemplateList{})
	return err
}

// Patch applies the patch and returns the patched cASTemplate.
func (c *FakeCASTemplates) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.CASTemplate, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(castemplatesResource, name, data, subresources...), &v1alpha1.CASTemplate{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CASTemplate), err
}
