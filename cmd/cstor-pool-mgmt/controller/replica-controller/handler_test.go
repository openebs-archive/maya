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
	"context"
	"testing"
	"time"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	openebsFakeClientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned/fake"
	informers "github.com/openebs/maya/pkg/client/generated/informers/externalversions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
)

// TestGetPoolResource checks if volume replica resource created
// is successfully got.
func TestGetVolumeReplicaResource(t *testing.T) {
	fakeKubeClient := fake.NewSimpleClientset()
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(fakeKubeClient, time.Second*30)
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)

	// Instantiate the cStor VolumeReplica controllers.
	volumeReplicaController := NewCStorVolumeReplicaController(fakeKubeClient, fakeOpenebsClient, kubeInformerFactory,
		openebsInformerFactory)

	testVolumeReplicaResource := map[string]struct {
		expectedName string
		test         *apis.CStorVolumeReplica
	}{
		"VolumeReplicaResource1": {
			expectedName: "VolumeReplicaResource1",
			test: &apis.CStorVolumeReplica{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "VolumeReplicaResource1",
					Namespace: "default",
				},
				Spec: apis.CStorVolumeReplicaSpec{
					TargetIP: "10.210.110.121",
					Capacity: "100MB",
				},
			},
		},
		"VolumeReplicaResource2": {
			expectedName: "VolumeReplicaResource2",
			test: &apis.CStorVolumeReplica{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "VolumeReplicaResource2",
					Namespace: "default",
				},
				Spec: apis.CStorVolumeReplicaSpec{
					TargetIP: "10.210.110.121",
					Capacity: "100MB",
				},
			},
		},
	}
	for desc, ut := range testVolumeReplicaResource {
		// Create a volume-replica resource.
		_, err := volumeReplicaController.clientset.OpenebsV1alpha1().CStorVolumeReplicas(ut.test.ObjectMeta.Namespace).
			Create(context.TODO(), ut.test, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("Desc:%v, Unable to create resource : %v", desc, ut.test.ObjectMeta.Name)
		}

		// Get volume replica resource with name
		cStorVolumeReplicaObtained, err := volumeReplicaController.getVolumeReplicaResource(ut.test.ObjectMeta.Namespace + "/" + ut.test.ObjectMeta.Name)

		if cStorVolumeReplicaObtained.Name != ut.expectedName {
			t.Fatalf("Desc:%v, volName mismatch, Expected:%v, Got:%v", desc, ut.expectedName,
				cStorVolumeReplicaObtained.Name)
		}
	}
}
