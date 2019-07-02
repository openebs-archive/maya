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
package cspc

import (
	openebsFakeClientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned/fake"
	informers "github.com/openebs/maya/pkg/client/generated/informers/externalversions"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
	"time"
)

func TestNewControllerBuilder(t *testing.T) {
	fakeKubeClient := fake.NewSimpleClientset()
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)
	controller, err := NewControllerBuilder().
		withKubeClient(fakeKubeClient).
		withOpenEBSClient(fakeOpenebsClient).
		withCSPCSynced(openebsInformerFactory).
		withCSPCLister(openebsInformerFactory).
		withRecorder(fakeKubeClient).
		withWorkqueueRateLimiting().
		withEventHandler(openebsInformerFactory).
		Build()

	if err != nil {
		t.Fatalf("failed to build controller instance: %s", err)
	}
	// TestCase#1
	// Check for cache sync in controller instance.
	if controller.cspcSynced == nil {
		t.Errorf("No cspc cache sync in controller object")
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
		t.Errorf("CSPC controller object's kubeclientset mismatch")
	}

	// TestCase#5
	// Check for obenebsclientset in controller instance.
	if controller.clientset != fakeOpenebsClient {
		t.Errorf("CSPC controller object's openebsclientset mismatch")
	}
}
