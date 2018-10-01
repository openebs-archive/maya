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
	//openebsFakeClientset "github.com/openebs/maya/pkg/client/clientset/versioned/fake"
	openebsFakeClientset "github.com/openebs/maya/pkg/client/generated/clientset/internalclientset/fake"

	//informers "github.com/openebs/maya/pkg/client/informers/externalversions"
	informers "github.com/openebs/maya/pkg/client/generated/informer/externalversions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
	"time"
)

// TestEnqueueSpc function test enqueueSpc function to check wether the queueload is
// properly formed for enqueue into the workqueue

func TestEnqueueSpc(t *testing.T) {
	// fakeKubeClient, fakeOpenebsClient, kubeInformerFactory, and openebsInformerFactory
	// are arguments that is expected by the NewController function.
	fakeKubeClient := fake.NewSimpleClientset()
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(fakeKubeClient, time.Second*30)
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)

	// Make a map of string(key) to struct(value).
	// Key of map describes test case behaviour.
	// Value of map is the test object.
	tests := map[string]struct {
		// fakestoragepoolclaim holds the fake storagepoolcalim object in test cases.
		fakestoragepoolclaim *apis.StoragePoolClaim
		// queueload holds the queueLoad for the test case under run.
		queueload QueueLoad
		// expectedKey holds the key that should be extracted from queueload
		expectedKey string
	}{
		// TestCase#1
		// Make a queueload object
		// Function under test should utilize the object and extract name from storagepoolcalim object as key
		// Finally the queueload key filed should be filled with the extracted key.
		"Forming queueload object": {
			queueload: QueueLoad{
				"",
				"operation",
				&apis.StoragePoolClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pool2",
					},
				},
			},
			expectedKey: "pool2"},
	}

	// Iterate over whole map to run the test cases.
	for name, test := range tests {
		// Instantiate the controller by passing the valid arguments.
		controller := NewController(fakeKubeClient, fakeOpenebsClient, kubeInformerFactory,
			openebsInformerFactory)
		t.Run(name, func(t *testing.T) {
			// enqueueSpc is the function under test.
			controller.enqueueSpc(&test.queueload)
			// If the key in queueload does not match the expectedKey, test case fails.
			if test.queueload.Key != test.expectedKey {
				t.Errorf("Test case failed : expected '%v' but got '%v' ", test.expectedKey, test.queueload.Key)
			}
		})
	}
}

// TestGetSpcResource tests getSpcResource.

// getSpcResource receives the name of the object and fetches the object
// from kube apiserver

func TestGetSpcResource(t *testing.T) {
	// fakeKubeClient, fakeOpenebsClient, kubeInformerFactory, and openebsInformerFactory
	// are arguments that is expected by the NewController function.
	fakeKubeClient := fake.NewSimpleClientset()
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(fakeKubeClient, time.Second*30)
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)

	// Make a map of string(key) to struct(value).
	// Key of map describes test case behaviour.
	// Value of map is the test object.
	tests := map[string]struct {
		// querySpcName is the name of object that will be queried.
		querySpcName string
		// expectedError tells whether the error should be there or not.
		expectedError bool
		// SpcObject is storagepoolclaim object whose fake creation will be done.
		SpcObject *apis.StoragePoolClaim
	}{
		// TestCase#2
		// Make a storagepoolcalim object named pool1.
		// Query for pool1 object and there should be no error.
		"Create spc object with name pool1 and query for the same object": {
			querySpcName:  "pool1",
			expectedError: false,
			SpcObject: &apis.StoragePoolClaim{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name: "pool1",
					UID:  types.UID("abc"),
				},
				Spec: apis.StoragePoolClaimSpec{
					Disks: apis.DiskAttr{
						DiskList: []string{"disk-4268137899842721d2d4fc0c16c3b138"},
					},
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool1.cache",
						PoolType:         "mirror",
						OverProvisioning: false,
					},
				},
				Status: apis.StoragePoolClaimStatus{},
			},
		},

		// TestCase#2
		// Make a storagepoolcalim object named pool3.
		// Query for pool4 object which does not exists.
		// Still the function under test should not error out but
		// the error should be handled at runtime
		"Create spc object with name pool3 and query for pool4 object which does not exists": {
			querySpcName:  "pool4",
			expectedError: false,
			SpcObject: &apis.StoragePoolClaim{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name: "pool3",
					UID:  types.UID("abcd"),
				},
				Spec: apis.StoragePoolClaimSpec{
					Disks: apis.DiskAttr{
						DiskList: []string{"disk-49c3f6bfe9906e8db04adda12815375c"},
					},
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool3.cache",
						PoolType:         "striped",
						OverProvisioning: false,
					},
				},
				Status: apis.StoragePoolClaimStatus{},
			},
		},
	}

	// Iterate over whole map to run the test cases.
	for name, test := range tests {
		// Instantiate the controller by passing the valid arguments.
		controller := NewController(fakeKubeClient, fakeOpenebsClient, kubeInformerFactory,
			openebsInformerFactory)
		t.Run(name, func(t *testing.T) {
			resultError := false

			// Create the fake storagepoolclaim object
			_, err := controller.clientset.OpenebsV1alpha1().StoragePoolClaims().Create(test.SpcObject)
			if err != nil {
				t.Fatalf("Desc:%v, Unable to create resource : %v", name, test.SpcObject.ObjectMeta.Name)
			}

			_, err = controller.getSpcResource(test.querySpcName)

			// If any error occurs set resultError true
			if err != nil {
				resultError = true
			}

			// If expectedError does not matches resultError, test case fails.
			if test.expectedError != resultError {
				t.Errorf("Test case failed : expected '%v' but got '%v' ", test.expectedError, resultError)
			}

		})
	}

}

