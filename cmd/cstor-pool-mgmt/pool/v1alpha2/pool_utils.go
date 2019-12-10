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
	"strings"

	pool "github.com/openebs/maya/cmd/cstor-pool-mgmt/pool"
	ndmapis "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	zpool "github.com/openebs/maya/pkg/apis/openebs.io/zpool/v1alpha1"
	blockdevice "github.com/openebs/maya/pkg/blockdevice/v1alpha2"
	blockdeviceclaim "github.com/openebs/maya/pkg/blockdeviceclaim/v1alpha1"
	apiscspc "github.com/openebs/maya/pkg/cstor/poolcluster/v1alpha1"
	env "github.com/openebs/maya/pkg/env/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	zfs "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

func getPathForBdevList(bdevs []apis.CStorPoolClusterBlockDevice) (map[string][]string, error) {
	var err error

	vdev := make(map[string][]string, len(bdevs))
	for _, b := range bdevs {
		path, er := getPathForBDev(b.BlockDeviceName)
		if er != nil || len(path) == 0 {
			err = ErrorWrapf(err, "Failed to fetch path for bdev {%s} {%s}", b.BlockDeviceName, er.Error())
			continue
		}
		vdev[b.BlockDeviceName] = path
	}
	return vdev, err
}

func getPathForBDev(bdev string) ([]string, error) {
	var path []string

	// TODO
	// replace `NAMESPACE` with env variable from CSP deployment
	bd, err := blockdevice.NewKubeClient().
		WithNamespace(env.Get(env.Namespace)).
		Get(bdev, metav1.GetOptions{})
	if err != nil {
		return path, err
	}
	return getPathForBDevFromBlockDevice(bd), nil
}

func getPathForBDevFromBlockDevice(bd *ndmapis.BlockDevice) []string {
	var paths []string
	if len(bd.Spec.DevLinks) != 0 {
		for _, v := range bd.Spec.DevLinks {
			paths = append(paths, v.Links...)
		}
	}

	if len(bd.Spec.Path) != 0 {
		paths = append(paths, bd.Spec.Path)
	}
	return paths
}

