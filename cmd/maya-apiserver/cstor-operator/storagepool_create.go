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
	nodeselect "github.com/openebs/maya/pkg/algorithm/nodeselect/v1alpha1"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/client/k8s"
	"github.com/openebs/maya/pkg/storagepool"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

type poolCreateConfig struct {
	*nodeselect.Config
}

var poolconfig *poolCreateConfig

// Cas template is a custom resource which has a list of runTasks.

// runTasks are configmaps which has defined yaml templates for resources that needs
// to be created or deleted for a storagepool creation or deletion respectively.

// CreateStoragePool is a function that does following:
// 1. It receives storagepoolclaim object from the spc watcher event handler.
// 2. After successful validation, it will call a worker function for actual storage creation
//    via the cas template specified in storagepoolclaim.
func CreateStoragePool(spcGot *apis.StoragePoolClaim) error {
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
	poolconfig = &poolCreateConfig{
		nodeselect.NewConfig(spcGot),
	}
	pool, err := newClientSet.NewCasPool(spcGot, poolconfig)
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
		return errors.Wrapf(err, "NewCasPoolOperation failed error")
	}
	_, err = storagepoolOps.Create()
	if err != nil {
		return errors.Wrapf(err, "failed to create cas template based storagepool")

	}
	glog.Infof("Cas template based storagepool created successfully: name '%s'", pool.StoragePoolClaim)
	return nil
}

func (newClientSet *clientSet) NewCasPool(spc *apis.StoragePoolClaim, algorithmConfig *poolCreateConfig) (*apis.CasPool, error) {
	// Create a CasPool object and fill it with default values.
	casPool := &v1alpha1.CasPool{}
	casTemplateName := spc.Annotations[string(v1alpha1.CreatePoolCASTemplateKey)]
	casPool.CasCreateTemplate = casTemplateName
	casPool.StoragePoolClaim = spc.Name
	casPool.PoolType = spc.Spec.PoolSpec.PoolType
	// ToDo: Remove MinPools field as it is not being used.
	casPool.MinPools = spc.Spec.MinPools
	casPool.MaxPools = spc.Spec.MaxPools
	casPool.Type = spc.Spec.Type
	casPool.Annotations = spc.Annotations
	// After CasPool object is filled with default values, call casPoolBuilder to fill more specific values.
	casPool, err := newClientSet.casPoolBuilder(casPool, spc, algorithmConfig)
	return casPool, err
}

// casPoolBuilder builds the CasPool object by filling details like diskList,nodeName etc.
// Some of the fields of the CasPool object is passed to CAS engine.
// CasPool object(type) is the contract on which CAS engine is instantiated for cStor pool creation.
func (newClientSet *clientSet) casPoolBuilder(casPool *apis.CasPool, spc *apis.StoragePoolClaim, ac *poolCreateConfig) (*apis.CasPool, error) {
	// getDiskList will hold node and disks attached to it to be used for storagepool provisioning.
	nodeDisks, err := ac.NodeDiskSelector()
	if err != nil {
		return nil, errors.Wrapf(err, "aborting storagepool create operation as no node qualified")
	}
	if len(nodeDisks.Disks.Items) == 0 {
		return nil, errors.New("aborting storagepool create operation as no disk was found")
	}

	// Fill the node name to the CasPool object.
	casPool.NodeName = nodeDisks.NodeName
	// Fill the disks attached to this node to the CasPool object.
	//casPool.DiskList = nodeDisks.Disks.Items
	//TODO: Improve Following Code
	if spc.Spec.PoolSpec.PoolType == string(apis.PoolTypeStripedCPV) {
		for _, disk := range nodeDisks.Disks.Items {
			var diskList []apis.CspDisk
			var group apis.DiskGroup
			disk := apis.CspDisk{
				Name:        disk,
				InUseByPool: true,
			}
			devID, err := newClientSet.getDeviceID(disk.Name)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to get dev is for disk %s for spc %s", disk.Name, spc.Name)
			}
			disk.DeviceID = devID
			diskList = append(diskList, disk)
			group = apis.DiskGroup{
				Item: diskList,
			}
			casPool.DiskList = append(casPool.DiskList, group)
		}
		return casPool, nil
	}
	count := nodeselect.DefaultDiskCount[spc.Spec.PoolSpec.PoolType]
	for i := 0; i <= len(nodeDisks.Disks.Items)/count; i = i + count {
		var diskList []apis.CspDisk
		var group apis.DiskGroup
		for j := 0; j < count; j++ {

			disk := apis.CspDisk{
				Name:        nodeDisks.Disks.Items[i+j],
				InUseByPool: true,
			}
			devID, err := newClientSet.getDeviceID(disk.Name)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to get dev is for disk %s for spc %s", disk.Name, spc.Name)
			}
			disk.DeviceID = devID
			diskList = append(diskList, disk)
		}
		group = apis.DiskGroup{
			Item: diskList,
		}
		casPool.DiskList = append(casPool.DiskList, group)
	}
	return casPool, nil
}

// TODO: Move to disk package
func (newClientSet *clientSet) getDeviceID(diskName string) (string, error) {
	var deviceID string
	disk, err := newClientSet.oecs.OpenebsV1alpha1().Disks().Get(diskName, v1.GetOptions{})
	if err != nil {
		return "", err
	}
	if len(disk.Spec.DevLinks) != 0 && len(disk.Spec.DevLinks[0].Links) != 0 {
		deviceID = disk.Spec.DevLinks[0].Links[0]
	} else {
		deviceID = disk.Spec.Path
	}
	return deviceID, nil
}
