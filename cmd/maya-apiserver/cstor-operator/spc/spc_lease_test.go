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
	"context"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"

	"os"
	"strconv"
	"testing"

	openebsFakeClientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned/fake"
	env "github.com/openebs/maya/pkg/env/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/klog/v2"
)

// SpcCreator will create fake spc objects
func (focs *clientSet) SpcCreator(poolName string, SpcLeaseKeyPresent bool, SpcLeaseKeyValue string) *apis.StoragePoolClaim {
	var spcObject *apis.StoragePoolClaim
	if SpcLeaseKeyPresent {
		spcObject = &apis.StoragePoolClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name: poolName,
				Annotations: map[string]string{
					SpcLeaseKey: "{\"holder\":\"" + SpcLeaseKeyValue + "\",\"leaderTransition\":1}",
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
	spcGot, err := focs.oecs.OpenebsV1alpha1().StoragePoolClaims().
		Create(context.TODO(), spcObject, metav1.CreateOptions{})
	if err != nil {
		klog.Error(err)
	}
	return spcGot
}

// Create 5 fake pods that will compete to acquire lease on spc
func PodCreator(fakeKubeClient kubernetes.Interface, podName string) {
	for i := 1; i <= 5; i++ {
		podObjet := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: podName + strconv.Itoa(i),
			},
			Status: v1.PodStatus{
				Phase: v1.PodRunning,
			},
		}
		_, err := fakeKubeClient.CoreV1().Pods("openebs").
			Create(context.TODO(), podObjet, metav1.CreateOptions{})
		if err != nil {
			klog.Error("Fake pod object could not be created:", err)
		}
	}
}
func TestHold(t *testing.T) {
	// Get a fake openebs client set
	focs := &clientSet{
		oecs: openebsFakeClientset.NewSimpleClientset(),
	}

	fakeKubeClient := k8sfake.NewSimpleClientset()

	// Make a map of string(key) to struct(value).
	// Key of map describes test case behaviour.
	// Value of map is the test object.
	PodCreator(fakeKubeClient, "maya-apiserver")
	tests := map[string]struct {
		// fakestoragepoolclaim holds the fake storagepoolcalim object in test cases.
		fakestoragepoolclaim *apis.StoragePoolClaim
		storagePoolClaimName string
		podName              string
		podNamespace         string
		// expectedResult holds the expected error for the test case under run.
		expectedError bool
		// expectedResult holds the expected lease value the test case under run.
		expectedResult string
	}{
		// TestCase#1
		"SPC#1 Lease Not acquired": {
			fakestoragepoolclaim: focs.SpcCreator("pool1", false, ""),
			podName:              "maya-apiserver1",
			podNamespace:         "openebs",
			expectedError:        false,
			expectedResult:       "{\"holder\":\"openebs/maya-apiserver1\",\"leaderTransition\":1}",
		},

		// TestCase#2
		"SPC#2 Lease already acquired": {
			fakestoragepoolclaim: focs.SpcCreator("pool2", true, "openebs/maya-apiserver1"),
			podName:              "maya-apiserver2",
			podNamespace:         "openebs",
			expectedError:        true,
			expectedResult:       "{\"holder\":\"openebs/maya-apiserver1\",\"leaderTransition\":1}",
		},
		// TestCase#3
		"SPC#3 Lease already acquired": {
			fakestoragepoolclaim: focs.SpcCreator("pool3", true, "openebs/maya-apiserver6"),
			podName:              "maya-apiserver2",
			podNamespace:         "openebs",
			expectedError:        false,
			expectedResult:       "{\"holder\":\"openebs/maya-apiserver2\",\"leaderTransition\":2}",
		},
		// TestCase#4
		"SPC#4 Lease Not acquired": {
			fakestoragepoolclaim: focs.SpcCreator("pool4", true, ""),
			podName:              "maya-apiserver3",
			podNamespace:         "openebs",
			expectedError:        false,
			expectedResult:       "{\"holder\":\"openebs/maya-apiserver3\",\"leaderTransition\":2}",
		},
	}

	// Iterate over whole map to run the test cases.
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var newSpcLease Lease
			var gotError bool
			os.Setenv(string(env.OpenEBSMayaPodName), test.podName)
			os.Setenv(string(env.OpenEBSNamespace), test.podNamespace)
			newSpcLease = Lease{test.fakestoragepoolclaim, SpcLeaseKey, focs.oecs, fakeKubeClient}
			// Hold is the function under test.
			err := newSpcLease.Hold()
			if err == nil {
				gotError = false
			} else {
				gotError = true
			}
			//If the result does not matches expectedResult, test case fails.
			if gotError != test.expectedError {
				t.Errorf("Test case failed:expected nil error but got error:'%v'", err)
			}
			// Check for lease value
			spcGot, err := focs.oecs.OpenebsV1alpha1().StoragePoolClaims().
				Get(context.TODO(), test.fakestoragepoolclaim.Name, metav1.GetOptions{})
			if spcGot.Annotations[SpcLeaseKey] != test.expectedResult {
				t.Errorf("Test case failed: expected lease value '%v' but got '%v' ", test.expectedResult, spcGot.Annotations[SpcLeaseKey])

			}
			os.Unsetenv(string(env.OpenEBSMayaPodName))
			os.Unsetenv(string(env.OpenEBSNamespace))
		})
	}
}
