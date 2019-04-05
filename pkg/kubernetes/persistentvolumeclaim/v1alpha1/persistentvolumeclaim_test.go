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

func fakeBoundAPIPVCObject(pvcNames []string) []v1.PersistentVolumeClaim {
	plist := []v1.PersistentVolumeClaim{}
	for _, pvcName := range pvcNames {
		pvc := v1.PersistentVolumeClaim{}
		pvc.SetName(pvcName)
		pvc.Status.Phase = "Bound"
		plist = append(plist, pvc)
	}
	return plist
}

func fakeNonBoundPVCList(pvcNames []string) []v1.PersistentVolumeClaim {
	plist := []v1.PersistentVolumeClaim{}
	for _, pvcName := range pvcNames {
		pvc := v1.PersistentVolumeClaim{}
		pvc.SetName(pvcName)
		plist = append(plist, pvc)
	}
	return plist
}

func fakeAPIPVCListFromNameStatusMap(pvcs map[string]string) []*PVC {
	plist := []*PVC{}
	for k, v := range pvcs {
		p := &v1.PersistentVolumeClaim{}
		p.SetName(k)
		p.Status.Phase = v1.PersistentVolumeClaimPhase(v)
		plist = append(plist, &PVC{p})
	}
	return plist
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
		t.Run(name, func(t *testing.T) {
			b := NewListBuilder().WithAPIList(fakeAPIPVCList(mock.availablePVCs))
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
		t.Run(name, func(t *testing.T) {
			b := NewListBuilder().WithAPIObject(fakeAPIPVCList(mock.availablePVCs).Items...)
			if mock.expectedPVCLen != len(b.list.items) {
				t.Fatalf("Test %v failed: expected %v got %v", name, mock.availablePVCs, len(b.list.items))
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
		t.Run(name, func(t *testing.T) {
			b := NewListBuilder().WithAPIList(fakeAPIPVCList(mock.availablePVCs)).List().ToAPIList()
			if mock.expectedPVCLen != len(b.Items) {
				t.Fatalf("Test %v failed: expected %v got %v", name, mock.expectedPVCLen, len(b.Items))
			}
		})
	}
}

func TestFilterList(t *testing.T) {
	tests := map[string]struct {
		availablePVCs map[string]string
		filteredPVCs  []string
		filters       PredicateList
	}{
		"PVC Set 1": {
			availablePVCs: map[string]string{"PVC 1": "Bound", "PVC 2": "Waiting"},
			filteredPVCs:  []string{"PVC 1"},
			filters:       PredicateList{IsBound()},
		},
		"PVC Set 2": {
			availablePVCs: map[string]string{"PVC 1": "Bound", "PVC 2": "Bound"},
			filteredPVCs:  []string{"PVC 1", "PVC 2"},
			filters:       PredicateList{IsBound()},
		},

		"PVC Set 3": {
			availablePVCs: map[string]string{"PVC 1": "Waiting", "PVC 2": "Waiting", "PVC 3": "Waiting"},
			filteredPVCs:  []string{},
			filters:       PredicateList{IsBound()},
		},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			list := NewListBuilder().WithObject(fakeAPIPVCListFromNameStatusMap(mock.availablePVCs)...).WithFilter(mock.filters...).List()
			if len(list.items) != len(mock.filteredPVCs) {
				t.Fatalf("Test %v failed: expected %v got %v", name, len(mock.filteredPVCs), len(list.items))
			}
		})
	}
}
