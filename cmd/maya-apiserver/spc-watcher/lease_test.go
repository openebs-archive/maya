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
package spc

import (
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"

	"github.com/golang/glog"
	openebsFakeClientset "github.com/openebs/maya/pkg/client/generated/clientset/internalclientset/fake"
	menv "github.com/openebs/maya/pkg/env/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"testing"
)

// SpcCreator will create fake spc objects
func (focs *clientSet) SpcCreator(poolName string, SpcLeaseKeyPresent bool, SpcLeaseKeyValue string) (claim *apis.StoragePoolClaim) {
	var spcObject *apis.StoragePoolClaim
	if SpcLeaseKeyPresent {
		spcObject = &apis.StoragePoolClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name: poolName,
				Annotations: map[string]string{
					SpcLeaseKey: SpcLeaseKeyValue,
				},
			},
		}
	} else {
		spcObject = &apis.StoragePoolClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name: poolName,
			},
		}
	}
	spcGot, err := focs.oecs.OpenebsV1alpha1().StoragePoolClaims().Create(spcObject)
	if err != nil {
		glog.Error(err)
	}
	return spcGot
}
func TestGetLease(t *testing.T) {
	// Get a fake openebs client set
	focs := &clientSet{
		oecs: openebsFakeClientset.NewSimpleClientset(),
	}
	// Make a map of string(key) to struct(value).
	// Key of map describes test case behaviour.
	// Value of map is the test object.
	tests := map[string]struct {
		// fakestoragepoolclaim holds the fake storagepoolcalim object in test cases.
		fakestoragepoolclaim *apis.StoragePoolClaim
		// expectedResult holds the expected result for the test case under run.
		expectedResult string
	}{
		// TestCase#1
		"SPC#1 Lease acquired": {
			fakestoragepoolclaim: focs.SpcCreator("pool1", true, "openebs/maya-apiserver-6b4695c9f8-nbwl9"),
			expectedResult:       "",
		},

		// TestCase#2
		"SPC#2 Lease not acquired": {
			fakestoragepoolclaim: focs.SpcCreator("pool2", false, ""),
			expectedResult:       "/pool2",
		},
		// TestCase#3
		"SPC#3 Lease not acquired": {
			fakestoragepoolclaim: focs.SpcCreator("pool3", true, ""),
			expectedResult:       "/pool3",
		},
	}

	// Iterate over whole map to run the test cases.
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var newSpcLease spcLease
			os.Setenv(string(menv.OpenEBSMayaPodName), test.fakestoragepoolclaim.Name)
			os.Setenv(string(menv.OpenEBSNamespace), test.fakestoragepoolclaim.Namespace)
			newSpcLease = spcLease{test.fakestoragepoolclaim, SpcLeaseKey, focs.oecs}
			// newSpcLease is the function under test.
			result, _ := newSpcLease.GetLease()
			// If the result does not matches expectedResult, test case fails.
			if result != test.expectedResult {
				t.Errorf("Test case failed: expected '%v' but got '%v' ", test.expectedResult, result)
			}
			os.Unsetenv(string(menv.OpenEBSMayaPodName))
			os.Unsetenv(string(menv.OpenEBSNamespace))
		})
	}
}
func TestRemoveLease(t *testing.T) {
	// Get a fake openebs client set
	focs := &clientSet{
		oecs: openebsFakeClientset.NewSimpleClientset(),
	}
	// Make a map of string(key) to struct(value).
	// Key of map describes test case behaviour.
	// Value of map is the test object.
	tests := map[string]struct {
		// fakestoragepoolclaim holds the fake storagepoolcalim object in test cases.
		fakestoragepoolclaim *apis.StoragePoolClaim
		// expectedResult holds the expected result for the test case under run.
		expectedResult string
	}{
		// TestCase#1
		"SPC#1 Lease acquired": {
			fakestoragepoolclaim: focs.SpcCreator("pool1", true, "openebs/maya-apiserver-6b4695c9f8-nbwl9"),
			expectedResult:       "",
		},

		// TestCase#2
		"SPC#2 Lease not acquired": {
			fakestoragepoolclaim: focs.SpcCreator("pool2", false, ""),
			expectedResult:       "",
		},
		// TestCase#3
		"SPC#3 Lease not acquired": {
			fakestoragepoolclaim: focs.SpcCreator("pool3", true, ""),
			expectedResult:       "",
		},
	}

	// Iterate over whole map to run the test cases.
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			newSpcLease := spcLease{test.fakestoragepoolclaim, SpcLeaseKey, focs.oecs}
			// newSpcLease is the function under test.
			spcObject := newSpcLease.RemoveLease()
			result := spcObject.Annotations[SpcLeaseKey]
			// If the result does not matches expectedResult, test case fails.
			if result != test.expectedResult {
				t.Errorf("Test case failed: expected '%v' but got '%v' ", test.expectedResult, result)
			}
		})
	}
}
