/*
Copyright 2017 The OpenEBS Authors

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

package spc

import (
	"github.com/golang/glog"
	apisv1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/storagepool"
	"github.com/pkg/errors"
)

// Cas template is a custom resource which has a list of runTasks.

// runTasks are custom resource which has defined yaml templates for resources that needs
// to be created for a storagepool creation.

// CreateStoragePool is a function that does following:
// 1. It receives cstorpoolcluster object from the cspc watcher event handler.
// 2. Call GetCasPool method from nodeselect package to get the CasPool object which is an internal representation
//    of a cstor pool.
func (po *PoolOperation) CreateStoragePool(cspcGot *apisv1alpha1.CStorPoolCluster) error {
	newCasPool, err := po.GetCasPool()
	if err != nil {
		return errors.Wrapf(err, "failed to build cas pool for cspc %s", cspcGot.Name)
	}
	// Calling worker function to create storagepool
	err = poolCreateWorker(newCasPool)
	if err != nil {
		return err
	}

	return nil
}

func poolCreateWorker(pool *apisv1alpha1.CasPool) error {

	glog.Infof("Creating storagepool for cstorpoolcluster %s via CASTemplate", pool.CStorPoolCluster)

	storagepoolOps, err := storagepool.NewCasPoolOperation(pool)
	if err != nil {
		return errors.Wrapf(err, "NewCasPoolOperation failed error")
	}
	_, err = storagepoolOps.Create()
	if err != nil {
		return errors.Wrapf(err, "failed to create cas template based storagepool")

	}
	glog.Infof("Cas template based storagepool created successfully: name '%s'", pool.CStorPoolCluster)
	return nil
}
