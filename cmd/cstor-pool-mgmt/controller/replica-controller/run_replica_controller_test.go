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
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/openebs/maya/cmd/cstor-pool-mgmt/controller/common"
	"github.com/openebs/maya/cmd/cstor-pool-mgmt/pool"
	"github.com/openebs/maya/cmd/cstor-pool-mgmt/volumereplica"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"k8s.io/klog"

	openebsFakeClientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned/fake"
	informers "github.com/openebs/maya/pkg/client/generated/informers/externalversions"

	"github.com/openebs/maya/pkg/signals"
	"github.com/openebs/maya/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
)

// TestRun is to run cStorVolumeReplica controller and check if it crashes or return back.
func TestRun(t *testing.T) {
	fakeKubeClient := fake.NewSimpleClientset()
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(fakeKubeClient, time.Second*30)
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)

	// Instantiate the cStor Volume Replica controller.
	volumeReplicaController := NewCStorVolumeReplicaController(fakeKubeClient, fakeOpenebsClient, kubeInformerFactory,
		openebsInformerFactory)

	stopCh := signals.SetupSignalHandler()
	done := make(chan bool)
	go func(chan bool) {
		volumeReplicaController.Run(1, stopCh)
		done <- true
	}(done)

	select {
	case <-time.After(3 * time.Second):

	case <-done:
		t.Fatalf("cVR controller returned - failure")

	}
}

// TestProcessNextWorkItemAdd is to test a cStorPool resource for add event.
func TestProcessNextWorkItemAdd(t *testing.T) {
	fakeKubeClient := fake.NewSimpleClientset()
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(fakeKubeClient, time.Second*30)
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)

	// Instantiate the cStor Pool controllers.
	volumeReplicaController := NewCStorVolumeReplicaController(fakeKubeClient, fakeOpenebsClient, kubeInformerFactory,
		openebsInformerFactory)

	testPoolResource := map[string]struct {
		expectedOutput bool
		test           *apis.CStorVolumeReplica
	}{
		"img2PoolResource": {
			expectedOutput: true,
			test: &apis.CStorVolumeReplica{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name: "cvr1",
					UID:  types.UID("abcd"),
					Labels: map[string]string{
						"cstorpool.openebs.io/uid": "123abc",
					},
					Finalizers: []string{"cstorpool.openebs.io/finalizer"},
				},
				Spec: apis.CStorVolumeReplicaSpec{
					TargetIP: "10.210.102.206",
					Capacity: "10MB",
				},
				Status: apis.CStorVolumeReplicaStatus{Phase: "init"},
			},
		},
	}
	_, err := volumeReplicaController.clientset.OpenebsV1alpha1().CStorVolumeReplicas("default").Create(testPoolResource["img2PoolResource"].test)
	if err != nil {
		t.Fatalf("Unable to create resource : %v", testPoolResource["img2PoolResource"].test.ObjectMeta.Name)
	}
	pool.RunnerVar = TestRunner{}
	volumereplica.RunnerVar = TestRunner{}
	RunnerVar = TestRunner{}
	var q common.QueueLoad
	q.Key = "cvr1"
	q.Operation = "add"
	volumeReplicaController.workqueue.AddRateLimited(q)

	obtainedOutput := volumeReplicaController.processNextWorkItem()
	if obtainedOutput != testPoolResource["img2PoolResource"].expectedOutput {
		t.Fatalf("Expected:%v, Got:%v", testPoolResource["img2PoolResource"].expectedOutput,
			obtainedOutput)
	}
}

