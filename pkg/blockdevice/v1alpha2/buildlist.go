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

package v1alpha2

//TODO: While using these packages UnitTest must be written to corresponding function

import (
	ndm "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
)

// ListBuilder is the builder object for BlockDeviceList
type ListBuilder struct {
	BDL *BlockDeviceList
}

// NewListBuilder returns a new instance of ListBuilder object.
func NewListBuilder() *ListBuilder {
	return &ListBuilder{
		BDL: &BlockDeviceList{
			ObjectList: &ndm.BlockDeviceList{},
		},
	}
}

// WithList builds the list based on the provided *BlockDeviceList instances.
func (b *ListBuilder) WithList(bdl *BlockDeviceList) *ListBuilder {
	if bdl == nil {
		return b
	}
	b.BDL.ObjectList.Items = append(
		b.BDL.ObjectList.Items,
		bdl.ObjectList.Items...)
	return b
}

// WithAPIList builds the list based on the provided *apisv1alpha1.CStorPoolList.
func (b *ListBuilder) WithAPIList(bdl *ndm.BlockDeviceList) *ListBuilder {
	if bdl == nil {
		return b
	}
	b.BDL.ObjectList.Items = append(
		b.BDL.ObjectList.Items,
		bdl.Items...)
	return b
}

// List returns the list of block device instances
// that were built by this builder.
func (b *ListBuilder) List() *BlockDeviceList {
	return b.BDL
}

// NewListBuilderForObjectList builds the list based on the provided
// *BlockDeviceList instances.
func NewListBuilderForObjectList(bdl *BlockDeviceList) *ListBuilder {
	newLB := NewListBuilder()
	newLB.BDL.ObjectList.Items = append(newLB.BDL.ObjectList.Items, bdl.ObjectList.Items...)
	return newLB
}

// NewListBuilderForAPIList returns a new instance of ListBuilderForApiList
// object based on block device api list.
func NewListBuilderForAPIList(bdAPIList *ndm.BlockDeviceList) *ListBuilder {
	newLb := NewListBuilder()
	newLb.BDL.ObjectList.Items = append(newLb.BDL.ObjectList.Items, bdAPIList.Items...)
	return newLb
}
