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
	apis "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestFilter(t *testing.T) {
	tests := map[string]struct {
		// fakeCasPool holds the fake fakeCasPool object in test cases.
		diskList        *DiskList
		filterPredicate []string
		// expectedDiskListLength holds the length of disk list
		expectedDiskCount int
	}{
		"EmptyDiskList1": {
			diskList:          nil,
			filterPredicate:   []string{FilterInactive},
			expectedDiskCount: 0,
		},
		"EmptyDiskList2": {
			diskList: &DiskList{
				DiskList: nil,
				errs:     nil,
			},
			filterPredicate:   []string{FilterInactive},
			expectedDiskCount: 0,
		},
		"EmptyDiskList3": {
			diskList: &DiskList{
				DiskList: &apis.DiskList{},
				errs:     nil,
			},
			filterPredicate:   []string{FilterInactive},
			expectedDiskCount: 0,
		},
		// Test Case #1
		"diskList3": {
			diskList: &DiskList{
				DiskList: &apis.DiskList{
					TypeMeta: v1.TypeMeta{},
					ListMeta: v1.ListMeta{},
					Items: []apis.Disk{
						{
							TypeMeta:   v1.TypeMeta{},
							ObjectMeta: v1.ObjectMeta{},
							Spec: apis.DiskSpec{
								Path: "/dev/sda",
							},
							Status: apis.DiskStatus{
								State: "Inactive",
							},
						},
						{
							TypeMeta:   v1.TypeMeta{},
							ObjectMeta: v1.ObjectMeta{},
							Spec: apis.DiskSpec{
								Path: "/dev/sda",
							},
							Status: apis.DiskStatus{
								State: "Inactive",
							},
						},
						{
							TypeMeta:   v1.TypeMeta{},
							ObjectMeta: v1.ObjectMeta{},
							Spec: apis.DiskSpec{
								Path: "/dev/sda",
							},
							Status: apis.DiskStatus{
								State: "Inactive",
							},
						},
					},
				},
				errs: nil,
			},
			filterPredicate:   []string{FilterInactive},
			expectedDiskCount: 3,
		},
		"diskList4": {
			diskList: &DiskList{
				DiskList: &apis.DiskList{
					TypeMeta: v1.TypeMeta{},
					ListMeta: v1.ListMeta{},
					Items: []apis.Disk{
						{
							TypeMeta:   v1.TypeMeta{},
							ObjectMeta: v1.ObjectMeta{},
							Spec: apis.DiskSpec{
								Path: "/dev/sda",
							},
							Status: apis.DiskStatus{
								State: "Inactive",
							},
						},
						{
							TypeMeta:   v1.TypeMeta{},
							ObjectMeta: v1.ObjectMeta{},
							Spec: apis.DiskSpec{
								Path: "/dev/sda",
							},
							Status: apis.DiskStatus{
								State: "Inactive",
							},
						},
						{
							TypeMeta:   v1.TypeMeta{},
							ObjectMeta: v1.ObjectMeta{},
							Spec: apis.DiskSpec{
								Path: "/dev/sda",
							},
							Status: apis.DiskStatus{
								State: "Inactive",
							},
						},
					},
				},
				errs: nil,
			},
			filterPredicate:   []string{FilterInactiveReverse},
			expectedDiskCount: 0,
		},
		"diskList5": {
			diskList: &DiskList{
				DiskList: &apis.DiskList{
					TypeMeta: v1.TypeMeta{},
					ListMeta: v1.ListMeta{},
					Items: []apis.Disk{
						{
							TypeMeta:   v1.TypeMeta{},
							ObjectMeta: v1.ObjectMeta{},
							Spec: apis.DiskSpec{
								Path: "/dev/sda",
							},
							Status: apis.DiskStatus{
								State: "Inactive",
							},
						},
						{
							TypeMeta:   v1.TypeMeta{},
							ObjectMeta: v1.ObjectMeta{},
							Spec: apis.DiskSpec{
								Path: "/dev/sda",
							},
							Status: apis.DiskStatus{
								State: "Inactive",
							},
						},
						{
							TypeMeta:   v1.TypeMeta{},
							ObjectMeta: v1.ObjectMeta{},
							Spec: apis.DiskSpec{
								Path: "/dev/sda",
							},
							Status: apis.DiskStatus{
								State: "Inactive",
							},
						},
					},
				},
				errs: nil,
			},
			filterPredicate:   []string{FilterInactiveReverse, FilterInactive},
			expectedDiskCount: 0,
		},
		"diskList6": {
			diskList: &DiskList{
				DiskList: &apis.DiskList{
					TypeMeta: v1.TypeMeta{},
					ListMeta: v1.ListMeta{},
					Items: []apis.Disk{
						{
							TypeMeta:   v1.TypeMeta{},
							ObjectMeta: v1.ObjectMeta{},
							Spec: apis.DiskSpec{
								Path: "/dev/sda",
							},
							Status: apis.DiskStatus{
								State: "Inactive",
							},
						},
						{
							TypeMeta:   v1.TypeMeta{},
							ObjectMeta: v1.ObjectMeta{},
							Spec: apis.DiskSpec{
								Path: "/dev/sda",
							},
							Status: apis.DiskStatus{
								State: "Inactive",
							},
						},
						{
							TypeMeta:   v1.TypeMeta{},
							ObjectMeta: v1.ObjectMeta{},
							Spec: apis.DiskSpec{
								Path: "/dev/sda",
							},
							Status: apis.DiskStatus{
								State: "Inactive",
							},
						},
					},
				},
				errs: nil,
			},
			filterPredicate:   []string{FilterInactive, FilterInactiveReverse},
			expectedDiskCount: 0,
		},
		"diskList7": {
			diskList: &DiskList{
				DiskList: &apis.DiskList{
					TypeMeta: v1.TypeMeta{},
					ListMeta: v1.ListMeta{},
					Items: []apis.Disk{
						{
							TypeMeta:   v1.TypeMeta{},
							ObjectMeta: v1.ObjectMeta{},
							Spec: apis.DiskSpec{
								Path: "/dev/sda",
							},
							Status: apis.DiskStatus{
								State: "Inactive",
							},
						},
						{
							TypeMeta:   v1.TypeMeta{},
							ObjectMeta: v1.ObjectMeta{},
							Spec: apis.DiskSpec{
								Path: "/dev/sda",
							},
							Status: apis.DiskStatus{
								State: "Inactive",
							},
						},
						{
							TypeMeta:   v1.TypeMeta{},
							ObjectMeta: v1.ObjectMeta{},
							Spec: apis.DiskSpec{
								Path: "/dev/sda",
							},
							Status: apis.DiskStatus{
								State: "Inactive",
							},
						},
					},
				},
				errs: nil,
			},
			filterPredicate:   []string{FilterInactive, FilterInactiveReverse},
			expectedDiskCount: 0,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			filtteredDiskList := test.diskList.Filter(test.filterPredicate...)
			if len(filtteredDiskList.Items) != test.expectedDiskCount {
				t.Errorf("Test case failed as expected disk object count %d but got %d", test.expectedDiskCount, len(filtteredDiskList.Items))
			}
		})
	}
}

