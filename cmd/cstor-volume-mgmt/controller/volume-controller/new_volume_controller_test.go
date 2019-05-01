// Copyright © 2017-2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package volumecontroller

import (
	"testing"
	"time"

	//openebsFakeClientset "github.com/openebs/maya/pkg/client/clientset/versioned/fake"
	openebsFakeClientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned/fake"

	//informers "github.com/openebs/maya/pkg/client/informers/externalversions"
	informers "github.com/openebs/maya/pkg/client/generated/informers/externalversions"

	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
)

// TestNewCStorVolumeController tests if the kubernetes and openebs configs
// are present in volume controller instance.
func TestNewCStorVolumeController(t *testing.T) {
	fakeKubeClient := fake.NewSimpleClientset()
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(fakeKubeClient, time.Second*30)
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)

	// Instantiate the cStor Volume controllers.
	volumeController := NewCStorVolumeController(fakeKubeClient, fakeOpenebsClient, kubeInformerFactory,
		openebsInformerFactory)

	if volumeController.kubeclientset != fakeKubeClient {
		t.Fatalf("Volume controller object's kubeclientset mismatch")
	}
	if volumeController.clientset != fakeOpenebsClient {
		t.Fatalf("Volume controller object's OpenebsClientset mismatch")
	}
}
