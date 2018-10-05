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
	"errors"
	"fmt"
	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/client/k8s"
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

func (c *Controller) CreateStoragePool(spcGot *apis.StoragePoolClaim, reSync bool, pendingPoolCount int) error {

	if reSync {
		glog.Infof("Storagepool resync event received for storagepoolclaim %s", spcGot.ObjectMeta.Name)
	} else {
		glog.Infof("Storagepool create event received for storagepoolclaim %s", spcGot.ObjectMeta.Name)
	}

	// Check wether the spc object has been processed for storagepool creation
	if spcGot.Status.Phase == onlineStatus && !reSync {
		glog.Infof("Storagepool already exists since the status on storagepoolclaim object %s is Online", spcGot.Name)
		return nil
	}
	var newSpcLease Leases
	newSpcLease = &Lease{spcGot, SpcLeaseKey, c.clientset, c.kubeclientset}
	err := newSpcLease.Hold()
	if err != nil {
		glog.Errorf("Could not acquire lease on spc object:%v", err)
		return err
	}
	glog.Info("Lease acquired successfully on storagepoolclaim %s ", spcGot.Name)

	defer newSpcLease.Release()

	// Get kubernetes clientset
	// namespaces is not required, hence passed empty.
	newK8sClient, err := k8s.NewK8sClient("")

	if err != nil {
		return err
	}
	// Get openebs clientset using a getter method (i.e. GetOECS() ) as
	// the openebs clientset is not exported.
	newOecsClient := newK8sClient.GetOECS()

	// Create instance of clientset struct defined above which binds
	// ListDisk method and fill it with openebs clienset (i.e.newOecsClient ).
	newClientSet := clientSet{
		oecs: newOecsClient,
	}
	// Get a CasPool object
	err, pool := newClientSet.newCasPool(spcGot, reSync, pendingPoolCount)
	if err != nil {
		return err
	}

	// Calling worker function to create storagepool
	err = poolCreateWorker(pool)
	if err != nil {
		return err
	}

	return nil
}

// poolCreateWorker is a worker function which will create a storagepool

func poolCreateWorker(pool *apis.CasPool) error {

	glog.Infof("Creating storagepool for storagepoolclaim %s via CASTemplate", pool.StoragePoolClaim)

	storagepoolOps, err := storagepool.NewCasPoolOperation(pool)
	if err != nil {
		return fmt.Errorf("NewCasPoolOperation failed error '%s'", err.Error())

	}
	_, err = storagepoolOps.Create()
	if err != nil {
		return fmt.Errorf("failed to create cas template based storagepool: error '%s'", err.Error())

	}

	glog.Infof("Cas template based storagepool created successfully: name '%s'", pool.StoragePoolClaim)
	return nil
}

// newCasPool will return a CasPool object
func (newClientSet *clientSet) newCasPool(spcGot *apis.StoragePoolClaim, reSync bool, pendingPoolCount int) (error, *apis.CasPool) {
	// Validations for poolType
	poolType := spcGot.Spec.PoolSpec.PoolType
	if poolType == "" {
		return errors.New("aborting storagepool create operation as no poolType is specified"), nil
	}

	if !(poolType == string(v1alpha1.PoolTypeStripedCPV) || poolType == string(v1alpha1.PoolTypeMirroredCPV)) {
		return fmt.Errorf("aborting storagepool create operation as specified poolType is %s which is invalid", poolType), nil
	}

	diskType := spcGot.Spec.Type
	if !(diskType == string(v1alpha1.TypeSparseCPV) || diskType == string(v1alpha1.TypeDiskCPV)) {
		return fmt.Errorf("aborting storagepool create operation as specified type is %s which is invalid", diskType), nil
	}
	// The name of cas template should be provided as annotation in storagepoolclaim yaml
	// so that it can be used.

	// Fill spc annotations to CasPool
	casTemplateName := spcGot.Annotations[string(v1alpha1.SPCreateCASTemplateCK)]

	pool := &v1alpha1.CasPool{}
	pool.StoragePoolClaim = spcGot.Name
	pool.CasCreateTemplate = casTemplateName
	pool.PoolType = spcGot.Spec.PoolSpec.PoolType
	pool.MinPools = spcGot.Spec.MinPools
	pool.MaxPools = spcGot.Spec.MaxPools
	pool.Type = spcGot.Spec.Type
	pool.ReSync = reSync
	pool.PendingPoolCount = pendingPoolCount
	pool.Annotations = spcGot.Annotations

	// Fill the object with the disks list
	pool.DiskList = spcGot.Spec.Disks.DiskList
	// Check for disks
	diskList := spcGot.Spec.Disks.DiskList
	// If no disk are specified pool will be provisioned dynamically
	if len(diskList) == 0 {
		// newDisksList is the list of disks over which pool will be provisioned
		err, newDisksList := newClientSet.getCasPoolDisk(pool)
		if err != nil {
			return err, nil
		}
		// Fill the object with the new disks list
		pool.DiskList = newDisksList
	}
	return nil, pool
}

// getCasPoolDisk is a wrapper that will call getDiskList function to get the disk lists
// that will be used to provision a storagepool dynamically

func (newClientSet *clientSet) getCasPoolDisk(cp *apis.CasPool) (error, []string) {
	// Performing valdations against CasPool fields
	if cp.MaxPools <= 0 {
		return fmt.Errorf("aborting storagepool create operation as no maxPool field is specified"), nil
	}
	// if no minimum pools were specified it will default to 1.
	if cp.MinPools <= 0 {
		glog.Warning("invalid or 0 min pool specified, defaulting to 1")
		cp.MinPools = 1
	}
	if cp.MaxPools < cp.MinPools {
		return fmt.Errorf("aborting storagepool create operation as maxPool cannot be less than minPool"), nil
	}
	// If it is a resync event, MaxPool is the pending pool to be provisioned
	if cp.ReSync {
		// if min pool was not provisioned try to provision again the minimum number of pool
		// else set min pool to 1 as in this case min pool was provisioned.
		if !(cp.MaxPools == cp.PendingPoolCount) {
			cp.MinPools = 1
		}
		cp.MaxPools = cp.PendingPoolCount
	}
	// getDiskList will get the disks to be used for storagepool provisioning
	newDisksList, err := newClientSet.nodeDiskAlloter(cp)

	if err != nil {
		return fmt.Errorf("aborting storagepool create operation as no node qualified: %v", err), nil
	}

	if len(newDisksList) == 0 {
		return fmt.Errorf("aborting storagepool create operation as no disk was found"), nil
	}
	return nil, newDisksList
}
