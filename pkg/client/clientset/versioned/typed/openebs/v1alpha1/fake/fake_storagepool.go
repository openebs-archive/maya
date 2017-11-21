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

// FakeStoragePools implements StoragePoolInterface
type FakeStoragePools struct {
	Fake *FakeOpenebsV1alpha1
	ns   string
}

var storagepoolsResource = schema.GroupVersionResource{Group: "openebs.io", Version: "v1alpha1", Resource: "storagepools"}

var storagepoolsKind = schema.GroupVersionKind{Group: "openebs.io", Version: "v1alpha1", Kind: "StoragePool"}

// Get takes name of the storagePool, and returns the corresponding storagePool object, and an error if there is any.
func (c *FakeStoragePools) Get(name string, options v1.GetOptions) (result *v1alpha1.StoragePool, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(storagepoolsResource, c.ns, name), &v1alpha1.StoragePool{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.StoragePool), err
}

// List takes label and field selectors, and returns the list of StoragePools that match those selectors.
func (c *FakeStoragePools) List(opts v1.ListOptions) (result *v1alpha1.StoragePoolList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(storagepoolsResource, storagepoolsKind, c.ns, opts), &v1alpha1.StoragePoolList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.StoragePoolList{}
	for _, item := range obj.(*v1alpha1.StoragePoolList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested storagePools.
func (c *FakeStoragePools) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(storagepoolsResource, c.ns, opts))

}

// Create takes the representation of a storagePool and creates it.  Returns the server's representation of the storagePool, and an error, if there is any.
func (c *FakeStoragePools) Create(storagePool *v1alpha1.StoragePool) (result *v1alpha1.StoragePool, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(storagepoolsResource, c.ns, storagePool), &v1alpha1.StoragePool{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.StoragePool), err
}

// Update takes the representation of a storagePool and updates it. Returns the server's representation of the storagePool, and an error, if there is any.
func (c *FakeStoragePools) Update(storagePool *v1alpha1.StoragePool) (result *v1alpha1.StoragePool, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(storagepoolsResource, c.ns, storagePool), &v1alpha1.StoragePool{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.StoragePool), err
}

// Delete takes name of the storagePool and deletes it. Returns an error if one occurs.
func (c *FakeStoragePools) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(storagepoolsResource, c.ns, name), &v1alpha1.StoragePool{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeStoragePools) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(storagepoolsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.StoragePoolList{})
	return err
}

// Patch applies the patch and returns the patched storagePool.
func (c *FakeStoragePools) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.StoragePool, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(storagepoolsResource, c.ns, name, data, subresources...), &v1alpha1.StoragePool{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.StoragePool), err
}
