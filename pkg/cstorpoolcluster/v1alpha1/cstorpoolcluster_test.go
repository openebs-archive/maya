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

package v1beta1

import (
	apisv1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"testing"
)

func mockAlwaysTrue(*CSPC) bool  { return true }
func mockAlwaysFalse(*CSPC) bool { return false }

func TestAll(t *testing.T) {
	tests := map[string]struct {
		Predicates     predicateList
		expectedOutput bool
	}{
		// Positive predicates
		"Positive Predicate 1": {[]Predicate{mockAlwaysTrue}, true},
		"Positive Predicate 2": {[]Predicate{mockAlwaysTrue, mockAlwaysTrue}, true},
		"Positive Predicate 3": {[]Predicate{mockAlwaysTrue, mockAlwaysTrue, mockAlwaysTrue}, true},
		// Negative Predicates
		"Negative Predicate 1": {[]Predicate{mockAlwaysFalse}, false},
		"Negative Predicate 2": {[]Predicate{mockAlwaysTrue, mockAlwaysFalse}, false},
		"Negative Predicate 3": {[]Predicate{mockAlwaysFalse, mockAlwaysTrue}, false},
		"Negative Predicate 4": {[]Predicate{mockAlwaysFalse, mockAlwaysFalse}, false},
		"Negative Predicate 5": {[]Predicate{mockAlwaysFalse, mockAlwaysTrue, mockAlwaysTrue}, false},
		"Negative Predicate 6": {[]Predicate{mockAlwaysTrue, mockAlwaysFalse, mockAlwaysTrue}, false},
		"Negative Predicate 7": {[]Predicate{mockAlwaysTrue, mockAlwaysTrue, mockAlwaysFalse}, false},
		"Negative Predicate 8": {[]Predicate{mockAlwaysTrue, mockAlwaysFalse, mockAlwaysFalse}, false},
		"Negative Predicate 9": {[]Predicate{mockAlwaysFalse, mockAlwaysFalse, mockAlwaysFalse}, false},
	}
	for name, mock := range tests {
		// pin the variables
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			if output := mock.Predicates.all(&CSPC{}); output != mock.expectedOutput {
				t.Fatalf("test %q failed: expected %v \n got : %v \n", name, mock.expectedOutput, output)
			}
		})
	}
}

func TestHasAnnotation(t *testing.T) {
	tests := map[string]struct {
		availableAnnotations       map[string]string
		checkForKey, checkForValue string
		hasAnnotation              bool
	}{
		"Test 1": {map[string]string{"Anno 1": "Val 1"}, "Anno 1", "Val 1", true},
		"Test 2": {map[string]string{"Anno 1": "Val 1"}, "Anno 1", "Val 2", false},
		"Test 3": {map[string]string{"Anno 1": "Val 1", "Anno 2": "Val 2"}, "Anno 0", "Val 2", false},
		"Test 4": {map[string]string{"Anno 1": "Val 1", "Anno 2": "Val 2"}, "Anno 1", "Val 1", true},
		"Test 5": {map[string]string{"Anno 1": "Val 1", "Anno 2": "Val 2", "Anno 3": "Val 3"}, "Anno 1", "Val 1", true},
	}

	for name, test := range tests {
		// pin the variables
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			fakecspc := &CSPC{&apisv1alpha1.CStorPoolCluster{ObjectMeta: metav1.ObjectMeta{Annotations: test.availableAnnotations}}}
			ok := HasAnnotation(test.checkForKey, test.checkForValue)(fakecspc)
			if ok != test.hasAnnotation {
				t.Fatalf("Test %v failed, Expected %v but got %v", name, test.availableAnnotations, fakecspc.Object.GetAnnotations())
			}
		})
	}
}

func TestIsProvisioningAuto(t *testing.T) {
	tests := map[string]struct {
		nodeSpec           []apisv1alpha1.CStorPoolClusterNodeSpec
		isAutoProvisioning bool
	}{
		"Test#1:Empty Node Spec": {[]apisv1alpha1.CStorPoolClusterNodeSpec{}, true},

		"Test#2:Non-empty Node Spec with empty items": {[]apisv1alpha1.CStorPoolClusterNodeSpec{
			{}, {},
		}, false},
	}

	for name, test := range tests {
		// pin the variables
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			fakecspc := &CSPC{&apisv1alpha1.CStorPoolCluster{Spec: apisv1alpha1.CStorPoolClusterSpec{Nodes: test.nodeSpec}}}
			ok := IsProvisioningAuto()(fakecspc)
			if ok != test.isAutoProvisioning {
				t.Fatalf("Test %v failed, Expected %v but got %v", name, test.isAutoProvisioning, ok)
			}
		})
	}
}

