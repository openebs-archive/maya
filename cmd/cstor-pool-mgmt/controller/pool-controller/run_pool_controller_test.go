package poolcontroller

import (
	"testing"
	"time"

	"github.com/openebs/maya/cmd/cstor-pool-mgmt/controller/common"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	openebsFakeClientset "github.com/openebs/maya/pkg/client/clientset/versioned/fake"
	informers "github.com/openebs/maya/pkg/client/informers/externalversions"
	"github.com/openebs/maya/pkg/signals"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
)

// TestRun is to run cStorPool controller and check if it crashes or return back.
func TestRun(t *testing.T) {
	fakeKubeClient := fake.NewSimpleClientset()
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(fakeKubeClient, time.Second*30)
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)

	// Instantiate the cStor Pool controllers.
	poolController := NewCStorPoolController(fakeKubeClient, fakeOpenebsClient, kubeInformerFactory,
		openebsInformerFactory)

	stopCh := signals.SetupSignalHandler()
	done := make(chan bool)
	go func(chan bool) {
		poolController.Run(2, stopCh)
		done <- true
	}(done)

	select {
	case <-time.After(3 * time.Second):

	case <-done:
		t.Fatalf("CStorPool controller returned - failure")

	}
}

// TestProcessNextWorkItemAdd is to test a cStorPool resource for add event.
func TestProcessNextWorkItemAdd(t *testing.T) {
	fakeKubeClient := fake.NewSimpleClientset()
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(fakeKubeClient, time.Second*30)
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)

	// Instantiate the cStor Pool controllers.
	poolController := NewCStorPoolController(fakeKubeClient, fakeOpenebsClient, kubeInformerFactory,
		openebsInformerFactory)

	testPoolResource := map[string]struct {
		expectedOutput bool
		test           *apis.CStorPool
	}{
		"img2PoolResource": {
			expectedOutput: true,
			test: &apis.CStorPool{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:       "pool2",
					UID:        types.UID("abcd"),
					Finalizers: []string{"cstorpool.openebs.io/finalizer"},
				},
				Spec: apis.CStorPoolSpec{
					Disks: apis.DiskAttr{
						DiskList: []string{"/tmp/img2.img"},
					},
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool2.cache",
						PoolType:         "striped",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{Phase: "init"},
			},
		},
	}
	_, err := poolController.clientset.OpenebsV1alpha1().CStorPools().Create(testPoolResource["img2PoolResource"].test)
	if err != nil {
		t.Fatalf("Unable to create resource : %v", testPoolResource["img2PoolResource"].test.ObjectMeta.Name)
	}

	var q common.QueueLoad
	q.Key = "pool2"
	q.Operation = "add"
	poolController.workqueue.AddRateLimited(q)

	obtainedOutput := poolController.processNextWorkItem()
	if obtainedOutput != testPoolResource["img2PoolResource"].expectedOutput {
		t.Fatalf("Expected:%v, Got:%v", testPoolResource["img2PoolResource"].expectedOutput,
			obtainedOutput)
	}
}

// TestProcessNextWorkItemModify is to test a cStorPool resource for modify event.
func TestProcessNextWorkItemModify(t *testing.T) {
	fakeKubeClient := fake.NewSimpleClientset()
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(fakeKubeClient, time.Second*30)
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)

	// Instantiate the cStor Pool controllers.
	poolController := NewCStorPoolController(fakeKubeClient, fakeOpenebsClient, kubeInformerFactory,
		openebsInformerFactory)

	testPoolResource := map[string]struct {
		expectedOutput bool
		test           *apis.CStorPool
	}{
		"img2PoolResource": {
			expectedOutput: true,
			test: &apis.CStorPool{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:       "pool2",
					UID:        types.UID("abcd"),
					Finalizers: []string{"cstorpool.openebs.io/finalizer"},
				},
				Spec: apis.CStorPoolSpec{
					Disks: apis.DiskAttr{
						DiskList: []string{"/tmp/img2.img"},
					},
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool2.cache",
						PoolType:         "striped",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
		},
	}

	_, err := poolController.clientset.OpenebsV1alpha1().CStorPools().Create(testPoolResource["img2PoolResource"].test)
	if err != nil {
		t.Fatalf("Unable to create resource : %v", testPoolResource["img2PoolResource"].test.ObjectMeta.Name)
	}

	var q common.QueueLoad
	q.Key = "pool2"
	q.Operation = "modify"
	poolController.workqueue.AddRateLimited(q)

	obtainedOutput := poolController.processNextWorkItem()
	if obtainedOutput != testPoolResource["img2PoolResource"].expectedOutput {
		t.Fatalf("Expected:%v, Got:%v", testPoolResource["img2PoolResource"].expectedOutput,
			obtainedOutput)
	}
}

// TestProcessNextWorkItemDestroy is to test a cStorPool resource for destroy event.
func TestProcessNextWorkItemDestroy(t *testing.T) {
	fakeKubeClient := fake.NewSimpleClientset()
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(fakeKubeClient, time.Second*30)
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)

	// Instantiate the cStor Pool controllers.
	poolController := NewCStorPoolController(fakeKubeClient, fakeOpenebsClient, kubeInformerFactory,
		openebsInformerFactory)

	testPoolResource := map[string]struct {
		expectedOutput bool
		test           *apis.CStorPool
	}{
		"img2PoolResource": {
			expectedOutput: true,
			test: &apis.CStorPool{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:       "pool2",
					UID:        types.UID("abcd"),
					Finalizers: []string{"cstorpool.openebs.io/finalizer"},
				},
				Spec: apis.CStorPoolSpec{
					Disks: apis.DiskAttr{
						DiskList: []string{"/tmp/img2.img"},
					},
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool2.cache",
						PoolType:         "striped",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
		},
	}

	_, err := poolController.clientset.OpenebsV1alpha1().CStorPools().Create(testPoolResource["img2PoolResource"].test)
	if err != nil {
		t.Fatalf("Unable to create resource : %v", testPoolResource["img2PoolResource"].test.ObjectMeta.Name)
	}

	var q common.QueueLoad
	q.Key = "pool2"
	q.Operation = "destroy"
	poolController.workqueue.AddRateLimited(q)

	obtainedOutput := poolController.processNextWorkItem()
	if obtainedOutput != testPoolResource["img2PoolResource"].expectedOutput {
		t.Fatalf("Expected:%v, Got:%v", testPoolResource["img2PoolResource"].expectedOutput,
			obtainedOutput)
	}
}
