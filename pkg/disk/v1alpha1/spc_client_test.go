/*
Copyright 2018 The OpenEBS Authors

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
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	ndmFakeClientset "github.com/openebs/maya/pkg/client/generated/openebs.io/ndm/v1alpha1/clientset/internalclientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"strconv"
	"testing"
)

func TestSpcClientGet(t *testing.T) {
	var diskK8s DiskInterface
	var diskSpc DiskInterface
	// Get a fake openebs client set
	focs := &KubernetesClient{
		Kubeclientset: fake.NewSimpleClientset(),
		NDMClientset:  ndmFakeClientset.NewSimpleClientset(),
	}
	spcClient := &SpcObjectClient{
		KubernetesClient: focs,
		Spc: &apis.StoragePoolClaim{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name: "pool1",
			},
			Spec: apis.StoragePoolClaimSpec{
				Disks: apis.DiskAttr{
					DiskList: []string{"mydisk1", "mydisk4"},
				},
			},
		},
	}
	diskK8s = focs
	diskSpc = spcClient
	diskObj, err := New().WithName("mydisk1").Build()
	if err != nil {
		t.Fatalf("Could not build disk object:%s", err)
	}
	diskK8s.Create(diskObj)
	diskObj, err = New().WithName("mydisk2").Build()
	if err != nil {
		t.Fatalf("Could not build disk object:%s", err)
	}
	diskK8s.Create(diskObj)
	tests := map[string]struct {
		diskName         string
		expectedDiskName string
		expectedErr      bool
	}{
		"disk present in spc as well as k8s": {
			diskName:         "mydisk1",
			expectedDiskName: "mydisk1",
			expectedErr:      false,
		},
		"disk not present in both spc and k8s": {
			diskName:         "mydisk2",
			expectedDiskName: "mydisk2",
			expectedErr:      true,
		},
		"disk not present in spc but present in k8s": {
			diskName:         "mydisk3",
			expectedDiskName: "mydisk3",
			expectedErr:      true,
		},
		"disk present in spc but not present in k8s": {
			diskName:         "mydisk4",
			expectedDiskName: "mydisk4",
			expectedErr:      true,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			disk, err := diskSpc.Get(test.diskName)
			if test.expectedErr {
				if err == nil {
					t.Error("Test case failed as got nil error")
				}
			} else {

				if disk.Disk == nil {
					t.Fatalf("Test case failed as got nil disk:%v", err)
				}
				if disk.Disk.Name != test.expectedDiskName {
					t.Errorf("Test case failed as expected disk name %s but got %s", disk.Disk.Name, test.expectedDiskName)
				}
			}

		})
	}
}

func TestSpcClientList(t *testing.T) {
	var diskK8s DiskInterface
	// Get a fake openebs client set
	focs := &KubernetesClient{
		Kubeclientset: fake.NewSimpleClientset(),
		NDMClientset:  ndmFakeClientset.NewSimpleClientset(),
	}
	diskK8s = focs
	diskObj, err := New().WithName("mydisk1").Build()
	if err != nil {
		t.Fatalf("Could not build disk object:%s", err)
	}
	diskK8s.Create(diskObj)
	diskObj, err = New().WithName("mydisk2").Build()
	if err != nil {
		t.Fatalf("Could not build disk object:%s", err)
	}
	diskK8s.Create(diskObj)

	tests := map[string]struct {
		// fakeCasPool holds the fake fakeCasPool object in test cases.
		diskCount    int
		spcClientObj *SpcObjectClient
		// expectedDiskListLength holds the length of disk list
		expectedErr bool
	}{
		// Test Case #1
		"All disks specified in spc is present in k8s": {
			diskCount: 2,
			spcClientObj: &SpcObjectClient{
				KubernetesClient: focs,
				Spc: &apis.StoragePoolClaim{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Name: "pool1",
					},
					Spec: apis.StoragePoolClaimSpec{
						Disks: apis.DiskAttr{
							DiskList: []string{"mydisk1", "mydisk2"},
						},
					},
				},
			},
			expectedErr: false,
		},
		"Some disks specified in spc is present in k8s": {
			diskCount: 2,
			spcClientObj: &SpcObjectClient{
				KubernetesClient: focs,
				Spc: &apis.StoragePoolClaim{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Name: "pool1",
					},
					Spec: apis.StoragePoolClaimSpec{
						Disks: apis.DiskAttr{
							DiskList: []string{"mydisk1", "mydisk2", "mydisk3", "mydisk4"},
						},
					},
				},
			},
			expectedErr: false,
		},
		"All disks specified in spc is not present in k8s": {
			diskCount: 0,
			spcClientObj: &SpcObjectClient{
				KubernetesClient: focs,
				Spc: &apis.StoragePoolClaim{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Name: "pool1",
					},
					Spec: apis.StoragePoolClaimSpec{
						Disks: apis.DiskAttr{
							DiskList: []string{"mydisk3", "mydisk4", "mydisk5", "mydisk6"},
						},
					},
				},
			},
			expectedErr: false,
		},
		"No disks specified in spc": {
			diskCount: 0,
			spcClientObj: &SpcObjectClient{
				KubernetesClient: focs,
				Spc: &apis.StoragePoolClaim{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Name: "pool1",
					},
					Spec: apis.StoragePoolClaimSpec{},
				},
			},
			expectedErr: true,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			diskList, err := test.spcClientObj.List(metav1.ListOptions{})
			gotErr := false
			if err != nil {
				gotErr = true
			}
			if test.expectedErr {
				if gotErr != test.expectedErr {
					t.Errorf("Test case failed as expected error %t but got %t", test.expectedErr, gotErr)
				}
			} else {
				if gotErr != test.expectedErr {
					t.Fatalf("Test case failed as the expected error %t but got %t:%v", test.expectedErr, gotErr, err)
				}
				if len(diskList.DiskList.Items) != test.diskCount {
					t.Errorf("Test case failed as expected disk object count %d but got %d", test.diskCount, len(diskList.DiskList.Items))
				}
			}

		})
	}
}

func TestSpcClientFilteredList(t *testing.T) {
	var diskK8s DiskInterface
	// Get a fake openebs client set
	focs := &KubernetesClient{
		Kubeclientset: fake.NewSimpleClientset(),
		NDMClientset:  ndmFakeClientset.NewSimpleClientset(),
	}
	diskK8s = focs
	// Create some disk objects
	for i := 1; i <= 5; i++ {
		var diskState string
		if i%2 == 0 {
			diskState = "Inactive"
		} else {
			diskState = "Active"
		}
		diskObj, errs := New().WithName("mydisk" + strconv.Itoa(i)).WithState(diskState).Build()
		if errs != nil {
			t.Fatalf("Could not build disk object:%s", errs)
		}
		disk, err := diskK8s.Create(diskObj)
		if disk == nil {
			t.Fatalf("Failed to create disk object:%v", err)
		}
	}

	tests := map[string]struct {
		// fakeCasPool holds the fake fakeCasPool object in test cases.
		diskCount    int
		spcClientObj *SpcObjectClient
		filterWith   []string
		// expectedDiskListLength holds the length of disk list
		expectedErr bool
	}{
		// Test Case #1
		"All disks specified in spc is present in k8s": {
			diskCount: 2,
			spcClientObj: &SpcObjectClient{
				KubernetesClient: focs,
				Spc: &apis.StoragePoolClaim{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Name: "pool1",
					},
					Spec: apis.StoragePoolClaimSpec{
						Disks: apis.DiskAttr{
							DiskList: []string{"mydisk1", "mydisk2", "mydisk3", "mydisk4", "mydisk5"},
						},
					},
				},
			},
			expectedErr: false,
			filterWith:  []string{FilterInactive},
		},
		"Some disks specified in spc is present in k8s": {
			diskCount: 1,
			spcClientObj: &SpcObjectClient{
				KubernetesClient: focs,
				Spc: &apis.StoragePoolClaim{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Name: "pool1",
					},
					Spec: apis.StoragePoolClaimSpec{
						Disks: apis.DiskAttr{
							DiskList: []string{"mydisk1", "mydisk2", "mydisk3", "mydisk6"},
						},
					},
				},
			},
			expectedErr: false,
		},
		"All disks specified in spc is not present in k8s": {
			diskCount: 0,
			spcClientObj: &SpcObjectClient{
				KubernetesClient: focs,
				Spc: &apis.StoragePoolClaim{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Name: "pool1",
					},
					Spec: apis.StoragePoolClaimSpec{
						Disks: apis.DiskAttr{
							DiskList: []string{"mydisk6", "mydisk7", "mydisk8", "mydisk9"},
						},
					},
				},
			},
			expectedErr: false,
		},
		"No disks specified in spc": {
			diskCount: 0,
			spcClientObj: &SpcObjectClient{
				KubernetesClient: focs,
				Spc: &apis.StoragePoolClaim{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Name: "pool1",
					},
					Spec: apis.StoragePoolClaimSpec{},
				},
			},
			expectedErr: true,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			diskList, err := test.spcClientObj.List(metav1.ListOptions{})
			diskList = diskList.Filter(FilterInactive)
			gotErr := false
			if err != nil {
				gotErr = true
			}
			if gotErr != test.expectedErr {
				t.Fatalf("Test case failed as the expected error %v but got %v", test.expectedErr, gotErr)
			}
			if len(diskList.DiskList.Items) != test.diskCount {
				t.Errorf("Test case failed as expected disk object count %d but got %d", test.diskCount, len(diskList.DiskList.Items))
			}
		})
	}
}