func TestIsProvisioningManual(t *testing.T) {
	tests := map[string]struct {
		nodeSpec           []apisv1alpha1.CStorPoolClusterNodeSpec
		isAutoProvisioning bool
	}{
		"Test#1:Empty Node Spec": {[]apisv1alpha1.CStorPoolClusterNodeSpec{}, false},

		"Test#2:Non-empty Node Spec with empty items": {[]apisv1alpha1.CStorPoolClusterNodeSpec{
			{}, {},
		}, true},
	}

	for name, test := range tests {
		// pin the variables
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			fakecspc := &CSPC{&apisv1alpha1.CStorPoolCluster{Spec: apisv1alpha1.CStorPoolClusterSpec{Nodes: test.nodeSpec}}}
			ok := IsProvisioningManual()(fakecspc)
			if ok != test.isAutoProvisioning {
				t.Fatalf("Test %v failed, Expected %v but got %v", name, test.isAutoProvisioning, ok)
			}
		})
	}
}

func TestIsSparse(t *testing.T) {
	tests := map[string]struct {
		diskType     string
		isSparseType bool
	}{
		"Test#1": {"sparse", true},

		"Test#2": {"disk", false},

		"Test#3": {"invalid_value", false},
	}

	for name, test := range tests {
		// pin the variables
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			fakecspc := &CSPC{&apisv1alpha1.CStorPoolCluster{Spec: apisv1alpha1.CStorPoolClusterSpec{Type: test.diskType}}}
			ok := IsSparse()(fakecspc)
			if ok != test.isSparseType {
				t.Fatalf("Test %v failed, Expected %v but got %v", name, test.isSparseType, ok)
			}
		})
	}
}

func TestIsDisk(t *testing.T) {
	tests := map[string]struct {
		diskType   string
		isDiskType bool
	}{
		"Test#1": {"sparse", false},

		"Test#2": {"disk", true},

		"Test#3": {"invalid_value", false},
	}

	for name, test := range tests {
		// pin the variables
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			fakecspc := &CSPC{&apisv1alpha1.CStorPoolCluster{Spec: apisv1alpha1.CStorPoolClusterSpec{Type: test.diskType}}}
			ok := IsDisk()(fakecspc)
			if ok != test.isDiskType {
				t.Fatalf("Test %v failed, Expected %v but got %v", name, test.isDiskType, ok)
			}
		})
	}
}

func TestCSPC_GetAnnotations(t *testing.T) {
	tests := map[string]struct {
		availableAnnotations map[string]string
		expectedAnnotations  map[string]string
	}{
		"Test 1": {map[string]string{"Anno 1": "Val 1"}, map[string]string{"Anno 1": "Val 1"}},
		"Test 2": {map[string]string{"Anno 1": "Val 1", "Anno 2": "Val 2", "Anno 3": "Val 3"}, map[string]string{"Anno 1": "Val 1", "Anno 2": "Val 2", "Anno 3": "Val 3"}},
		"Test 3": {nil, nil},
	}

	for name, test := range tests {
		// pin the variables
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			fakecspc := &CSPC{&apisv1alpha1.CStorPoolCluster{ObjectMeta: metav1.ObjectMeta{Annotations: test.availableAnnotations}}}
			gotAnnotations := fakecspc.GetAnnotations()
			if !reflect.DeepEqual(gotAnnotations, test.expectedAnnotations) {
				t.Fatalf("Test %v failed, Expected %v but got %v", name, test.expectedAnnotations, gotAnnotations)
			}
		})
	}
}

