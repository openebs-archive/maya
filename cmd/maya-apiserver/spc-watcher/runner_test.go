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
	openebsFakeClientset "github.com/openebs/maya/pkg/client/clientset/versioned/fake"
	informers "github.com/openebs/maya/pkg/client/informers/externalversions"
	"github.com/openebs/maya/pkg/signals"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
	"time"
)

// TestRun is to run spc controller and check if it crashes or returns back.
func TestRun(t *testing.T) {
	// fakeKubeClient, fakeOpenebsClient, kubeInformerFactory, and openebsInformerFactory
	// are arguments that is expected by the NewController function.
	fakeKubeClient := fake.NewSimpleClientset()
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(fakeKubeClient, time.Second*30)
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)

	// Instantiate the controller by passing the valid arguments.
	controller := NewController(fakeKubeClient, fakeOpenebsClient, kubeInformerFactory,
		openebsInformerFactory)

	stopCh := signals.SetupSignalHandler()
	done := make(chan bool)
	go func(chan bool) {
		// Run is the function under test.
		controller.Run(2, stopCh)
		done <- true
	}(done)

	select {
	case <-time.After(3 * time.Second):

	case <-done:
		t.Fatalf("Controller returned - failure")

	}
}

// TestProcessNextWorkItemAdd is to test a spc resource for add event.
func TestProcessNextWorkItem(t *testing.T) {
	// fakeKubeClient, fakeOpenebsClient, kubeInformerFactory, and openebsInformerFactory
	// are arguments that is expected by the NewController function.
	fakeKubeClient := fake.NewSimpleClientset()
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(fakeKubeClient, time.Second*30)
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)

	// sampleTimestamp will be used to set deletion timestamp in object under test
	sampleTimestamp := time.Now()

	// Make a map of string(key) to struct(value).
	// Key of map describes test case behaviour.
	// Value of map is the test object.
	testSpcResource := map[string]struct {
		// expectedOutput holds the expected return value by the function under test.
		expectedOutput bool
		// spcObject holds the fake storagepoolclaim object.
		spcObject *apis.StoragePoolClaim
		// operation tells the type of event for the storagepoolclaim object.
		operation string
	}{
		// For any event type of spc resource the function should always return true.
		// TestCase#1
		"Add event of spc resource": {
			expectedOutput: true,
			operation:      addEvent,
			spcObject: &apis.StoragePoolClaim{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name: "pool2",
					UID:  types.UID("abcd"),
				},
				Spec: apis.StoragePoolClaimSpec{
					Disks: apis.DiskAttr{
						DiskList: []string{"disk-49c3f6bfe9906e8db04adda12815375c", "disk-99cde73d1defa35375029e8164e974e0"},
					},
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool2.cache",
						PoolType:         "striped",
						OverProvisioning: false,
					},
				},
				Status: apis.StoragePoolClaimStatus{},
			},
		},

		// TestCase#2
		"Delete event of spc resource": {
			expectedOutput: true,
			operation:      deleteEvent,
			spcObject: &apis.StoragePoolClaim{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:              "pool1",
					UID:               types.UID("abcd"),
					DeletionTimestamp: &metav1.Time{sampleTimestamp},
				},
				Spec: apis.StoragePoolClaimSpec{
					Disks: apis.DiskAttr{
						DiskList: []string{"disk-49c3f6bfe9906e8db04adda12815375c", "disk-99cde73d1defa35375029e8164e974e0"},
					},
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool1.cache",
						PoolType:         "striped",
						OverProvisioning: false,
					},
				},
				Status: apis.StoragePoolClaimStatus{},
			},
		},
	}

	// Iterate over whole map to run the test cases.
	for name, test := range testSpcResource {
		// Instantiate the controller by passing the valid arguments.
		controller := NewController(fakeKubeClient, fakeOpenebsClient, kubeInformerFactory,
			openebsInformerFactory)

		t.Run(name, func(t *testing.T) {
			_, err := controller.clientset.OpenebsV1alpha1().StoragePoolClaims().Create(test.spcObject)
			if err != nil {
				t.Fatalf("Unable to create resource : %v", test.spcObject.ObjectMeta.Name)
			}
			// Forming the queueload object
			q := &QueueLoad{}
			q.Key = test.spcObject.Name
			q.Operation = test.operation
			q.Object = test.spcObject
			// Adding to queue
			controller.workqueue.AddRateLimited(q)
			// processNextWorkItem is the function under test
			obtainedOutput := controller.processNextWorkItem()
			if obtainedOutput != test.expectedOutput {
				t.Fatalf("Expected:%v, Got:%v", test.expectedOutput,
					obtainedOutput)
			}
		})

	}
}
