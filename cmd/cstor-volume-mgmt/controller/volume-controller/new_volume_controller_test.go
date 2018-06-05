package volumecontroller

import (
	"testing"
	"time"

	openebsFakeClientset "github.com/openebs/maya/pkg/client/clientset/versioned/fake"
	informers "github.com/openebs/maya/pkg/client/informers/externalversions"
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