// TestProcessNextWorkItemAdd is to test a cStorPool resource for add event.
func TestProcessNextWorkItemModify(t *testing.T) {
	fakeKubeClient := fake.NewSimpleClientset()
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(fakeKubeClient, time.Second*30)
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)

	// Instantiate the cStor Pool controllers.
	volumeReplicaController := NewCStorVolumeReplicaController(fakeKubeClient, fakeOpenebsClient, kubeInformerFactory,
		openebsInformerFactory)

	testPoolResource := map[string]struct {
		expectedOutput bool
		test           *apis.CStorVolumeReplica
	}{
		"img2PoolResource": {
			expectedOutput: true,
			test: &apis.CStorVolumeReplica{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name: "cvr1",
					UID:  types.UID("abcd"),
					Labels: map[string]string{
						"cstorpool.openebs.io/uid": "123abc",
					},
					Finalizers: []string{"cstorpool.openebs.io/finalizer"},
				},
				Spec: apis.CStorVolumeReplicaSpec{
					TargetIP: "10.210.102.206",
					Capacity: "10MB",
				},
				Status: apis.CStorVolumeReplicaStatus{Phase: "init"},
			},
		},
	}
	_, err := volumeReplicaController.clientset.OpenebsV1alpha1().CStorVolumeReplicas("default").Create(testPoolResource["img2PoolResource"].test)
	if err != nil {
		t.Fatalf("Unable to create resource : %v", testPoolResource["img2PoolResource"].test.ObjectMeta.Name)
	}
	pool.RunnerVar = TestRunner{}
	volumereplica.RunnerVar = TestRunner{}
	RunnerVar = TestRunner{}
	var q common.QueueLoad
	q.Key = "cvr1"
	q.Operation = "modify"
	volumeReplicaController.workqueue.AddRateLimited(q)

	obtainedOutput := volumeReplicaController.processNextWorkItem()
	if obtainedOutput != testPoolResource["img2PoolResource"].expectedOutput {
		t.Fatalf("Expected:%v, Got:%v", testPoolResource["img2PoolResource"].expectedOutput,
			obtainedOutput)
	}
}

var RunnerVar util.Runner

type TestRunner struct{}

// RunCommandWithTimeoutContext is to mock real runner exec with stdoutpipe.
func (r TestRunner) RunCommandWithTimeoutContext(timeout time.Duration, command string, args ...string) ([]byte, error) {
	return []byte("success"), nil
}

// RunCombinedOutput is to mock binaries with fake test binaries.
func (r TestRunner) RunCombinedOutput(command string, args ...string) ([]byte, error) {
	var cs []string
	var cmd *exec.Cmd
	switch args[0] {
	case "create":
		cs = append([]string{"-test.run=TestCreaterProcess", "--"})
		//	cmd.Env = append([]string{"createErr=nil"})
		break
	case "import":
		cs = []string{"-test.run=TestImporterProcess", "--"}
		cmd.Env = []string{"importErr=nil"}
		break
	case "destroy":
		cs = []string{"-test.run=TestDestroyerProcess", "--"}
		cmd.Env = []string{"destroyErr=nil"}
		break
	case "labelclear":
		cs = []string{"-test.run=TestLabelClearerProcess", "--"}
		cmd.Env = []string{"labelClearErr=nil"}
		break
	case "status":
		cs = []string{"-test.run=TestStatusProcess", "--"}
		cmd.Env = []string{"StatusErr=nil"}
		break
	}
	cs = append(cs, args...)
	cmd = exec.Command(os.Args[0], cs...)
	stdout, err := cmd.CombinedOutput()
	return stdout, err
}

// RunStdoutPipe is to mock real runner exec with stdoutpipe.
func (r TestRunner) RunStdoutPipe(command string, args ...string) ([]byte, error) {
	var cs []string
	var cmd *exec.Cmd
	switch args[0] {
	case "get":
		cs = append([]string{"-test.run=TestGetterProcess", "--"})
		break
	}
	cs = append(cs, args...)
	cmd = exec.Command(os.Args[0], cs...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		klog.Errorf(err.Error())
		return []byte{}, err
	}
	if err := cmd.Start(); err != nil {
		klog.Errorf(err.Error())
		return []byte{}, err
	}
	data, _ := ioutil.ReadAll(stdout)
	if err := cmd.Wait(); err != nil {
		klog.Errorf(err.Error())
		return []byte{}, err
	}
	return data, nil
}

// RunCommandWithLog is to mock real runner exec with stdoutpipe
func (r TestRunner) RunCommandWithLog(common string, args ...string) ([]byte, error) {
	return []byte("success"), nil
}

// TestGetterProcess mocks zpool get.
func TestGetterProcess(*testing.T) {
	defer os.Exit(0)
	fmt.Println("cstor-123abc")
}

// TestCreaterProcess mocks zpool create.
func TestCreaterProcess(*testing.T) {
	fmt.Println(nil)
	defer os.Exit(0)

}

// TestDestroyerProcess mocks zpool destroy.
func TestDestroyerProcess(*testing.T) {
	if os.Getenv("destroyErr") != "nil" {
		return
	}
	defer os.Exit(0)
	fmt.Println(nil)
}
