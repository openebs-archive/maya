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
	var cspK8s CstorpoolInterface
	// Get a fake openebs client set
	focs := &KubernetesClient{
		Kubeclientset: fake.NewSimpleClientset(),
		Clientset:     openebsFakeClientset.NewSimpleClientset(),
	}
	cspK8s = focs
	cspObj, err := New().WithName("csp1").Build()
	if err != nil {
		t.Fatalf("Could not build csp object:%s", err)
	}
	cspK8s.Create(cspObj)
	tests := map[string]struct {
		cspName string
		err     bool
	}{
		// Test Case #1
		"csp": {
			"csp1",
			false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			csp, err := cspK8s.Get(test.cspName)
			gotErr := false
			if err != nil {
				gotErr = true
			}
			if gotErr != test.err {
				t.Fatalf("Test case failed as the expected error %v but got %v: %v", test.err, gotErr, err)
			}
			if csp.CStorPool == nil {
				t.Errorf("Test case failed as nil csp object")
			}
		})
	}
}

func TestCreate(t *testing.T) {
	var cspK8s CstorpoolInterface
	// Get a fake openebs client set
	focs := &KubernetesClient{
		Kubeclientset: fake.NewSimpleClientset(),
		Clientset:     openebsFakeClientset.NewSimpleClientset(),
	}
	cspK8s = focs
	tests := map[string]struct {
		cspName string
		err     bool
	}{
		// Test Case #1
		"csp": {
			"mycsp1",
			false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cspObj, errs := New().WithName("mycsp1").Build()
			if errs != nil {
				t.Fatalf("Could not build csp object:%s", errs)
			}
			csp, err := cspK8s.Create(cspObj)
			gotErr := false
			if err != nil {
				gotErr = true
			}
			if gotErr != test.err {
				t.Fatalf("Test case failed as the expected error %v but got %v", test.err, err)
			}
			if csp.CStorPool == nil {
				t.Errorf("Test case failed as nil csp object found")
			}
		})
	}
}

func TestList(t *testing.T) {
	var cspK8s CstorpoolInterface
	// Get a fake openebs client set
	focs := &KubernetesClient{
		Kubeclientset: fake.NewSimpleClientset(),
		Clientset:     openebsFakeClientset.NewSimpleClientset(),
	}
	cspK8s = focs
	// Create some csp objects
	for i := 1; i <= 5; i++ {
		cspObj, errs := New().WithName("mycsp" + strconv.Itoa(i)).Build()
		if errs != nil {
			t.Fatalf("Could not build csp object:%s", errs)
		}
		csp, err := cspK8s.Create(cspObj)
		if csp == nil {
			t.Fatalf("Failed to create csp object:%v", err)
		}

	}

	tests := map[string]struct {
		// fakeCasPool holds the fake fakeCasPool object in test cases.
		cspCount int
		// expectedcspListLength holds the length of csp list
		err bool
	}{
		// Test Case #1
		"csp": {
			5,
			false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cspList, err := cspK8s.List(metav1.ListOptions{})
			cspList.filter()
			gotErr := false
			if err != nil {
				gotErr = true
			}
			if gotErr != test.err {
				t.Fatalf("Test case failed as the expected error %v but got %v", test.err, gotErr)
			}
			if len(cspList.CStorPoolList.Items) != test.cspCount {
				t.Errorf("Test case failed as expected csp object count %d but got %d", test.cspCount, len(cspList.CStorPoolList.Items))
			}
		})
	}
}

func TestFilteredList(t *testing.T) {
	var cspK8s CstorpoolInterface
	// Get a fake openebs client set
	focs := &KubernetesClient{
		Kubeclientset: fake.NewSimpleClientset(),
		Clientset:     openebsFakeClientset.NewSimpleClientset(),
	}
	cspK8s = focs
	// Create some csp objects
	for i := 1; i <= 5; i++ {
		var cspPhase string
		if i%2 == 0 {
			cspPhase = "Healthy"
		} else {
			cspPhase = "Degraded"
		}
		cspObj, errs := New().WithName("mycsp" + strconv.Itoa(i)).WithPhase(cspPhase).Build()
		if errs != nil {
			t.Fatalf("Could not build csp object:%s", errs)
		}
		csp, err := cspK8s.Create(cspObj)
		if csp == nil {
			t.Fatalf("Failed to create csp object:%v", err)
		}
	}

	tests := map[string]struct {
		// fakeCasPool holds the fake fakeCasPool object in test cases.
		cspCount int
		// expectedcspListLength holds the length of csp list
		err bool
	}{
		// Test Case #1
		"csp": {
			2,
			false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cspList, err := cspK8s.List(metav1.ListOptions{})
			cspList.filter(filterHealthyPredicateKey)
			gotErr := false
			if err != nil {
				gotErr = true
			}
			if gotErr != test.err {
				t.Fatalf("Test case failed as the expected error %v but got %v", test.err, gotErr)
			}
			if len(cspList.FilteredCStorPoolList.Items) != test.cspCount {
				t.Errorf("Test case failed as expected csp object count %d but got %d", test.cspCount, len(cspList.FilteredCStorPoolList.Items))
			}
		})
	}
}
