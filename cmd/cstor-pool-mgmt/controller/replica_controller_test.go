package controller

import (
	"testing"
	"time"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	openebsFakeClientset "github.com/openebs/maya/pkg/client/clientset/versioned/fake"
	informers "github.com/openebs/maya/pkg/client/informers/externalversions"
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
		expectedVolName string
		test            *apis.CStorVolumeReplica
	}{
		"VolumeReplicaResource1": {
			expectedVolName: "vol1",
			test: &apis.CStorVolumeReplica{
				ObjectMeta: metav1.ObjectMeta{Name: "VolumeReplicaResource1"},
				Spec: apis.CStorVolumeReplicaSpec{
					CStorControllerIP: "10.210.110.121",
					VolName:           "vol1",
					Capacity:          "100MB",
				},
			},
		},
		"VolumeReplicaResource2": {
			expectedVolName: "abcdefgh_Volume_2",
			test: &apis.CStorVolumeReplica{
				ObjectMeta: metav1.ObjectMeta{Name: "VolumeReplicaResource2"},
				Spec: apis.CStorVolumeReplicaSpec{
					CStorControllerIP: "10.210.110.121",
					VolName:           "abcdefgh_Volume_2",
					Capacity:          "100MB",
				},
			},
		},
	}
	for desc, ut := range testVolumeReplicaResource {
		// Create a volume-replica resource.
		_, err := volumeReplicaController.clientset.OpenebsV1alpha1().CStorVolumeReplicas().Create(ut.test)
		if err != nil {
			t.Fatalf("Desc:%v, Unable to create resource : %v", desc, ut.test.ObjectMeta.Name)
		}

		// Get volume replica resource with name
		cStorVolumeReplicaObtained, err := volumeReplicaController.getVolumeReplicaResource(ut.test.ObjectMeta.Name, "")

		if cStorVolumeReplicaObtained.Spec.VolName != ut.expectedVolName {
			t.Fatalf("Desc:%v, volName mismatch, Expected:%v, Got:%v", desc, ut.expectedVolName,
				cStorVolumeReplicaObtained.Spec.VolName)
		}
	}
}
