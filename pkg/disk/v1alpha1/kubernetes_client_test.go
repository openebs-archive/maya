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
	"k8s.io/client-go/kubernetes/fake"
	"strconv"
	"testing"

	ndmFakeClientset "github.com/openebs/maya/pkg/client/generated/openebs.io/ndm/v1alpha1/clientset/internalclientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGet(t *testing.T) {
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
	tests := map[string]struct {
		diskName string
		err      bool
	}{
		// Test Case #1
		"disk": {
			"mydisk1",
			false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			disk, err := diskK8s.Get(test.diskName)
			gotErr := false
			if err != nil {
				gotErr = true
			}
			if gotErr != test.err {
				t.Fatalf("Test case failed as the expected error %v but got %v: %v", test.err, gotErr, err)
			}
			if disk.Disk == nil {
				t.Errorf("Test case failed as nil disk object")
			}
		})
	}
}

func TestCreate(t *testing.T) {
	var diskK8s DiskInterface
	// Get a fake openebs client set
	focs := &KubernetesClient{
		Kubeclientset: fake.NewSimpleClientset(),
		NDMClientset:  ndmFakeClientset.NewSimpleClientset(),
	}
	diskK8s = focs
	tests := map[string]struct {
		diskName string
		err      bool
	}{
		// Test Case #1
		"disk": {
			"mydisk1",
			false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			diskObj, errs := New().WithName("mydisk1").Build()
			if errs != nil {
				t.Fatalf("Could not build disk object:%s", errs)
			}
			disk, err := diskK8s.Create(diskObj)
			gotErr := false
			if err != nil {
				gotErr = true
			}
			if gotErr != test.err {
				t.Fatalf("Test case failed as the expected error %v but got %v", test.err, err)
			}
			if disk.Disk == nil {
				t.Errorf("Test case failed as nil disk object found")
			}
		})
	}
}

func TestList(t *testing.T) {
	var diskK8s DiskInterface
	// Get a fake openebs client set
	focs := &KubernetesClient{
		Kubeclientset: fake.NewSimpleClientset(),
		NDMClientset:  ndmFakeClientset.NewSimpleClientset(),
	}
	diskK8s = focs
	// Create some disk objects
	for i := 1; i <= 5; i++ {
		diskObj, errs := New().WithName("mydisk" + strconv.Itoa(i)).Build()
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
		diskCount int
		// expectedDiskListLength holds the length of disk list
		err bool
	}{
		// Test Case #1
		"disk": {
			5,
			false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			diskList, err := diskK8s.List(metav1.ListOptions{})
			gotErr := false
			if err != nil {
				gotErr = true
			}
			if gotErr != test.err {
				t.Fatalf("Test case failed as the expected error %v but got %v", test.err, gotErr)
			}
			if len(diskList.Items) != test.diskCount {
				t.Errorf("Test case failed as expected disk object count %d but got %d", test.diskCount, len(diskList.Items))
			}
		})
	}
}

func TestFilteredList(t *testing.T) {
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
		diskCount int
		// expectedDiskListLength holds the length of disk list
		err bool
	}{
		// Test Case #1
		"disk": {
			2,
			false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			diskList, err := diskK8s.List(metav1.ListOptions{})
			filtteredDiskList := diskList.Filter(FilterInactive)
			gotErr := false
			if err != nil {
				gotErr = true
			}
			if gotErr != test.err {
				t.Fatalf("Test case failed as the expected error %v but got %v", test.err, gotErr)
			}
			if len(filtteredDiskList.Items) != test.diskCount {
				t.Errorf("Test case failed as expected disk object count %d but got %d", test.diskCount, len(filtteredDiskList.Items))
			}
		})
	}
}
