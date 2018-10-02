/*
Copyright 2018 The OpenEBS Authors.

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
package replicacontroller

import (
	"testing"
	"time"

	//openebsFakeClientset "github.com/openebs/maya/pkg/client/clientset/versioned/fake"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	openebsFakeClientset "github.com/openebs/maya/pkg/client/generated/clientset/internalclientset/fake"

	//informers "github.com/openebs/maya/pkg/client/informers/externalversions"
	informers "github.com/openebs/maya/pkg/client/generated/informer/externalversions"

	"github.com/openebs/maya/cmd/cstor-pool-mgmt/controller/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
)

// TestNewCStorVolumeReplicaController tests if the kubernetes and openebs
// configs are present in volume replica controller instance.
func TestNewCStorVolumeReplicaController(t *testing.T) {
	fakeKubeClient := fake.NewSimpleClientset()
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(fakeKubeClient, time.Second*30)
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)

	// Instantiate the cStor VolumeReplica controllers.
	volumeReplicaController := NewCStorVolumeReplicaController(fakeKubeClient, fakeOpenebsClient, kubeInformerFactory,
		openebsInformerFactory)

	if volumeReplicaController.kubeclientset != fakeKubeClient {
		t.Fatalf("Pool controller object's kubeclientset mismatch")
	}
	if volumeReplicaController.clientset != fakeOpenebsClient {
		t.Fatalf("Pool controller object's OpenebsClientset mismatch")
	}
}

func TestEnqueueCStorReplica(t *testing.T) {
	fakeKubeClient := fake.NewSimpleClientset()
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(fakeKubeClient, time.Second*30)
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)

	// Instantiate the cStor VolumeReplica controllers.
	volumeReplicaController := NewCStorVolumeReplicaController(fakeKubeClient, fakeOpenebsClient, kubeInformerFactory,
		openebsInformerFactory)
	test := apis.CStorVolumeReplica{
		ObjectMeta: metav1.ObjectMeta{
			Name: "VolumeReplicaResource1",
			UID:  "abcd123",
			Labels: map[string]string{
				"cstorpool.openebs.io/uid":  "123abc",
				"cstorpool.openebs.io/name": "cstor-123abc",
			},
		},
		Spec: apis.CStorVolumeReplicaSpec{
			TargetIP: "127.0.0.1",
			Capacity: "100MB",
		},
	}
	q := common.QueueLoad{}
	volumeReplicaController.enqueueCStorReplica(&test, q)
	if volumeReplicaController.workqueue.Len() != 1 {
		t.Fatalf("Queue is empty")
	}
}
