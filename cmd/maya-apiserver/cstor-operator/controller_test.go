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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	//openebsFakeClientset "github.com/openebs/maya/pkg/client/clientset/versioned/fake"
	openebsFakeClientset "github.com/openebs/maya/pkg/client/generated/clientset/internalclientset/fake"

	//informers "github.com/openebs/maya/pkg/client/informers/externalversions"
	informers "github.com/openebs/maya/pkg/client/generated/informer/externalversions"

	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"reflect"
	"testing"
	"time"
)

// TestIsDeleteEvent function tests IsDeleteEvent function.

// IsDeleteEvent function takes a kubernetes object as argument
// and returns true if the object is scheduled for deletion else false.

func TestIsDeleteEvent(t *testing.T) {
	// Make a map of string(key) to struct(value).
	// Key of map describes test case behaviour.
	// Value of map is the test object.
	tests := map[string]struct {
		// fakestoragepoolclaim holds the fake storagepoolcalim object in test cases.
		fakestoragepoolclaim *apis.StoragePoolClaim
		// expectedResult holds the expected result for the test case under run.
		expectedResult bool
	}{
		// TestCase#1
		// Make a storagepoolcalim object and set its DeletionTimestamp to nil
		// nil DeletionTimestamp will mean that the object is not scheduled to be deleted
		// Hence, expected result is false
		"DeletionTimestamp is nil": {
			fakestoragepoolclaim: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					DeletionTimestamp: nil,
				},
			},
			expectedResult: false,
		},

		// TestCase#2
		// Make a storagepoolcalim object and set its DeletionTimestamp to current time
		// nil DeletionTimestamp will mean that the object is not scheduled to be deleted
		// Hence, expected result is true
		"DeletionTimestamp is not nil": {
			fakestoragepoolclaim: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					DeletionTimestamp: &metav1.Time{time.Now()},
				},
			},
			expectedResult: true,
		},
	}

	// Iterate over whole map to run the test cases.
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// IsDeleteEvent is the function under test.
			// Pass the fake storagepoolclaim(i.e. test.fakestoragepoolclaim ) object to the function
			result := IsDeleteEvent(test.fakestoragepoolclaim)
			// If the result does not matches expectedResult, test case fails.
			if result != test.expectedResult {
				t.Errorf("Test case failed: expected '%v' but got '%v' ", test.expectedResult, result)
			}
		})
	}
}

// TestNewController function tests NewController function

// NewController function returns a controller instance with
// some init objects that are required for watcher functionality

// For e.g the kubernetes and openebs clientsets, workqueue, recorder
// and other required objects are present in controller instance.

func TestNewController(t *testing.T) {
	// fakeKubeClient, fakeOpenebsClient, kubeInformerFactory, and openebsInformerFactory
	// are arguments that is expected by the NewController function.
	fakeKubeClient := fake.NewSimpleClientset()
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(fakeKubeClient, time.Second*30)
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)

	// Instantiate the controller by passing the valid arguments.
	controller := NewController(fakeKubeClient, fakeOpenebsClient, kubeInformerFactory,
		openebsInformerFactory)

	// TestCase#1
	// Check for cache sync in controller instance.
	if controller.spcSynced == nil {
		t.Errorf("No spc cache sync in controller object")
	}

	// TestCase#2
	// Check for workqueue in controller instance
	if controller.workqueue == nil {
		t.Errorf("No workqueue in controller object")
	}

	// TestCase#3
	// Check for recoder in controller instance.
	if controller.recorder == nil {
		t.Errorf("No recorder in controller object")
	}

	// TestCase#4
	// Check for kubeclientset in controller instance.
	if controller.kubeclientset != fakeKubeClient {
		t.Errorf("SPC controller object's kubeclientset mismatch")
	}

	// TestCase#5
	// Check for obenebsclientset in controller instance.
	if controller.clientset != fakeOpenebsClient {
		t.Errorf("SPC controller object's openebsclientset mismatch")
	}
}

// TestAddSpc function tests if addSpc function is properly forming the queueload that
// is pushed into the workqueue, as according to add event of storagepoolclaim object

