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

package v1alpha2

import (
	"strings"

	"github.com/golang/glog"
	"github.com/openebs/maya/cmd/cstor-pool-mgmt/controller/common"
	"github.com/openebs/maya/cmd/cstor-pool-mgmt/volumereplica"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	api "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha2"
	zpool "github.com/openebs/maya/pkg/apis/openebs.io/zpool/v1alpha1"
	blockdevice "github.com/openebs/maya/pkg/blockdevice/v1alpha2"
	env "github.com/openebs/maya/pkg/env/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	zfs "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	//	clientset1 "github.com/openebs/maya/pkg/client/generated/clientset/versioned"

	clientset2 "github.com/openebs/maya/pkg/client/generated/openebs.io/v1alpha2/clientset/internalclientset"
	"k8s.io/client-go/kubernetes"
)

// ImportedCStorPools is a map of imported cstor pools API config identified via their UID
var ImportedCStorPools map[string]*api.CStorNPool

//PoolAddEventHandled is a flag representing if the pool has been initially imported or created
var PoolAddEventHandled = false

// RunnerVar the runner variable for executing binaries.
var RunnerVar util.Runner

var KubeClient kubernetes.Interface
var OpenEbsClient2 clientset2.Interface

// PoolPrefix is prefix for pool name
const (
	PoolPrefix string = "cstor-"

	StatusNoPoolsAvailable = "no pools available"
	PoolStatusDegraded     = "DEGRADED"
	PoolStatusFaulted      = "FAULTED"
	PoolStatusOffline      = "OFFLINE"
	PoolStatusOnline       = "ONLINE"
	PoolStatusRemoved      = "REMOVED"
	PoolStatusUnavail      = "UNAVAIL"
)

// PoolName return pool name for given CSP object
func PoolName(csp *api.CStorNPool) string {
	return string(PoolPrefix) + string(csp.ObjectMeta.UID)
}

// Delete will destroy the pool for given csp.
// It will also perform labelclear for pool disk.
func Delete(csp *api.CStorNPool) error {
	glog.Infof("Destroying a pool for %+v", csp)

	// First delete a pool
	ret, err := zfs.NewPoolDestroy().
		WithPool(PoolName(csp)).
		Execute()
	if err != nil {
		glog.Errorf("Failed to destroy a pool : %s : %s", ret, err.Error())
		return err
	}

	// We successfully deleted the pool.
	// We also need to clear the label for attached disk
	for _, r := range csp.Spec.RaidGroups {
		vlist, err := getPathForCSPBdevList(r.BlockDevices)
		if err != nil {
			glog.Errorf("Failed to fetch vdev path, skipping labelclear : %s", err.Error())
		}
		for _, v := range vlist {
			if _, err := zfs.NewPoolLabelClear().
				WithForceFully(true).
				WithVdev(v).Execute(); err != nil {
				glog.Errorf("Failed to perform label clear for disk {%s}", v)
			}
		}
	}

	return nil
}

