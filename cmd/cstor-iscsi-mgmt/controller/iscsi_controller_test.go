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

// TestNewCStorIscsiController tests if the kubernetes and openebs configs
// are present in iscsi controller instance.
func TestNewCStorIscsiController(t *testing.T) {
	fakeKubeClient := fake.NewSimpleClientset()
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(fakeKubeClient, time.Second*30)
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)

	// Instantiate the cStor Iscsi controllers.
	iscsiController := NewCStorIscsiController(fakeKubeClient, fakeOpenebsClient, kubeInformerFactory,
		openebsInformerFactory)

	if iscsiController.kubeclientset != fakeKubeClient {
		t.Fatalf("Iscsi controller object's kubeclientset mismatch")
	}
	if iscsiController.clientset != fakeOpenebsClient {
		t.Fatalf("Iscsi controller object's OpenebsClientset mismatch")
	}
}

// TestGetIscsiResource checks if iscsi resource created is successfully got.
func TestGetIscsiResource(t *testing.T) {
	fakeKubeClient := fake.NewSimpleClientset()
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(fakeKubeClient, time.Second*30)
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)

	// Instantiate the cStor Iscsi controllers.
	iscsiController := NewCStorIscsiController(fakeKubeClient, fakeOpenebsClient, kubeInformerFactory,
		openebsInformerFactory)

	testIscsiResource := map[string]struct {
		expectedIscsiName string
		test             *apis.CStorIscsi
	}{
		"iscsiResource1": {
			expectedIscsiName: "iscsi1",
			test: &apis.CStorIscsi{
				ObjectMeta: metav1.ObjectMeta{Name: "iscsiResource1"},
				Spec: apis.CStorIscsiSpec{
					Disks: apis.DiskAttr{
						DiskList: []string{"/tmp/img1.img"},
					},
					IscsiSpec: apis.CStorIscsiAttr{
						IscsiName:  "iscsi1",
						CacheFile: "/tmp/iscsi1.cache",
						IscsiType:  "mirror",
					},
				},
			},
		},
		"iscsiResource2": {
			expectedIscsiName: "iscsi2",
			test: &apis.CStorIscsi{
				ObjectMeta: metav1.ObjectMeta{Name: "iscsiResource2"},
				Spec: apis.CStorIscsiSpec{
					Disks: apis.DiskAttr{
						DiskList: []string{"/tmp/img2.img"},
					},
					IscsiSpec: apis.CStorIscsiAttr{
						IscsiName:  "iscsi2",
						CacheFile: "/tmp/iscsi2.cache",
						IscsiType:  "mirror",
					},
				},
			},
		},
	}
	for desc, ut := range testIscsiResource {
		// Create Iscsi resource
		_, err := iscsiController.clientset.OpenebsV1alpha1().CStorIscsis().Create(ut.test)
		if err != nil {
			t.Fatalf("Desc:%v, Unable to create resource : %v", desc, ut.test.ObjectMeta.Name)
		}

		// Get the created iscsi resource using name
		cStorIscsiObtained, err := iscsiController.getIscsiResource(ut.test.ObjectMeta.Name, "")

		if cStorIscsiObtained.Spec.IscsiSpec.IscsiName != ut.expectedIscsiName {
			t.Fatalf("Desc:%v, IscsiName mismatch, Expected:%v, Got:%v", desc, ut.expectedIscsiName,
				cStorIscsiObtained.Spec.IscsiSpec.IscsiName)
		}
	}
}
