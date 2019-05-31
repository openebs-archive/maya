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

package v1alpha2

import (
	"testing"

	apisv1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func mockAlwaysTrue(*Disk) bool  { return true }
func mockAlwaysFalse(*Disk) bool { return false }

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
			if output := mock.Predicates.all(&Disk{}); output != mock.expectedOutput {
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
			fakedisk := &Disk{&apisv1alpha1.Disk{ObjectMeta: metav1.ObjectMeta{Annotations: test.availableAnnotations}}}
			ok := HasAnnotation(test.checkForKey, test.checkForValue)(fakedisk)
			if ok != test.hasAnnotation {
				t.Fatalf("Test %v failed, Expected %v but got %v", name, test.availableAnnotations, fakedisk.Object.GetAnnotations())
			}
		})
	}
}

// TODO: Add more Unit test
