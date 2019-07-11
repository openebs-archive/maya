/*
Copyright 2019 The OpenEBS Authors

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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha2 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeCStorNPools implements CStorNPoolInterface
type FakeCStorNPools struct {
	Fake *FakeOpenebsV1alpha2
	ns   string
}

var cstornpoolsResource = schema.GroupVersionResource{Group: "openebs.io", Version: "v1alpha2", Resource: "cstornpools"}

var cstornpoolsKind = schema.GroupVersionKind{Group: "openebs.io", Version: "v1alpha2", Kind: "CStorNPool"}

// Get takes name of the cStorNPool, and returns the corresponding cStorNPool object, and an error if there is any.
func (c *FakeCStorNPools) Get(name string, options v1.GetOptions) (result *v1alpha2.CStorNPool, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(cstornpoolsResource, c.ns, name), &v1alpha2.CStorNPool{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.CStorNPool), err
}

// List takes label and field selectors, and returns the list of CStorNPools that match those selectors.
func (c *FakeCStorNPools) List(opts v1.ListOptions) (result *v1alpha2.CStorNPoolList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(cstornpoolsResource, cstornpoolsKind, c.ns, opts), &v1alpha2.CStorNPoolList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha2.CStorNPoolList{ListMeta: obj.(*v1alpha2.CStorNPoolList).ListMeta}
	for _, item := range obj.(*v1alpha2.CStorNPoolList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested cStorNPools.
func (c *FakeCStorNPools) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(cstornpoolsResource, c.ns, opts))

}

// Create takes the representation of a cStorNPool and creates it.  Returns the server's representation of the cStorNPool, and an error, if there is any.
func (c *FakeCStorNPools) Create(cStorNPool *v1alpha2.CStorNPool) (result *v1alpha2.CStorNPool, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(cstornpoolsResource, c.ns, cStorNPool), &v1alpha2.CStorNPool{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.CStorNPool), err
}

// Update takes the representation of a cStorNPool and updates it. Returns the server's representation of the cStorNPool, and an error, if there is any.
func (c *FakeCStorNPools) Update(cStorNPool *v1alpha2.CStorNPool) (result *v1alpha2.CStorNPool, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(cstornpoolsResource, c.ns, cStorNPool), &v1alpha2.CStorNPool{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.CStorNPool), err
}

// Delete takes name of the cStorNPool and deletes it. Returns an error if one occurs.
func (c *FakeCStorNPools) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(cstornpoolsResource, c.ns, name), &v1alpha2.CStorNPool{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeCStorNPools) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(cstornpoolsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha2.CStorNPoolList{})
	return err
}

// Patch applies the patch and returns the patched cStorNPool.
func (c *FakeCStorNPools) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha2.CStorNPool, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(cstornpoolsResource, c.ns, name, data, subresources...), &v1alpha2.CStorNPool{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.CStorNPool), err
}