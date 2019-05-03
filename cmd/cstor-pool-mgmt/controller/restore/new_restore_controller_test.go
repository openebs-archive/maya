/*
Copyright 2019 The OpenEBS Authors.

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
package restorecontroller

import (
	"testing"
	"time"

	openebsFakeClientset "github.com/openebs/maya/pkg/client/generated/clientset/internalclientset/fake"
	informers "github.com/openebs/maya/pkg/client/generated/informer/externalversions"

	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
)

// TestNewCStorRestoreController tests if the kubernetes and openebs
// configs are present in restore instance.
func TestNewCStorRestoreController(t *testing.T) {
	fakeKubeClient := fake.NewSimpleClientset()
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(fakeKubeClient, time.Second*30)
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)

	// Instantiate the cStor Restore controllers.
	restoreController := NewCStorRestoreController(fakeKubeClient, fakeOpenebsClient, kubeInformerFactory,
		openebsInformerFactory)

	if restoreController.kubeclientset != fakeKubeClient {
		t.Fatalf("Pool controller object's kubeclientset mismatch")
	}
	if restoreController.clientset != fakeOpenebsClient {
		t.Fatalf("Pool controller object's OpenebsClientset mismatch")
	}
}
