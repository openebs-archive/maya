package v1alpha2

import (
	"github.com/golang/glog"
	api "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha2"
	zfs "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1"
)

// Create will create the pool for given csp object
func Create(csp *api.CStorNPool) error {
	var err error
	raidGroups := csp.Spec.RaidGroups

	glog.Infof("Creating a pool for %s %s", csp.Name, PoolName(csp))

	// First create a pool
	// TODO, IsWriteCache, IsSpare, IsReadCache should be disable for actual pool?

	// Lets say we need to execute following command
	// -- zpool create newpool mirror v0 v1 mirror v2 v3 log mirror v4 v5
	// Above command we will execute using following steps:
	// 1. zpool create newpool mirror v0 v1
	// 2. zpool add newpool log mirror v4 v5
	// 3. zpool add newpool mirror v2 v3
	for i, r := range raidGroups {
		if !r.IsReadCache && !r.IsSpare && !r.IsWriteCache {
			// we found the main raidgroup. let's create the pool
			err := createPool(csp, &r)
			if err != nil {
				glog.Errorf("Failed to create pool {%s} : %s", PoolName(csp), err.Error())
				return err
			}
			// Remove this raidGroup
			raidGroups = append(raidGroups[:i], raidGroups[i+1:]...)
			break
		}
	}

	// We created the pool
	// Lets update it with extra config, if provided
	for _, r := range raidGroups {
		if e := addRaidGroup(csp, &r); e != nil {
			err = ErrorWrapf(err, "Failed to add raidGroup{%s}.. %s", r.Name, e.Error())
		}
	}

	// TODO, should we delete the pool?
	if err != nil {
		glog.Errorf("Failed to add supporting device to pool {%s} : {%s}", PoolName(csp), err.Error())
	} else {
		// Add entry to imported pool list
		ImportedCStorPools[string(csp.GetUID())] = csp
	}

	// We created the pool successfully
	// Let's set cachefile for this pool, if it is provided in csp object
	if len(csp.Spec.PoolConfig.CacheFile) != 0 && err == nil {
		if _, err := zfs.NewPoolSProperty().
			WithProperty("cachefile", csp.Spec.PoolConfig.CacheFile).
			WithPool(PoolName(csp)).
			Execute(); err != nil {
			//TODO, If cachefile set failed, do we need to delete the pool?
			glog.Errorf("Failed to set cachefile for pool {%s} : %s", PoolName(csp), err.Error())
		}
		err = nil
		glog.Infof("Set cachefile successful for pool {%s}", PoolName(csp))
	}

	return err
}

func createPool(csp *api.CStorNPool, r *api.RaidGroup) error {
	ptype := r.Type
	if len(ptype) == 0 {
		// type is not mentioned in raidGroup,
		// We will use default raidGroupType from poolConfig
		ptype = csp.Spec.PoolConfig.DefaultRaidGroupType
	}

	vlist, err := getPathForCSPBdevList(r.BlockDevices)
	if err != nil {
		glog.Errorf("Failed to get list of disk-path : %s", err.Error())
		return err
	}

	_, err = zfs.NewPoolCreate().
		WithType(ptype).
		WithPool(PoolName(csp)).
		WithVdevList(vlist).
		Execute()
	return err
}
