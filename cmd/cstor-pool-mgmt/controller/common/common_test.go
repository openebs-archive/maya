package common

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/golang/glog"
	"github.com/openebs/maya/cmd/cstor-pool-mgmt/pool"
	openebsFakeClientset "github.com/openebs/maya/pkg/client/clientset/versioned/fake"
	//github.com/openebs/maya/vendor/k8s.io/client-go/kubernetes/fake
)

// TestCheckForCStorPoolCRD validates if CStorPool CRD operations
// can be done.
func TestCheckForCStorPoolCRD(t *testing.T) {
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()
	done := make(chan bool)

	go func(done chan bool) {
		CheckForCStorPoolCRD(fakeOpenebsClient)
		done <- true
	}(done)

	select {
	case <-time.After(10 * time.Second):
		t.Fatalf("Timeout - CStorPool is unknown")
	case <-done:
	}
}

// TestCheckForCStorVolumeReplicaCRD validates if CStorVolumeReplica CRD
// operations can be done.
func TestCheckForCStorVolumeReplicaCRD(t *testing.T) {
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()
	done := make(chan bool)

	go func(done chan bool) {
		CheckForCStorVolumeReplicaCRD(fakeOpenebsClient)
		done <- true
	}(done)

	select {
	case <-time.After(10 * time.Second):
		t.Fatalf("Timeout - CStorVolumeReplica is unknown")
	case <-done:
	}
}

type TestRunner struct{}

// RunCombinedOutput is to mock binaries with fake test binaries.
func (r TestRunner) RunCombinedOutput(command string, args ...string) ([]byte, error) {
	var cs []string
	var cmd *exec.Cmd
	cs = append(cs, args...)
	cmd = exec.Command(os.Args[0], cs...)
	switch args[0] {
	case "create":
		cs = []string{"-test.run=TestCreaterProcess", "--"}
		cmd.Env = []string{"createErr=nil"}
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
	stdout, err := cmd.CombinedOutput()
	return stdout, err
}

// RunStdoutPipe is to mock binaries requiring pipes with fake test binaries.
func (r TestRunner) RunStdoutPipe(command string, args ...string) ([]byte, error) {
	var cs []string
	var cmd *exec.Cmd
	cs = append(cs, args...)
	cmd = exec.Command(os.Args[0], cs...)
	switch args[0] {
	case "get":
		cs = []string{"-test.run=TestGetterProcess", "--"}
		cmd.Env = []string{"poolName=cstor-123abc"}
		break
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		glog.Errorf(err.Error())
		return []byte{}, err
	}
	if err := cmd.Start(); err != nil {
		glog.Errorf(err.Error())
		return []byte{}, err
	}
	data, _ := ioutil.ReadAll(stdout)
	if err := cmd.Wait(); err != nil {
		glog.Errorf(err.Error())
		return []byte{}, err
	}
	return data, nil
}

// TestGetterProcess is to fake get process
func TestGetterProcess(*testing.T) {
	if os.Getenv("poolName") != "cstor-123abc" {
		return
	}
	defer os.Exit(0)
	fmt.Println("cstor-123abc")
}

// TestPoolNameHandler is test temporary blocking call for pool availability.
func TestPoolNameHandler(t *testing.T) {
	testPoolResource := map[string]struct {
		expectedPoolName string
		expectedError    error
	}{
		"img1PoolResource": {
			expectedPoolName: "cstor-123abc\n",
			expectedError:    nil,
		},
	}
	pool.RunnerVar = TestRunner{}
	obtainedPoolName, obtainedErr := PoolNameHandler(1)
	//	obtainedPoolName, obtainedErr := GetPoolName()
	fmt.Println(obtainedPoolName, obtainedErr)
	if testPoolResource["img1PoolResource"].expectedPoolName != obtainedPoolName {
		t.Fatalf("Expected: %v, Got: %v", testPoolResource["img1PoolResource"].expectedPoolName, obtainedPoolName)
	}
	if testPoolResource["img1PoolResource"].expectedError != obtainedErr {
		t.Fatalf("Expected: %v, Got: %v", testPoolResource["img1PoolResource"].expectedError, obtainedErr)
	}
}
