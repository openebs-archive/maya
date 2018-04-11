package pool

import (
	"fmt"
	"reflect"
	"testing"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
)

// TestCheckValidPool tests pool related operations
func TestCheckValidPool(t *testing.T) {
	testPoolResource := map[string]struct {
		expectedPoolName string
		expectedError    error
		test             *apis.CStorPool
	}{
		"Valid-img1PoolResource": {
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
		"Invalid-poolNameEmpty": {
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
		"Invalid-DiskListEmpty": {
			expectedPoolName: "",
			expectedError:    fmt.Errorf("Disk name(s) cannot be empty"),
			test: &apis.CStorPool{
				Spec: apis.CStorPoolSpec{
					Disks: apis.DiskAttr{
						DiskList: []string{""},
					},
					PoolSpec: apis.CStorPoolAttr{
						PoolName:  "pool1",
						CacheFile: "/tmp/pool1.cache",
						PoolType:  "mirror",
					},
				},
			},
		},
	}

	for desc, ut := range testPoolResource {
		Obtainederr := CheckValidPool(ut.test)
		if Obtainederr != nil {
			if Obtainederr.Error() == ut.expectedError.Error() {
				return
			}
			t.Fatalf("Desc : %v, Expected error: %v, Got : %v",
				desc, ut.expectedError, Obtainederr)
		}

	}
}

// TestCheckValidPool tests pool related operations
func TestImportPoolBuilder(t *testing.T) {
	testPoolResource := map[string]struct {
		expectedCmd []string
		test        *apis.CStorPool
	}{
		"img1PoolResource": {
			expectedCmd: []string{PoolOperator + " import -c cachefile=/tmp/pool1.cache pool1"},
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

		"img2PoolResource": {
			expectedCmd: []string{PoolOperator + " import -c cachefile=/tmp/pool2.cache pool2"},
			test: &apis.CStorPool{
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
			},
		},
	}

	for desc, ut := range testPoolResource {
		obtainedCmd := importPoolBuilder(ut.test)
		if reflect.DeepEqual(ut.expectedCmd, obtainedCmd.Args) {
			t.Fatalf("desc:%v, Commands mismatch, expected:%v, Got:%v", desc,
				ut.expectedCmd, obtainedCmd.Args[0])
		}
	}
}

// TestCheckValidPool tests pool related operations
func TestCreatePoolBuilder(t *testing.T) {
	testPoolResource := map[string]struct {
		expectedCmd []string
		test        *apis.CStorPool
	}{
		"img1PoolResource": {
			expectedCmd: []string{PoolOperator + " create -f -o cachefile=/tmp/pool1.cache pool1"},
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

		"img2PoolResource": {
			expectedCmd: []string{PoolOperator + " create -f -o cachefile=/tmp/pool2.cache pool2"},
			test: &apis.CStorPool{
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
			},
		},
	}

	for desc, ut := range testPoolResource {
		obtainedCmd := createPoolBuilder(ut.test)
		if reflect.DeepEqual(ut.expectedCmd, obtainedCmd.Args) {
			t.Fatalf("desc:%v, Commands mismatch, expected:%v, Got:%v", desc,
				ut.expectedCmd, obtainedCmd.Args[0])
		}
	}
}
