package pool

import (
	"fmt"
	"testing"
	"time"

	"github.com/openebs/maya/cmd/cstor-pool-mgmt/cstorops/uzfs"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
)

// TestCStorPool tests pool related operations
func TestCStorPool(t *testing.T) {
	// Uncomment this in case of zrepl crash
	/*
		RemoveStale()
		time.Sleep(5 * time.Second)
		go func() {
			err := StartZrepl()
			fmt.Println("at start zrepl")
			if err != nil {
				t.Fatalf(err.Error())
			}
		}()
	*/
	done := make(chan bool)
	go func() {
		uzfs.CheckForZrepl()
		fmt.Println("uzfs is running")
		done <- true
	}()
	select {
	case <-time.After(20 * time.Second):
		t.Fatalf("Timeout error")
	case <-done:

	}

	testPoolResource := map[string]struct {
		expectedPoolName string
		expectedError    error
		test             *apis.CStorPool
	}{
		"img1PoolResource": {
			expectedPoolName: "pool1",
			expectedError:    nil,
			test: &apis.CStorPool{
				Spec: apis.CStorPoolSpec{
					Disks: apis.DiskAttr{
						DiskList: []string{"/tmp/img1.img"},
					},
					PoolSpec: apis.CStorPoolAttr{
						PoolName:  "pool1",
						CacheFile: "/tmp/pool1.cache",
						PoolType:  "mirror",
					},
				},
			},
		},
		"poolNameEmpty": {
			expectedPoolName: "",
			expectedError:    fmt.Errorf("Poolname cannot be empty"),
			test: &apis.CStorPool{
				Spec: apis.CStorPoolSpec{
					Disks: apis.DiskAttr{
						DiskList: []string{"/tmp/img1.img"},
					},
					PoolSpec: apis.CStorPoolAttr{
						PoolName:  "",
						CacheFile: "/tmp/pool1.cache",
						PoolType:  "mirror",
					},
				},
			},
		},
	}

	for desc, ut := range testPoolResource {
		olderPoolName, err := GetPoolName()
		if err == nil {
			DeletePool(olderPoolName)
		}

		Obtainederr := CheckValidPool(ut.test)
		if Obtainederr != nil {
			if Obtainederr.Error() == ut.expectedError.Error() {
				return
			}
			t.Fatalf("Desc : %v, Expected error: %v, Got : %v",
				desc, ut.expectedError, Obtainederr)
		}

		CreatePool(ut.test)
		actualPoolName, err := GetPoolName()
		if err != nil {
			t.Errorf("Desc : %v, Unable to get pool name", desc)
		}
		DeletePool(actualPoolName)
		if actualPoolName != ut.expectedPoolName {
			t.Fatalf("Desc : %v, expectedPoolName: %v, Got: %v ",
				desc, ut.expectedPoolName, actualPoolName)
		}
	}
}
