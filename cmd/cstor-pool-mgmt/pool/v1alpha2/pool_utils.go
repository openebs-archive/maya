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

	api "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha2"
	zpool "github.com/openebs/maya/pkg/apis/openebs.io/zpool/v1alpha1"
	blockdevice "github.com/openebs/maya/pkg/blockdevice/v1alpha2"
	env "github.com/openebs/maya/pkg/env/v1alpha1"
	zfs "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// TODO
	// CSPC should set ENV variable for following constant

	// SparseDir sparse file location
	SparseDir = "/var/openebs/sparse"
	// DevDir /dev/ path
	DevDir = "/dev"
)

func getPathForBdevList(bdevs []api.CStorPoolClusterBlockDevice) ([]string, error) {
	var vdev []string
	var err error

	for _, b := range bdevs {
		path, er := getPathForBDev(b.BlockDeviceName)
		if er != nil {
			err = ErrorWrapf(err, "Failed to fetch path for bdev {%s} {%s}", b.BlockDeviceName, er.Error())
			continue
		}
		vdev = append(vdev, path)
	}
	return vdev, err
}

func getPathForBDev(bdev string) (string, error) {
	// TODO
	// replace `NAMESPACE` with env variable from CSP deployment
	bd, err := blockdevice.NewKubeClient().
		WithNamespace(env.Get("NAMESPACE")).
		Get(bdev, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	if len(bd.Spec.DevLinks) != 0 && len(bd.Spec.DevLinks[0].Links) != 0 {
		return bd.Spec.DevLinks[0].Links[0], nil
	}
	return bd.Spec.Path, nil
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

func isBdevPathChanged(bdev api.CStorPoolClusterBlockDevice) (string, bool, error) {
	var err error
	var isPathChanged bool

	newPath, er := getPathForBDev(bdev.BlockDeviceName)
	if er != nil {
		err = errors.Errorf("Failed to get bdev {%s} path err {%s}", bdev.BlockDeviceName, er.Error())
	}

	if err == nil && newPath != bdev.DevLink {
		isPathChanged = true
	}
	return newPath, isPathChanged, err
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

func checkIfPoolNotImported(csp *api.CStorNPool) (string, bool, error) {
	ret, err := zfs.NewPoolImport().
		WithDirectory(SparseDir).
		WithDirectory(DevDir).
		Execute()
	if err != nil {
		return string(ret), false, err
	}
	if strings.Contains(string(ret), PoolName(csp)) {
		return string(ret), true, nil
	}
	return string(ret), false, nil
}

// getDeviceType will return type of device from raidGroup
// It can be either log/cache/stripe(/"")
func getDeviceType(r api.RaidGroup) string {
	if r.IsReadCache {
		return DeviceTypeReadCache
	} else if r.IsSpare {
		return DeviceTypeSpare
	} else if r.IsWriteCache {
		return DeviceTypeWriteCache
	} else {
		return DeviceTypeEmpty
	}
}
