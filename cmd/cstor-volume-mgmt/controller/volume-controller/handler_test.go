package volumecontroller

import (
	"os"
	"testing"
	"time"

	"k8s.io/api/core/v1"

	"github.com/openebs/maya/cmd/cstor-volume-mgmt/controller/common"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"

	//openebsFakeClientset "github.com/openebs/maya/pkg/client/clientset/versioned/fake"
	openebsFakeClientset "github.com/openebs/maya/pkg/client/generated/clientset/internalclientset/fake"

	//informers "github.com/openebs/maya/pkg/client/informers/externalversions"
	informers "github.com/openebs/maya/pkg/client/generated/informer/externalversions"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
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
					Name:      "volume1",
					UID:       types.UID("abc"),
					Namespace: string(common.DefaultNameSpace),
				},
				Spec: apis.CStorVolumeSpec{
					TargetIP: "0.0.0.0",
					Capacity: "5G",
					Status:   "init",
				},
				Status: apis.CStorVolumeStatus{},
			},
		},
		"img2VolumeResource": {
			expectedVolumeName: "abcd",
			test: &apis.CStorVolume{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "volume2",
					UID:       types.UID("abcd"),
					Namespace: string(common.DefaultNameSpace),
				},
				Spec: apis.CStorVolumeSpec{
					TargetIP: "0.0.0.0",
					Capacity: "15G",
					Status:   "init",
				},
				Status: apis.CStorVolumeStatus{},
			},
		},
	}
	for desc, ut := range testVolumeResource {
		// Create Volume resource
		_, err := volumeController.clientset.OpenebsV1alpha1().CStorVolumes(string(common.DefaultNameSpace)).Create(ut.test)
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
					Namespace:  string(common.DefaultNameSpace),
				},
				Spec: apis.CStorVolumeSpec{
					TargetIP: "0.0.0.0",
					Capacity: "15G",
					Status:   "init",
				},
				Status: apis.CStorVolumeStatus{},
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
					Namespace:  string(common.DefaultNameSpace),
				},
				Spec: apis.CStorVolumeSpec{
					TargetIP: "0.0.0.0",
					Capacity: "15G",
					Status:   "init",
				},
				Status: apis.CStorVolumeStatus{},
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

func TestCreateSyncUpdateEvent(t *testing.T) {
	tests := map[string]struct {
		client      kubernetes.Interface
		cstorVolume *apis.CStorVolume
		event       *v1.Event
		name, msg   string
		wantErr     bool
	}{
		"event does not exist": {
			client: fake.NewSimpleClientset(),
			cstorVolume: &apis.CStorVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "csv-1",
					UID:       types.UID("abcd"),
					Namespace: string(common.DefaultNameSpace),
				},
				TypeMeta: metav1.TypeMeta{
					Kind:       "CstorVolume",
					APIVersion: "v1alpha1",
				},
			},
			name: "csv-1.Init",
			msg:  "Volume is in Init state",
		},
		"event already created": {
			cstorVolume: &apis.CStorVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "csv-1",
					UID:       types.UID("abcd"),
					Namespace: string(common.DefaultNameSpace),
				},
				TypeMeta: metav1.TypeMeta{
					Kind:       "CstorVolume",
					APIVersion: "v1alpha1",
				},
			},
			client: fake.NewSimpleClientset(),
			event: &v1.Event{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "csv-1.Init",
					Namespace: string(common.DefaultNameSpace),
				},
			},
			name: "csv-1.Init",
			msg:  "Volume is in Init state",
		},
	}

	for desc, ut := range tests {
		t.Run(desc, func(t *testing.T) {
			if ut.event != nil {
				ut.client.CoreV1().Events(ut.cstorVolume.Namespace).Create(ut.event)
			}
			cvController := CStorVolumeController{kubeclientset: ut.client}
			err := cvController.createSyncUpdateEvent(ut.cstorVolume, ut.name, ut.msg)
			if (err != nil) != ut.wantErr {
				t.Errorf("Error creating event, wantErr=%v gotErr=%v", ut.wantErr, err != nil)
			}
		})
	}
}
