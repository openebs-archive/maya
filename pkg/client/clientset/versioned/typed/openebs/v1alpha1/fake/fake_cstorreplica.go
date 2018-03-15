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

// FakeCstorReplicas implements CstorReplicaInterface
type FakeCstorReplicas struct {
	Fake *FakeOpenebsV1alpha1
}

var cstorreplicasResource = schema.GroupVersionResource{Group: "openebs.io", Version: "v1alpha1", Resource: "cstorreplicas"}

var cstorreplicasKind = schema.GroupVersionKind{Group: "openebs.io", Version: "v1alpha1", Kind: "CstorReplica"}

// Get takes name of the cstorReplica, and returns the corresponding cstorReplica object, and an error if there is any.
func (c *FakeCstorReplicas) Get(name string, options v1.GetOptions) (result *v1alpha1.CstorReplica, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(cstorreplicasResource, name), &v1alpha1.CstorReplica{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CstorReplica), err
}

// List takes label and field selectors, and returns the list of CstorReplicas that match those selectors.
func (c *FakeCstorReplicas) List(opts v1.ListOptions) (result *v1alpha1.CstorReplicaList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(cstorreplicasResource, cstorreplicasKind, opts), &v1alpha1.CstorReplicaList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.CstorReplicaList{}
	for _, item := range obj.(*v1alpha1.CstorReplicaList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested cstorReplicas.
func (c *FakeCstorReplicas) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(cstorreplicasResource, opts))
}

// Create takes the representation of a cstorReplica and creates it.  Returns the server's representation of the cstorReplica, and an error, if there is any.
func (c *FakeCstorReplicas) Create(cstorReplica *v1alpha1.CstorReplica) (result *v1alpha1.CstorReplica, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(cstorreplicasResource, cstorReplica), &v1alpha1.CstorReplica{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CstorReplica), err
}

// Update takes the representation of a cstorReplica and updates it. Returns the server's representation of the cstorReplica, and an error, if there is any.
func (c *FakeCstorReplicas) Update(cstorReplica *v1alpha1.CstorReplica) (result *v1alpha1.CstorReplica, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(cstorreplicasResource, cstorReplica), &v1alpha1.CstorReplica{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CstorReplica), err
}

// Delete takes name of the cstorReplica and deletes it. Returns an error if one occurs.
func (c *FakeCstorReplicas) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteAction(cstorreplicasResource, name), &v1alpha1.CstorReplica{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeCstorReplicas) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(cstorreplicasResource, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.CstorReplicaList{})
	return err
}

// Patch applies the patch and returns the patched cstorReplica.
func (c *FakeCstorReplicas) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.CstorReplica, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(cstorreplicasResource, name, data, subresources...), &v1alpha1.CstorReplica{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CstorReplica), err
}