func TestFilterAny(t *testing.T) {
	tests := map[string]struct {
		// fakeCasPool holds the fake fakeCasPool object in test cases.
		diskList        *DiskList
		filterPredicate []string
		// expectedDiskListLength holds the length of disk list
		expectedDiskCount int
	}{
		"EmptyDiskList1": {
			diskList:          nil,
			filterPredicate:   []string{FilterInactive},
			expectedDiskCount: 0,
		},
		"EmptyDiskList2": {
			diskList: &DiskList{
				DiskList: nil,
				errs:     nil,
			},
			filterPredicate:   []string{FilterInactive},
			expectedDiskCount: 0,
		},
		"EmptyDiskList3": {
			diskList: &DiskList{
				DiskList: &apis.DiskList{},
				errs:     nil,
			},
			filterPredicate:   []string{FilterInactive},
			expectedDiskCount: 0,
		},
		// Test Case #1
		"diskList3": {
			diskList: &DiskList{
				DiskList: &apis.DiskList{
					TypeMeta: v1.TypeMeta{},
					ListMeta: v1.ListMeta{},
					Items: []apis.Disk{
						{
							TypeMeta:   v1.TypeMeta{},
							ObjectMeta: v1.ObjectMeta{},
							Spec: apis.DiskSpec{
								Path: "/dev/sda",
							},
							Status: apis.DiskStatus{
								State: "Inactive",
							},
						},
						{
							TypeMeta:   v1.TypeMeta{},
							ObjectMeta: v1.ObjectMeta{},
							Spec: apis.DiskSpec{
								Path: "/dev/sda",
							},
							Status: apis.DiskStatus{
								State: "Inactive",
							},
						},
						{
							TypeMeta:   v1.TypeMeta{},
							ObjectMeta: v1.ObjectMeta{},
							Spec: apis.DiskSpec{
								Path: "/dev/sda",
							},
							Status: apis.DiskStatus{
								State: "Inactive",
							},
						},
					},
				},
				errs: nil,
			},
			filterPredicate:   []string{FilterInactive},
			expectedDiskCount: 3,
		},
		"diskList4": {
			diskList: &DiskList{
				DiskList: &apis.DiskList{
					TypeMeta: v1.TypeMeta{},
					ListMeta: v1.ListMeta{},
					Items: []apis.Disk{
						{
							TypeMeta:   v1.TypeMeta{},
							ObjectMeta: v1.ObjectMeta{},
							Spec: apis.DiskSpec{
								Path: "/dev/sda",
							},
							Status: apis.DiskStatus{
								State: "Inactive",
							},
						},
						{
							TypeMeta:   v1.TypeMeta{},
							ObjectMeta: v1.ObjectMeta{},
							Spec: apis.DiskSpec{
								Path: "/dev/sda",
							},
							Status: apis.DiskStatus{
								State: "Inactive",
							},
						},
						{
							TypeMeta:   v1.TypeMeta{},
							ObjectMeta: v1.ObjectMeta{},
							Spec: apis.DiskSpec{
								Path: "/dev/sda",
							},
							Status: apis.DiskStatus{
								State: "Inactive",
							},
						},
					},
				},
				errs: nil,
			},
			filterPredicate:   []string{FilterInactiveReverse},
			expectedDiskCount: 0,
		},
		"diskList5": {
			diskList: &DiskList{
				DiskList: &apis.DiskList{
					TypeMeta: v1.TypeMeta{},
					ListMeta: v1.ListMeta{},
					Items: []apis.Disk{
						{
							TypeMeta:   v1.TypeMeta{},
							ObjectMeta: v1.ObjectMeta{},
							Spec: apis.DiskSpec{
								Path: "/dev/sda",
							},
							Status: apis.DiskStatus{
								State: "Inactive",
							},
						},
						{
							TypeMeta:   v1.TypeMeta{},
							ObjectMeta: v1.ObjectMeta{},
							Spec: apis.DiskSpec{
								Path: "/dev/sda",
							},
							Status: apis.DiskStatus{
								State: "Inactive",
							},
						},
						{
							TypeMeta:   v1.TypeMeta{},
							ObjectMeta: v1.ObjectMeta{},
							Spec: apis.DiskSpec{
								Path: "/dev/sda",
							},
							Status: apis.DiskStatus{
								State: "Inactive",
							},
						},
					},
				},
				errs: nil,
			},
			filterPredicate:   []string{FilterInactiveReverse, FilterInactive},
			expectedDiskCount: 3,
		},
		"diskList6": {
			diskList: &DiskList{
				DiskList: &apis.DiskList{
					TypeMeta: v1.TypeMeta{},
					ListMeta: v1.ListMeta{},
					Items: []apis.Disk{
						{
							TypeMeta:   v1.TypeMeta{},
							ObjectMeta: v1.ObjectMeta{},
							Spec: apis.DiskSpec{
								Path: "/dev/sda",
							},
							Status: apis.DiskStatus{
								State: "Inactive",
							},
						},
						{
							TypeMeta:   v1.TypeMeta{},
							ObjectMeta: v1.ObjectMeta{},
							Spec: apis.DiskSpec{
								Path: "/dev/sda",
							},
							Status: apis.DiskStatus{
								State: "Inactive",
							},
						},
						{
							TypeMeta:   v1.TypeMeta{},
							ObjectMeta: v1.ObjectMeta{},
							Spec: apis.DiskSpec{
								Path: "/dev/sda",
							},
							Status: apis.DiskStatus{
								State: "Inactive",
							},
						},
					},
				},
				errs: nil,
			},
			filterPredicate:   []string{FilterInactive, FilterInactiveReverse},
			expectedDiskCount: 3,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			filtteredDiskList := test.diskList.FilterAny(test.filterPredicate...)
			if len(filtteredDiskList.Items) != test.expectedDiskCount {
				t.Errorf("Test case failed as expected disk object count %d but got %d", test.expectedDiskCount, len(filtteredDiskList.Items))
			}
		})
	}
}
