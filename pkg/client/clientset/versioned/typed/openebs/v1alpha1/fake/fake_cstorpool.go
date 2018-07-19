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

// FakeCStorPools implements CStorPoolInterface
type FakeCStorPools struct {
	Fake *FakeOpenebsV1alpha1
}

var cstorpoolsResource = schema.GroupVersionResource{Group: "openebs.io", Version: "v1alpha1", Resource: "cstorpools"}

var cstorpoolsKind = schema.GroupVersionKind{Group: "openebs.io", Version: "v1alpha1", Kind: "CStorPool"}

// Get takes name of the cStorPool, and returns the corresponding cStorPool object, and an error if there is any.
func (c *FakeCStorPools) Get(name string, options v1.GetOptions) (result *v1alpha1.CStorPool, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(cstorpoolsResource, name), &v1alpha1.CStorPool{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CStorPool), err
}

// List takes label and field selectors, and returns the list of CStorPools that match those selectors.
func (c *FakeCStorPools) List(opts v1.ListOptions) (result *v1alpha1.CStorPoolList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(cstorpoolsResource, cstorpoolsKind, opts), &v1alpha1.CStorPoolList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.CStorPoolList{}
	for _, item := range obj.(*v1alpha1.CStorPoolList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested cStorPools.
func (c *FakeCStorPools) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(cstorpoolsResource, opts))
}

// Create takes the representation of a cStorPool and creates it.  Returns the server's representation of the cStorPool, and an error, if there is any.
func (c *FakeCStorPools) Create(cStorPool *v1alpha1.CStorPool) (result *v1alpha1.CStorPool, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(cstorpoolsResource, cStorPool), &v1alpha1.CStorPool{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CStorPool), err
}

// Update takes the representation of a cStorPool and updates it. Returns the server's representation of the cStorPool, and an error, if there is any.
func (c *FakeCStorPools) Update(cStorPool *v1alpha1.CStorPool) (result *v1alpha1.CStorPool, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(cstorpoolsResource, cStorPool), &v1alpha1.CStorPool{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CStorPool), err
}

// Delete takes name of the cStorPool and deletes it. Returns an error if one occurs.
func (c *FakeCStorPools) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteAction(cstorpoolsResource, name), &v1alpha1.CStorPool{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeCStorPools) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(cstorpoolsResource, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.CStorPoolList{})
	return err
}

// Patch applies the patch and returns the patched cStorPool.
func (c *FakeCStorPools) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.CStorPool, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(cstorpoolsResource, name, data, subresources...), &v1alpha1.CStorPool{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CStorPool), err
}
