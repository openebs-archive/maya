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
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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
	case "set":
		cs = []string{"-test.run=TestSetCachefileProcess", "--"}
		cmd.Env = []string{"SetErr=nil"}
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

// TestStatusProcess mocks zpool status.
func TestStatusProcess(*testing.T) {
	if os.Getenv("StatusErr") != "nil" {
		return
	}
	defer os.Exit(0)
	fmt.Println(nil)
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
					Disks: apis.DiskAttr{
						DiskList: []string{"/tmp/img1.img", "/tmp/img2.img", "/tmp/img3.img", "/tmp/img4.img"},
					},
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
					Disks: apis.DiskAttr{
						DiskList: []string{"/tmp/img1.img", "/tmp/img2.img", "/tmp/img3.img", "/tmp/img4.img"},
					},
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool1.cache",
						PoolType:         "mirror",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
		},
	}
	RunnerVar = TestRunner{}
	for desc, ut := range testPoolResource {
		obtainedErr := CreatePool(ut.test)
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
					Disks: apis.DiskAttr{
						DiskList: []string{"/tmp/img1.img"},
					},
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool1.cache",
						PoolType:         "mirror",
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
					Disks: apis.DiskAttr{
						DiskList: []string{"/tmp/img1.img"},
					},
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool1.cache",
						PoolType:         "mirror",
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
					Disks: apis.DiskAttr{
						DiskList: []string{"/tmp/img1.img", "/tmp/img2.img", "/tmp/img3.img", "/tmp/img4.img"},
					},
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

// TestCheckForZrepl is to test zrepl running.
func TestCheckForZrepl(t *testing.T) {
	done := make(chan bool)
	RunnerVar = TestRunner{}
	go func(done chan bool) {
		CheckForZrepl()
		done <- true
	}(done)

	select {
	case <-time.After(3 * time.Second):
		t.Fatalf("Zrepl test failure - Timed out")
	case <-done:

	}
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
		expectedError error
		test          *apis.CStorPool
	}{
		"Invalid-poolNameEmpty": {
			expectedError: fmt.Errorf("Poolname/UID cannot be empty"),
			test: &apis.CStorPool{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					Name: "pool1-abc",
					UID:  types.UID(""),
				},
				Spec: apis.CStorPoolSpec{
					Disks: apis.DiskAttr{
						DiskList: []string{"/tmp/img1.img"},
					},
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool1.cache",
						PoolType:         "mirror",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
		},
		"Invalid-DiskListEmpty": {
			expectedError: fmt.Errorf("Disk name(s) cannot be empty"),
			test: &apis.CStorPool{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					UID: types.UID("abc"),
				},
				Spec: apis.CStorPoolSpec{
					Disks: apis.DiskAttr{
						DiskList: []string{},
					},
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool1.cache",
						PoolType:         "mirror",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
		},
		"Invalid-MirrorOddDisks": {
			expectedError: fmt.Errorf("Mirror poolType needs even number of disks"),
			test: &apis.CStorPool{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					UID: types.UID("abc"),
				},
				Spec: apis.CStorPoolSpec{
					Disks: apis.DiskAttr{
						DiskList: []string{"/dev/sdb", "/dev/sdc", "/dev/sdd"},
					},
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool1.cache",
						PoolType:         "mirror",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
		},
		"Valid-Pool": {
			expectedError: nil,
			test: &apis.CStorPool{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					UID: types.UID("abc"),
				},
				Spec: apis.CStorPoolSpec{
					Disks: apis.DiskAttr{
						DiskList: []string{"/dev/sdb", "/dev/sdc"},
					},
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool1.cache",
						PoolType:         "mirror",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
		},
	}
	for desc, ut := range testPoolResource {
		Obtainederr := CheckValidPool(ut.test)
		if Obtainederr != nil {
			if Obtainederr.Error() != ut.expectedError.Error() {
				t.Fatalf("Desc : %v, Expected error: %v, Got : %v",
					desc, ut.expectedError, Obtainederr)
			}
		}
	}
}
