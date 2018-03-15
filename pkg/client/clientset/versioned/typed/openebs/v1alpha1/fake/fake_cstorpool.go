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

// FakeCstorPools implements CstorPoolInterface
type FakeCstorPools struct {
	Fake *FakeOpenebsV1alpha1
}

var cstorpoolsResource = schema.GroupVersionResource{Group: "openebs.io", Version: "v1alpha1", Resource: "cstorpools"}

var cstorpoolsKind = schema.GroupVersionKind{Group: "openebs.io", Version: "v1alpha1", Kind: "CstorPool"}

// Get takes name of the cstorPool, and returns the corresponding cstorPool object, and an error if there is any.
func (c *FakeCstorPools) Get(name string, options v1.GetOptions) (result *v1alpha1.CstorPool, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(cstorpoolsResource, name), &v1alpha1.CstorPool{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CstorPool), err
}

// List takes label and field selectors, and returns the list of CstorPools that match those selectors.
func (c *FakeCstorPools) List(opts v1.ListOptions) (result *v1alpha1.CstorPoolList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(cstorpoolsResource, cstorpoolsKind, opts), &v1alpha1.CstorPoolList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.CstorPoolList{}
	for _, item := range obj.(*v1alpha1.CstorPoolList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested cstorPools.
func (c *FakeCstorPools) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(cstorpoolsResource, opts))
}

// Create takes the representation of a cstorPool and creates it.  Returns the server's representation of the cstorPool, and an error, if there is any.
func (c *FakeCstorPools) Create(cstorPool *v1alpha1.CstorPool) (result *v1alpha1.CstorPool, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(cstorpoolsResource, cstorPool), &v1alpha1.CstorPool{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CstorPool), err
}

// Update takes the representation of a cstorPool and updates it. Returns the server's representation of the cstorPool, and an error, if there is any.
func (c *FakeCstorPools) Update(cstorPool *v1alpha1.CstorPool) (result *v1alpha1.CstorPool, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(cstorpoolsResource, cstorPool), &v1alpha1.CstorPool{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CstorPool), err
}

// Delete takes name of the cstorPool and deletes it. Returns an error if one occurs.
func (c *FakeCstorPools) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteAction(cstorpoolsResource, name), &v1alpha1.CstorPool{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeCstorPools) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(cstorpoolsResource, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.CstorPoolList{})
	return err
}

// Patch applies the patch and returns the patched cstorPool.
func (c *FakeCstorPools) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.CstorPool, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(cstorpoolsResource, name, data, subresources...), &v1alpha1.CstorPool{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CstorPool), err
}
