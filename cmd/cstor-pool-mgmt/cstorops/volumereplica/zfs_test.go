package volumereplica

import (
	"fmt"
	"testing"
	"time"

	"github.com/openebs/maya/cmd/cstor-pool-mgmt/cstorops/pool"
	"github.com/openebs/maya/cmd/cstor-pool-mgmt/cstorops/uzfs"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
)

// TestVolumeReplica tests cstor volume replica operations
func TestVolumeReplica(t *testing.T) {
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

	pool2 := &apis.CStorPool{
		Spec: apis.CStorPoolSpec{
			Disks: apis.DiskAttr{
				DiskList: []string{"/tmp/img2.img"},
			},
			PoolSpec: apis.CStorPoolAttr{
				PoolName:  "pool2",
				CacheFile: "/tmp/pool2.cache",
				PoolType:  "mirror",
			},
		},
	}

	actualPoolName, err := pool.GetPoolName()
	if err == nil {
		pool.DeletePool(actualPoolName)
	}

	time.Sleep(3 * time.Second)
	err = pool.CreatePool(pool2)
	if err != nil {
		t.Fatalf("Unable to create pool, %v", err.Error())
	}

	testVolumeResource := map[string]struct {
		expectedVolumeName string
		expectedError      error
		test               *apis.CStorVolumeReplica
	}{
		"volReplicaResource_1": {
			expectedVolumeName: pool2.Spec.PoolSpec.PoolName + "/" + "vol1",
			expectedError:      nil,
			test: &apis.CStorVolumeReplica{
				Spec: apis.CStorVolumeReplicaSpec{
					CStorControllerIP: "10.210.110.120",
					VolName:           "vol1",
					Capacity:          "100MB",
				},
			},
		},

		"volReplicaResource_2": {
			expectedVolumeName: pool2.Spec.PoolSpec.PoolName + "/" + "abcdefgh_Volume_2",
			expectedError:      nil,
			test: &apis.CStorVolumeReplica{
				Spec: apis.CStorVolumeReplicaSpec{
					CStorControllerIP: "10.210.110.121",
					VolName:           "abcdefgh_Volume_2",
					Capacity:          "100MB",
				},
			},
		},
	}

	for desc, ut := range testVolumeResource {
		Obtainederr := CheckValidVolumeReplica(ut.test)
		if Obtainederr != nil {
			if Obtainederr == ut.expectedError {
				return
			}
			t.Fatalf("Desc : %v, Expected error: %v, Got : %v",
				desc, ut.expectedError, Obtainederr)
		}

		time.Sleep(2 * time.Second)
		actualPoolName, err := pool.GetPoolName()
		if err != nil {
			t.Fatalf("Desc : %v, Unable to get pool name", desc)
		}
		time.Sleep(2 * time.Second)
		volNames := GetVolumes()

		time.Sleep(2 * time.Second)
		err = CreateVolume(ut.test, actualPoolName+"/"+ut.test.Spec.VolName)
		if err != nil {
			t.Fatalf("Unable to create volume replica: %v", err.Error())
		}

		volNames = GetVolumes()
		var availableFlag = false
		for _, volName := range volNames {
			if volName == ut.expectedVolumeName {
				availableFlag = true
				break
			}
		}
		if !availableFlag {
			t.Errorf("desc: %v, Fail : %v is not available", desc, ut.expectedVolumeName)
		}
	}
	pool.DeletePool(pool2.Spec.PoolSpec.PoolName)
	pool.DeletePool(pool2.Spec.PoolSpec.PoolName)

}
