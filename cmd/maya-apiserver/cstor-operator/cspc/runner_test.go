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

package cspc

import (
	openebsFakeClientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned/fake"
	informers "github.com/openebs/maya/pkg/client/generated/informers/externalversions"
	"github.com/openebs/maya/pkg/signals"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
	"time"
)

// TestRun is to run cspc controller and check if it crashes or returns back.
func TestRun(t *testing.T) {
	// fakeKubeClient, fakeOpenebsClient, kubeInformerFactory, and openebsInformerFactory
	// are arguments that is expected by the NewController function.
	fakeKubeClient := fake.NewSimpleClientset()
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)
	// Instantiate the controller by passing the valid arguments.
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
