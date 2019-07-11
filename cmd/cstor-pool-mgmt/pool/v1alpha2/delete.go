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
	api "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha2"
	zfs "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1"
)

// Delete will destroy the pool for given csp.
// It will also perform labelclear for pool disk.
func Delete(csp *api.CStorNPool) error {
	glog.Infof("Destroying a pool {%s}", PoolName(csp))

	// Let's check if pool exists or not
	if poolExist := checkIfPoolPresent(PoolName(csp)); !poolExist {
		return nil
	}

	// First delete a pool
	ret, err := zfs.NewPoolDestroy().
		WithPool(PoolName(csp)).
		Execute()
	if err != nil {
		glog.Errorf("Failed to destroy a pool {%s}.. %s", ret, err.Error())
		return err
	}

	// We successfully deleted the pool.
	// We also need to clear the label for attached disk
	for _, r := range csp.Spec.RaidGroups {
		vlist, err := getPathForBdevList(r.BlockDevices)
		if err != nil {
			glog.Errorf("Failed to fetch vdev path, skipping labelclear.. %s", err.Error())
		}
		for _, v := range vlist {
			if _, err := zfs.NewPoolLabelClear().
				WithForceFully(true).
				WithVdev(v).Execute(); err != nil {
				glog.Errorf("Failed to perform label clear for disk {%s}.. %s", v, err.Error())
			}
		}
	}

	return nil
}
