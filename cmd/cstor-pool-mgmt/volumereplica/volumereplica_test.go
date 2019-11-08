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
	"time"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

type TestRunner struct{}

const (
	testZfsStatusUnknown  = "Unknown"
	testZfsReconstructing = "Reconstructing"
	testNonQuorumDegraded = "NonQuorumDegraded"
	testStatusType        = "StatusType"
)

// RunCombinedOutput is to mock Real runner exec.
func (r TestRunner) RunCombinedOutput(command string, args ...string) ([]byte, error) {
	var cs []string
	var env []string
	var cmd *exec.Cmd
	switch args[0] {
	case "create":
		cs = []string{"-test.run=TestCreaterProcess", "--"}
		env = []string{"createErr=nil"}
	case "destroy":
		cs = []string{"-test.run=TestDestroyerProcess", "--"}
		env = []string{"destroyErr=nil"}
	case "get":
		// Create command arguments
		cs = []string{"-test.run=TestCapacityHelperProcess", "--", command}
		// Set env varibles for the 'TestCapacityHelperProcess' function which runs as a process.
		env = []string{"GO_WANT_CAPACITY_HELPER_PROCESS=1"}
	case StatsCmd:
		// Create command arguments
		cs = []string{"-test.run=TestStatusHelperProcess", "--", command}
		// Set env varibles for the 'TestStatusHelperProcess' function which runs as a process.
		env = []string{"GO_WANT_STATUS_HELPER_PROCESS=1", "StatusType=" + os.Getenv("StatusType")}
	}
	cs = append(cs, args...)
	cmd = exec.Command(os.Args[0], cs...)
	cmd.Env = env
	stdout, err := cmd.CombinedOutput()
	return stdout, err
}

