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
package volumereplica

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"testing"

	"github.com/golang/glog"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TestRunner struct{}

// RunCombinedOutput is to mock Real runner exec.
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
	case "destroy":
		cs = []string{"-test.run=TestDestroyerProcess", "--"}
		cmd.Env = []string{"destroyErr=nil"}
		break
	}
	stdout, err := cmd.CombinedOutput()
	return stdout, err
}

// RunStdoutPipe is to mock real runner exec with stdoutpipe.
func (r TestRunner) RunStdoutPipe(command string, args ...string) ([]byte, error) {
	var cs []string
	var cmd *exec.Cmd
	cs = append(cs, args...)
	cmd = exec.Command(os.Args[0], cs...)
	switch args[0] {
	case "get":
		cs = []string{"-test.run=TestGetterProcess", "--"}
		cmd.Env = []string{"volName=cstor-123abc/cba"}
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

// TestCreaterProcess mocks zpool create.
func TestCreaterProcess(*testing.T) {
	if os.Getenv("createErr") != "nil" {
		return
	}
	fmt.Println(nil)
	defer os.Exit(0)

}

// TestGetterProcess mocks zpool get.
func TestGetterProcess(*testing.T) {
	if os.Getenv("volName") != "cstor-123abc/cba" {
		return
	}
	defer os.Exit(0)
	fmt.Println("cstor-123abc/cba")
}

// TestDestroyerProcess mocks zpool destroy.
func TestDestroyerProcess(*testing.T) {
	if os.Getenv("destroyErr") != "nil" {
		return
	}
	defer os.Exit(0)
	fmt.Println(nil)
}

// TestCreateVolume is to test cStorVolumeReplica creation.
func TestCreateVolume(t *testing.T) {
	testPoolResource := map[string]struct {
		expectedError error
		test          *apis.CStorVolumeReplica
	}{
		"Valid-vol1Resource": {
			expectedError: nil,
			test: &apis.CStorVolumeReplica{
				ObjectMeta: metav1.ObjectMeta{
					Name: "VolumeReplicaResource0",
					UID:  "abcd123",
					Labels: map[string]string{
						"cstorpool.openebs.io/name": "cstor-ab12",
					},
				},
				Spec: apis.CStorVolumeReplicaSpec{
					TargetIPSpec: apis.TargetIPSpec{
						"10.210.110.121",
					},
					CapacitySpec: apis.CapacitySpec{
						"10MB",
					},
				},
			},
		},
	}
	RunnerVar = TestRunner{}
	obtainedErr := CreateVolume(testPoolResource["Valid-vol1Resource"].test, "abcd123/dcba")
	if testPoolResource["Valid-vol1Resource"].expectedError != obtainedErr {
		t.Fatalf("Expected: %v, Got: %v", testPoolResource["Valid-vol1Resource"].expectedError, obtainedErr)
	}
}

// TestDeleteVolume is to test cStorVolumeReplica deletion.
func TestDeleteVolume(t *testing.T) {
	testPoolResource := map[string]struct {
		expectedError error
		test          *apis.CStorVolumeReplica
	}{
		"Valid-vol1Resource": {
			expectedError: nil,
			test: &apis.CStorVolumeReplica{
				ObjectMeta: metav1.ObjectMeta{
					Name: "VolumeReplicaResource0",
					UID:  "abcd123",
					Labels: map[string]string{
						"cstorpool.openebs.io/name": "cstor-ab12",
					},
				},
				Spec: apis.CStorVolumeReplicaSpec{
					TargetIPSpec: apis.TargetIPSpec{
						"10.210.110.121",
					},
					CapacitySpec: apis.CapacitySpec{
						"100MB",
					},
				},
			},
		},
	}
	RunnerVar = TestRunner{}
	obtainedErr := DeleteVolume(string(testPoolResource["Valid-vol1Resource"].test.UID))
	if testPoolResource["Valid-vol1Resource"].expectedError != obtainedErr {
		t.Fatalf("Expected: %v, Got: %v", testPoolResource["Valid-vol1Resource"].expectedError, obtainedErr)
	}
}

// TestGetVolume tests get zfs volumes
func TestGetVolume(t *testing.T) {
	testVolResource := map[string]struct {
		expectedVolName []string
		expectedError   error
	}{
		"Vol1Resource": {
			expectedVolName: []string{"cstor-123abc/cba", ""},
			expectedError:   nil,
		},
	}
	RunnerVar = TestRunner{}
	obtainedVolName, obtainedErr := GetVolumes()
	if !reflect.DeepEqual(testVolResource["Vol1Resource"].expectedVolName, obtainedVolName) {
		t.Fatalf("Expected: %v, Got: %v", testVolResource["Vol1Resource"].expectedVolName, obtainedVolName)
	}
	if testVolResource["Vol1Resource"].expectedError != obtainedErr {
		t.Fatalf("Expected: %v, Got: %v", testVolResource["Vol1Resource"].expectedError, obtainedErr)
	}
}

// TestCheckValidVolumeReplica tests VolumeReplica related operations
func TestCheckValidVolumeReplica(t *testing.T) {
	testVolumeReplicaResource := map[string]struct {
		expectedError error
		test          *apis.CStorVolumeReplica
	}{
		"Invalid-VolumeNameEmpty": {
			expectedError: fmt.Errorf("Volume Name/UID cannot be empty"),
			test: &apis.CStorVolumeReplica{
				ObjectMeta: metav1.ObjectMeta{
					Name: "VolumeReplicaResource1",
				},
				Spec: apis.CStorVolumeReplicaSpec{
					TargetIPSpec: apis.TargetIPSpec{
						"",
					},
					CapacitySpec: apis.CapacitySpec{
						"10MB",
					},
				},
			},
		},
		"Invalid-controllerIpEmpty": {
			expectedError: fmt.Errorf("TargetIP cannot be empty"),
			test: &apis.CStorVolumeReplica{
				ObjectMeta: metav1.ObjectMeta{
					Name: "VolumeReplicaResource1",
					Labels: map[string]string{
						"cstorvolume.openebs.io/name": "cstor-ab12",
					},
				},
				Spec: apis.CStorVolumeReplicaSpec{
					TargetIPSpec: apis.TargetIPSpec{
						"",
					},
					CapacitySpec: apis.CapacitySpec{
						"100MB",
					},
				},
			},
		},
		"Invalid-CapacityEmpty": {
			expectedError: fmt.Errorf("Capacity cannot be empty"),
			test: &apis.CStorVolumeReplica{
				ObjectMeta: metav1.ObjectMeta{
					Name: "VolumeReplicaResource2",
					Labels: map[string]string{
						"cstorvolume.openebs.io/name": "cstor-ab12",
					},
				},
				Spec: apis.CStorVolumeReplicaSpec{
					TargetIPSpec: apis.TargetIPSpec{
						"10.210.110.121",
					},
					CapacitySpec: apis.CapacitySpec{
						"",
					},
				},
			},
		},
		"Invalid-poolNameEmpty": {
			expectedError: fmt.Errorf("Pool cannot be empty"),
			test: &apis.CStorVolumeReplica{
				ObjectMeta: metav1.ObjectMeta{
					Name: "VolumeReplicaResource2",
					Labels: map[string]string{
						"cstorvolume.openebs.io/name": "cstor-ab12",
					},
				},
				Spec: apis.CStorVolumeReplicaSpec{
					TargetIPSpec: apis.TargetIPSpec{
						"10.210.110.121",
					},
					CapacitySpec: apis.CapacitySpec{
						"100MB",
					},
				},
			},
		},
	}

	for desc, ut := range testVolumeReplicaResource {
		Obtainederr := CheckValidVolumeReplica(ut.test)
		if Obtainederr != nil {
			if Obtainederr.Error() != ut.expectedError.Error() {
				t.Fatalf("Desc : %v, Expected error: %v, Got : %v",
					desc, ut.expectedError, Obtainederr)
			}

		}
	}
}
