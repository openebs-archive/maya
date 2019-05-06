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
	"reflect"
	"testing"

	apisv1beta1 "github.com/openebs/maya/pkg/apis/openebs.io/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func mockAlwaysTrue(*SPC) bool  { return true }
func mockAlwaysFalse(*SPC) bool { return false }

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
			if output := mock.Predicates.all(&SPC{}); output != mock.expectedOutput {
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
			fakespc := &SPC{&apisv1beta1.StoragePoolClaim{ObjectMeta: metav1.ObjectMeta{Annotations: test.availableAnnotations}}}
			ok := HasAnnotation(test.checkForKey, test.checkForValue)(fakespc)
			if ok != test.hasAnnotation {
				t.Fatalf("Test %v failed, Expected %v but got %v", name, test.availableAnnotations, fakespc.Object.GetAnnotations())
			}
		})
	}
}

func TestIsProvisioningAuto(t *testing.T) {
	tests := map[string]struct {
		nodeSpec           []apisv1beta1.StoragePoolClaimNodeSpec
		isAutoProvisioning bool
	}{
		"Test#1:Empty Node Spec": {[]apisv1beta1.StoragePoolClaimNodeSpec{}, true},

		"Test#2:Non-empty Node Spec with empty items": {[]apisv1beta1.StoragePoolClaimNodeSpec{
			{}, {},
		}, false},
	}

	for name, test := range tests {
		// pin the variables
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			fakespc := &SPC{&apisv1beta1.StoragePoolClaim{Spec: apisv1beta1.StoragePoolClaimSpec{Nodes: test.nodeSpec}}}
			ok := IsProvisioningAuto()(fakespc)
			if ok != test.isAutoProvisioning {
				t.Fatalf("Test %v failed, Expected %v but got %v", name, test.isAutoProvisioning, ok)
			}
		})
	}
}

func TestIsProvisioningManual(t *testing.T) {
	tests := map[string]struct {
		nodeSpec           []apisv1beta1.StoragePoolClaimNodeSpec
		isAutoProvisioning bool
	}{
		"Test#1:Empty Node Spec": {[]apisv1beta1.StoragePoolClaimNodeSpec{}, false},

		"Test#2:Non-empty Node Spec with empty items": {[]apisv1beta1.StoragePoolClaimNodeSpec{
			{}, {},
		}, true},
	}

	for name, test := range tests {
		// pin the variables
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			fakespc := &SPC{&apisv1beta1.StoragePoolClaim{Spec: apisv1beta1.StoragePoolClaimSpec{Nodes: test.nodeSpec}}}
			ok := IsProvisioningManual()(fakespc)
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
			fakespc := &SPC{&apisv1beta1.StoragePoolClaim{Spec: apisv1beta1.StoragePoolClaimSpec{Type: test.diskType}}}
			ok := IsSparse()(fakespc)
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
			fakespc := &SPC{&apisv1beta1.StoragePoolClaim{Spec: apisv1beta1.StoragePoolClaimSpec{Type: test.diskType}}}
			ok := IsDisk()(fakespc)
			if ok != test.isDiskType {
				t.Fatalf("Test %v failed, Expected %v but got %v", name, test.isDiskType, ok)
			}
		})
	}
}

func TestSPC_GetAnnotations(t *testing.T) {
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
			fakespc := &SPC{&apisv1beta1.StoragePoolClaim{ObjectMeta: metav1.ObjectMeta{Annotations: test.availableAnnotations}}}
			gotAnnotations := fakespc.GetAnnotations()
			if !reflect.DeepEqual(gotAnnotations, test.expectedAnnotations) {
				t.Fatalf("Test %v failed, Expected %v but got %v", name, test.expectedAnnotations, gotAnnotations)
			}
		})
	}
}

func TestSPC_GetNodeDisk(t *testing.T) {
	tests := map[string]struct {
		spcNodeSpec       []apisv1beta1.StoragePoolClaimNodeSpec
		expectedNodeNames []string
	}{
		"Test 1": {[]apisv1beta1.StoragePoolClaimNodeSpec{{Name: "worker-node-1"}, {Name: "worker-node-2"}}, []string{"worker-node-1", "worker-node-2"}},
		"Test 2": {[]apisv1beta1.StoragePoolClaimNodeSpec{{}}, []string{""}},
		"Test 3": {[]apisv1beta1.StoragePoolClaimNodeSpec{}, nil},
		"Test 4": {[]apisv1beta1.StoragePoolClaimNodeSpec{{}, {}}, []string{"", ""}},
	}

	for name, test := range tests {
		// pin the variables
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			fakespc := &SPC{&apisv1beta1.StoragePoolClaim{Spec: apisv1beta1.StoragePoolClaimSpec{Nodes: test.spcNodeSpec}}}
			gotNodeNames := fakespc.GetNodeNames()
			if !reflect.DeepEqual(gotNodeNames, test.expectedNodeNames) {
				t.Fatalf("Test %v failed, Expected %v but got %v", name, test.expectedNodeNames, gotNodeNames)
			}
		})
	}
}