func TestCSPC_GetNodeDisk(t *testing.T) {
	tests := map[string]struct {
		cspcNodeSpec      []apisv1alpha1.CStorPoolClusterNodeSpec
		expectedNodeNames []string
	}{
		"Test 1": {[]apisv1alpha1.CStorPoolClusterNodeSpec{{Name: "worker-node-1"}, {Name: "worker-node-2"}}, []string{"worker-node-1", "worker-node-2"}},
		"Test 2": {[]apisv1alpha1.CStorPoolClusterNodeSpec{{}}, []string{""}},
		"Test 3": {[]apisv1alpha1.CStorPoolClusterNodeSpec{}, nil},
		"Test 4": {[]apisv1alpha1.CStorPoolClusterNodeSpec{{}, {}}, []string{"", ""}},
	}

	for name, test := range tests {
		// pin the variables
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			fakecspc := &CSPC{&apisv1alpha1.CStorPoolCluster{Spec: apisv1alpha1.CStorPoolClusterSpec{Nodes: test.cspcNodeSpec}}}
			gotNodeNames := fakecspc.GetNodeNames()
			if !reflect.DeepEqual(gotNodeNames, test.expectedNodeNames) {
				t.Fatalf("Test %v failed, Expected %v but got %v", name, test.expectedNodeNames, gotNodeNames)
			}
		})
	}
}

func TestCSPC_GetCASTName(t *testing.T) {
	tests := map[string]struct {
		castAnnotation   map[string]string
		expectedcastName string
	}{
		"Test 1": {map[string]string{"Anno 1": "Val 1"}, ""},
		"Test 2": {map[string]string{string(apisv1alpha1.CreatePoolCASTemplateKey): "Val 1"}, "Val 1"},
		"Test 3": {map[string]string{}, ""},
		"Test 4": {nil, ""},
		"Test 5": {map[string]string{string(apisv1alpha1.CreatePoolCASTemplateKey): "Val 1", "Anno 1": "Val 1"}, "Val 1"},
	}

	for name, test := range tests {
		// pin the variables
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			fakecspc := &CSPC{&apisv1alpha1.CStorPoolCluster{ObjectMeta: metav1.ObjectMeta{Annotations: test.castAnnotation}}}
			gotCastname := fakecspc.GetCASTName()
			if gotCastname != test.expectedcastName {
				t.Fatalf("Test %v failed, Expected %v but got %v", name, test.expectedcastName, gotCastname)
			}
		})
	}
}

func TestCSPC_GetPoolType(t *testing.T) {
	tests := map[string]struct {
		cspcNodeSpec []apisv1alpha1.CStorPoolClusterNodeSpec

		expectedPoolType string
		nodeName         string
	}{
		"Test 1": {[]apisv1alpha1.CStorPoolClusterNodeSpec{
			{Name: "worker-node-1", PoolSpec: apisv1alpha1.CStorPoolClusterSpecAttr{PoolType: "striped"}}},
			"striped", "worker-node-1"},
		"Test 2": {[]apisv1alpha1.CStorPoolClusterNodeSpec{
			{Name: "worker-node-1", PoolSpec: apisv1alpha1.CStorPoolClusterSpecAttr{PoolType: "mirrored"}},
			{Name: "worker-node-2", PoolSpec: apisv1alpha1.CStorPoolClusterSpecAttr{PoolType: "striped"}}},
			"mirrored", "worker-node-1"},
		"Test 3": {[]apisv1alpha1.CStorPoolClusterNodeSpec{
			{Name: "worker-node-1", PoolSpec: apisv1alpha1.CStorPoolClusterSpecAttr{PoolType: "mirrored"}},
			{Name: "worker-node-2", PoolSpec: apisv1alpha1.CStorPoolClusterSpecAttr{PoolType: "striped"}}},
			"", "worker-node-4"},
		"Test 4": {[]apisv1alpha1.CStorPoolClusterNodeSpec{},
			"", "worker-node-4"},
	}

	for name, test := range tests {
		// pin the variables
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			fakecspc := &CSPC{&apisv1alpha1.CStorPoolCluster{Spec: apisv1alpha1.CStorPoolClusterSpec{Nodes: test.cspcNodeSpec}}}
			gotPoolType := fakecspc.GetPoolTypeForNode(test.nodeName)
			if gotPoolType != test.expectedPoolType {
				t.Fatalf("Test %v failed, Expected %v but got %v", name, test.expectedPoolType, gotPoolType)
			}
		})
	}
}

func TestCSPCList_Len(t *testing.T) {
	tests := map[string]struct {
		fakeCSPCList *CSPCList

		expectedLen int
	}{
		"Test 1": {NewListBuilder().CSPCList, 0},
		// TODO: Add more test cases
	}

	for name, test := range tests {
		// pin the variables
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			gotLen := test.fakeCSPCList.Len()
			if gotLen != test.expectedLen {
				t.Fatalf("Test %v failed, Expected %v but got %v", name, test.expectedLen, gotLen)
			}
		})
	}
}

