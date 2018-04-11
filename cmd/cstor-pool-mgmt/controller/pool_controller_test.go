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

// TestGetPoolResource checks if pool resource created is successfully got.
func TestGetPoolResource(t *testing.T) {
	fakeKubeClient := fake.NewSimpleClientset()
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(fakeKubeClient, time.Second*30)
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)

	// Instantiate the cStor Pool controllers.
	poolController := NewCStorPoolController(fakeKubeClient, fakeOpenebsClient, kubeInformerFactory,
		openebsInformerFactory)

	testPoolResource := map[string]struct {
		expectedPoolName string
		test             *apis.CStorPool
	}{
		"poolResource1": {
			expectedPoolName: "pool1",
			test: &apis.CStorPool{
				ObjectMeta: metav1.ObjectMeta{Name: "poolResource1"},
				Spec: apis.CStorPoolSpec{
					Disks: apis.DiskAttr{
						DiskList: []string{"/tmp/img1.img"},
					},
					PoolSpec: apis.CStorPoolAttr{
						PoolName:  "pool1",
						CacheFile: "/tmp/pool1.cache",
						PoolType:  "mirror",
					},
				},
			},
		},
		"poolResource2": {
			expectedPoolName: "pool2",
			test: &apis.CStorPool{
				ObjectMeta: metav1.ObjectMeta{Name: "poolResource2"},
				Spec: apis.CStorPoolSpec{
					Disks: apis.DiskAttr{
						DiskList: []string{"/tmp/img2.img"},
					},
					PoolSpec: apis.CStorPoolAttr{
						PoolName:  "pool2",
						CacheFile: "/tmp/pool2.cache",
						PoolType:  "mirror",
					},
				},
			},
		},
	}
	for desc, ut := range testPoolResource {
		// Create Pool resource
		_, err := poolController.clientset.OpenebsV1alpha1().CStorPools().Create(ut.test)
		if err != nil {
			t.Fatalf("Desc:%v, Unable to create resource : %v", desc, ut.test.ObjectMeta.Name)
		}

		// Get the created pool resource using name
		cStorPoolObtained, err := poolController.getPoolResource(ut.test.ObjectMeta.Name, "")

		if cStorPoolObtained.Spec.PoolSpec.PoolName != ut.expectedPoolName {
			t.Fatalf("Desc:%v, PoolName mismatch, Expected:%v, Got:%v", desc, ut.expectedPoolName,
				cStorPoolObtained.Spec.PoolSpec.PoolName)
		}
	}
}
