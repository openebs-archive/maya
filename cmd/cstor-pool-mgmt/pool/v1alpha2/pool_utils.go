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
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	zpool "github.com/openebs/maya/pkg/apis/openebs.io/zpool/v1alpha1"
	blockdevice "github.com/openebs/maya/pkg/blockdevice/v1alpha2"
	env "github.com/openebs/maya/pkg/env/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	zfs "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		WithNamespace(env.Get("NAMESPACE")).
		Get(bdev, metav1.GetOptions{})
	if err != nil {
		return path, err
	}

	if len(bd.Spec.DevLinks) != 0 {
		for _, v := range bd.Spec.DevLinks {
			path = append(path, v.Links...)
		}
	}

	if len(bd.Spec.Path) != 0 {
		path = append(path, bd.Spec.Path)

	}

	return path, nil
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

func checkIfPoolNotImported(cspi *apis.CStorPoolInstance) (string, bool, error) {
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