func TestCSPCList_IsEmpty(t *testing.T) {
	tests := map[string]struct {
		fakeCSPCList *CSPCList

		isEmpty bool
	}{
		"Test 1": {NewListBuilder().CSPCList, true},
		// TODO: Add more test cases
	}

	for name, test := range tests {
		// pin the variables
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			ok := test.fakeCSPCList.IsEmpty()
			if ok != test.isEmpty {
				t.Fatalf("Test %v failed, Expected %v but got %v", name, test.isEmpty, ok)
			}
		})
	}
}

func TestBuilder_WithDiskType(t *testing.T) {
	tests := map[string]struct {
		diskType string
	}{
		"Test 1": {"sparse"},
		"Test 2": {"disk"},
	}
	for name, test := range tests {
		// pin the variables
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			cspc := NewBuilder().WithDiskType(test.diskType)
			gotDiskType := cspc.CSPC.Object.Spec.Type
			if gotDiskType != test.diskType {
				t.Fatalf("Test %v failed, Expected %v but got %v", name, test.diskType, gotDiskType)
			}
		})
	}
}

func TestBuilder_WithMaxPool(t *testing.T) {
	tests := map[string]struct {
		maxPool int
	}{
		"Test 1": {1},
		"Test 2": {3},
	}
	for name, test := range tests {
		// pin the variables
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			cspc := NewBuilder().WithMaxPool(test.maxPool)
			gotMaxPool := cspc.CSPC.Object.Spec.MaxPools
			if *gotMaxPool != test.maxPool {
				t.Fatalf("Test %v failed, Expected %v but got %v", name, test.maxPool, *gotMaxPool)
			}
		})
	}
}

func TestBuilder_WithName(t *testing.T) {
	tests := map[string]struct {
		cspcName string
	}{
		"Test 1": {"sparse-claim"},
		"Test 2": {"disk-claim"},
	}
	for name, test := range tests {
		// pin the variables
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			cspc := NewBuilder().WithName(test.cspcName)
			gotCSPCName := cspc.CSPC.Object.Spec.Name
			if gotCSPCName != test.cspcName {
				t.Fatalf("Test %v failed, Expected %v but got %v", name, test.cspcName, gotCSPCName)
			}
		})
	}
}

func TestBuilder_WithPoolType(t *testing.T) {
	tests := map[string]struct {
		poolType string
	}{
		"Test 1": {"striped"},
		"Test 2": {"mirrored"},
		"Test 3": {"raidz"},
		"Test 4": {"raidz2"},
	}
	for name, test := range tests {
		// pin the variables
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			cspc := NewBuilder().WithPoolType(test.poolType)
			gotPoolType := cspc.CSPC.Object.Spec.PoolSpec.PoolType
			if gotPoolType != test.poolType {
				t.Fatalf("Test %v failed, Expected %v but got %v", name, test.poolType, gotPoolType)
			}
		})
	}
}

func TestBuilder_WithOverProvisioning(t *testing.T) {
	tests := map[string]struct {
		overProvisioning bool
	}{
		"Test 1": {true},
		"Test 2": {false},
	}
	for name, test := range tests {
		// pin the variables
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			cspc := NewBuilder().WithOverProvisioning(test.overProvisioning)
			gotOverProvisioning := cspc.CSPC.Object.Spec.PoolSpec.OverProvisioning
			if gotOverProvisioning != test.overProvisioning {
				t.Fatalf("Test %v failed, Expected %v but got %v", name, test.overProvisioning, gotOverProvisioning)
			}
		})
	}
}

func TestWithAPIList(t *testing.T) {
	tests := map[string]struct {
		expectedPoolName []string
	}{
		"Test 1": {[]string{"pool1", "pool2", "pool3"}},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			poolItems := &apisv1alpha1.CStorPoolClusterList{}
			for _, p := range mock.expectedPoolName {
				poolItems.Items = append(poolItems.Items, apisv1alpha1.CStorPoolCluster{ObjectMeta: metav1.ObjectMeta{Name: p}})
			}

			b := NewListBuilderForAPIList(poolItems)
			for index, ob := range b.CSPCList.ObjectList.Items {
				if !reflect.DeepEqual(ob, poolItems.Items[index]) {
					t.Fatalf("test %q failed: expected %v \n got : %v \n", name, poolItems.Items[index], ob)
				}
			}
		})
	}

}