// RunCommandWithTimeoutContext is to mock real runner exec with stdoutpipe.
func (r TestRunner) RunCommandWithTimeoutContext(timeout time.Duration, command string, args ...string) ([]byte, error) {
	return []byte("success"), nil
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

// TestStatusHelperProcess is a function that is run as a process to get the mocked std output
func TestStatusHelperProcess(*testing.T) {
	// Following constants are different mocked output for `zfs status` command for
	// different statuses.
	const (
		mockedStatusOutputHealthy = `{
  "stats": [
    {
      "name": "cstor-183f17c6-ed8b-11e8-87fd-42010a800087/pvc-9e91f938-ee23-11e8-87fd-42010a800087",
      "status": "Healthy",
      "rebuildStatus": "INIT",
      "isIOAckSenderCreated": 0,
      "isIOReceiverCreated": 0,
      "runningIONum": 0,
      "checkpointedIONum": 0,
      "degradedCheckpointedIONum": 0,
      "quorum": 1,
      "checkpointedTime": 0,
      "rebuildBytes": 0,
      "rebuildCnt": 0,
      "rebuildDoneCnt": 0,
      "rebuildFailedCnt": 0
    }
  ]
}`
		mockedStatusOutputOffline = `{
  "stats": [
    {
      "name": "cstor-183f17c6-ed8b-11e8-87fd-42010a800087/pvc-9e91f938-ee23-11e8-87fd-42010a800087",
      "status": "Offline",
      "rebuildStatus": "INIT",
      "isIOAckSenderCreated": 0,
      "isIOReceiverCreated": 0,
      "runningIONum": 0,
      "checkpointedIONum": 0,
      "degradedCheckpointedIONum": 0,
      "quorum": 1,
      "checkpointedTime": 0,
      "rebuildBytes": 0,
      "rebuildCnt": 0,
      "rebuildDoneCnt": 0,
      "rebuildFailedCnt": 0
    }
  ]
}`
		mockedStatusOutputRebuilding = `{
  "stats": [
    {
      "name": "cstor-183f17c6-ed8b-11e8-87fd-42010a800087/pvc-9e91f938-ee23-11e8-87fd-42010a800087",
      "status": "Rebuilding",
      "rebuildStatus": "INIT",
      "isIOAckSenderCreated": 0,
      "isIOReceiverCreated": 0,
      "runningIONum": 0,
      "checkpointedIONum": 0,
      "degradedCheckpointedIONum": 0,
      "quorum": 1,
      "checkpointedTime": 0,
      "rebuildBytes": 0,
      "rebuildCnt": 0,
      "rebuildDoneCnt": 0,
      "rebuildFailedCnt": 0
    }
  ]
}`
		mockedStatusOutputDegraded = `{
  "stats": [
    {
      "name": "cstor-183f17c6-ed8b-11e8-87fd-42010a800087/pvc-9e91f938-ee23-11e8-87fd-42010a800087",
      "status": "Degraded",
      "rebuildStatus": "INIT",
      "isIOAckSenderCreated": 0,
      "isIOReceiverCreated": 0,
      "runningIONum": 0,
      "checkpointedIONum": 0,
      "degradedCheckpointedIONum": 0,
      "quorum": 1,
      "checkpointedTime": 0,
      "rebuildBytes": 0,
      "rebuildCnt": 0,
      "rebuildDoneCnt": 0,
      "rebuildFailedCnt": 0
    }
  ]
}`
		mockedStatusOutputReconstructing = `{
  "stats": [
    {
      "name": "cstor-183f17c6-ed8b-11e8-87fd-42010a800087/pvc-9e91f938-ee23-11e8-87fd-42010a800087",
      "status": "Rebuilding",
      "rebuildStatus": "SNAP REBUILD INPROGRESS",
      "quorum": 0
    }
  ]
}`
		mockedStatusOutputNonQuorumDegraded = `{
  "stats": [
    {
      "name": "cstor-183f17c6-ed8b-11e8-87fd-42010a800087/pvc-9e91f938-ee23-11e8-87fd-42010a800087",
      "status": "Degraded",
      "rebuildStatus": "INIT",
      "quorum": 0
    }
  ]
}`
		mockedStatusOutputUnknown = `{
  "stats": [
    {
      "name": "cstor-183f17c6-ed8b-11e8-87fd-42010a800087/pvc-9e91f938-ee23-11e8-87fd-42010a800087",
      "status": "Error",
      "rebuildStatus": "Failed",
      "quorum": 0
    }
  ]
}`
	)
	if os.Getenv("GO_WANT_STATUS_HELPER_PROCESS") != "1" {
		return
	}
	// Pick the mocked output as specified by the env variable.
	if os.Getenv(testStatusType) == ZfsStatusHealthy {
		fmt.Fprint(os.Stdout, mockedStatusOutputHealthy)
	}
	if os.Getenv(testStatusType) == ZfsStatusDegraded {
		fmt.Fprint(os.Stdout, mockedStatusOutputDegraded)
	}
	if os.Getenv(testStatusType) == ZfsStatusOffline {
		fmt.Fprint(os.Stdout, mockedStatusOutputOffline)
	}
	if os.Getenv(testStatusType) == ZfsStatusRebuilding {
		fmt.Fprint(os.Stdout, mockedStatusOutputRebuilding)
	}
	if os.Getenv(testStatusType) == testZfsStatusUnknown {
		fmt.Fprint(os.Stdout, mockedStatusOutputUnknown)
	}
	if os.Getenv(testStatusType) == testZfsReconstructing {
		fmt.Fprint(os.Stdout, mockedStatusOutputReconstructing)
	}
	if os.Getenv(testStatusType) == testNonQuorumDegraded {
		fmt.Fprint(os.Stdout, mockedStatusOutputNonQuorumDegraded)
	}

	defer os.Exit(0)
}

// TestCapacityHelperProcess is a function that is run as a process to get the mocked std output
func TestCapacityHelperProcess(*testing.T) {
	// Following constants are different mocked output for `zfs get` command for capacity.
	const (
		mockedCapacityOutput = `NAME                                                PROPERTY     VALUE  SOURCE
cstor-d82bd105-f3a8-11e8-87fd-42010a800087/pvc-1b2a7d4b-f3a9-11e8-87fd-42010a800087  used         10K     -
cstor-d82bd105-f3a8-11e8-87fd-42010a800087/pvc-1b2a7d4b-f3a9-11e8-87fd-42010a800087  logicalused  6K     -
`
	)
	if os.Getenv("GO_WANT_CAPACITY_HELPER_PROCESS") != "1" {
		return
	}
	fmt.Fprint(os.Stdout, mockedCapacityOutput)
	defer os.Exit(0)
}

// TestSetCachefileProcess mocks zpool set cachefile.
func TestSetCachefileProcess(*testing.T) {
	if os.Getenv("SetErr") != "nil" {
		return
	}
	defer os.Exit(0)
	fmt.Println(nil)
}

// TestCreateVolumeReplica is to test cStorVolumeReplica creation.
func TestCreateVolumeReplica(t *testing.T) {
	fakeQuorums := []bool{true, false}
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
					TargetIP: "10.210.110.121",
					Capacity: "10MB",
				},
			},
		},
	}
	RunnerVar = TestRunner{}
	for _, fakeQuorum := range fakeQuorums {
		obtainedErr := CreateVolumeReplica(testPoolResource["Valid-vol1Resource"].test, "abcd123/dcba", fakeQuorum)
		if testPoolResource["Valid-vol1Resource"].expectedError != obtainedErr {
			t.Fatalf("Expected: %v, Got: %v", testPoolResource["Valid-vol1Resource"].expectedError, obtainedErr)
		}
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
					TargetIP: "10.210.110.121",
					Capacity: "100MB",
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
					TargetIP: "",
					Capacity: "100MB",
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
					TargetIP: "",
					Capacity: "100MB",
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
					TargetIP: "10.210.110.121",
					Capacity: "",
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
					TargetIP: "10.210.110.121",
					Capacity: "100MB",
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

// TestParseCapacityUnit tests parseCapacityUnit function which
// provide backward compatibility of capacity units.
func TestParseCapacityUnit(t *testing.T) {
	testVolumeCapacity := map[string]struct {
		volumeCapacity         string
		expectedVolumeCapacity string
	}{
		"capacity#1": {
			volumeCapacity: "1Ei",
			//expectedVolumeCapacity: "1.15292E",
			expectedVolumeCapacity: "1E",
		},
		"capacity#2": {
			volumeCapacity: "1Pi",
			//expectedVolumeCapacity: "1.12590P",
			expectedVolumeCapacity: "1P",
		},
		"capacity#3": {
			volumeCapacity: "1Ti",
			//expectedVolumeCapacity: "1.09951T",
			expectedVolumeCapacity: "1T",
		},
		"capacity#4": {
			volumeCapacity: "1Gi",
			//expectedVolumeCapacity: "1.07374G",
			expectedVolumeCapacity: "1G",
		},
		"capacity#5": {
			volumeCapacity: "1Mi",
			//expectedVolumeCapacity: "1.04858M",
			expectedVolumeCapacity: "1M",
		},
		"capacity#6": {
			volumeCapacity: "1Ki",
			//expectedVolumeCapacity: "1.024K",
			expectedVolumeCapacity: "1K",
		},
	}
	for name, test := range testVolumeCapacity {
		t.Run(name, func(t *testing.T) {
			gotVolumeCapacity := parseCapacityUnit(test.volumeCapacity)
			if gotVolumeCapacity != test.expectedVolumeCapacity {
				t.Errorf("Test case failed as expected capacity '%v' but got '%v'", test.expectedVolumeCapacity, gotVolumeCapacity)
			}
		})
	}
}

// TestVolumeStatus tests Status function which retunr cvr status.
func TestVolumeStatus(t *testing.T) {
	testPoolResource := map[string]struct {
		// cvrName holds the name of zfs volume(not the cvr object name).
		volumeName string
		// MockedOutputType holds the type for which the mocked output should be taken e.g.
		// for 'ZfsStatusHealthy' type a mocked output of `zfs stats` command for Healthy type is taken.
		mockedOutputType string
		// expectedStatus is the status that is expected for the test case.
		expectedStatus apis.CStorVolumeReplicaPhase
	}{
		// ToDo : Test case for error status.
		"#1 OnlineVolumeStatus": {
			volumeName:       "cstor-183f17c6-ed8b-11e8-87fd-42010a800087/pvc-9e91f938-ee23-11e8-87fd-42010a800087",
			mockedOutputType: ZfsStatusHealthy,
			expectedStatus:   apis.CVRStatusOnline,
		},
		"#2 OfflineVolumeStatus": {
			volumeName:       "cstor-183f17c6-ed8b-11e8-87fd-42010a800087/pvc-9e91f938-ee23-11e8-87fd-42010a800087",
			mockedOutputType: ZfsStatusOffline,
			expectedStatus:   apis.CVRStatusOffline,
		},
		"#3 DegradedVolumeStatus": {
			volumeName:       "cstor-183f17c6-ed8b-11e8-87fd-42010a800087/pvc-9e91f938-ee23-11e8-87fd-42010a800087",
			mockedOutputType: ZfsStatusDegraded,
			expectedStatus:   apis.CVRStatusDegraded,
		},
		"#4 RebuildingVolumeStatus": {
			volumeName:       "cstor-183f17c6-ed8b-11e8-87fd-42010a800087/pvc-9e91f938-ee23-11e8-87fd-42010a800087",
			mockedOutputType: ZfsStatusRebuilding,
			expectedStatus:   apis.CVRStatusRebuilding,
		},
		"#5 ErrorVolumeStatus": {
			volumeName:       "cstor-183f17c6-ed8b-11e8-87fd-42010a800087/pvc-9e91f938-ee23-11e8-87fd-42010a800087",
			mockedOutputType: testZfsStatusUnknown,
			expectedStatus:   apis.CVRStatusError,
		},
		"#6 ReconstructingVolumeStatus": {
			volumeName:       "cstor-183f17c6-ed8b-11e8-87fd-42010a800087/pvc-9e91f938-ee23-11e8-87fd-42010a800087",
			mockedOutputType: testZfsReconstructing,
			expectedStatus:   apis.CVRStatusReconstructing,
		},
		"#7 NonQuorumDegradedStatus": {
			volumeName:       "cstor-183f17c6-ed8b-11e8-87fd-42010a800087/pvc-9e91f938-ee23-11e8-87fd-42010a800087",
			mockedOutputType: testNonQuorumDegraded,
			expectedStatus:   apis.CVRStatusNonQuorumDegraded,
		},
	}
	for name, test := range testPoolResource {
		test := test
		name := name
		t.Run(name, func(t *testing.T) {
			// Set env variable "StatusType" to "mockedOutputType"
			// It will help to decide which mocked output should be considered as a std output.
			os.Setenv(testStatusType, test.mockedOutputType)
			RunnerVar = TestRunner{}
			gotStatus, err := Status(test.volumeName)
			if err != nil {
				t.Fatal("Some error occured in getting volume status:", err)
			}
			if string(test.expectedStatus) != gotStatus {
				t.Errorf("Test case failed as expected status '%s' but got '%s'", test.expectedStatus, gotStatus)
			}
			// Unset the "StatusType" env variable
			os.Unsetenv("StatusType")
		})
	}
}

// TestVolumeCapacity tests Capacity function.
func TestVolumeCapacity(t *testing.T) {
	testVolumeResource := map[string]struct {
		// volumeName holds the name of zfs volume. This name is the actual zfs volume name but not the cvr name.
		// However, volume name is trivial here as the the ouptut of 'zfs get' is being mocked and
		// changing the volume name to any value won't effect but the volume name is required by function
		// which is under test.
		volumeName string
		// expectedCapacity is the capacity that is expected for the test case.
		expectedCapacity *apis.CStorVolumeCapacityAttr
	}{
		"#1 VolumeCapacity": {
			volumeName: "cstor-530c9c4f-e0df-11e8-94a8-42010a80013b",
			expectedCapacity: &apis.CStorVolumeCapacityAttr{
				TotalAllocated: "10K",
				Used:           "6K",
			},
		},
	}
	for name, test := range testVolumeResource {
		t.Run(name, func(t *testing.T) {
			RunnerVar = TestRunner{}
			gotCapacity, err := Capacity(test.volumeName)
			if err != nil {
				t.Fatal("Some error occured in getting volume capacity:", err)
			}
			if !(reflect.DeepEqual(test.expectedCapacity, gotCapacity)) {
				t.Errorf("Test case failed as expected object: %v but got object:%v", test.expectedCapacity, gotCapacity)
			}
		})
	}
}
