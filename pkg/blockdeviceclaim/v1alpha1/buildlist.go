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

package v1alpha1

import (
	ndm "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
)

//TODO: While using these packages UnitTest must be written to corresponding function

// ListBuilder is the builder object for BlockDeviceClaimList
type ListBuilder struct {
	BlockDeviceClaimList *BlockDeviceClaimList
}

// NewListBuilder returns a new instance of ListBuilder object.
func NewListBuilder() *ListBuilder {
	return &ListBuilder{
		BlockDeviceClaimList: &BlockDeviceClaimList{
			ObjectList: &ndm.BlockDeviceClaimList{},
		},
	}
}

// ListBuilderFromList builds the list based on the
// provided *BlockDeviceClaimList instances.
func ListBuilderFromList(bdcl *BlockDeviceClaimList) *ListBuilder {
	lb := NewListBuilder()
	if bdcl == nil {
		return lb
	}
	lb.BlockDeviceClaimList.ObjectList.Items =
		append(lb.BlockDeviceClaimList.ObjectList.Items,
			bdcl.ObjectList.Items...)
	return lb
}

// ListBuilderFromAPIList builds the list based on the provided APIBDC List
func ListBuilderFromAPIList(bdcl *ndm.BlockDeviceClaimList) *ListBuilder {
	lb := NewListBuilder()
	if bdcl == nil {
		return lb
	}
	lb.BlockDeviceClaimList.ObjectList.Items = append(
		lb.BlockDeviceClaimList.ObjectList.Items,
		bdcl.Items...)
	return lb
}

// List returns the list of block device claim
// instances that were built by this builder.
func (b *ListBuilder) List() *BlockDeviceClaimList {
	return b.BlockDeviceClaimList
}
