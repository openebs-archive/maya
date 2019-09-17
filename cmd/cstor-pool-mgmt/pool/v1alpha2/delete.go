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
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	zfs "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1"
	"k8s.io/klog"
)

// Delete will destroy the pool for given cspi.
// It will also perform labelclear for pool disk.
func Delete(cspi *apis.CStorPoolInstance) error {
	klog.Infof("Destroying a pool {%s}", PoolName(cspi))

	// Let's check if pool exists or not
	if poolExist := checkIfPoolPresent(PoolName(cspi)); !poolExist {
		klog.Infof("Pool %s not imported.. so, can't destroy", PoolName(cspi))
		return nil
	}

	// First delete a pool
	ret, err := zfs.NewPoolDestroy().
		WithPool(PoolName(cspi)).
		Execute()
	if err != nil {
		klog.Errorf("Failed to destroy a pool {%s}.. %s", ret, err.Error())
		return err
	}

	// We successfully deleted the pool.
	// We also need to clear the label for attached disk
	for _, r := range cspi.Spec.RaidGroups {
		disklist, err := getPathForBdevList(r.BlockDevices)
		if err != nil {
			klog.Errorf("Failed to fetch vdev path, skipping labelclear.. %s", err.Error())
		}
		for _, v := range disklist {
			if _, err := zfs.NewPoolLabelClear().
				WithForceFully(true).
				WithVdev(v[0]).Execute(); err != nil {
				klog.Errorf("Failed to perform label clear for disk {%s}.. %s", v, err.Error())
			}
		}
	}

	return nil
}
