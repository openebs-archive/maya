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

package cstorpoolclusterblockdevice

import (
	"github.com/pkg/errors"
)

// ListBuilder enables building an instance of
// CSPCBlockDevices
type ListBuilder struct {
	list *CSPCBlockDeviceList
	errs []error
}

// NewListBuilder returns an instance of ListBuilder
func NewListBuilder() *ListBuilder {
	return &ListBuilder{list: &CSPCBlockDeviceList{}}
}

// ListBuilderForObjectList builds the ListBuilder object based on CSPCList
func ListBuilderForObjectList(cspcBDList *CSPCBlockDeviceList) *ListBuilder {
	b := NewListBuilder()
	if cspcBDList == nil {
		b.errs = append(
			b.errs,
			errors.New("failed to build cspc blockdevice list: missing object list"),
		)
		return b
	}
	b.list = cspcBDList
	return b
}

// ListBuilderForObjectNew builds the new ListBuilder object
// based on CSPCBlockDevice
func ListBuilderForObjectNew(cspcBD *CSPCBlockDevice) *ListBuilder {
	lb := &ListBuilder{list: &CSPCBlockDeviceList{}}
	if cspcBD == nil {
		lb.errs = append(
			lb.errs,
			errors.New("failed to build cspc blockdevice list: missing object cspc blockdevice"),
		)
		return lb
	}
	lb.list.items = append(lb.list.items, cspcBD)
	return lb
}

// ListBuilderForObject adds the CSPCBlockDevice to existing ListBuilder
func (lb *ListBuilder) ListBuilderForObject(cspcBD *CSPCBlockDevice) *ListBuilder {
	if cspcBD == nil {
		lb.errs = append(
			lb.errs,
			errors.New("failed to build cspc blockdevice list: missing object cspc blockdevice"),
		)
		return lb
	}
	if lb == nil {
		return ListBuilderForObjectNew(cspcBD)
	}
	lb.list.items = append(lb.list.items, cspcBD)
	return lb
}

// GetRequiredBlockDevices returns the n number of blockdevices
func (lb *ListBuilder) GetRequiredBlockDevices(count int) *ListBuilder {
	newListBuilder := &ListBuilder{list: &CSPCBlockDeviceList{}}
	for i, obj := range lb.list.items {
		obj := obj
		newListBuilder.list.items = append(newListBuilder.list.items, obj)
		if i == count {
			break
		}
	}
	return newListBuilder
}

// Len returns the count of CSPC blockdevices
func (lb *ListBuilder) Len() int {
	return lb.list.Len()
}

//TODO: Get good function name from reviews

// ToObjectList converts the ListBuilder object into array of custom objects
func (lb *ListBuilder) ToObjectList() ([]*CSPCBlockDevice, error) {
	if len(lb.errs) > 0 {
		return nil, errors.Errorf("%+v", lb.errs)
	}
	return lb.list.items, nil
}
