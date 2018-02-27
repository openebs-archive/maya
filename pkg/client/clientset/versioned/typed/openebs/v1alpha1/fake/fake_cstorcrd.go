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

// FakeCstorCrds implements CstorCrdInterface
type FakeCstorCrds struct {
	Fake *FakeOpenebsV1alpha1
	ns   string
}

var cstorcrdsResource = schema.GroupVersionResource{Group: "openebs.io", Version: "v1alpha1", Resource: "cstorcrds"}

var cstorcrdsKind = schema.GroupVersionKind{Group: "openebs.io", Version: "v1alpha1", Kind: "CstorCrd"}

// Get takes name of the cstorCrd, and returns the corresponding cstorCrd object, and an error if there is any.
func (c *FakeCstorCrds) Get(name string, options v1.GetOptions) (result *v1alpha1.CstorCrd, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(cstorcrdsResource, c.ns, name), &v1alpha1.CstorCrd{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CstorCrd), err
}

// List takes label and field selectors, and returns the list of CstorCrds that match those selectors.
func (c *FakeCstorCrds) List(opts v1.ListOptions) (result *v1alpha1.CstorCrdList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(cstorcrdsResource, cstorcrdsKind, c.ns, opts), &v1alpha1.CstorCrdList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.CstorCrdList{}
	for _, item := range obj.(*v1alpha1.CstorCrdList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested cstorCrds.
func (c *FakeCstorCrds) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(cstorcrdsResource, c.ns, opts))

}

// Create takes the representation of a cstorCrd and creates it.  Returns the server's representation of the cstorCrd, and an error, if there is any.
func (c *FakeCstorCrds) Create(cstorCrd *v1alpha1.CstorCrd) (result *v1alpha1.CstorCrd, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(cstorcrdsResource, c.ns, cstorCrd), &v1alpha1.CstorCrd{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CstorCrd), err
}

// Update takes the representation of a cstorCrd and updates it. Returns the server's representation of the cstorCrd, and an error, if there is any.
func (c *FakeCstorCrds) Update(cstorCrd *v1alpha1.CstorCrd) (result *v1alpha1.CstorCrd, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(cstorcrdsResource, c.ns, cstorCrd), &v1alpha1.CstorCrd{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CstorCrd), err
}

// Delete takes name of the cstorCrd and deletes it. Returns an error if one occurs.
func (c *FakeCstorCrds) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(cstorcrdsResource, c.ns, name), &v1alpha1.CstorCrd{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeCstorCrds) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(cstorcrdsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.CstorCrdList{})
	return err
}

// Patch applies the patch and returns the patched cstorCrd.
func (c *FakeCstorCrds) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.CstorCrd, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(cstorcrdsResource, c.ns, name, data, subresources...), &v1alpha1.CstorCrd{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CstorCrd), err
}
