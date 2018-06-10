package volumecontroller

import (
	"os"
	"testing"
	"time"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	openebsFakeClientset "github.com/openebs/maya/pkg/client/clientset/versioned/fake"
	informers "github.com/openebs/maya/pkg/client/informers/externalversions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
)

// TestGetVolumeResource checks if volume resource created is successfully got.
func TestGetVolumeResource(t *testing.T) {
	fakeKubeClient := fake.NewSimpleClientset()
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(fakeKubeClient, time.Second*30)
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)

	// Instantiate the cStor Volume controllers.
	volumeController := NewCStorVolumeController(fakeKubeClient, fakeOpenebsClient, kubeInformerFactory,
		openebsInformerFactory)

	testVolumeResource := map[string]struct {
		expectedVolumeName string
		test               *apis.CStorVolume
	}{
		"img1VolumeResource": {
			expectedVolumeName: "abc",
			test: &apis.CStorVolume{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name: "volume1",
					UID:  types.UID("abc"),
				},
				Spec: apis.CStorVolumeSpec{
					CStorControllerIP: "0.0.0.0",
					VolumeName:        "abc",
					VolumeID:          "abc",
					Capacity:          "5G",
					Status:            "init",
				},
				Status: apis.CStorVolumePhase{},
			},
		},
		"img2VolumeResource": {
			expectedVolumeName: "abcd",
			test: &apis.CStorVolume{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name: "volume2",
					UID:  types.UID("abcd"),
				},
				Spec: apis.CStorVolumeSpec{
					CStorControllerIP: "0.0.0.0",
					VolumeName:        "abcd",
					VolumeID:          "abcd",
					Capacity:          "15G",
					Status:            "init",
				},
				Status: apis.CStorVolumePhase{},
			},
		},
	}
	for desc, ut := range testVolumeResource {
		// Create Volume resource
		_, err := volumeController.clientset.OpenebsV1alpha1().CStorVolumes().Create(ut.test)
		if err != nil {
			t.Fatalf("Desc:%v, Unable to create resource : %v", desc, ut.test.ObjectMeta.Name)
		}
		// Get the created volume resource using name
		cStorVolumeObtained, err := volumeController.getVolumeResource(ut.test.ObjectMeta.Name)
		if string(cStorVolumeObtained.ObjectMeta.UID) != ut.expectedVolumeName {
			t.Fatalf("Desc:%v, VolumeName mismatch, Expected:%v, Got:%v", desc, ut.expectedVolumeName,
				string(cStorVolumeObtained.ObjectMeta.UID))
		}
	}
}

// TestIsValidCStorVolumeMgmt is to check if right sidecar does operation with env match.
func TestIsValidCStorVolumeMgmt(t *testing.T) {
	testVolumeResource := map[string]struct {
		expectedOutput bool
		test           *apis.CStorVolume
	}{
		"img2VolumeResource": {
			expectedOutput: true,
			test: &apis.CStorVolume{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:       "volume2",
					UID:        types.UID("abcd"),
					Finalizers: []string{"cstorvolume.openebs.io/finalizer"},
				},
				Spec: apis.CStorVolumeSpec{
					CStorControllerIP: "0.0.0.0",
					VolumeName:        "abcd",
					VolumeID:          "abcd",
					Capacity:          "15G",
					Status:            "init",
				},
				Status: apis.CStorVolumePhase{},
			},
		},
	}
	for desc, ut := range testVolumeResource {
		os.Setenv("OPENEBS_IO_CSTOR_VOLUME_ID", string(ut.test.UID))
		obtainedOutput := IsValidCStorVolumeMgmt(ut.test)
		if obtainedOutput != ut.expectedOutput {
			t.Fatalf("Desc:%v, Expected:%v, Got:%v", desc, ut.expectedOutput,
				obtainedOutput)
		}
		os.Setenv("OPENEBS_IO_CSTOR_VOLUME_ID", "")
	}
}

// TestIsValidCStorVolumeMgmtNegative is to check if right sidecar does operation with env match.
func TestIsValidCStorVolumeMgmtNegative(t *testing.T) {
	testVolumeResource := map[string]struct {
		expectedOutput bool
		test           *apis.CStorVolume
	}{
		"img2VolumeResource": {
			expectedOutput: false,
			test: &apis.CStorVolume{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:       "volume2",
					UID:        types.UID("abcd"),
					Finalizers: []string{"cstorvolume.openebs.io/finalizer"},
				},
				Spec: apis.CStorVolumeSpec{
					CStorControllerIP: "0.0.0.0",
					VolumeName:        "abcd",
					VolumeID:          "abcd",
					Capacity:          "15G",
					Status:            "init",
				},
				Status: apis.CStorVolumePhase{},
			},
		},
	}
	for desc, ut := range testVolumeResource {
		os.Setenv("OPENEBS_IO_CSTOR_VOLUME_ID", string("awer"))
		obtainedOutput := IsValidCStorVolumeMgmt(ut.test)
		if obtainedOutput != ut.expectedOutput {
			t.Fatalf("Desc:%v, Expected:%v, Got:%v", desc, ut.expectedOutput,
				obtainedOutput)
		}
		os.Setenv("OPENEBS_IO_CSTOR_VOLUME_ID", "")
	}
}
