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
package pool

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"testing"
	"time"

	"github.com/golang/glog"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type TestRunner struct {
	expectedError string
}

// RunCommandWithTimeoutContext is to mock Real runner exec.
func (r TestRunner) RunCommandWithTimeoutContext(timeout time.Duration, command string, args ...string) ([]byte, error) {
	return []byte("success"), nil
}

// RunCombinedOutput is to mock Real runner exec.
func (r TestRunner) RunCombinedOutput(command string, args ...string) ([]byte, error) {
	var cs []string
	var env []string
	var cmd *exec.Cmd
	switch args[0] {
	case "create":
		cs = []string{"-test.run=TestCreaterProcess", "--"}
		env = []string{"createErr=nil"}
		break
	case "import":
		cs = []string{"-test.run=TestImporterProcess", "--"}
		env = []string{"importErr=nil"}
		break
	case "destroy":
		cs = []string{"-test.run=TestDestroyerProcess", "--"}
		env = []string{"destroyErr=nil"}
		break
	case "labelclear":
		cs = []string{"-test.run=TestLabelClearerProcess", "--"}
		env = []string{"labelClearErr=nil"}
		break
	case "status":
		if len(r.expectedError) != 0 {
			return []byte(r.expectedError), nil
		}
		// Create command arguments
		cs = []string{"-test.run=TestStatusHelperProcess", "--", command}
		// Set env varibles for the 'TestStatusHelperProcess' function which runs as a process.
		env = []string{"GO_WANT_STATUS_HELPER_PROCESS=1", "StatusType=" + os.Getenv("StatusType")}
	case "get":
		if len(r.expectedError) != 0 {
			return []byte(r.expectedError), nil
		}
		// Create command arguments
		cs = []string{"-test.run=TestCapacityHelperProcess", "--", command}
		// Set env varibles for the 'TestCapacityHelperProcess' function which runs as a process.
		env = []string{"GO_WANT_CAPACITY_HELPER_PROCESS=1"}
	case "set":
		cs = []string{"-test.run=TestSetCachefileProcess", "--"}
		env = []string{"SetErr=nil"}
		break
	}
	cs = append(cs, args...)
	cmd = exec.Command(os.Args[0], cs...)
	cmd.Env = env
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

// TestCreaterProcess mocks zpool create.
func TestCreaterProcess(*testing.T) {
	if os.Getenv("createErr") != "nil" {
		return
	}
	fmt.Println(nil)
	defer os.Exit(0)

}

// TestImporterProcess mocks zpool import.
func TestImporterProcess(*testing.T) {
	if os.Getenv("importErr") != "nil" {
		return
	}
	defer os.Exit(0)
}

// TestGetterProcess mocks zpool get.
func TestGetterProcess(*testing.T) {
	if os.Getenv("poolName") != "cstor-123abc" {
		return
	}
	defer os.Exit(0)
	fmt.Println("cstor-123abc")
}

// TestDestroyerProcess mocks zpool destroy.
func TestDestroyerProcess(*testing.T) {
	if os.Getenv("destroyErr") != "nil" {
		return
	}
	defer os.Exit(0)
	fmt.Println(nil)
}

// TestLabelClearerProcess mocks zpool labelclear.
func TestLabelClearerProcess(*testing.T) {
	if os.Getenv("labelClearErr") != "nil" {
		return
	}
	defer os.Exit(0)
	fmt.Println(nil)
}

// TestStatusHelperProcess is a function that is run as a process to get the mocked std output
func TestStatusHelperProcess(*testing.T) {
	// Following constants are different mocked output for `zpool status` command for
	// different statuses.
	const (
		mockedStatusOutputOnline = `pool: cstor-530c9c4f-e0df-11e8-94a8-42010a80013b
		 state: ONLINE
		  scan: none requested
		config:

			NAME                                        STATE     READ WRITE CKSUM
			cstor-530c9c4f-e0df-11e8-94a8-42010a80013b  ONLINE       0     0     0
			  scsi-0Google_PersistentDisk_ashu-disk2    ONLINE       0     0     0

		errors: No known data errors`
		mockedStatusOutputOffline = `pool: cstor-530c9c4f-e0df-11e8-94a8-42010a80013b
		 state: OFFLINE
		  scan: none requested
		config:

			NAME                                        STATE     READ WRITE CKSUM
			cstor-530c9c4f-e0df-11e8-94a8-42010a80013b  OFFLINE       0     0     0
			  scsi-0Google_PersistentDisk_ashu-disk2    OFFLINE       0     0     0

		errors: No known data errors`
		mockedStatusOutputRemoved = `pool: cstor-530c9c4f-e0df-11e8-94a8-42010a80013b
		 state: REMOVED
		  scan: none requested
		config:

			NAME                                        STATE     READ WRITE CKSUM
			cstor-530c9c4f-e0df-11e8-94a8-42010a80013b  REMOVED       0     0     0
			  scsi-0Google_PersistentDisk_ashu-disk2    REMOVED       0     0     0

		errors: No known data errors`
		mockedStatusOutputUnavail = `pool: cstor-530c9c4f-e0df-11e8-94a8-42010a80013b
		 state: UNAVAIL
		  scan: none requested
		config:

			NAME                                        STATE     READ WRITE CKSUM
			cstor-530c9c4f-e0df-11e8-94a8-42010a80013b  UNAVAIL       0     0     0
			  scsi-0Google_PersistentDisk_ashu-disk2    UNAVAIL       0     0     0

		errors: No known data errors`
		mockedStatusOutputFaulted = `pool: cstor-530c9c4f-e0df-11e8-94a8-42010a80013b
		 state: FAULTED
		  scan: none requested
		config:

			NAME                                        STATE     READ WRITE CKSUM
			cstor-530c9c4f-e0df-11e8-94a8-42010a80013b  FAULTED       0     0     0
			  scsi-0Google_PersistentDisk_ashu-disk2    FAULTED       0     0     0

		errors: No known data errors`
		mockedStatusOutputDegraded = `pool: cstor-530c9c4f-e0df-11e8-94a8-42010a80013b
		 state: DEGRADED
		  scan: none requested
		config:

			NAME                                        STATE     READ WRITE CKSUM
			cstor-530c9c4f-e0df-11e8-94a8-42010a80013b  DEGRADED       0     0     0
			  scsi-0Google_PersistentDisk_ashu-disk2    DEGRADED       0     0     0

		errors: No known data errors`
	)
	if os.Getenv("GO_WANT_STATUS_HELPER_PROCESS") != "1" {
		return
	}
	// Pick the mocked output as specified by the env variable.
	if os.Getenv("StatusType") == ZpoolStatusOnline {
		fmt.Fprint(os.Stdout, mockedStatusOutputOnline)
	}
	if os.Getenv("StatusType") == ZpoolStatusOffline {
		fmt.Fprint(os.Stdout, mockedStatusOutputOffline)
	}
	if os.Getenv("StatusType") == ZpoolStatusDegraded {
		fmt.Fprint(os.Stdout, mockedStatusOutputDegraded)
	}
	if os.Getenv("StatusType") == ZpoolStatusFaulted {
		fmt.Fprint(os.Stdout, mockedStatusOutputFaulted)
	}
	if os.Getenv("StatusType") == ZpoolStatusRemoved {
		fmt.Fprint(os.Stdout, mockedStatusOutputRemoved)
	}
	if os.Getenv("StatusType") == ZpoolStatusUnavail {
		fmt.Fprint(os.Stdout, mockedStatusOutputUnavail)
	}
	defer os.Exit(0)
}

// TestCapacityHelperProcess is a function that is run as a process to get the mocked std output
func TestCapacityHelperProcess(*testing.T) {
	// Following constants are different mocked output for `zpool get` command for capacity.
	const (
		mockedCapacityOutput = `NAME       PROPERTY   VALUE  SOURCE
cstor-2ebe403a-f2e2-11e8-87fd-42010a800087  size       9.94G  -
cstor-2ebe403a-f2e2-11e8-87fd-42010a800087  free       9.94G  -
cstor-2ebe403a-f2e2-11e8-87fd-42010a800087  allocated  202K   -
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

// TestCreatePool is to test cStorPool creation.
func TestCreatePool(t *testing.T) {
	testPoolResource := map[string]struct {
		expectedError error
		test          *apis.CStorPool
	}{
		"img1PoolResource": {
			expectedError: nil,
			test: &apis.CStorPool{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					UID: types.UID("abc"),
				},
				Spec: apis.CStorPoolSpec{
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool1.cache",
						PoolType:         "striped",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
		},
		"img2PoolResource": {
			expectedError: nil,
			test: &apis.CStorPool{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					UID: types.UID("abc"),
				},
				Spec: apis.CStorPoolSpec{
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool1.cache",
						PoolType:         "mirrored",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
		},
		"img3PoolResource": {
			expectedError: nil,
			test: &apis.CStorPool{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					UID: types.UID("abc"),
				},
				Spec: apis.CStorPoolSpec{
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool1.cache",
						PoolType:         "raidz",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
		},
		"img4PoolResource": {
			expectedError: nil,
			test: &apis.CStorPool{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					UID: types.UID("abc"),
				},
				Spec: apis.CStorPoolSpec{
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool1.cache",
						PoolType:         "raidz2",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
		},
	}
	RunnerVar = TestRunner{}
	for desc, ut := range testPoolResource {
		obtainedErr := CreatePool(ut.test, []string{"test"})
		if ut.expectedError != obtainedErr {
			t.Fatalf("Desc: %v, Expected: %v, Got: %v", desc, ut.expectedError, obtainedErr)
		}
	}
}

// TestImportPool is to test cStorPool import.
func TestImportPool(t *testing.T) {
	testPoolResource := map[string]struct {
		expectedError error
		test          *apis.CStorPool
		cachefileFlag bool
	}{
		"img1PoolResource": {
			expectedError: nil,
			cachefileFlag: true,
			test: &apis.CStorPool{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					UID: types.UID("abc"),
				},
				Spec: apis.CStorPoolSpec{
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool1.cache",
						PoolType:         "mirrored",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
		},
		"img2PoolResource": {
			expectedError: nil,
			cachefileFlag: false,
			test: &apis.CStorPool{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					UID: types.UID("abc"),
				},
				Spec: apis.CStorPoolSpec{
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool1.cache",
						PoolType:         "mirrored",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
		},
	}
	RunnerVar = TestRunner{}
	for desc, ut := range testPoolResource {
		obtainedErr := ImportPool(ut.test, ut.cachefileFlag)
		if ut.expectedError != obtainedErr {
			t.Fatalf("desc:%v, Expected: %v, Got: %v", desc, ut.expectedError, obtainedErr)
		}
	}
}

// TestDeletePool is to test cStorPool delete.
func TestDeletePool(t *testing.T) {
	testPoolResource := map[string]struct {
		expectedError error
		poolName      string
	}{
		"img1PoolResource": {
			expectedError: nil,
			poolName:      "pool1-a2b",
		},
	}
	RunnerVar = TestRunner{}
	obtainedErr := DeletePool(testPoolResource["img1PoolResource"].poolName)
	if testPoolResource["img1PoolResource"].expectedError != obtainedErr {
		t.Fatalf("Expected: %v, Got: %v", testPoolResource["img1PoolResource"].expectedError, obtainedErr)
	}
}

// TestDeletePool is to test cStorPool delete.
func TestLabelClear(t *testing.T) {
	testResource := map[string]struct {
		expectedError error
		disks         []string
	}{
		"Resource1": {
			expectedError: nil,
			disks:         []string{"/dev/sdb1"},
		},
	}
	RunnerVar = TestRunner{}
	obtainedErr := LabelClear(testResource["Resource1"].disks)
	if testResource["Resource1"].expectedError != obtainedErr {
		t.Fatalf("Expected: %v, Got: %v", testResource["Resource1"].expectedError, obtainedErr)
	}
}

// TestSetCacheFile is to test cachefile set for pool.
func TestSetCacheFile(t *testing.T) {
	testPoolResource := map[string]struct {
		expectedError error
		test          *apis.CStorPool
	}{
		"img1PoolResource": {
			expectedError: nil,
			test: &apis.CStorPool{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					UID: types.UID("abc"),
				},
				Spec: apis.CStorPoolSpec{
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool1.cache",
						PoolType:         "striped",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
		},
	}
	RunnerVar = TestRunner{}
	for desc, ut := range testPoolResource {
		obtainedErr := SetCachefile(ut.test)
		if ut.expectedError != obtainedErr {
			t.Fatalf("Desc: %v, Expected: %v, Got: %v", desc, ut.expectedError, obtainedErr)
		}
	}
}

// TestCheckForZreplInitialis to test zrepl running.
func TestCheckForZreplInitial(t *testing.T) {
	done := make(chan bool)
	RunnerVar = TestRunner{}
	go func(done chan bool) {
		CheckForZreplInitial(3 * time.Second)
		done <- true
	}(done)

	select {
	case <-time.After(3 * time.Second):
		t.Fatalf("Check for Zrepl initial test failure - Timed out")
	case <-done:

	}
}

// TestCheckForZreplContinuous is to test zrepl running.
func TestCheckForZreplContinuous(t *testing.T) {
	PoolAddEventHandled = true
	testcases := map[string]struct {
		expectedError error
		runner        util.Runner
	}{
		"statusSuccess": {
			expectedError: nil,
			runner:        TestRunner{},
		},
		"statusNoPoolsAvailable": {
			expectedError: fmt.Errorf(StatusNoPoolsAvailable),
			runner:        TestRunner{expectedError: StatusNoPoolsAvailable},
		},
	}

	for desc, ut := range testcases {
		done := make(chan bool)
		RunnerVar = ut.runner
		go func(done chan bool) {
			CheckForZreplContinuous(1 * time.Second)
			done <- true
		}(done)
		if ut.expectedError != nil {
			select {
			case <-done:
			case <-time.After(3 * time.Second):
				t.Fatalf("Check for Zrepl continuous test failure for expected error case %s", desc)
			}
		} else {
			select {
			case <-time.After(3 * time.Second):
			case <-done:
				t.Fatalf("Check for Zrepl continuous test failure")
			}
		}
	}
	PoolAddEventHandled = false
}

// TestGetPool is to test zrepl running.
func TestGetPool(t *testing.T) {
	testPoolResource := map[string]struct {
		expectedPoolName []string
		expectedError    error
	}{
		"img1PoolResource": {
			expectedPoolName: []string{"cstor-123abc"},
			expectedError:    nil,
		},
	}
	RunnerVar = TestRunner{}
	obtainedPoolName, obtainedErr := GetPoolName()
	if reflect.DeepEqual(testPoolResource["img1PoolResource"].expectedPoolName, obtainedPoolName) {
		t.Fatalf("Expected: %v, Got: %v", testPoolResource["img1PoolResource"].expectedPoolName, obtainedPoolName)
	}
	if testPoolResource["img1PoolResource"].expectedError != obtainedErr {
		t.Fatalf("Expected: %v, Got: %v", testPoolResource["img1PoolResource"].expectedError, obtainedErr)
	}
}

// TestCheckValidPool tests pool related operations.
func TestCheckValidPool(t *testing.T) {
	testPoolResource := map[string]struct {
		test          *apis.CStorPool
		deviceIDs     []string
		expectedError bool
	}{
		"Invalid-poolNameEmpty": {
			expectedError: true,
			test: &apis.CStorPool{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					Name: "pool1-abc",
					UID:  types.UID(""),
				},
				Spec: apis.CStorPoolSpec{
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool1.cache",
						PoolType:         "mirrored",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
			deviceIDs: []string{"/var/img-0", "/var/img-1"},
		},
		"Valid-StripedDisks1": {
			expectedError: false,
			test: &apis.CStorPool{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					UID: types.UID("abc"),
				},
				Spec: apis.CStorPoolSpec{
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool1.cache",
						PoolType:         "striped",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
			deviceIDs: []string{"/var/img-0"},
		},
		"Valid-StripedDisks2": {
			expectedError: false,
			test: &apis.CStorPool{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					UID: types.UID("abc"),
				},
				Spec: apis.CStorPoolSpec{
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool1.cache",
						PoolType:         "striped",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
			deviceIDs: []string{"/var/img-0", "var/img-1"},
		},

		"Invalid-StripedDisks": {
			expectedError: true,
			test: &apis.CStorPool{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					UID: types.UID("abc"),
				},
				Spec: apis.CStorPoolSpec{
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool1.cache",
						PoolType:         "striped",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
			deviceIDs: []string{},
		},

		"Invalid-DiskListEmpty": {
			expectedError: true,
			test: &apis.CStorPool{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					UID: types.UID("abc"),
				},
				Spec: apis.CStorPoolSpec{
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool1.cache",
						PoolType:         "mirrored",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
			deviceIDs: []string{},
		},
		"Invalid-MirrorOddDisks": {
			expectedError: true,
			test: &apis.CStorPool{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					UID: types.UID("abc"),
				},
				Spec: apis.CStorPoolSpec{
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool1.cache",
						PoolType:         "mirrored",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
			deviceIDs: []string{"/var/img-0", "/var/img-1", "/var/img-2"},
		},
		"Valid-Pool": {
			expectedError: false,
			test: &apis.CStorPool{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					UID: types.UID("abc"),
				},
				Spec: apis.CStorPoolSpec{
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool1.cache",
						PoolType:         "mirrored",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
			deviceIDs: []string{"/var/img-0", "/var/img-1"},
		},
		"Valid-RaidzDisks": {
			expectedError: false,
			test: &apis.CStorPool{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					UID: types.UID("abc"),
				},
				Spec: apis.CStorPoolSpec{
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool1.cache",
						PoolType:         "raidz",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
			deviceIDs: []string{"/var/img-0", "/var/img-1", "/var/img-2"},
		},
		"Invalid-NoOfRaidzDisks": {
			expectedError: true,
			test: &apis.CStorPool{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					UID: types.UID("abc"),
				},
				Spec: apis.CStorPoolSpec{
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool1.cache",
						PoolType:         "raidz",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
			deviceIDs: []string{"/var/img-0", "/var/img-1"},
		},
		"Valid-Raidz2Disks": {
			expectedError: false,
			test: &apis.CStorPool{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					UID: types.UID("abc"),
				},
				Spec: apis.CStorPoolSpec{
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool1.cache",
						PoolType:         "raidz2",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
			deviceIDs: []string{"/var/img-0", "/var/img-1", "/var/img-2", "/var/img-3", "/var/img-4", "/var/img-5"},
		},
		"Invalid-Raidz2Disks": {
			expectedError: true,
			test: &apis.CStorPool{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					UID: types.UID("abc"),
				},
				Spec: apis.CStorPoolSpec{
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool1.cache",
						PoolType:         "raidz2",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
			deviceIDs: []string{"/var/img-0", "/var/img-1", "/var/img-2", "/var/img-3", "/var/img-4"},
		},
		"Valid-RaidzDisks2": {
			expectedError: true,
			test: &apis.CStorPool{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					UID: types.UID("abc"),
				},
				Spec: apis.CStorPoolSpec{
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool1.cache",
						PoolType:         "raidz",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
			deviceIDs: []string{"/var/img-0", "/var/img-1", "/var/img-2", "/var/img-3", "/var/img-4", "/var/img-5", "/var/img-6"},
		},
		"InValid-RaidzDisks3": {
			expectedError: true,
			test: &apis.CStorPool{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					UID: types.UID("abc"),
				},
				Spec: apis.CStorPoolSpec{
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool1.cache",
						PoolType:         "raidz",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
			deviceIDs: []string{"/var/img-0", "/var/img-1", "/var/img-2", "/var/img-3", "/var/img-4"},
		},
	}
	for name, ut := range testPoolResource {
		name := name
		ut := ut
		t.Run(name, func(t *testing.T) {
			Obtainederr := ValidatePool(ut.test, ut.deviceIDs)
			if ut.expectedError && Obtainederr == nil {
				t.Fatalf("Desc : %q, Expected error not be nil", name)
			}
			if !ut.expectedError && Obtainederr != nil {
				t.Fatalf("Desc : %q, Expected error to be nil got: %v", name, Obtainederr)
			}
		})
	}
}

// TestPoolStatus tests PoolStatus function.
func TestPoolStatus(t *testing.T) {
	testPoolResource := map[string]struct {
		// PoolName holds the name of pool. This name is the actual zpool name but not the spc name.
		// However, pool name is trivial here as the the ouptut of zpool status is being mocked and
		// changing the pool name to any value won't effect but the pool name is required by function
		// which is under test.
		poolName string
		// MockedOutputType holds the type for which the mocked output should be taken e.g.
		// for 'ONLINE' type a mocked output of `zpool status` command for ONLINE type is taken.
		mockedOutputType string
		// expectedStatus is the status that is expected for the test case.
		expectedStatus string
	}{
		"#1 OnlinePoolStatus": {
			poolName:         "cstor-530c9c4f-e0df-11e8-94a8-42010a80013b",
			mockedOutputType: ZpoolStatusOnline,
			expectedStatus:   "Healthy",
		},
		"#2 OfflinePoolStatus": {
			poolName:         "cstor-530c9c4f-e0df-11e8-94a8-42010a80013b",
			mockedOutputType: ZpoolStatusOffline,
			expectedStatus:   "Offline",
		},
		"#3 UnavailPoolStatus": {
			poolName:         "cstor-530c9c4f-e0df-11e8-94a8-42010a80013b",
			mockedOutputType: ZpoolStatusUnavail,
			expectedStatus:   "Offline",
		},
		"#4 RemovedPoolStatus": {
			poolName:         "cstor-530c9c4f-e0df-11e8-94a8-42010a80013b",
			mockedOutputType: ZpoolStatusRemoved,
			expectedStatus:   "Degraded",
		},
		"#5 FaultedPoolStatus": {
			poolName:         "cstor-530c9c4f-e0df-11e8-94a8-42010a80013b",
			mockedOutputType: ZpoolStatusFaulted,
			expectedStatus:   "Offline",
		},
		"#6 DegradedPoolStatus": {
			poolName:         "cstor-530c9c4f-e0df-11e8-94a8-42010a80013b",
			mockedOutputType: ZpoolStatusDegraded,
			expectedStatus:   "Degraded",
		},
	}
	for name, test := range testPoolResource {
		t.Run(name, func(t *testing.T) {
			// Set env variable "StatusType" to "mockedOutputType"
			// It will help to decide which mocked output should be considered as a std output.
			os.Setenv("StatusType", test.mockedOutputType)
			RunnerVar = TestRunner{}
			gotStatus, err := Status(test.poolName)
			if err != nil {
				t.Fatal("Some error occured in getting pool status:", err)
			}
			if test.expectedStatus != gotStatus {
				t.Errorf("Test case failed as expected status '%s' but got '%s'", test.expectedStatus, gotStatus)
			}
			// Unset the "StatusType" env variable
			os.Unsetenv("StatusType")
		})
	}
}

// TestPoolCapacity tests Capacity function.
func TestPoolCapacity(t *testing.T) {
	testPoolResource := map[string]struct {
		// PoolName holds the name of pool. This name is the actual zpool name but not the spc name.
		// However, pool name is trivial here as the the ouptut of 'zpool get' is being mocked and
		// changing the pool name to any value won't effect but the pool name is required by function
		// which is under test.
		poolName string
		// expectedCapacity is the capacity that is expected for the test case.
		expectedCapacity *apis.CStorPoolCapacityAttr
	}{
		"#1 OnlinePoolStatus": {
			poolName: "cstor-530c9c4f-e0df-11e8-94a8-42010a80013b",
			expectedCapacity: &apis.CStorPoolCapacityAttr{
				"9.94G",
				"9.94G",
				"202K",
			},
		},
	}
	for name, test := range testPoolResource {
		t.Run(name, func(t *testing.T) {
			RunnerVar = TestRunner{}
			gotCapacity, err := Capacity(test.poolName)
			if err != nil {
				t.Fatal("Some error occured in getting pool capacity:", err)
			}
			if !(reflect.DeepEqual(test.expectedCapacity, gotCapacity)) {
				t.Errorf("Test case failed as expected object: %v but got object:%v", test.expectedCapacity, gotCapacity)
			}
		})
	}
}
