package volumecontroller

import (
	"testing"
	"time"

	"github.com/openebs/maya/cmd/cstor-volume-mgmt/controller/common"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	openebsFakeClientset "github.com/openebs/maya/pkg/client/clientset/versioned/fake"
	informers "github.com/openebs/maya/pkg/client/informers/externalversions"
	"github.com/openebs/maya/pkg/signals"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
)

// TestRun is to run cStorVolume controller and check if it crashes or return back.
func TestRun(t *testing.T) {
	fakeKubeClient := fake.NewSimpleClientset()
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(fakeKubeClient, time.Second*30)
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)

	// Instantiate the cStor Volume controllers.
	volumeController := NewCStorVolumeController(fakeKubeClient, fakeOpenebsClient, kubeInformerFactory,
		openebsInformerFactory)

	stopCh := signals.SetupSignalHandler()
	done := make(chan bool)
	go func(chan bool) {
		volumeController.Run(2, stopCh)
		done <- true
	}(done)

	select {
	case <-time.After(3 * time.Second):

	case <-done:
		t.Fatalf("CStorVolume controller returned - failure")

	}
}

// TestProcessNextWorkItemAdd is to test a cStorVolume resource for add event.
func TestProcessNextWorkItemAdd(t *testing.T) {
	fakeKubeClient := fake.NewSimpleClientset()
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(fakeKubeClient, time.Second*30)
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)

	// Instantiate the cStor Volume controllers.
	volumeController := NewCStorVolumeController(fakeKubeClient, fakeOpenebsClient, kubeInformerFactory,
		openebsInformerFactory)

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
					VolumeName:        "abc",
					VolumeID:          "abc",
					Capacity:          "5G",
					Status:            "init",
				},
				Status: apis.CStorVolumePhase{Phase: "init"},
			},
		},
	}
	_, err := volumeController.clientset.OpenebsV1alpha1().CStorVolumes().Create(testVolumeResource["img2VolumeResource"].test)
	if err != nil {
		t.Fatalf("Unable to create resource : %v", testVolumeResource["img2VolumeResource"].test.ObjectMeta.Name)
	}

	var q common.QueueLoad
	q.Key = "volume2"
	q.Operation = common.QOpAdd
	volumeController.workqueue.AddRateLimited(q)

	obtainedOutput := volumeController.processNextWorkItem()
	if obtainedOutput != testVolumeResource["img2VolumeResource"].expectedOutput {
		t.Fatalf("Expected:%v, Got:%v", testVolumeResource["img2VolumeResource"].expectedOutput,
			obtainedOutput)
	}
}

// TestProcessNextWorkItemModify is to test a cStorVolume resource for modify event.
func TestProcessNextWorkItemModify(t *testing.T) {
	fakeKubeClient := fake.NewSimpleClientset()
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(fakeKubeClient, time.Second*30)
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)

	// Instantiate the cStor Volume controllers.
	volumeController := NewCStorVolumeController(fakeKubeClient, fakeOpenebsClient, kubeInformerFactory,
		openebsInformerFactory)

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
					VolumeName:        "abc",
					VolumeID:          "abc",
					Capacity:          "5G",
					Status:            "init",
				},
				Status: apis.CStorVolumePhase{},
			},
		},
	}

	_, err := volumeController.clientset.OpenebsV1alpha1().CStorVolumes().Create(testVolumeResource["img2VolumeResource"].test)
	if err != nil {
		t.Fatalf("Unable to create resource : %v", testVolumeResource["img2VolumeResource"].test.ObjectMeta.Name)
	}

	var q common.QueueLoad
	q.Key = "volume2"
	q.Operation = common.QOpModify
	volumeController.workqueue.AddRateLimited(q)

	obtainedOutput := volumeController.processNextWorkItem()
	if obtainedOutput != testVolumeResource["img2VolumeResource"].expectedOutput {
		t.Fatalf("Expected:%v, Got:%v", testVolumeResource["img2VolumeResource"].expectedOutput,
			obtainedOutput)
	}
}

// TestProcessNextWorkItemDestroy is to test a cStorVolume resource for destroy event.
func TestProcessNextWorkItemDestroy(t *testing.T) {
	fakeKubeClient := fake.NewSimpleClientset()
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(fakeKubeClient, time.Second*30)
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)

	// Instantiate the cStor Volume controllers.
	volumeController := NewCStorVolumeController(fakeKubeClient, fakeOpenebsClient, kubeInformerFactory,
		openebsInformerFactory)

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
					VolumeName:        "abc",
					VolumeID:          "abc",
					Capacity:          "5G",
					Status:            "init",
				},
				Status: apis.CStorVolumePhase{},
			},
		},
	}

	_, err := volumeController.clientset.OpenebsV1alpha1().CStorVolumes().Create(testVolumeResource["img2VolumeResource"].test)
	if err != nil {
		t.Fatalf("Unable to create resource : %v", testVolumeResource["img2VolumeResource"].test.ObjectMeta.Name)
	}

	var q common.QueueLoad
	q.Key = "volume2"
	q.Operation = common.QOpDestroy
	volumeController.workqueue.AddRateLimited(q)

	obtainedOutput := volumeController.processNextWorkItem()
	if obtainedOutput != testVolumeResource["img2VolumeResource"].expectedOutput {
		t.Fatalf("Expected:%v, Got:%v", testVolumeResource["img2VolumeResource"].expectedOutput,
			obtainedOutput)
	}
}