// Import will import pool for given CSP object.
// It will also set `cachefile` property for that pool
// if it is mentioned in object
func Import(csp *api.CStorNPool) (string, bool, error) {
	ret, err := zfs.NewPoolImport().
		WithCachefile(csp.Spec.PoolConfig.CacheFile).
		Execute()
	if err != nil {
		glog.Errorf("Failed to import pool : %s : %s", ret, err.Error())
		// We return error as nil because pool doesn't exist
		return "", false, nil
	}

	// We imported pool successfully
	// Let's set cachefile for this pool, if it is provided in csp object
	if len(csp.Spec.PoolConfig.CacheFile) != 0 {
		if _, err := zfs.NewPoolSProperty().
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

// Create will create the pool for given csp object
func Create(csp *api.CStorNPool) error {
	var err error
	var poolCreated bool

	glog.Infof("Creating a pool for %+v", csp)

	// First create a pool
	// TODO, IsWriteCache, IsSpare, IsReadCache should be disable for actual pool?

	// Lets say we need to execute following command
	// -- zpool create newpool mirror v0 v1 mirror v2 v3 log mirror v4 v5
	// Above command we will execute using following steps:
	// 1. zpool create newpool mirror v0 v1
	// 2. zpool add newpool log mirror v4 v5
	// 3. zpool add newpool mirror v2 v3
	for _, r := range csp.Spec.RaidGroups {
		if !r.IsReadCache && !r.IsSpare && !r.IsWriteCache {
			if poolCreated {
				// uhh.. We already created the pool..
				// buggy config!
				return errors.New("invalid config")
			}
			// we found the main raidgroup. let's create the pool
			err := createPool(csp, &r)
			if err != nil {
				glog.Errorf("Failed to create pool {%s} : %s", PoolName(csp), err.Error())
				return err
			}
			poolCreated = true
		}
	}

	// We created the pool
	// Lets update it with extra config, if provided
	for _, r := range csp.Spec.RaidGroups {
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
	if len(csp.Spec.PoolConfig.CacheFile) != 0 && err != nil {
		if _, err := zfs.NewPoolSProperty().
			WithProperty("cachefile", csp.Spec.PoolConfig.CacheFile).
			Execute(); err != nil {
			//TODO, If cachefile set failed, do we need to delete the pool?
			glog.Errorf("Failed to set cachefile for pool {%s} : %s", PoolName(csp), err.Error())
		}
		err = nil
		glog.Infof("Set cachefile successful for pool {%s}", PoolName(csp))
	}

	return err
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

func getPathForCSPBdevList(bdevs []api.CStorPoolClusterBlockDevice) ([]string, error) {
	var vdev []string
	var err error

	for _, b := range bdevs {
		glog.Infof("bdev is %+v", b)
		path, er := getPathForBDev(b.BlockDeviceName)
		if er != nil {
			er = ErrorWrapf(err, "Failed to fetch path for bdev {%s} {%s}", b.BlockDeviceName, err.Error())
			continue
		}
		vdev = append(vdev, path)
	}
	return vdev, err
}

func getPathForBDev(bdev string) (string, error) {
	bd, err := blockdevice.NewKubeClient().
		WithNamespace(env.Get(env.OpenEBSNamespace)).
		Get(bdev, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return bd.Spec.Path, nil
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

func addRaidGroup(csp *api.CStorNPool, r *api.RaidGroup) error {
	var err error

	if r.IsReadCache {
		if e := addReadCacheVdev(csp, r); e != nil {
			err = ErrorWrapf(err, "Failed add readcache {%s}", e.Error())
		}
	} else if r.IsSpare {
		if e := addSpareVdev(csp, r); e != nil {
			err = ErrorWrapf(err, "Failed to add spare disk {%s}", e.Error())
		}
	} else if r.IsWriteCache {
		if e := addWriteCacheVdev(csp, r); e != nil {
			err = ErrorWrapf(err, "Failed to add write cache {%s}", e.Error())
		}
	} else {
		if e := addNewVdev(csp, r); e != nil {
			err = ErrorWrapf(err, "Failed to add additional disk {%s}", e.Error())
		}
	}
	return err
}

func addReadCacheVdev(csp *api.CStorNPool, r *api.RaidGroup) error {
	vlist, err := getPathForCSPBdevList(r.BlockDevices)
	if err != nil {
		glog.Errorf("Failed to get list of disk-path : %s", err.Error())
		return err
	}

	_, err = zfs.NewPoolExpansion().
		WithDeviceType("cache").
		WithPool(PoolName(csp)).
		WithVdevList(vlist).
		Execute()
	return err
}

func addSpareVdev(csp *api.CStorNPool, r *api.RaidGroup) error {
	vlist, err := getPathForCSPBdevList(r.BlockDevices)
	if err != nil {
		glog.Errorf("Failed to get list of disk-path : %s", err.Error())
		return err
	}

	_, err = zfs.NewPoolExpansion().
		WithDeviceType("spare").
		WithPool(PoolName(csp)).
		WithVdevList(vlist).
		Execute()
	return err
}

func addWriteCacheVdev(csp *api.CStorNPool, r *api.RaidGroup) error {
	vlist, err := getPathForCSPBdevList(r.BlockDevices)
	if err != nil {
		glog.Errorf("Failed to get list of disk-path : %s", err.Error())
		return err
	}

	_, err = zfs.NewPoolExpansion().
		WithDeviceType("cache").
		WithPool(PoolName(csp)).
		WithVdevList(vlist).
		Execute()
	return err
}

func addNewVdev(csp *api.CStorNPool, r *api.RaidGroup) error {
	ptype := r.Type
	if len(ptype) == 0 {
		return errors.Errorf("No type mentioned in raidGroup for pool {%s}", PoolName(csp))
	}

	vlist, err := getPathForCSPBdevList(r.BlockDevices)
	if err != nil {
		glog.Errorf("Failed to get list of disk-path : %s", err.Error())
		return err
	}

	_, err = zfs.NewPoolExpansion().
		WithType(ptype).
		WithPool(PoolName(csp)).
		WithVdevList(vlist).
		Execute()
	return err
}

func checkIfPoolPresent(name string) bool {
	if _, err := zfs.NewPoolGProperty().
		WithParsableMode(true).
		WithScriptedMode(true).
		WithField("name").
		WithProperty("name").
		WithPool(name).
		Execute(); err != nil {
		return false
	}
	return true
}

// GetStatus return status of the pool
func GetStatus(csp *api.CStorNPool) (string, error) {
	ret, err := zfs.NewPoolStatus().
		WithPool(PoolName(csp)).
		Execute()
	if err != nil {
		return "", err
	}

	switch parsePoolStatus(string(ret)) {
	case PoolStatusDegraded:
		return string(apis.CStorPoolStatusDegraded), nil
	case PoolStatusFaulted:
		return string(apis.CStorPoolStatusOffline), nil
	case PoolStatusOffline:
		return string(apis.CStorPoolStatusOffline), nil
	case PoolStatusOnline:
		return string(apis.CStorPoolStatusOnline), nil
	case PoolStatusRemoved:
		return string(apis.CStorPoolStatusDegraded), nil
	case PoolStatusUnavail:
		return string(apis.CStorPoolStatusOffline), nil
	default:
		return string(apis.CStorPoolStatusError), nil
	}
}

// parsePoolStatus parse output of `zpool status` command to extract the status of the pool.
// ToDo: Need to find some better way e.g contract for zpool command outputs.
func parsePoolStatus(output string) string {
	var outputStr []string
	var poolStatus string
	if strings.TrimSpace(string(output)) != "" {
		outputStr = strings.Split(string(output), "\n")
		if !(len(outputStr) < 2) {
			poolStatusArr := strings.Split(outputStr[1], ":")
			if !(len(outputStr) < 2) {
				poolStatus = strings.TrimSpace(poolStatusArr[1])
			}
		}
	}
	return poolStatus
}

// Update will update the deployed pool according to given csp object
func Update(csp *api.CStorNPool) error {
	var err error
	var isObjChanged bool

	// first we will check if there any bdev is replaced or removed
	for raidIndex, raidGroup := range csp.Spec.RaidGroups {
		for bdevIndex, bdev := range raidGroup.BlockDevices {
			glog.Infof("got bdev %s", bdev.BlockDeviceName)

			// Let's check if bdev name is empty
			// if yes then remove relevant disk from pool
			if IsEmpty(bdev.BlockDeviceName) {
				// block device name is empty
				// Let's remove it
				// TODO should we offline it only?
				if er := removePoolVdev(csp, bdev); er != nil {
					err = ErrorWrapf(err, "Failed to remove bdev {%s}.. %s", bdev.DevLink, er.Error())
				}
			}

			// Let's check if bdev path is changed or not
			newpath, isChanged, er := isBdevPathChanged(csp, &bdev)
			if er != nil {
				err = ErrorWrapf(err, "Failed to check bdev change {%s}.. %s", bdev.BlockDeviceName, er.Error())
			} else if isChanged {
				if er := replacePoolVdev(csp, bdev, newpath); err != nil {
					err = ErrorWrapf(err, "Failed to replace bdev for {%s}.. %s", bdev.BlockDeviceName, er.Error())
				} else {
					// Let's update devLink with new path for this bdev
					csp.Spec.RaidGroups[raidIndex].BlockDevices[bdevIndex].DevLink = newpath
					isObjChanged = true
				}
			}
		}
	}

	if er := addNewVdevFromCSP(csp); er != nil {
		err = ErrorWrapf(err, "Failed to execute add operation.. %s", er.Error())
	}

	if isObjChanged {
		if _, er := OpenEbsClient2.
			OpenebsV1alpha2().
			CStorNPools(csp.Namespace).
			Update(csp); er != nil {
			err = ErrorWrapf(err, "Failed to update object.. err {%s}", er.Error())
		}
	}
	return err
}

func removePoolVdev(csp *api.CStorNPool, bdev api.CStorPoolClusterBlockDevice) error {
	_, err := zfs.NewPoolRemove().
		WithDevice(bdev.DevLink).
		WithPool(PoolName(csp)).
		Execute()
	return err
}

func replacePoolVdev(csp *api.CStorNPool, bdev api.CStorPoolClusterBlockDevice, npath string) error {
	if IsEmpty(npath) || IsEmpty(bdev.DevLink) {
		return errors.Errorf("Empty path for bdev {%s}", bdev.BlockDeviceName)
	}

	_, err := zfs.NewPoolDiskReplace().
		WithOldVdev(bdev.DevLink).
		WithNewVdev(npath).
		WithPool(PoolName(csp)).
		WithForcefully(true).Execute()
	return err
}

func isBdevPathChanged(csp *api.CStorNPool, bdev *api.CStorPoolClusterBlockDevice) (string, bool, error) {
	var err error
	var isPathChanged bool

	newPath, er := getPathForBDev(bdev.BlockDeviceName)
	if er != nil {
		err = errors.Errorf("Failed to get bdev {%s} path err {%s}", bdev.BlockDeviceName, err.Error())
	}

	if err != nil && newPath != bdev.DevLink {
		isPathChanged = true
	}
	return newPath, isPathChanged, err
}

func addNewVdevFromCSP(csp *api.CStorNPool) error {
	var err error

	poolTopology, err := zfs.NewPoolDump().WithPool(PoolName(csp)).Execute()
	if err != nil {
		return errors.Errorf("Failed to fetch pool topology.. %s", err.Error())
	}

	for _, raidGroup := range csp.Spec.RaidGroups {
		wholeGroup := true
		var devlist []string

		for _, bdev := range raidGroup.BlockDevices {
			if isUsed := checkIfDeviceUsed(bdev.DevLink, poolTopology); !isUsed {
				devlist = append(devlist, bdev.DevLink)
			} else {
				wholeGroup = false
			}
		}
		if wholeGroup {
			if er := addRaidGroup(csp, &raidGroup); er != nil {
				err = ErrorWrapf(err, "Failed to add raidGroup{%s}.. %s", raidGroup.Name, er.Error())
			}
		} else if len(devlist) != 0 {
			if _, er := zfs.NewPoolExpansion().
				WithVdevList(devlist).
				WithPool(PoolName(csp)).
				Execute(); er != nil {
				err = ErrorWrapf(err, "Failed to add devlist %v.. err {%s}", devlist, err.Error())
			}
		}
	}

	return err
}

func compareDisk(path string, d []zpool.Vdev) bool {
	for _, v := range d {
		if path == v.Path {
			return true
		}
		for _, p := range v.Children {
			if path == p.Path {
				return true
			}
			if r := compareDisk(path, p.Children); r {
				return true
			}
		}
	}
	return false
}

func checkIfDeviceUsed(path string, t zpool.Topology) bool {
	var isUsed bool

	if isUsed = compareDisk(path, t.VdevTree.Topvdev); isUsed {
		return isUsed
	}

	if isUsed = compareDisk(path, t.VdevTree.Spares); isUsed {
		return isUsed
	}

	if isUsed = compareDisk(path, t.VdevTree.Readcache); isUsed {
		return isUsed
	}
	return isUsed
}
