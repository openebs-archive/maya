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
	"testing"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
	openebsFakeClientset "github.com/openebs/maya/pkg/client/clientset/versioned/fake"
	informers "github.com/openebs/maya/pkg/client/informers/externalversions"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
)

// TestIsDeleteEvent function tests IsDeleteEvent function

// IsDeleteEvent function takes a kubernetes object as argument
// and returns a boolean value which tells whether the object is
// scheduled for deletion or not

func TestIsDeleteEvent(t *testing.T) {

	tests := map[string]struct {
		fakestoragepoolclaim *apis.StoragePoolClaim
		expectedResult       bool
	}{
		"DeletionTimestamp is nil": {
			fakestoragepoolclaim: &apis.StoragePoolClaim{
				ObjectMeta: api_meta.ObjectMeta{
					DeletionTimestamp: nil,
				},
			},
			expectedResult: false},
		"DeletionTimestamp is not nil": {
			fakestoragepoolclaim: &apis.StoragePoolClaim{
				ObjectMeta: api_meta.ObjectMeta{
					DeletionTimestamp: &api_meta.Time{time.Now(),
					},
				},
			},
			expectedResult: true},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsDeleteEvent(test.fakestoragepoolclaim)
			if result != test.expectedResult {
				t.Errorf("Test case failed: expected '%v' but got '%v' ", test.expectedResult, result)
			}
		})
	}
}

// TestNewCStorPoolController function tests if the kubernetes, openebs clientsets
// and other required objects are present in controller instance.

func TestNewCStorPoolController(t *testing.T) {
	fakeKubeClient := fake.NewSimpleClientset()
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(fakeKubeClient, time.Second*30)
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)
	// Instantiate the controller
	controller := NewController(fakeKubeClient, fakeOpenebsClient, kubeInformerFactory,
		openebsInformerFactory)

	if controller.spcSynced == nil{
		t.Errorf("No spc cache sync in controller object")
	}

	if controller.workqueue == nil{
		t.Errorf("No workqueue in controller object")
	}
	if controller.recorder == nil{
		t.Errorf("No recorder in controller object")
	}
	if controller.kubeclientset != fakeKubeClient {
		t.Errorf("SPC controller object's kubeclientset mismatch")
	}
	if controller.clientset != fakeOpenebsClient {
		t.Errorf("SPC controller object's openebsclientset mismatch")
	}
}
