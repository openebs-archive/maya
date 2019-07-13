/*
Copyright 2019 The OpenEBS Authors.

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

package v1alpha2

import (
	"github.com/golang/glog"
	"github.com/openebs/maya/cmd/cstor-pool-mgmt/controller/common"
	"github.com/openebs/maya/cmd/cstor-pool-mgmt/volumereplica"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	zfs "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1"
)

// Import will import pool for given CSP object.
// It will also set `cachefile` property for that pool
// if it is mentioned in object
// It will return -
// - If pool is imported or not
// - If any error occurred during import operation
func Import(csp *apis.NewTestCStorPool) (bool, error) {
	if poolExist := checkIfPoolPresent(PoolName(csp)); poolExist {
		return true, nil
	}

	// Pool is not imported.. Let's update the syncResource
	common.SyncResources.IsImported = false

	if ret, er := zfs.NewPoolImport().
		WithCachefile(csp.Spec.PoolConfig.CacheFile).
		WithProperty("cachefile", csp.Spec.PoolConfig.CacheFile).
		WithDirectory(SparseDir).
		WithDirectory(DevDir).
		WithPool(PoolName(csp)).
		Execute(); er != nil {
		glog.Errorf("Failed to import pool : %s : %s", ret, er.Error())
		// We return error as nil because pool doesn't exist
		return false, nil
	}

	glog.Infof("Pool Import successful: %v", string(csp.GetUID()))
	return true, nil
}

// CheckImportedPoolVolume will notify CVR controller
// for new imported pool's volumes
func CheckImportedPoolVolume() {
	var err error

	if common.SyncResources.IsImported {
		return
	}

	// GetVolumes is called because, while importing a pool, volumes corresponding
	// to the pool are also imported. This needs to be handled and made visible
	// to cvr controller.
	common.InitialImportedPoolVol, err = volumereplica.GetVolumes()
	if err != nil {
		common.SyncResources.IsImported = false
		return
	}

	// make a check if initialImportedPoolVol is not empty, then notify cvr controller
	// through channel.
	if len(common.InitialImportedPoolVol) != 0 {
		common.SyncResources.IsImported = true
	} else {
		common.SyncResources.IsImported = false
	}
}
