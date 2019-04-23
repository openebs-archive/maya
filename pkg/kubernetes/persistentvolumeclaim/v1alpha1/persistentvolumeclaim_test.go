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

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

func fakeAPIPVCList(pvcNames []string) *corev1.PersistentVolumeClaimList {
	if len(pvcNames) == 0 {
		return nil
	}
	list := &corev1.PersistentVolumeClaimList{}
	for _, name := range pvcNames {
		pvc := corev1.PersistentVolumeClaim{}
		pvc.SetName(name)
		list.Items = append(list.Items, pvc)
	}
	return list
}

func fakeAPIPVCListFromNameStatusMap(pvcs map[string]corev1.PersistentVolumeClaimPhase) *corev1.PersistentVolumeClaimList {
	list := &corev1.PersistentVolumeClaimList{}
	for k, v := range pvcs {
		pvc := corev1.PersistentVolumeClaim{}
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
		availablePVCs map[string]corev1.PersistentVolumeClaimPhase
		filteredPVCs  []string
		filters       PredicateList
	}{
		"PVC Set 1": {
			availablePVCs: map[string]corev1.PersistentVolumeClaimPhase{"PVC5": corev1.ClaimBound, "PVC6": corev1.ClaimPending, "PVC7": corev1.ClaimLost},
			filteredPVCs:  []string{"PVC5"},
			filters:       PredicateList{IsBound()},
		},

		"PVC Set 2": {
			availablePVCs: map[string]corev1.PersistentVolumeClaimPhase{"PVC3": corev1.ClaimBound, "PVC4": corev1.ClaimBound},
			filteredPVCs:  []string{"PVC2", "PVC4"},
			filters:       PredicateList{IsBound()},
		},

		"PVC Set 3": {
			availablePVCs: map[string]corev1.PersistentVolumeClaimPhase{"PVC1": corev1.ClaimLost, "PVC2": corev1.ClaimPending, "PVC3": corev1.ClaimPending},
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

func TestPVCWithName(t *testing.T) {
	tests := map[string]struct {
		name      string
		pvc       *PVC
		expectErr bool
	}{
		"Test PVC with name": {
			name:      "PVC1",
			pvc:       &PVC{Object: &corev1.PersistentVolumeClaim{}},
			expectErr: false,
		},
		"Test PVC without name": {
			name:      "",
			pvc:       &PVC{Object: &corev1.PersistentVolumeClaim{}},
			expectErr: true,
		},
		"Test with PVC error": {
			name:      "PVC2",
			pvc:       &PVC{Err: errors.New("PVC not built")},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			pvcObj := mock.pvc.WithName(mock.name)
			if mock.expectErr && pvcObj.Err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && pvcObj.Err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestPVCWithNamespace(t *testing.T) {
	tests := map[string]struct {
		namespace string
		pvc       *PVC
		expectErr bool
	}{
		"Test PVC with namespae": {
			namespace: "jiva-ns",
			pvc:       &PVC{Object: &corev1.PersistentVolumeClaim{}},
			expectErr: false,
		},
		"Test PVC without namespace": {
			namespace: "",
			pvc:       &PVC{Object: &corev1.PersistentVolumeClaim{}},
			expectErr: false,
		},
		"Test with PVC error": {
			namespace: "cstor-ns",
			pvc:       &PVC{Err: errors.New("PVC shouldn't be empty")},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			pvcObj := mock.pvc.WithNamespace(mock.namespace)
			if mock.expectErr && pvcObj.Err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && pvcObj.Err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestPVCWithAnnotations(t *testing.T) {
	tests := map[string]struct {
		annotations map[string]string
		pvc         *PVC
		expectErr   bool
	}{
		"Test PVC with annotations": {
			annotations: map[string]string{"persistent-volume": "PV", "application": "percona"},
			pvc:         &PVC{Object: &corev1.PersistentVolumeClaim{}},
			expectErr:   false,
		},
		"Test PVC without annotations": {
			annotations: map[string]string{},
			pvc:         &PVC{Object: &corev1.PersistentVolumeClaim{}},
			expectErr:   false,
		},
		"Test with PVC error": {
			annotations: map[string]string{"persistent-volume": "PV"},
			pvc:         &PVC{Err: errors.New("PVC name shouldn't be nil")},
			expectErr:   true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			pvcObj := mock.pvc.WithAnnotations(mock.annotations)
			if mock.expectErr && pvcObj.Err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && pvcObj.Err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestPVCWithLabels(t *testing.T) {
	tests := map[string]struct {
		labels    map[string]string
		pvc       *PVC
		expectErr bool
	}{
		"Test PVC with labels": {
			labels:    map[string]string{"persistent-volume": "PV", "application": "percona"},
			pvc:       &PVC{Object: &corev1.PersistentVolumeClaim{}},
			expectErr: false,
		},
		"Test PVC without labels": {
			labels:    map[string]string{},
			pvc:       &PVC{Object: &corev1.PersistentVolumeClaim{}},
			expectErr: false,
		},
		"Test with PVC error": {
			labels:    map[string]string{"persistent-volume": "PV"},
			pvc:       &PVC{Err: errors.New("PVC name shouldn't be nil")},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			pvcObj := mock.pvc.WithLabels(mock.labels)
			if mock.expectErr && pvcObj.Err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && pvcObj.Err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestPVCWithAccessModes(t *testing.T) {
	tests := map[string]struct {
		accessModes []corev1.PersistentVolumeAccessMode
		pvc         *PVC
		expectErr   bool
	}{
		"Test PVC with accessModes": {
			accessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce, corev1.ReadOnlyMany},
			pvc:         &PVC{Object: &corev1.PersistentVolumeClaim{}},
			expectErr:   false,
		},
		"Test PVC without accessModes": {
			accessModes: []corev1.PersistentVolumeAccessMode{},
			pvc:         &PVC{Object: &corev1.PersistentVolumeClaim{}},
			expectErr:   true,
		},
		"Test with PVC accessModes": {
			accessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce, corev1.ReadWriteMany},
			pvc:         &PVC{Err: errors.New("PVC name shouldn't be nil")},
			expectErr:   true,
		},
	}

	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			pvcObj := mock.pvc.WithAccessModes(mock.accessModes)
			if mock.expectErr && pvcObj.Err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && pvcObj.Err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestPVCWithStorageClass(t *testing.T) {
	tests := map[string]struct {
		scName    string
		pvc       *PVC
		expectErr bool
	}{
		"Test PVC with SC": {
			scName:    "single-replica",
			pvc:       &PVC{Object: &corev1.PersistentVolumeClaim{}},
			expectErr: false,
		},
		"Test PVC without SC": {
			scName:    "",
			pvc:       &PVC{Object: &corev1.PersistentVolumeClaim{}},
			expectErr: true,
		},
		"Test with PVC error": {
			scName:    "multi-replica",
			pvc:       &PVC{Err: errors.New("PVC not built")},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			pvcObj := mock.pvc.WithStorageClass(mock.scName)
			if mock.expectErr && pvcObj.Err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && pvcObj.Err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestPVCWithCapacity(t *testing.T) {
	tests := map[string]struct {
		capacity  string
		pvc       *PVC
		expectErr bool
	}{
		"Test PVC with capacity": {
			capacity:  "5G",
			pvc:       &PVC{Object: &corev1.PersistentVolumeClaim{}},
			expectErr: false,
		},
		"Test PVC without capacity": {
			capacity:  "",
			pvc:       &PVC{Object: &corev1.PersistentVolumeClaim{}},
			expectErr: true,
		},
		"Test with PVC error": {
			capacity:  "10Ti",
			pvc:       &PVC{Err: errors.New("PVC not built")},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			pvcObj := mock.pvc.WithCapacity(mock.capacity)
			if mock.expectErr && pvcObj.Err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && pvcObj.Err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}
