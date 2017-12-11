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
	openebs_io_v1 "github.com/openebs/maya/pkg/storagepool-apis/openebs.io/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeStoragepools implements StoragepoolInterface
type FakeStoragepools struct {
	Fake *FakeExampleV1
	ns   string
}

var storagepoolsResource = schema.GroupVersionResource{Group: "example.com", Version: "v1", Resource: "storagepools"}

var storagepoolsKind = schema.GroupVersionKind{Group: "example.com", Version: "v1", Kind: "Storagepool"}

// Get takes name of the storagepool, and returns the corresponding storagepool object, and an error if there is any.
func (c *FakeStoragepools) Get(name string, options v1.GetOptions) (result *openebs_io_v1.Storagepool, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(storagepoolsResource, c.ns, name), &openebs_io_v1.Storagepool{})

	if obj == nil {
		return nil, err
	}
	return obj.(*openebs_io_v1.Storagepool), err
}

// List takes label and field selectors, and returns the list of Storagepools that match those selectors.
func (c *FakeStoragepools) List(opts v1.ListOptions) (result *openebs_io_v1.StoragepoolList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(storagepoolsResource, storagepoolsKind, c.ns, opts), &openebs_io_v1.StoragepoolList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &openebs_io_v1.StoragepoolList{}
	for _, item := range obj.(*openebs_io_v1.StoragepoolList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested storagepools.
func (c *FakeStoragepools) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(storagepoolsResource, c.ns, opts))

}

// Create takes the representation of a storagepool and creates it.  Returns the server's representation of the storagepool, and an error, if there is any.
func (c *FakeStoragepools) Create(storagepool *openebs_io_v1.Storagepool) (result *openebs_io_v1.Storagepool, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(storagepoolsResource, c.ns, storagepool), &openebs_io_v1.Storagepool{})

	if obj == nil {
		return nil, err
	}
	return obj.(*openebs_io_v1.Storagepool), err
}

// Update takes the representation of a storagepool and updates it. Returns the server's representation of the storagepool, and an error, if there is any.
func (c *FakeStoragepools) Update(storagepool *openebs_io_v1.Storagepool) (result *openebs_io_v1.Storagepool, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(storagepoolsResource, c.ns, storagepool), &openebs_io_v1.Storagepool{})

	if obj == nil {
		return nil, err
	}
	return obj.(*openebs_io_v1.Storagepool), err
}

// Delete takes name of the storagepool and deletes it. Returns an error if one occurs.
func (c *FakeStoragepools) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(storagepoolsResource, c.ns, name), &openebs_io_v1.Storagepool{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeStoragepools) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(storagepoolsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &openebs_io_v1.StoragepoolList{})
	return err
}

// Patch applies the patch and returns the patched storagepool.
func (c *FakeStoragepools) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *openebs_io_v1.Storagepool, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(storagepoolsResource, c.ns, name, data, subresources...), &openebs_io_v1.Storagepool{})

	if obj == nil {
		return nil, err
	}
	return obj.(*openebs_io_v1.Storagepool), err
}
