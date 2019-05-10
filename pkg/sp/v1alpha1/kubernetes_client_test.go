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

	openebsFakeClientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGet(t *testing.T) {
	var spK8s StoragepoolInterface
	// Get a fake openebs client set
	focs := &KubernetesClient{
		Kubeclientset: fake.NewSimpleClientset(),
		Clientset:     openebsFakeClientset.NewSimpleClientset(),
	}
	spK8s = focs
	spObj, err := New().WithName("sp1").Build()
	if err != nil {
		t.Fatalf("Could not build sp object:%s", err)
	}
	spK8s.Create(spObj)
	tests := map[string]struct {
		spName string
		err    bool
	}{
		// Test Case #1
		"sp": {
			"sp1",
			false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			sp, err := spK8s.Get(test.spName)
			gotErr := false
			if err != nil {
				gotErr = true
			}
			if gotErr != test.err {
				t.Fatalf("Test case failed as the expected error %v but got %v: %v", test.err, gotErr, err)
			}
			if sp.StoragePool == nil {
				t.Errorf("Test case failed as nil sp object")
			}
		})
	}
}

func TestCreate(t *testing.T) {
	var spK8s StoragepoolInterface
	// Get a fake openebs client set
	focs := &KubernetesClient{
		Kubeclientset: fake.NewSimpleClientset(),
		Clientset:     openebsFakeClientset.NewSimpleClientset(),
	}
	spK8s = focs
	tests := map[string]struct {
		spName string
		err    bool
	}{
		// Test Case #1
		"sp": {
			"mysp1",
			false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			spObj, errs := New().WithName("mysp1").Build()
			if errs != nil {
				t.Fatalf("Could not build sp object:%s", errs)
			}
			sp, err := spK8s.Create(spObj)
			gotErr := false
			if err != nil {
				gotErr = true
			}
			if gotErr != test.err {
				t.Fatalf("Test case failed as the expected error %v but got %v", test.err, err)
			}
			if sp.StoragePool == nil {
				t.Errorf("Test case failed as nil sp object found")
			}
		})
	}
}

func TestList(t *testing.T) {
	var spK8s StoragepoolInterface
	// Get a fake openebs client set
	focs := &KubernetesClient{
		Kubeclientset: fake.NewSimpleClientset(),
		Clientset:     openebsFakeClientset.NewSimpleClientset(),
	}
	spK8s = focs
	// Create some sp objects
	for i := 1; i <= 5; i++ {
		spObj, errs := New().WithName("mysp" + strconv.Itoa(i)).Build()
		if errs != nil {
			t.Fatalf("Could not build sp object:%s", errs)
		}
		sp, err := spK8s.Create(spObj)
		if sp == nil {
			t.Fatalf("Failed to create sp object:%v", err)
		}

	}

	tests := map[string]struct {
		// fakeCasPool holds the fake fakeCasPool object in test cases.
		spCount int
		// expectedspListLength holds the length of sp list
		err bool
	}{
		// Test Case #1
		"sp": {
			5,
			false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			spList, err := spK8s.List(metav1.ListOptions{})
			spList.filter()
			gotErr := false
			if err != nil {
				gotErr = true
			}
			if gotErr != test.err {
				t.Fatalf("Test case failed as the expected error %v but got %v", test.err, gotErr)
			}
			if len(spList.StoragePoolList.Items) != test.spCount {
				t.Errorf("Test case failed as expected sp object count %d but got %d", test.spCount, len(spList.StoragePoolList.Items))
			}
		})
	}
}

func TestFilteredList(t *testing.T) {
	var spK8s StoragepoolInterface
	// Get a fake openebs client set
	focs := &KubernetesClient{
		Kubeclientset: fake.NewSimpleClientset(),
		Clientset:     openebsFakeClientset.NewSimpleClientset(),
	}
	spK8s = focs
	// Create some sp objects
	for i := 1; i <= 5; i++ {
		var poolType string
		if i%2 == 0 {
			poolType = "striped"
		} else {
			poolType = "mirrored"
		}
		spObj, errs := New().WithName("mysp" + strconv.Itoa(i)).WithPoolType(poolType).Build()
		if errs != nil {
			t.Fatalf("Could not build sp object:%s", errs)
		}
		sp, err := spK8s.Create(spObj)
		if sp == nil {
			t.Fatalf("Failed to create sp object:%v", err)
		}
	}

	tests := map[string]struct {
		// fakeCasPool holds the fake fakeCasPool object in test cases.
		spCount int
		// expectedspListLength holds the length of sp list
		err bool
	}{
		// Test Case #1
		"sp": {
			2,
			false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			spList, err := spK8s.List(metav1.ListOptions{})
			spList.filter(filterStripedPredicateKey)
			gotErr := false
			if err != nil {
				gotErr = true
			}
			if gotErr != test.err {
				t.Fatalf("Test case failed as the expected error %v but got %v", test.err, gotErr)
			}
			if len(spList.FilteredStoragePoolList.Items) != test.spCount {
				t.Errorf("Test case failed as expected sp object count %d but got %d", test.spCount, len(spList.FilteredStoragePoolList.Items))
			}
		})
	}
}
