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
	ret, err := zfs.NewPoolImport().
		WithCachefile(csp.Spec.PoolConfig.CacheFile).
		WithPool(PoolName(csp)).
		Execute()
	if err != nil {
		glog.Errorf("Failed to import pool : %s : %s", ret, err.Error())
		// We return error as nil because pool doesn't exist
		return "", false, nil
	}

	// We imported pool successfully
	// Let's set cachefile for this pool, if it is provided in csp object
	if len(csp.Spec.PoolConfig.CacheFile) != 0 {
		if _, err = zfs.NewPoolSProperty().
			WithProperty("cachefile", csp.Spec.PoolConfig.CacheFile).
			Execute(); err != nil {
			//TODO, If cachefile set failed, do we need to return status as offline?
			glog.Errorf("Failed to set cachefile for pool {%s} : %s", PoolName(csp), err.Error())
			common.SyncResources.IsImported = false
			return string(apis.CStorPoolStatusOffline), true, err
		}
		glog.Infof("Set cachefile successful for pool {%s}", PoolName(csp))
	}

	// TODO: audit required
	// GetVolumes is called because, while importing a pool, volumes corresponding
	// to the pool are also imported. This needs to be handled and made visible
	// to cvr controller.
	common.InitialImportedPoolVol, err = volumereplica.GetVolumes()
	if err != nil {
		common.SyncResources.IsImported = false
		return string(apis.CStorPoolStatusOffline), true, err
	}

	glog.Infof("Import Pool with cachefile successful: %v", string(csp.GetUID()))
	common.SyncResources.IsImported = true

	// make a check if initialImportedPoolVol is not empty, then notify cvr controller
	// through channel.
	if len(common.InitialImportedPoolVol) != 0 {
		common.SyncResources.IsImported = true
	} else {
		common.SyncResources.IsImported = false
	}

	// Add entry to imported pool list
	ImportedCStorPools[string(csp.GetUID())] = csp

	return string(apis.CStorPoolStatusOnline), true, nil
}

// IsPoolImported check if pool is imported or not
func IsPoolImported(csp *api.CStorNPool, shouldWait bool) (string, bool) {
	/* TODO: audit
	If pool is already present.
	Pool CR status is online. This means pool (main car) is running successfully,
	but watcher container got restarted.
	Pool CR status is init/online. If entire pod got restarted, both zrepl and watcher
	are started.
	a) Zrepl could have come up first, in this case, watcher will update after
	the specified interval of (2*30) = 60s.
	b) Watcher could have come up first, in this case, there is a possibility
	that zrepl goes down and comes up and the watcher sees that no pool is there,
	so it will break the loop and attempt to import the pool. */

	// cnt is no of attempts to wait and handle in case of already present pool.
	/*
		cnt := common.NoOfPoolWaitAttempts
		existingPool, _ := GetPoolName()
		isPoolExists := len(existingPool) != 0
	*/
	//TODO check if we need to wait for zrepl
	//common.InitialImportedPoolVol, _ = volumereplica.GetVolumes()
	cnt := common.NoOfPoolWaitAttempts
	isPoolExists := checkIfPoolPresent(PoolName(csp))

	// There is no need of loop here, if the GetPoolName returns poolname with cStorPoolGot.GetUID.
	// It is going to stay forever until zrepl restarts
	for i := 0; !isPoolExists && shouldWait && i < cnt; i++ {
		// GetVolumes is called because, while importing a pool, volumes corresponding
		// to the pool are also imported. This needs to be handled and made visible
		// to cvr controller.
		common.InitialImportedPoolVol, _ = volumereplica.GetVolumes()
		// GetPoolName is to get pool name for particular no. of attempts.
		isPoolExists = checkIfPoolPresent(PoolName(csp))
	}

	if isPoolExists {
		if IsPendingStatus(csp) || IsEmptyStatus(csp) {
			// Pool CR status is init. This means pool deployment was done
			// successfully, but before updating the CR to Online status,
			// the watcher container got restarted.
			glog.Infof("Pool %s is online", PoolName(csp))
			common.SyncResources.IsImported = true
			return string(apis.CStorPoolStatusOnline), true
		}
		glog.Warningf("Pool %v already present", PoolName(csp))
		common.SyncResources.IsImported = true
		return string(apis.CStorPoolStatusErrorDuplicate), true
	}
	return "", false
}
