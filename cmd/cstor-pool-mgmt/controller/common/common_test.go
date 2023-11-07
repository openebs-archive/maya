// Copyright © 2017-2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package common

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/openebs/maya/cmd/cstor-pool-mgmt/pool"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"k8s.io/klog/v2"

	//openebsFakeClientset "github.com/openebs/maya/pkg/client/clientset/versioned/fake"
	openebsFakeClientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		break
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

// RunCommandWithTimeoutContext is to mock Real runner exec.
func (r TestRunner) RunCommandWithTimeoutContext(timeout time.Duration, command string, args ...string) ([]byte, error) {
	return []byte("success"), nil
}

// RunCommandWithLog is to mock Real runner exec.
func (r TestRunner) RunCommandWithLog(command string, args ...string) ([]byte, error) {
	return []byte("success"), nil
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
	testResource := map[string]struct {
		expectedFlag bool
		test         *apis.CStorVolumeReplica
	}{
		"cVR": {
			expectedFlag: true,
			test: &apis.CStorVolumeReplica{
				ObjectMeta: metav1.ObjectMeta{
					Name: "VolumeReplicaResource1",
					UID:  "abcd123",
					Labels: map[string]string{
						"cstorpool.openebs.io/uid":  "123abc",
						"cstorpool.openebs.io/name": "cstor-123abc",
					},
				},
				Spec: apis.CStorVolumeReplicaSpec{
					TargetIP: "127.0.0.1",
					Capacity: "100MB",
				},
			},
		},
	}
	pool.RunnerVar = TestRunner{}
	obtainedFlag := PoolNameHandler(testResource["cVR"].test, 1)
	if testResource["cVR"].expectedFlag != obtainedFlag {
		t.Fatalf("Expected: %v, Got: %v", testResource["cVR"].expectedFlag, obtainedFlag)
	}
}

// TestCheckForInitialImportedPoolVol tests if pool/vols are already imported.
func TestCheckForInitialImportedPoolVol(t *testing.T) {
	testResource := map[string]struct {
		expectedPresentFlag        bool
		expectedDeletedEntryFlag   bool
		InitialImportedPoolVolName []string
		fullvolname                string
	}{
		"pool1/vol1-Resource": {
			expectedPresentFlag:        true,
			expectedDeletedEntryFlag:   false,
			InitialImportedPoolVolName: []string{"pool1/vol1", "pool1/vol2"},
			fullvolname:                "pool1/vol1",
		},
	}
	for desc, ut := range testResource {
		obtainedPresentFlag := CheckForInitialImportedPoolVol(ut.InitialImportedPoolVolName, ut.fullvolname)
		obtainedDeletedEntryFlag := CheckForInitialImportedPoolVol(ut.InitialImportedPoolVolName, ut.fullvolname)

		if obtainedPresentFlag != ut.expectedPresentFlag {
			t.Fatalf("Desc:%v, Test case failure, Expected:%v, Got:%v", desc, ut.expectedPresentFlag,
				obtainedPresentFlag)
		}
		if obtainedDeletedEntryFlag != ut.expectedDeletedEntryFlag {
			t.Fatalf("Desc:%v, Test case failure, Expected:%v, Got:%v", desc, ut.expectedDeletedEntryFlag,
				obtainedDeletedEntryFlag)
		}
	}
}

// CheckIfPresent tests if pool/vols are already imported.
func TestCheckIfPresent(t *testing.T) {
	testResource := map[string]struct {
		expectedFlag bool
		arrStr       []string
		searchStr    string
	}{
		"pool1/vol1-Resource": {
			expectedFlag: true,
			arrStr:       []string{"pool1/vol1", "pool1/vol2"},
			searchStr:    "pool1/vol1",
		},
	}
	for desc, ut := range testResource {
		obtainedFlag := CheckIfPresent(ut.arrStr, ut.searchStr)

		if obtainedFlag != ut.expectedFlag {
			t.Fatalf("Desc:%v, Test case failure, Expected:%v, Got:%v", desc, ut.expectedFlag,
				obtainedFlag)
		}
	}
}

// TestCheckForCStorPool validates if CStorVolumeReplica CRD
// operations can be done.
func TestCheckForCStorPool(t *testing.T) {
	done := make(chan bool)
	pool.RunnerVar = TestRunner{}
	go func(done chan bool) {
		CheckForCStorPool()
		done <- true
	}(done)

	select {
	case <-time.After(10 * time.Second):
		t.Fatalf("CStorPool not found")
	case <-done:
	}
}