func TestSPC_GetCASTName(t *testing.T) {
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
			fakespc := &SPC{&apisv1beta1.StoragePoolClaim{ObjectMeta: metav1.ObjectMeta{Annotations: test.castAnnotation}}}
			gotCastname := fakespc.GetCASTName()
			if gotCastname != test.expectedcastName {
				t.Fatalf("Test %v failed, Expected %v but got %v", name, test.expectedcastName, gotCastname)
			}
		})
	}
}

func TestSPC_GetPoolType(t *testing.T) {
	tests := map[string]struct {
		spcNodeSpec []apisv1beta1.StoragePoolClaimNodeSpec

		expectedPoolType string
		nodeName         string
	}{
		"Test 1": {[]apisv1beta1.StoragePoolClaimNodeSpec{
			{Name: "worker-node-1", PoolSpec: apisv1beta1.CStorPoolAttr{PoolType: "striped"}}},
			"striped", "worker-node-1"},
		"Test 2": {[]apisv1beta1.StoragePoolClaimNodeSpec{
			{Name: "worker-node-1", PoolSpec: apisv1beta1.CStorPoolAttr{PoolType: "mirrored"}},
			{Name: "worker-node-2", PoolSpec: apisv1beta1.CStorPoolAttr{PoolType: "striped"}}},
			"mirrored", "worker-node-1"},
		"Test 3": {[]apisv1beta1.StoragePoolClaimNodeSpec{
			{Name: "worker-node-1", PoolSpec: apisv1beta1.CStorPoolAttr{PoolType: "mirrored"}},
			{Name: "worker-node-2", PoolSpec: apisv1beta1.CStorPoolAttr{PoolType: "striped"}}},
			"", "worker-node-4"},
		"Test 4": {[]apisv1beta1.StoragePoolClaimNodeSpec{},
			"", "worker-node-4"},
	}

	for name, test := range tests {
		// pin the variables
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			fakespc := &SPC{&apisv1beta1.StoragePoolClaim{Spec: apisv1beta1.StoragePoolClaimSpec{Nodes: test.spcNodeSpec}}}
			gotPoolType := fakespc.GetPoolTypeForNode(test.nodeName)
			if gotPoolType != test.expectedPoolType {
				t.Fatalf("Test %v failed, Expected %v but got %v", name, test.expectedPoolType, gotPoolType)
			}
		})
	}
}

func TestSPCList_Len(t *testing.T) {
	tests := map[string]struct {
		fakeSPCList *SPCList

		expectedLen int
	}{
		"Test 1": {NewListBuilder().SPCList, 0},
		// TODO: Add more test cases
	}

	for name, test := range tests {
		// pin the variables
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			gotLen := test.fakeSPCList.Len()
			if gotLen != test.expectedLen {
				t.Fatalf("Test %v failed, Expected %v but got %v", name, test.expectedLen, gotLen)
			}
		})
	}
}

func TestSPCList_IsEmpty(t *testing.T) {
	tests := map[string]struct {
		fakeSPCList *SPCList

		isEmpty bool
	}{
		"Test 1": {NewListBuilder().SPCList, true},
		// TODO: Add more test cases
	}

	for name, test := range tests {
		// pin the variables
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			ok := test.fakeSPCList.IsEmpty()
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
			spc := NewBuilder().WithDiskType(test.diskType)
			gotDiskType := spc.SPC.Object.Spec.Type
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
			spc := NewBuilder().WithMaxPool(test.maxPool)
			gotMaxPool := spc.SPC.Object.Spec.MaxPools
			if *gotMaxPool != test.maxPool {
				t.Fatalf("Test %v failed, Expected %v but got %v", name, test.maxPool, *gotMaxPool)
			}
		})
	}
}

func TestBuilder_WithName(t *testing.T) {
	tests := map[string]struct {
		spcName string
	}{
		"Test 1": {"sparse-claim"},
		"Test 2": {"disk-claim"},
	}
	for name, test := range tests {
		// pin the variables
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			spc := NewBuilder().WithName(test.spcName)
			gotSPCName := spc.SPC.Object.Spec.Name
			if gotSPCName != test.spcName {
				t.Fatalf("Test %v failed, Expected %v but got %v", name, test.spcName, gotSPCName)
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
			spc := NewBuilder().WithPoolType(test.poolType)
			gotPoolType := spc.SPC.Object.Spec.PoolSpec.PoolType
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
			spc := NewBuilder().WithOverProvisioning(test.overProvisioning)
			gotOverProvisioning := spc.SPC.Object.Spec.PoolSpec.OverProvisioning
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
			poolItems := &apisv1beta1.StoragePoolClaimList{}
			for _, p := range mock.expectedPoolName {
				poolItems.Items = append(poolItems.Items, apisv1beta1.StoragePoolClaim{ObjectMeta: metav1.ObjectMeta{Name: p}})
			}

			b := NewListBuilderForAPIList(poolItems)
			for index, ob := range b.SPCList.ObjectList.Items {
				if !reflect.DeepEqual(ob, poolItems.Items[index]) {
					t.Fatalf("test %q failed: expected %v \n got : %v \n", name, poolItems.Items[index], ob)
				}
			}
		})
	}

}
