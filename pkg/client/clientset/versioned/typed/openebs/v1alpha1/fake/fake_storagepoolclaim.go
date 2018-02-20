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

// FakeStoragePoolClaims implements StoragePoolClaimInterface
type FakeStoragePoolClaims struct {
	Fake *FakeOpenebsV1alpha1
}

var storagepoolclaimsResource = schema.GroupVersionResource{Group: "openebs.io", Version: "v1alpha1", Resource: "storagepoolclaims"}

var storagepoolclaimsKind = schema.GroupVersionKind{Group: "openebs.io", Version: "v1alpha1", Kind: "StoragePoolClaim"}

// Get takes name of the storagePoolClaim, and returns the corresponding storagePoolClaim object, and an error if there is any.
func (c *FakeStoragePoolClaims) Get(name string, options v1.GetOptions) (result *v1alpha1.StoragePoolClaim, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(storagepoolclaimsResource, name), &v1alpha1.StoragePoolClaim{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.StoragePoolClaim), err
}

// List takes label and field selectors, and returns the list of StoragePoolClaims that match those selectors.
func (c *FakeStoragePoolClaims) List(opts v1.ListOptions) (result *v1alpha1.StoragePoolClaimList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(storagepoolclaimsResource, storagepoolclaimsKind, opts), &v1alpha1.StoragePoolClaimList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.StoragePoolClaimList{}
	for _, item := range obj.(*v1alpha1.StoragePoolClaimList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested storagePoolClaims.
func (c *FakeStoragePoolClaims) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(storagepoolclaimsResource, opts))
}

// Create takes the representation of a storagePoolClaim and creates it.  Returns the server's representation of the storagePoolClaim, and an error, if there is any.
func (c *FakeStoragePoolClaims) Create(storagePoolClaim *v1alpha1.StoragePoolClaim) (result *v1alpha1.StoragePoolClaim, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(storagepoolclaimsResource, storagePoolClaim), &v1alpha1.StoragePoolClaim{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.StoragePoolClaim), err
}

// Update takes the representation of a storagePoolClaim and updates it. Returns the server's representation of the storagePoolClaim, and an error, if there is any.
func (c *FakeStoragePoolClaims) Update(storagePoolClaim *v1alpha1.StoragePoolClaim) (result *v1alpha1.StoragePoolClaim, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(storagepoolclaimsResource, storagePoolClaim), &v1alpha1.StoragePoolClaim{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.StoragePoolClaim), err
}

// Delete takes name of the storagePoolClaim and deletes it. Returns an error if one occurs.
func (c *FakeStoragePoolClaims) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteAction(storagepoolclaimsResource, name), &v1alpha1.StoragePoolClaim{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeStoragePoolClaims) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(storagepoolclaimsResource, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.StoragePoolClaimList{})
	return err
}

// Patch applies the patch and returns the patched storagePoolClaim.
func (c *FakeStoragePoolClaims) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.StoragePoolClaim, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(storagepoolclaimsResource, name, data, subresources...), &v1alpha1.StoragePoolClaim{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.StoragePoolClaim), err
}
