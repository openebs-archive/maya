// Copyright © 2018-2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"testing"

	v1 "k8s.io/api/core/v1"
)

func fakeAPIPVCList(pvcNames []string) *v1.PersistentVolumeClaimList {
	if len(pvcNames) == 0 {
		return nil
	}
	list := &v1.PersistentVolumeClaimList{}
	for _, name := range pvcNames {
		pvc := v1.PersistentVolumeClaim{}
		pvc.SetName(name)
		list.Items = append(list.Items, pvc)
	}
	return list
}

func fakeAPIPVCListFromNameStatusMap(pvcs map[string]v1.PersistentVolumeClaimPhase) *v1.PersistentVolumeClaimList {
	list := &v1.PersistentVolumeClaimList{}
	for k, v := range pvcs {
		pvc := v1.PersistentVolumeClaim{}
		pvc.SetName(k)
		pvc.Status.Phase = v
		list.Items = append(list.Items, pvc)
	}
	return list
}

func TestListBuilderWithAPIList(t *testing.T) {
	tests := map[string]struct {
		availablePVCs  []string
		expectedPVCLen int
	}{
		"PVC set 1":  {[]string{}, 0},
		"PVC set 2":  {[]string{"pvc1"}, 1},
		"PVC set 3":  {[]string{"pvc1", "pvc2"}, 2},
		"PVC set 4":  {[]string{"pvc1", "pvc2", "pvc3"}, 3},
		"PVC set 5":  {[]string{"pvc1", "pvc2", "pvc3", "pvc4"}, 4},
		"PVC set 6":  {[]string{"pvc1", "pvc2", "pvc3", "pvc4", "pvc5"}, 5},
		"PVC set 7":  {[]string{"pvc1", "pvc2", "pvc3", "pvc4", "pvc5", "pvc6"}, 6},
		"PVC set 8":  {[]string{"pvc1", "pvc2", "pvc3", "pvc4", "pvc5", "pvc6", "pvc7"}, 7},
		"PVC set 9":  {[]string{"pvc1", "pvc2", "pvc3", "pvc4", "pvc5", "pvc6", "pvc7", "pvc8"}, 8},
		"PVC set 10": {[]string{"pvc1", "pvc2", "pvc3", "pvc4", "pvc5", "pvc6", "pvc7", "pvc8", "pvc9"}, 9},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := ListBuilderForAPIObjects(fakeAPIPVCList(mock.availablePVCs))
			if mock.expectedPVCLen != len(b.list.items) {
				t.Fatalf("Test %v failed: expected %v got %v", name, mock.expectedPVCLen, len(b.list.items))
			}
		})
	}
}

func TestListBuilderWithAPIObjects(t *testing.T) {
	tests := map[string]struct {
		availablePVCs  []string
		expectedPVCLen int
	}{
		"PVC set 1":  {[]string{}, 0},
		"PVC set 2":  {[]string{"pvc1"}, 1},
		"PVC set 3":  {[]string{"pvc1", "pvc2"}, 2},
		"PVC set 4":  {[]string{"pvc1", "pvc2", "pvc3"}, 3},
		"PVC set 5":  {[]string{"pvc1", "pvc2", "pvc3", "pvc4"}, 4},
		"PVC set 6":  {[]string{"pvc1", "pvc2", "pvc3", "pvc4", "pvc5"}, 5},
		"PVC set 7":  {[]string{"pvc1", "pvc2", "pvc3", "pvc4", "pvc5", "pvc6"}, 6},
		"PVC set 8":  {[]string{"pvc1", "pvc2", "pvc3", "pvc4", "pvc5", "pvc6", "pvc7"}, 7},
		"PVC set 9":  {[]string{"pvc1", "pvc2", "pvc3", "pvc4", "pvc5", "pvc6", "pvc7", "pvc8"}, 8},
		"PVC set 10": {[]string{"pvc1", "pvc2", "pvc3", "pvc4", "pvc5", "pvc6", "pvc7", "pvc8", "pvc9"}, 9},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b, _ := ListBuilderForAPIObjects(fakeAPIPVCList(mock.availablePVCs)).APIList()
			if mock.expectedPVCLen != len(b.Items) {
				t.Fatalf("Test %v failed: expected %v got %v", name, mock.availablePVCs, len(b.Items))
			}
		})
	}
}

func TestListBuilderToAPIList(t *testing.T) {
	tests := map[string]struct {
		availablePVCs  []string
		expectedPVCLen int
	}{
		"PVC set 1":  {[]string{}, 0},
		"PVC set 2":  {[]string{"pvc1"}, 1},
		"PVC set 3":  {[]string{"pvc1", "pvc2"}, 2},
		"PVC set 4":  {[]string{"pvc1", "pvc2", "pvc3"}, 3},
		"PVC set 5":  {[]string{"pvc1", "pvc2", "pvc3", "pvc4"}, 4},
		"PVC set 6":  {[]string{"pvc1", "pvc2", "pvc3", "pvc4", "pvc5"}, 5},
		"PVC set 7":  {[]string{"pvc1", "pvc2", "pvc3", "pvc4", "pvc5", "pvc6"}, 6},
		"PVC set 8":  {[]string{"pvc1", "pvc2", "pvc3", "pvc4", "pvc5", "pvc6", "pvc7"}, 7},
		"PVC set 9":  {[]string{"pvc1", "pvc2", "pvc3", "pvc4", "pvc5", "pvc6", "pvc7", "pvc8"}, 8},
		"PVC set 10": {[]string{"pvc1", "pvc2", "pvc3", "pvc4", "pvc5", "pvc6", "pvc7", "pvc8", "pvc9"}, 9},
	}
	for name, mock := range tests {

		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := ListBuilderForAPIObjects(fakeAPIPVCList(mock.availablePVCs)).List().ToAPIList()
			if mock.expectedPVCLen != len(b.Items) {
				t.Fatalf("Test %v failed: expected %v got %v", name, mock.expectedPVCLen, len(b.Items))
			}
		})
	}
}

func TestFilterList(t *testing.T) {
	tests := map[string]struct {
		availablePVCs map[string]v1.PersistentVolumeClaimPhase
		filteredPVCs  []string
		filters       PredicateList
	}{
		"PVC Set 1": {
			availablePVCs: map[string]v1.PersistentVolumeClaimPhase{"PVC5": v1.ClaimBound, "PVC6": v1.ClaimPending, "PVC7": v1.ClaimLost},
			filteredPVCs:  []string{"PVC5"},
			filters:       PredicateList{IsBound()},
		},

		"PVC Set 2": {
			availablePVCs: map[string]v1.PersistentVolumeClaimPhase{"PVC3": v1.ClaimBound, "PVC4": v1.ClaimBound},
			filteredPVCs:  []string{"PVC2", "PVC4"},
			filters:       PredicateList{IsBound()},
		},

		"PVC Set 3": {
			availablePVCs: map[string]v1.PersistentVolumeClaimPhase{"PVC1": v1.ClaimLost, "PVC2": v1.ClaimPending, "PVC3": v1.ClaimPending},
			filteredPVCs:  []string{},
			filters:       PredicateList{IsBound()},
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			list := ListBuilderForAPIObjects(fakeAPIPVCListFromNameStatusMap(mock.availablePVCs)).WithFilter(mock.filters...).List()
			if len(list.items) != len(mock.filteredPVCs) {
				t.Fatalf("Test %v failed: expected %v got %v", name, len(mock.filteredPVCs), len(list.items))
			}
		})
	}
}