// TestSyncHandler function tests syncHandler function which call the business handlers

func TestSyncHandler(t *testing.T) {
	// fakeKubeClient, fakeOpenebsClient, kubeInformerFactory, and openebsInformerFactory
	// are arguments that is expected by the NewController function.
	fakeKubeClient := fake.NewSimpleClientset()
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(fakeKubeClient, time.Second*30)
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)

	// Make a map of string(key) to struct(value).
	// Key of map describes test case behaviour.
	// Value of map is the test object.
	tests := map[string]struct {
		// key holds the name of storagepoolcalim object
		key string
		// operations holds the type of operation that happened for storagepoolclaim object
		operation string
		// fakestoragepoolclaim holds the fake storagepoolcalim object in test cases.
		fakestoragepoolclaim *apis.StoragePoolClaim
		// expectedError tells whether the error should be there or not.
		expectedError bool
	}{
		// Function under test expects key,operation,and storagepoolcalim object as an argument
		// If the event is deleteEvent or addEvent , creation of storagepool and deletion of storagepool
		// should be attempted.
		// The attempt will fail because of the absence of cas template in go environment
		// Hence deleteEvent and addEvent should error out.
		// For all other events there is no such attempt and no error should occur.

		// TestCase#1
		"Sync Operation for spc delete event": {
			key:       "pool1",
			operation: deleteEvent,
			fakestoragepoolclaim: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pool1",
				},
			},
			expectedError: true,
		},

		// TestCase#2
		"Sync Operation for spc for any unrecognized operation tag": {
			key:       "pool1",
			operation: "Default Case In Switch",
			fakestoragepoolclaim: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pool1",
				},
			},
			expectedError: false,
		},
		// TestCase#3
		"Sync Operation for spc add event": {
			key:       "pool1",
			operation: addEvent,
			fakestoragepoolclaim: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pool1",
				},
			},
			expectedError: true,
		},

		// TestCase#4
		"Sync Operation for spc update event": {
			key:       "pool2",
			operation: updateEvent,
			fakestoragepoolclaim: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pool2",
				},
			},
			expectedError: false,
		},

		// TestCase#5
		"Sync Operation for spc ignored event": {
			key:       "pool2",
			operation: ignoreEvent,
			fakestoragepoolclaim: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pool2",
				},
			},
			expectedError: false,
		},

		// TestCase#6
		"Sync Operation for spc delete event when the function receives nil object": {
			key:                  "pool2",
			operation:            deleteEvent,
			fakestoragepoolclaim: nil,
			expectedError:        true,
		},
	}

	for name, test := range tests {
		// Instantiate the controller by passing the valid arguments.
		controller := NewController(fakeKubeClient, fakeOpenebsClient, kubeInformerFactory,
			openebsInformerFactory)
		t.Run(name, func(t *testing.T) {
			if test.operation == addEvent || test.operation == updateEvent {
				// For addEvent and updateEvent storagepoolclaim object should exist
				// Hence creating the objects
				_, err := controller.clientset.OpenebsV1alpha1().StoragePoolClaims().Create(test.fakestoragepoolclaim)
				if err != nil {
					t.Fatalf("Desc:%v, Unable to create resource : %s", name, test.key)
				}
			}

			resultError := false
			err := controller.syncHandler(test.key, test.operation, test.fakestoragepoolclaim)

			if err != nil {
				resultError = true
			}

			if test.expectedError != resultError {
				t.Errorf("Test case failed : expected '%v' but got '%v' ", err, test.expectedError)
			}
		})
	}
}
