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
	"reflect"
	"testing"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
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
			fakespc := &SPC{&apis.StoragePoolClaim{ObjectMeta: metav1.ObjectMeta{Annotations: test.availableAnnotations}}}
			ok := HasAnnotation(test.checkForKey, test.checkForValue)(fakespc)
			if ok != test.hasAnnotation {
				t.Fatalf("Test %v failed, Expected %v but got %v", name, test.availableAnnotations, fakespc.Object.GetAnnotations())
			}
		})
	}
}

func TestFilterUIDs(t *testing.T) {
	tests := map[string]struct {
		Predicates     predicateList
		UIDs           []string
		expectedOutput []string
	}{
		// With all Positive predicates
		"Positive 1": {[]Predicate{mockAlwaysTrue}, []string{"uid1", "uid2", "uid3"}, []string{"uid1", "uid2", "uid3"}},
		"Positive 2": {[]Predicate{mockAlwaysTrue, mockAlwaysTrue}, []string{"uid1", "uid2", "uid3"}, []string{"uid1", "uid2", "uid3"}},
		"Positive 3": {[]Predicate{mockAlwaysTrue, mockAlwaysTrue}, []string{"uid1", "uid2"}, []string{"uid1", "uid2"}},
		//  With all negative predicates
		"Negative 1": {[]Predicate{mockAlwaysFalse}, []string{"uid1", "uid2", "uid3"}, []string{}},
		"Negative 2": {[]Predicate{mockAlwaysFalse, mockAlwaysFalse}, []string{"uid1", "uid2", "uid3"}, []string{}},
		"Negative 3": {[]Predicate{mockAlwaysFalse, mockAlwaysFalse}, []string{"uid1", "uid2", "uid3"}, []string{}},
	}
	for name, mock := range tests {
		// pin the variables
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			cspL := NewListBuilder().WithUIDs(mock.UIDs...).List()
			output := cspL.Filter(mock.Predicates...)
			if len(mock.expectedOutput) != len(output.GetPoolUIDs()) {
				t.Fatalf("test %q failed: expected %v \n got : %v \n", name, mock.expectedOutput, output.GetPoolUIDs())
			}
			for index, val := range output.GetPoolUIDs() {
				if val != mock.expectedOutput[index] {
					t.Fatalf("test %q failed: expected %v \n got : %v \n", name, mock.expectedOutput, output.GetPoolUIDs())
				}
			}
		})
	}
}

func TestStoragePoolClaimList(t *testing.T) {
	tests := map[string]struct {
		expectedUIDs []string
	}{
		"UID set 1":  {[]string{}},
		"UID set 2":  {[]string{"uid1"}},
		"UID set 3":  {[]string{"uid1", "uid2"}},
		"UID set 4":  {[]string{"uid1", "uid2", "uid3"}},
		"UID set 5":  {[]string{"uid1", "uid2", "uid3", "uid4"}},
		"UID set 6":  {[]string{"uid1", "uid2", "uid3", "uid4", "uid5"}},
		"UID set 7":  {[]string{"uid1", "uid2", "uid3", "uid4", "uid5", "uid6"}},
		"UID set 8":  {[]string{"uid1", "uid2", "uid3", "uid4", "uid5", "uid6", "uid7"}},
		"UID set 9":  {[]string{"uid1", "uid2", "uid3", "uid4", "uid5", "uid6", "uid7", "uid8"}},
		"UID set 10": {[]string{"uid1", "uid2", "uid3", "uid4", "uid5", "uid6", "uid7", "uid8", "uid9"}},
	}

	for name, mock := range tests {
		// pin the variables
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			lb := NewListBuilder().WithUIDs(mock.expectedUIDs...).List()
			if len(lb.ObjectList.Items) != len(mock.expectedUIDs) {
				t.Fatalf("test %q failed: expected %v \n got : %v \n", name, mock.expectedUIDs, lb.ObjectList.Items)
			}
			for index, val := range lb.ObjectList.Items {
				if string(val.GetUID()) != mock.expectedUIDs[index] {
					t.Fatalf("test %q failed: expected %v \n got : %v \n", name, mock.expectedUIDs[index], string(val.GetUID()))
				}
			}
		})
	}
}

