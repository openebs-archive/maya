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

package storagepoolactions

import (
	"errors"
	"fmt"
	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/storagepool"
)

const (
	onlineStatus = "Online"
)

// Cas template is a custom resource which has a list of runTasks.

// runTasks are configmaps which has defined yaml templates for resources that needs
// to be created or deleted for a storagepool creation or deletion respectively.

// CreateStoragePool is a function that does following:
// 1. It receives storagepoolclaim object from the spc watcher event handler.

// 2. After successful validation, it will call a worker function for actual storage creation
//    via the cas template specified in storagepoolclaim.

func CreateStoragePool(spcGot *apis.StoragePoolClaim) error {

	glog.Infof("Storagepool create event received for storagepoolclaim %s", spcGot.ObjectMeta.Name)

	// Check wether the spc object has been processed for storagepool creation
	if spcGot.Status.Phase == onlineStatus {
		return errors.New("Storagepool already exists since the status on storagepoolclaim object is Online")
	}

	// Check for poolType
	poolType := spcGot.Spec.PoolSpec.PoolType
	if poolType == "" {
		return errors.New("Aborting storagepool create operation as no poolType is specified")
	}

	// Check for disks
	diskList := spcGot.Spec.Disks.DiskList
	if len(diskList) == 0 {
		return errors.New("Aborting storagepool create operation as no disk is specified")
	}

	// The name of cas template should be provided as annotation in storagepoolclaim yaml
	// so that it can be used.

	// Check for cas template
	casTemplateName := spcGot.Annotations[string(v1alpha1.SPCreateCASTemplateCK)]
	if casTemplateName == "" {
		return errors.New("Aborting storagepool create operation as no cas template is specified")
	}

	// Calling worker function to create storagepool
	err := poolCreateWorker(spcGot, casTemplateName)

	if err != nil {
		return err
	}

	return nil
}

// poolCreateWorker is a worker function which will create a storagepool

func poolCreateWorker(spcGot *apis.StoragePoolClaim, casTemplateName string) error {

	glog.Infof("Creating storagepool for storagepoolclaim %s via CASTemplate", spcGot.ObjectMeta.Name)
	// Create an empty CasPool object
	pool := &v1alpha1.CasPool{}
	// Fill the object with storagepoolclaim object name
	pool.StoragePoolClaim = spcGot.Name
	// Fill the object with casTemplateName
	pool.CasCreateTemplate = casTemplateName

	storagepoolOps, err := storagepool.NewCasPoolOperation(pool)
	if err != nil {
		return fmt.Errorf("NewCasPoolOperation failed error '%s'", err.Error())

	}
	_, err = storagepoolOps.Create()
	if err != nil {
		return fmt.Errorf("Failed to create cas template based storagepool: error '%s'", err.Error())

	}

	glog.Infof("Cas template based storagepool created successfully: name '%s'", spcGot.Name)
	return nil
}
