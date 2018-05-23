package poolcontroller

import (
	"testing"
	"time"

	openebsFakeClientset "github.com/openebs/maya/pkg/client/clientset/versioned/fake"
	informers "github.com/openebs/maya/pkg/client/informers/externalversions"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
)

// TestNewCStorPoolController tests if the kubernetes and openebs configs
// are present in pool controller instance.
func TestNewCStorPoolController(t *testing.T) {
	fakeKubeClient := fake.NewSimpleClientset()
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(fakeKubeClient, time.Second*30)
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)

	// Instantiate the cStor Pool controllers.
	poolController := NewCStorPoolController(fakeKubeClient, fakeOpenebsClient, kubeInformerFactory,
		openebsInformerFactory)

	if poolController.kubeclientset != fakeKubeClient {
		t.Fatalf("Pool controller object's kubeclientset mismatch")
	}
	if poolController.clientset != fakeOpenebsClient {
		t.Fatalf("Pool controller object's OpenebsClientset mismatch")
	}
}
