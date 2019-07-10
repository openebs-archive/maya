package v1alpha2

import (
	"github.com/golang/glog"
	"github.com/openebs/maya/cmd/cstor-pool-mgmt/controller/common"
	"github.com/openebs/maya/cmd/cstor-pool-mgmt/volumereplica"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	api "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha2"
	zfs "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1"
)

// Import will import pool for given CSP object.
// It will also set `cachefile` property for that pool
// if it is mentioned in object
func Import(csp *api.CStorNPool) (string, bool, error) {
	var err error

	if poolExist := checkIfPoolPresent(PoolName(csp)); poolExist {
		return "", true, nil
	}

	if ret, er := zfs.NewPoolImport().
		WithCachefile(csp.Spec.PoolConfig.CacheFile).
		WithProperty("cachefile", csp.Spec.PoolConfig.CacheFile).
		WithPool(PoolName(csp)).
		Execute(); er != nil {
		glog.Errorf("Failed to import pool : %s : %s", ret, er.Error())
		// We return error as nil because pool doesn't exist
		return "", false, nil
	}

	glog.Infof("Pool Import successful: %v", string(csp.GetUID()))
	common.SyncResources.IsImported = true

	// TODO: audit required
	// GetVolumes is called because, while importing a pool, volumes corresponding
	// to the pool are also imported. This needs to be handled and made visible
	// to cvr controller.
	common.InitialImportedPoolVol, err = volumereplica.GetVolumes()
	if err != nil {
		common.SyncResources.IsImported = false
		return string(apis.CStorPoolStatusOffline), true, err
	}

	// make a check if initialImportedPoolVol is not empty, then notify cvr controller
	// through channel.
	if len(common.InitialImportedPoolVol) != 0 {
		common.SyncResources.IsImported = true
	} else {
		common.SyncResources.IsImported = false
	}

	return string(apis.CStorPoolStatusOnline), true, nil
}