func checkIfPoolPresent(name string) bool {
	if _, err := zfs.NewPoolGetProperty().
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

/*
func isBdevPathChanged(bdev apis.CStorPoolClusterBlockDevice) ([]string, bool, error) {
	var err error
	var isPathChanged bool

	newPath, er := getPathForBDev(bdev.BlockDeviceName)
	if er != nil {
		err = errors.Errorf("Failed to get bdev {%s} path err {%s}", bdev.BlockDeviceName, er.Error())
	}

	if err == nil && !util.ContainsString(newPath, bdev.DevLink) {
		isPathChanged = true
	}

	return newPath, isPathChanged, err
}
*/

func compareDisk(path []string, d []zpool.Vdev) (string, bool) {
	for _, v := range d {
		if util.ContainsString(path, v.Path) {
			return v.Path, true
		}
		for _, p := range v.Children {
			if util.ContainsString(path, p.Path) {
				return p.Path, true
			}
			if path, r := compareDisk(path, p.Children); r {
				return path, true
			}
		}
	}
	return "", false
}

func checkIfDeviceUsed(path []string, t zpool.Topology) (string, bool) {
	var isUsed bool
	var usedPath string

	if usedPath, isUsed = compareDisk(path, t.VdevTree.Topvdev); isUsed {
		return usedPath, isUsed
	}

	if usedPath, isUsed = compareDisk(path, t.VdevTree.Spares); isUsed {
		return usedPath, isUsed
	}

	if usedPath, isUsed = compareDisk(path, t.VdevTree.Readcache); isUsed {
		return usedPath, isUsed
	}
	return usedPath, isUsed
}

// checkIfPoolIsImportable checks if the pool is imported or not. If the pool
// is present on the disk but  not imported it returns true as the pool can be
// imported. It also returns false if pool is not found on the disk.
func checkIfPoolIsImportable(cspi *apis.CStorPoolInstance) (string, bool, error) {
	var cmdOut []byte
	var err error

	bdPath, err := getPathForBDev(cspi.Spec.RaidGroups[0].BlockDevices[0].BlockDeviceName)
	if err != nil {
		return "", false, err
	}

	devID := pool.GetDevPathIfNotSlashDev(bdPath[0])
	if len(devID) != 0 {
		cmdOut, err = zfs.NewPoolImport().WithDirectory(devID).Execute()
		if strings.Contains(string(cmdOut), PoolName(cspi)) {
			return string(cmdOut), true, nil
		}
	}
	// there are some cases when import is succesful but zpool command return
	// noisy errors, hence better to check contains before return error
	cmdOut, err = zfs.NewPoolImport().Execute()
	if strings.Contains(string(cmdOut), PoolName(cspi)) {
		return string(cmdOut), true, nil
	}
	return string(cmdOut), false, err
}

// getDeviceType will return type of device from raidGroup
// It can be either log/cache/stripe(/"")
func getDeviceType(r apis.RaidGroup) string {
	if r.IsReadCache {
		return DeviceTypeReadCache
	} else if r.IsSpare {
		return DeviceTypeSpare
	} else if r.IsWriteCache {
		return DeviceTypeWriteCache
	}
	return DeviceTypeEmpty
}

// getBlockDeviceClaimList returns list of block device claims based on the
// label passed to the function
func getBlockDeviceClaimList(key, value string) (
	*blockdeviceclaim.BlockDeviceClaimList, error) {
	namespace := env.Get(env.Namespace)
	bdcClient := blockdeviceclaim.NewKubeClient().
		WithNamespace(namespace)
	bdcAPIList, err := bdcClient.List(metav1.ListOptions{
		LabelSelector: key + "=" + value,
	})
	if err != nil {
		return nil, errors.Wrapf(err,
			"failed to list bdc related to key: %s value: %s",
			key,
			value,
		)
	}
	return &blockdeviceclaim.BlockDeviceClaimList{ObjectList: bdcAPIList}, nil
}

func executeZpoolDump(cspi *apis.CStorPoolInstance) (zpool.Topology, error) {
	return zfs.NewPoolDump().
		WithPool(PoolName(cspi)).
		WithStripVdevPath().
		Execute()
}

// isResilveringInProgress returns true if resilvering is inprogress at cstor
// pool
func isResilveringInProgress(
	executeCommand func(cspi *apis.CStorPoolInstance) (zpool.Topology, error),
	cspi *apis.CStorPoolInstance,
	path string) bool {
	poolTopology, err := executeCommand(cspi)
	if err != nil {
		// log error
		klog.Errorf("Failed to get pool topology error: %v", err)
		return true
	}
	vdev, isVdevExist := getVdevFromPath(path, poolTopology)
	if !isVdevExist {
		return true
	}
	// If device in raid group didn't got replaced then there won't be any info
	// related to scan stats
	if len(vdev.ScanStats) == 0 {
		return false
	}
	// If device didn't underwent resilvering then no.of scaned bytes will be
	// zero
	if vdev.VdevStats[zpool.VdevScanProcessedIndex] == 0 {
		return false
	}
	// To decide whether resilvering is completed then check following steps
	// 1. Current device should be child device.
	// 2. Device Scan State should be completed
	if len(vdev.Children) == 0 &&
		vdev.ScanStats[zpool.VdevScanStatsStateIndex] == uint64(zpool.PoolScanFinished) &&
		vdev.ScanStats[zpool.VdevScanStatsScanFuncIndex] == uint64(zpool.PoolScanFuncResilver) {
		return false
	}
	return true
}

func getVdevFromPath(path string, topology zpool.Topology) (zpool.Vdev, bool) {
	var vdev zpool.Vdev
	var isVdevExist bool

	if vdev, isVdevExist = zpool.
		VdevList(topology.VdevTree.Topvdev).
		GetVdevFromPath(path); isVdevExist {
		return vdev, isVdevExist
	}

	if vdev, isVdevExist = zpool.
		VdevList(topology.VdevTree.Spares).
		GetVdevFromPath(path); isVdevExist {
		return vdev, isVdevExist
	}

	if vdev, isVdevExist = zpool.
		VdevList(topology.VdevTree.Readcache).
		GetVdevFromPath(path); isVdevExist {
		return vdev, isVdevExist
	}
	return vdev, isVdevExist
}

//cleanUpReplacementMarks should be called only after resilvering is completed.
//It does the following work
// 1. RemoveFinalizer on old block device claim exists and delete the old block
//   device claim.
// 2. Remove link of old block device in new block device claim
// oldObj is block device claim of replaced block device object which is
// detached from pool
// newObj is block device claim of current block device object which is in use
// by pool
func cleanUpReplacementMarks(oldObj, newObj *ndmapis.BlockDeviceClaim) error {
	bdcClient := blockdeviceclaim.NewKubeClient().WithNamespace(newObj.Namespace)
	if oldObj != nil {
		updatedOldObj, err := blockdeviceclaim.
			BuilderForAPIObject(oldObj).BDC.RemoveFinalizer(apiscspc.CSPCFinalizer)
		if err != nil {
			return errors.Wrapf(err,
				"failed to remove finalizer on blockdeviceclaim {%s}",
				oldObj.Name,
			)
		}
		err = bdcClient.Delete(updatedOldObj.Name, &metav1.DeleteOptions{})
		if err != nil {
			return errors.Wrapf(
				err,
				"Failed to unclaim old blockdevice {%s}",
				oldObj.Spec.BlockDeviceName,
			)
		}
	}
	bdAnnotations := newObj.GetAnnotations()
	delete(bdAnnotations, string(apis.PredecessorBlockDeviceCPK))
	newObj.SetAnnotations(bdAnnotations)
	_, err := bdcClient.Update(newObj)
	if err != nil {
		return errors.Wrapf(
			err,
			"Failed to remove annotation {%s} from blockdeviceclaim {%s}",
			string(apis.PredecessorBlockDeviceCPK),
			newObj.Name,
		)
	}
	return nil
}