func TestStoragePoolClaimWithUIDs(t *testing.T) {
	tests := map[string]struct {
		expectedUIDs []string
	}{
		"UID set 1":  {[]string{}},
		"UID set 2":  {[]string{"uid1"}},
		"UID set 3":  {[]string{"uid1", "uid2"}},
		"UID set 4":  {[]string{"uid1", "uid2", "uid3"}},
		"UID set 5":  {[]string{"uid1", "uid2", "uid3", "uid4"}},
		"UID set 6":  {[]string{"uid1", "uid2", "uid3", "uid4", "uid5"}},
		"UID set 7":  {[]string{"uid1", "uid2", "uid3", "uid4", "uid5", "uid6"}},
		"UID set 8":  {[]string{"uid1", "uid2", "uid3", "uid4", "uid5", "uid6", "uid7"}},
		"UID set 9":  {[]string{"uid1", "uid2", "uid3", "uid4", "uid5", "uid6", "uid7", "uid8"}},
		"UID set 10": {[]string{"uid1", "uid2", "uid3", "uid4", "uid5", "uid6", "uid7", "uid8", "uid9"}},
	}

	for name, mock := range tests {
		// pin the variables
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			lb := NewListBuilder().WithUIDs(mock.expectedUIDs...)
			if len(lb.SpcList.ObjectList.Items) != len(mock.expectedUIDs) {
				t.Fatalf("test %q failed: expected %v \n got : %v \n", name, mock.expectedUIDs, lb.SpcList.ObjectList.Items)
			}
			for index, val := range lb.SpcList.ObjectList.Items {
				if string(val.GetUID()) != mock.expectedUIDs[index] {
					t.Fatalf("test %q failed: expected %v \n got : %v \n", name, mock.expectedUIDs[index], string(val.GetUID()))
				}
			}
		})
	}
}

func TestCStorPoolFilterUIDs(t *testing.T) {
	tests := map[string]struct {
		Predicates     predicateList
		UIDs           []string
		expectedOutput []string
	}{
		// With all Positive predicates
		"Positive 1": {[]Predicate{mockAlwaysTrue}, []string{"uid1", "uid2", "uid3"}, []string{"uid1", "uid2", "uid3"}},
		"Positive 2": {[]Predicate{mockAlwaysTrue, mockAlwaysTrue}, []string{"uid1", "uid2", "uid3"}, []string{"uid1", "uid2", "uid3"}},
		"Positive 3": {[]Predicate{mockAlwaysTrue, mockAlwaysTrue}, []string{"uid1", "uid2"}, []string{"uid1", "uid2"}},
		//  With all negative predicates
		"Negative 1": {[]Predicate{mockAlwaysFalse}, []string{"uid1", "uid2", "uid3"}, []string{}},
		"Negative 2": {[]Predicate{mockAlwaysFalse, mockAlwaysFalse}, []string{"uid1", "uid2", "uid3"}, []string{}},
		"Negative 3": {[]Predicate{mockAlwaysFalse, mockAlwaysFalse}, []string{"uid1", "uid2", "uid3"}, []string{}},
	}
	for name, mock := range tests {
		// pin the variables
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			cspL := NewListBuilder().WithUIDs(mock.UIDs...).List()
			output := cspL.Filter(mock.Predicates...)
			if len(mock.expectedOutput) != len(output.GetPoolUIDs()) {
				t.Fatalf("test %q failed: expected %v \n got : %v \n", name, mock.expectedOutput, output.GetPoolUIDs())
			}
			for index, val := range output.GetPoolUIDs() {
				if val != mock.expectedOutput[index] {
					t.Fatalf("test %q failed: expected %v \n got : %v \n", name, mock.expectedOutput, output.GetPoolUIDs())
				}
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
			poolItems := &apis.StoragePoolClaimList{}
			for _, p := range mock.expectedPoolName {
				poolItems.Items = append(poolItems.Items, apis.StoragePoolClaim{ObjectMeta: metav1.ObjectMeta{Name: p}})
			}

			b := NewListBuilder().WithAPIList(poolItems)
			for index, ob := range b.SpcList.ObjectList.Items {
				if !reflect.DeepEqual(ob, poolItems.Items[index]) {
					t.Fatalf("test %q failed: expected %v \n got : %v \n", name, poolItems.Items[index], ob)
				}
			}
		})
	}

}
