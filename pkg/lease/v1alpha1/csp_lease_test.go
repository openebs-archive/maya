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
package lease

import (
	"context"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"

	"os"
	"strconv"
	"testing"

	openebs "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
	openebsFakeClientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/klog/v2"
)

type fakeClientset struct {
	oecs openebs.Interface
}

// CspCreator will create fake csp objects
func (focs *fakeClientset) CspCreator(poolName string, CspLeaseKeyPresent bool, CspLeaseKeyValue string) *apis.CStorPool {
	var cspObject *apis.CStorPool
	if CspLeaseKeyPresent {
		cspObject = &apis.CStorPool{
			ObjectMeta: metav1.ObjectMeta{
				Name: poolName,
				Annotations: map[string]string{
					CspLeaseKey: "{\"holder\":\"" + CspLeaseKeyValue + "\",\"leaderTransition\":1}",
				},
			},
		}
	} else {
		cspObject = &apis.CStorPool{
			ObjectMeta: metav1.ObjectMeta{
				Name: poolName,
			},
		}
	}
	cspGot, err := focs.oecs.OpenebsV1alpha1().CStorPools().
		Create(context.TODO(), cspObject, metav1.CreateOptions{})
	if err != nil {
		klog.Error(err)
	}
	return cspGot
}

// Create 5 fake pods that will compete to acquire lease on csp
func PodCreator(fakeKubeClient kubernetes.Interface, podName string) {
	for i := 1; i <= 5; i++ {
		podObjet := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: podName + strconv.Itoa(i),
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
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
	focs := &fakeClientset{
		oecs: openebsFakeClientset.NewSimpleClientset(),
	}

	fakeKubeClient := k8sfake.NewSimpleClientset()

	// Make a map of string(key) to struct(value).
	// Key of map describes test case behaviour.
	// Value of map is the test object.
	PodCreator(fakeKubeClient, "pool-pod")
	tests := map[string]struct {
		// fakestoragepoolclaim holds the fake storagepoolcalim object in test cases.
		fakestoragepoolclaim *apis.CStorPool
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
			fakestoragepoolclaim: focs.CspCreator("pool1", false, ""),
			podName:              "pool-pod1",
			podNamespace:         "openebs",
			expectedError:        false,
			expectedResult:       "{\"holder\":\"openebs/pool-pod1\",\"leaderTransition\":1}",
		},

		// TestCase#2
		"SPC#2 Lease already acquired": {
			fakestoragepoolclaim: focs.CspCreator("pool2", true, "openebs/pool-pod1"),
			podName:              "pool-pod2",
			podNamespace:         "openebs",
			expectedError:        true,
			expectedResult:       "{\"holder\":\"openebs/pool-pod1\",\"leaderTransition\":1}",
		},
		// TestCase#3
		"SPC#3 Lease already acquired": {
			fakestoragepoolclaim: focs.CspCreator("pool3", true, "openebs/pool-pod6"),
			podName:              "pool-pod2",
			podNamespace:         "openebs",
			expectedError:        false,
			expectedResult:       "{\"holder\":\"openebs/pool-pod2\",\"leaderTransition\":2}",
		},
		// TestCase#4
		"SPC#4 Lease Not acquired": {
			fakestoragepoolclaim: focs.CspCreator("pool4", true, ""),
			podName:              "pool-pod3",
			podNamespace:         "openebs",
			expectedError:        false,
			expectedResult:       "{\"holder\":\"openebs/pool-pod3\",\"leaderTransition\":2}",
		},
	}

	// Iterate over whole map to run the test cases.
	for name, test := range tests {
		test := test //pin it
		t.Run(name, func(t *testing.T) {
			var newCspLease Lease
			var gotError bool
			os.Setenv(string(PodName), test.podName)
			os.Setenv(string(NameSpace), test.podNamespace)
			newCspLease = Lease{test.fakestoragepoolclaim, CspLeaseKey, focs.oecs, fakeKubeClient}
			// Hold is the function under test.
			_, err := newCspLease.Hold()
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
			cspGot, _ := focs.oecs.OpenebsV1alpha1().CStorPools().
				Get(context.TODO(), test.fakestoragepoolclaim.Name, metav1.GetOptions{})
			if cspGot.Annotations[CspLeaseKey] != test.expectedResult {
				t.Errorf("Test case failed: expected lease value '%v' but got '%v' ", test.expectedResult, cspGot.Annotations[CspLeaseKey])

			}
			os.Unsetenv(string(PodName))
			os.Unsetenv(string(NameSpace))
		})
	}
}
