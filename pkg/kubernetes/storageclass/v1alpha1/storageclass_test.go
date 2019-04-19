package v1alpha1

import (
	"testing"

	storagev1 "k8s.io/api/storage/v1"
)

func fakeStorageClassListEmpty(noOfStorageclasses int) *storagev1.StorageClassList {
	list := &storagev1.StorageClassList{}
	for i := 0; i < noOfStorageclasses; i++ {
		sc := storagev1.StorageClass{}
		list.Items = append(list.Items, sc)
	}
	return list
}

func fakeStorageClassListAPI(scNames []string) *storagev1.StorageClassList {
	if len(scNames) == 0 {
		return nil
	}
	list := &storagev1.StorageClassList{}
	for _, name := range scNames {
		name := name // Pin It
		sc := storagev1.StorageClass{}
		sc.SetName(name)
		list.Items = append(list.Items, sc)
	}
	return list
}

func fakeStorageClassInstances(scNames []string) []*StorageClass {
	if len(scNames) == 0 {
		return nil
	}
	list := []*StorageClass{}
	for _, name := range scNames {
		name := name // Pin It
		sc := storagev1.StorageClass{}
		sc.SetName(name)
		list = append(list, &StorageClass{&sc})
	}
	return list
}

func TestListBuilderFuncWithAPIList(t *testing.T) {
	tests := map[string]struct {
		availableSCs    []string
		expectedSCCount int
	}{
		"StorageClass set 1": {[]string{"sc1"}, 1},
		"StorageClass set 2": {[]string{"sc1", "sc2"}, 2},
		"StorageClass set 3": {[]string{"sc1", "sc2", "sc3"}, 3},
		"StorageClass set 4": {[]string{"sc1", "sc2", "sc3"}, 3},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			b := NewListBuilder().WithAPIList(fakeStorageClassListAPI(mock.availableSCs))
			if mock.expectedSCCount != len(b.list.items) {
				t.Fatalf("Test %v failed: expected %v got %v", name, mock.expectedSCCount, len(b.list.items))
			}
		})
	}
}

func TestListBuilderFuncWithEmptyStorageClassList(t *testing.T) {
	tests := map[string]struct {
		scCount, expectedSCCount int
	}{
		"StorageClass": {5, 5},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			b := NewListBuilder().WithAPIList(fakeStorageClassListEmpty(mock.scCount))
			if mock.expectedSCCount != len(b.list.items) {
				t.Fatalf("Test %v failed: expected %v got %v", name, mock.expectedSCCount, len(b.list.items))
			}
		})
	}
}

func TestListBuilderWithAPIObjects(t *testing.T) {
	tests := map[string]struct {
		availableSC     []string
		expectedSCCount int
	}{
		"StorageClass set 1": {[]string{"sc1"}, 1},
		"StorageClass set 2": {[]string{"sc1", "sc2"}, 2},
		"StorageClass set 3": {[]string{"sc1", "sc2", "sc3"}, 3},
		"StorageClass set 4": {[]string{"sc1", "sc2", "sc3"}, 3},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			b := NewListBuilder().WithAPIObject(fakeStorageClassListAPI(mock.availableSC).Items...)
			if mock.expectedSCCount != len(b.list.items) {
				t.Fatalf("Test %v failed: expected %v got %v", name, mock.expectedSCCount, len(b.list.items))
			}
		})
	}
}

func TestListBuilderToAPIList(t *testing.T) {
	tests := map[string]struct {
		availableSCs    []string
		expectedSCCount int
	}{
		"StorageClass set 1": {[]string{}, 0},
		"StorageClass set 2": {[]string{"sc1"}, 1},
		"StorageClass set 3": {[]string{"sc1", "sc2"}, 2},
		"StorageClass set 4": {[]string{"sc1", "sc2", "sc3"}, 3},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			b := NewListBuilder().WithAPIList(fakeStorageClassListAPI(mock.availableSCs)).List().ToAPIList()
			if mock.expectedSCCount != len(b.Items) {
				t.Fatalf("Test %v failed: expected %v got %v", name, mock.expectedSCCount, len(b.Items))
			}
		})
	}
}

func TestFilterList(t *testing.T) {
	tests := map[string]struct {
		availableSC []string
		filteredSCs []string
	}{
		"StorageClass Set 1": {
			availableSC: []string{"SC1", "SC2"},
			filteredSCs: []string{"SC1", "SC2"},
		},
		"StorageClass Set 2": {
			availableSC: []string{},
			filteredSCs: []string{},
		},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			sc := &ListBuilder{list: &StorageClassList{items: fakeStorageClassInstances(mock.availableSC)}}
			list := sc.List()
			if len(list.items) != len(mock.filteredSCs) {
				t.Fatalf("Test %v failed: expected %v got %v", name, len(mock.filteredSCs), len(list.items))
			}
		})
	}
}