func TestAddSpc(t *testing.T) {
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
		// expectedQueueLoad holds the expected queueLoad for the test case under run.
		expectedQueueLoad QueueLoad
	}{
		// TestCase#1
		// Make a storagepoolcalim object
		// Its creation should from the queueload as in the expectedQueueLoad
		"Add event of spc object": {
			fakestoragepoolclaim: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pool1",
				},
			},
			expectedQueueLoad: QueueLoad{"pool1", addEvent, &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pool1",
				},
			},
			},
		},
	}

	// Iterate over whole map to run the test cases.
	for name, test := range tests {
		// Instantiate the controller by passing the valid arguments.
		controller := NewController(fakeKubeClient, fakeOpenebsClient, kubeInformerFactory,
			openebsInformerFactory)
		t.Run(name, func(t *testing.T) {
			// addSpc is the function under test.
			controller.addSpc(test.fakestoragepoolclaim)
			// If the controller instance queueload does not matches expectedQueueLoad, test case fails.
			if !reflect.DeepEqual(test.expectedQueueLoad, controller.queueLoad) {
				t.Errorf("Test case failed: expected '%+v' but got '%+v' ", test.expectedQueueLoad, controller.queueLoad)
				t.Errorf("Object difference: expected '%v' but got '%v' ", test.expectedQueueLoad.Object.(*apis.StoragePoolClaim), controller.queueLoad.Object.(*apis.StoragePoolClaim))
			}
		})
	}
}

// TestUpdateSpc function tests if updateSpc function is properly forming the queueload that
// is pushed into the workqueue, as according to update event of storagepoolclaim object.

func TestUpdateSpc(t *testing.T) {
	// fakeKubeClient, fakeOpenebsClient, kubeInformerFactory, and openebsInformerFactory
	// are arguments that is expected by the NewController function.
	fakeKubeClient := fake.NewSimpleClientset()
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(fakeKubeClient, time.Second*30)
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)

	// Take a timestamp that will be used in test cases
	sampleTimestamp := time.Now()

	// Make a map of string(key) to struct(value).
	// Key of map describes test case behaviour.
	// Value of map is the test object.
	tests := map[string]struct {
		// fakestoragepoolclaimOld holds the fake old storagepoolcalim object in test cases.
		fakestoragepoolclaimOld *apis.StoragePoolClaim
		// fakestoragepoolclaimNew holds the fake new storagepoolcalim object in test cases.
		fakestoragepoolclaimNew *apis.StoragePoolClaim
		// expectedQueueLoad holds the expected queueLoad for the test case under run.
		expectedQueueLoad QueueLoad
	}{
		// TestCase#1
		// Make a two storagepoolcalim objects i.e. fakestoragepoolclaimOld & fakestoragepoolclaimNew.
		// Keep the resource version for both the objects same.
		// The expected queueload should be as it is in expectedQueueLoad
		"Update event of spc object when RV does not change": {
			fakestoragepoolclaimOld: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "pool1",
					ResourceVersion: "111232",
				},
			},
			fakestoragepoolclaimNew: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "pool1",
					ResourceVersion: "111232",
				},
			},
			expectedQueueLoad: QueueLoad{"pool1", syncEvent, &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "pool1",
					ResourceVersion: "111232",
				},
			},
			},
		},

		// TestCase#2
		// Make a two storagepoolcalim objects i.e. fakestoragepoolclaimOld & fakestoragepoolclaimNew.
		// Keep the resource version for both the objects different.
		// The expected queueload should be as it is in expectedQueueLoad
		"Update event of spc object when RV changes": {
			fakestoragepoolclaimOld: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "pool1",
					ResourceVersion: "111232",
				},
			},
			fakestoragepoolclaimNew: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "pool1",
					ResourceVersion: "111235",
				},
			},
			expectedQueueLoad: QueueLoad{"pool1", updateEvent, &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "pool1",
					ResourceVersion: "111235",
				},
			},
			},
		},

		// TestCase#3
		// Make a two storagepoolcalim objects i.e. fakestoragepoolclaimOld & fakestoragepoolclaimNew.
		// Keep the resource version for both the objects different.
		// Set the DeletionTimestamp for the fakestoragepoolclaimNew object.
		// The expected queueload should be as it is in expectedQueueLoad
		"Update event of spc object when deletion scheduled": {
			fakestoragepoolclaimOld: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "pool1",
					ResourceVersion:   "111235",
					DeletionTimestamp: nil,
				},
			},
			fakestoragepoolclaimNew: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "pool1",
					ResourceVersion:   "111935",
					DeletionTimestamp: &metav1.Time{sampleTimestamp},
				},
			},
			expectedQueueLoad: QueueLoad{"", ignoreEvent, nil},
		},
	}

	// Iterate over whole map to run the test cases.
	for name, test := range tests {
		// Instantiate the controller by passing the valid arguments.
		controller := NewController(fakeKubeClient, fakeOpenebsClient, kubeInformerFactory,
			openebsInformerFactory)
		t.Run(name, func(t *testing.T) {
			// updateSpc is the function under test.
			controller.updateSpc(test.fakestoragepoolclaimOld, test.fakestoragepoolclaimNew)
			// If the controller instance queueload does not matches expectedQueueLoad, test case fails.
			if !reflect.DeepEqual(test.expectedQueueLoad, controller.queueLoad) {
				t.Errorf("Test case failed: expected '%+v' but got '%+v:object ' ", test.expectedQueueLoad, controller.queueLoad)
			}
		})
	}
}
