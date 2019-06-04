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
	"testing"

	ndm "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFilter(t *testing.T) {
	tests := map[string]struct {
		// fakeCasPool holds the fake fakeCasPool object in test cases.
		blockDeviceList *BlockDeviceList
		filterPredicate []string
		// expectedBlockDeviceListLength holds the length of disk list
		expectedBlockDeviceCount int
	}{
		"EmptyBlockDeviceList1": {
			blockDeviceList:          nil,
			filterPredicate:          []string{FilterInactive},
			expectedBlockDeviceCount: 0,
		},
		"EmptyBlockDeviceList2": {
			blockDeviceList: &BlockDeviceList{
				BlockDeviceList: nil,
				errs:            nil,
			},
			filterPredicate:          []string{FilterInactive},
			expectedBlockDeviceCount: 0,
		},
		"EmptyBlockDeviceList3": {
			blockDeviceList: &BlockDeviceList{
				BlockDeviceList: &ndm.BlockDeviceList{},
				errs:            nil,
			},
			filterPredicate:          []string{FilterInactive},
			expectedBlockDeviceCount: 0,
		},
		"blockDeviceList3": {
			blockDeviceList: &BlockDeviceList{
				BlockDeviceList: &ndm.BlockDeviceList{
					TypeMeta: metav1.TypeMeta{},
					ListMeta: metav1.ListMeta{},
					Items: []ndm.BlockDevice{
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path: "/dev/sda",
							},
							Status: ndm.DeviceStatus{
								State: "Inactive",
							},
						},
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path: "/dev/sdb",
							},
							Status: ndm.DeviceStatus{
								State: "Inactive",
							},
						},
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path: "/dev/sdb",
							},
							Status: ndm.DeviceStatus{
								State: "Inactive",
							},
						},
					},
				},
				errs: nil,
			},
			filterPredicate:          []string{FilterInactive},
			expectedBlockDeviceCount: 3,
		},
		"blockDeviceList4": {
			blockDeviceList: &BlockDeviceList{
				BlockDeviceList: &ndm.BlockDeviceList{
					TypeMeta: metav1.TypeMeta{},
					ListMeta: metav1.ListMeta{},
					Items: []ndm.BlockDevice{
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path: "/dev/sda",
							},
							Status: ndm.DeviceStatus{
								State: "Inactive",
							},
						},
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path: "/dev/sdb",
							},
							Status: ndm.DeviceStatus{
								State: "Inactive",
							},
						},
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path: "/dev/sdb",
							},
							Status: ndm.DeviceStatus{
								State: "Inactive",
							},
						},
					},
				},
				errs: nil,
			},
			filterPredicate:          []string{FilterNonInactive},
			expectedBlockDeviceCount: 0,
		},
		"blockDeviceList5": {
			blockDeviceList: &BlockDeviceList{
				BlockDeviceList: &ndm.BlockDeviceList{
					TypeMeta: metav1.TypeMeta{},
					ListMeta: metav1.ListMeta{},
					Items: []ndm.BlockDevice{
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path: "/dev/sda",
							},
							Status: ndm.DeviceStatus{
								State: "Inactive",
							},
						},
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path: "/dev/sdb",
							},
							Status: ndm.DeviceStatus{
								State: "Inactive",
							},
						},
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path: "/dev/sdb",
							},
							Status: ndm.DeviceStatus{
								State: "Inactive",
							},
						},
					},
				},
				errs: nil,
			},
			filterPredicate:          []string{FilterNonInactive, FilterInactive},
			expectedBlockDeviceCount: 0,
		},
		"blockDeviceList6": {
			blockDeviceList: &BlockDeviceList{
				BlockDeviceList: &ndm.BlockDeviceList{
					TypeMeta: metav1.TypeMeta{},
					ListMeta: metav1.ListMeta{},
					Items: []ndm.BlockDevice{
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path: "/dev/sda",
							},
							Status: ndm.DeviceStatus{
								State: "Inactive",
							},
						},
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path: "/dev/sdb",
							},
							Status: ndm.DeviceStatus{
								State: "Inactive",
							},
						},
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path: "/dev/sdb",
							},
							Status: ndm.DeviceStatus{
								State: "Inactive",
							},
						},
					},
				},
				errs: nil,
			},
			filterPredicate:          []string{FilterInactive, FilterNonInactive},
			expectedBlockDeviceCount: 0,
		},
		"blockDeviceList7": {
			blockDeviceList: &BlockDeviceList{
				BlockDeviceList: &ndm.BlockDeviceList{
					TypeMeta: metav1.TypeMeta{},
					ListMeta: metav1.ListMeta{},
					Items: []ndm.BlockDevice{
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path: "/dev/sda",
							},
							Status: ndm.DeviceStatus{
								State: "Inactive",
							},
						},
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path: "/dev/sdb",
							},
							Status: ndm.DeviceStatus{
								State: "Active",
							},
						},
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path: "/dev/sdb",
							},
							Status: ndm.DeviceStatus{
								State: "Inactive",
							},
						},
					},
				},
				errs: nil,
			},
			filterPredicate:          []string{FilterNonInactive},
			expectedBlockDeviceCount: 1,
		},
		"blockDeviceList8": {
			blockDeviceList: &BlockDeviceList{
				BlockDeviceList: &ndm.BlockDeviceList{
					TypeMeta: metav1.TypeMeta{},
					ListMeta: metav1.ListMeta{},
					Items: []ndm.BlockDevice{
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path: "/dev/sda",
							},
							Status: ndm.DeviceStatus{
								State:      "Inactive",
								ClaimState: ndm.BlockDeviceClaimed,
							},
						},
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path: "/dev/sdb",
							},
							Status: ndm.DeviceStatus{
								State:      "Active",
								ClaimState: ndm.BlockDeviceUnclaimed,
							},
						},
						{
							TypeMeta:   metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{},
							Spec: ndm.DeviceSpec{
								Path: "/dev/sdb",
							},
							Status: ndm.DeviceStatus{
								State:      "Inactive",
								ClaimState: ndm.BlockDeviceClaimed,
							},
						},
					},
				},
				errs: nil,
			},
			filterPredicate:          []string{FilterClaimedDevices},
			expectedBlockDeviceCount: 2,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			filtteredBlockDeviceList := test.blockDeviceList.Filter(test.filterPredicate...)
			if len(filtteredBlockDeviceList.Items) != test.expectedBlockDeviceCount {
				t.Errorf("Test %q failed: expected block device object count %d but got %d", name, test.expectedBlockDeviceCount, len(filtteredBlockDeviceList.Items))
			}
		})
	}
}

func TestHasitems(t *testing.T) {
	tests := map[string]struct {
		blockDeviceList BlockDeviceList
		expectedOutput  bool
	}{
		"Nil block device list": {
			blockDeviceList: BlockDeviceList{nil, nil},
			expectedOutput:  false,
		},
		"Empty block device list": {
			blockDeviceList: BlockDeviceList{
				&ndm.BlockDeviceList{},
				nil,
			},
			expectedOutput: false,
		},
		"Empty block device items": {
			blockDeviceList: BlockDeviceList{
				&ndm.BlockDeviceList{
					Items: []ndm.BlockDevice{},
				},
				nil,
			},
			expectedOutput: true,
		},
		"Valid block device list": {
			blockDeviceList: BlockDeviceList{
				&ndm.BlockDeviceList{
					Items: []ndm.BlockDevice{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "blockdevice",
							},
						},
					},
				},
				nil,
			},
			expectedOutput: true,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			_, actual := test.blockDeviceList.Hasitems()
			if actual != test.expectedOutput {
				t.Errorf("Test %q failed expected blockdevice list items: %t and got: %t", name, test.expectedOutput, actual)
			}
		})
	}
}
