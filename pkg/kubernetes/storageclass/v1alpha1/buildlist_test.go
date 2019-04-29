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

package v1alpha1

import (
	"testing"
)

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
			b := ListBuilderForAPIList(fakeStorageClassListAPI(mock.availableSCs))
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
			b := ListBuilderForAPIList(fakeStorageClassListEmpty(mock.scCount))
			if mock.expectedSCCount != len(b.list.items) {
				t.Fatalf("Test %v failed: expected %v got %v", name, mock.expectedSCCount, len(b.list.items))
			}
		})
	}
}

func TestListBuilderForObjects(t *testing.T) {
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
			b := ListBuilderForObjects(fakeStorageClassListInstance(mock.availableSCs))
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
		expectedErr     bool
	}{
		"StorageClass set 1": {[]string{}, 0, true},
		"StorageClass set 2": {[]string{"sc1"}, 1, false},
		"StorageClass set 3": {[]string{"sc1", "sc2"}, 2, false},
		"StorageClass set 4": {[]string{"sc1", "sc2", "sc3"}, 3, false},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			b, err := ListBuilderForAPIList(fakeStorageClassListAPI(mock.availableSCs)).APIList()
			if err == nil && mock.expectedSCCount != len(b.Items) {
				t.Fatalf("Test %v failed: expected %v got %v", name, mock.expectedSCCount, len(b.Items))
			}
			if mock.expectedErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectedErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}
