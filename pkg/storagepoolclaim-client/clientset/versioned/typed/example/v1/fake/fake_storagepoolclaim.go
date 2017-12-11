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
	openebs_io_v1 "github.com/openebs/maya/pkg/storagepoolclaim-apis/openebs.io/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeStoragepoolclaims implements StoragepoolclaimInterface
type FakeStoragepoolclaims struct {
	Fake *FakeExampleV1
	ns   string
}

var storagepoolclaimsResource = schema.GroupVersionResource{Group: "example.com", Version: "v1", Resource: "storagepoolclaims"}

var storagepoolclaimsKind = schema.GroupVersionKind{Group: "example.com", Version: "v1", Kind: "Storagepoolclaim"}

// Get takes name of the storagepoolclaim, and returns the corresponding storagepoolclaim object, and an error if there is any.
func (c *FakeStoragepoolclaims) Get(name string, options v1.GetOptions) (result *openebs_io_v1.Storagepoolclaim, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(storagepoolclaimsResource, c.ns, name), &openebs_io_v1.Storagepoolclaim{})

	if obj == nil {
		return nil, err
	}
	return obj.(*openebs_io_v1.Storagepoolclaim), err
}

// List takes label and field selectors, and returns the list of Storagepoolclaims that match those selectors.
func (c *FakeStoragepoolclaims) List(opts v1.ListOptions) (result *openebs_io_v1.StoragepoolclaimList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(storagepoolclaimsResource, storagepoolclaimsKind, c.ns, opts), &openebs_io_v1.StoragepoolclaimList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &openebs_io_v1.StoragepoolclaimList{}
	for _, item := range obj.(*openebs_io_v1.StoragepoolclaimList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested storagepoolclaims.
func (c *FakeStoragepoolclaims) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(storagepoolclaimsResource, c.ns, opts))

}

// Create takes the representation of a storagepoolclaim and creates it.  Returns the server's representation of the storagepoolclaim, and an error, if there is any.
func (c *FakeStoragepoolclaims) Create(storagepoolclaim *openebs_io_v1.Storagepoolclaim) (result *openebs_io_v1.Storagepoolclaim, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(storagepoolclaimsResource, c.ns, storagepoolclaim), &openebs_io_v1.Storagepoolclaim{})

	if obj == nil {
		return nil, err
	}
	return obj.(*openebs_io_v1.Storagepoolclaim), err
}

// Update takes the representation of a storagepoolclaim and updates it. Returns the server's representation of the storagepoolclaim, and an error, if there is any.
func (c *FakeStoragepoolclaims) Update(storagepoolclaim *openebs_io_v1.Storagepoolclaim) (result *openebs_io_v1.Storagepoolclaim, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(storagepoolclaimsResource, c.ns, storagepoolclaim), &openebs_io_v1.Storagepoolclaim{})

	if obj == nil {
		return nil, err
	}
	return obj.(*openebs_io_v1.Storagepoolclaim), err
}

// Delete takes name of the storagepoolclaim and deletes it. Returns an error if one occurs.
func (c *FakeStoragepoolclaims) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(storagepoolclaimsResource, c.ns, name), &openebs_io_v1.Storagepoolclaim{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeStoragepoolclaims) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(storagepoolclaimsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &openebs_io_v1.StoragepoolclaimList{})
	return err
}

// Patch applies the patch and returns the patched storagepoolclaim.
func (c *FakeStoragepoolclaims) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *openebs_io_v1.Storagepoolclaim, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(storagepoolclaimsResource, c.ns, name, data, subresources...), &openebs_io_v1.Storagepoolclaim{})

	if obj == nil {
		return nil, err
	}
	return obj.(*openebs_io_v1.Storagepoolclaim), err
}
