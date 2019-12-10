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
	"github.com/pkg/errors"
	"k8s.io/klog"
)

// Create will create the pool for given csp object
func Create(csp *apis.CStorPoolInstance) error {
	var err error

	// Let's check if there is any disk having the pool config
	// If so then we will not create the pool
	ret, notImported, err := checkIfPoolIsImportable(csp)
	if err != nil {
		return errors.Errorf("Failed to check not imported pool %s", err.Error())
	}
	if notImported {
		return errors.Errorf("Pool {%s} is in faulty state.. %s", PoolName(csp), ret)
	}

	klog.Infof("Creating a pool for %s %s", csp.Name, PoolName(csp))

	// First create a pool
	// TODO, IsWriteCache, IsSpare, IsReadCache should be disable for actual pool?

	// Lets say we need to execute following command
	// -- zpool create newpool mirror v0 v1 mirror v2 v3 log mirror v4 v5
	// Above command we will execute using following steps:
	// 1. zpool create newpool mirror v0 v1
	// 2. zpool add newpool log mirror v4 v5
	// 3. zpool add newpool mirror v2 v3
	spec := csp.Spec.DeepCopy()
	raidGroups := spec.RaidGroups
	for i, r := range raidGroups {
		if !r.IsReadCache && !r.IsSpare && !r.IsWriteCache {
			// we found the main raidgroup. let's create the pool
			err = createPool(csp, r)
			if err != nil {
				return errors.Errorf("Failed to create pool {%s} : %s",
					PoolName(csp), err.Error())
			}
			// Remove this raidGroup
			raidGroups = append(raidGroups[:i], raidGroups[i+1:]...)
			break
		}
	}

	// We created the pool
	// Lets update it with extra config, if provided
	for _, r := range raidGroups {
		if e := addRaidGroup(csp, r); e != nil {
			err = ErrorWrapf(err, "Failed to add raidGroup{%#v}.. %s", r, e.Error())
		}
	}

	return err
}

func createPool(csp *apis.CStorPoolInstance, r apis.RaidGroup) error {
	var vdevlist []string

	ptype := r.Type
	if len(ptype) == 0 {
		// type is not mentioned in raidGroup,
		// We will use default raidGroupType from poolConfig
		ptype = csp.Spec.PoolConfig.DefaultRaidGroupType
	}

	disklist, err := getPathForBdevList(r.BlockDevices)
	if err != nil {
		return errors.Errorf("Failed to get list of disk-path : %s", err.Error())
	}

	for _, v := range disklist {
		vdevlist = append(vdevlist, v[0])
	}

	ret, err := zfs.NewPoolCreate().
		WithType(ptype).
		WithProperty("cachefile", csp.Spec.PoolConfig.CacheFile).
		WithPool(PoolName(csp)).
		WithVdevList(vdevlist).
		Execute()
	if err != nil {
		return errors.Errorf("Failed to create pool.. %s .. %s", string(ret), err.Error())
	}
	return nil
}
