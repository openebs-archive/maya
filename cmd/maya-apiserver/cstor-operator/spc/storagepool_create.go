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
	blockdevice "github.com/openebs/maya/pkg/blockdevice/v1alpha1"
	"github.com/openebs/maya/pkg/storagepool"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PoolCreateConfig is config object used to create a cstor pool.
type PoolCreateConfig struct {
	*nodeselect.Config
	*Controller
}

// CasPoolBuilder is the builder object for cas pool.
type CasPoolBuilder struct {
	CasPool *apis.CasPool
}

// Cas template is a custom resource which has a list of runTasks.

// runTasks are configmaps which has defined yaml templates for resources that needs
// to be created or deleted for a storagepool creation or deletion respectively.

// CreateStoragePool is a function that does following:
// 1. It receives storagepoolclaim object from the spc watcher event handler.
// 2. After successful validation, it will call a worker function for actual storage creation
//    via the cas template specified in storagepoolclaim.
func (c *Controller) CreateStoragePool(spcGot *apis.StoragePoolClaim) error {
	poolconfig := c.NewPoolCreateConfig(spcGot)
	newCasPool, err := poolconfig.getCasPool(spcGot)

	if err != nil {
		return errors.Wrapf(err, "failed to build cas pool for spc %s", spcGot.Name)
	}

	// Calling worker function to create storagepool
	err = poolCreateWorker(newCasPool)
	if err != nil {
		return err
	}

	return nil
}

// getCasPool returns a configured cas pool object.
func (pc *PoolCreateConfig) getCasPool(spc *apis.StoragePoolClaim) (*apis.CasPool, error) {
	casPool := NewCasPoolBuilder().
		withSpcName(spc.Name).
		withCasTemplateName(spc.Annotations[string(v1alpha1.CreatePoolCASTemplateKey)]).
		withDiskType(spc.Spec.Type).
		withPoolType(spc.Spec.PoolSpec.PoolType).
		withAnnotations(spc.Annotations).
		withMaxPool(spc).
		Build()
	casPoolWithDisks, err := pc.withDisks(casPool, spc)
	if err != nil {
		return nil, err
	}
	return casPoolWithDisks, nil
}

// NewPoolCreateConfig returns an instance of pool create config.
func (c *Controller) NewPoolCreateConfig(spc *apis.StoragePoolClaim) *PoolCreateConfig {
	poolconfig := &PoolCreateConfig{
		nodeselect.NewConfig(spc),
		c,
	}
	return poolconfig
}

// NewCasPoolBuilder returns an empty instance of CasPoolBuilder.
func NewCasPoolBuilder() *CasPoolBuilder {
	return &CasPoolBuilder{
		CasPool: &apis.CasPool{},
	}
}

func (cb *CasPoolBuilder) withCasTemplateName(casTemplateName string) *CasPoolBuilder {
	//casTemplateName := spc.Annotations[string(v1alpha1.CreatePoolCASTemplateKey)]
	cb.CasPool.CasCreateTemplate = casTemplateName
	return cb
}

func (cb *CasPoolBuilder) withSpcName(name string) *CasPoolBuilder {
	cb.CasPool.StoragePoolClaim = name
	return cb
}

func (cb *CasPoolBuilder) withPoolType(poolType string) *CasPoolBuilder {
	cb.CasPool.PoolType = poolType
	return cb
}

func (cb *CasPoolBuilder) withMaxPool(spc *apis.StoragePoolClaim) *CasPoolBuilder {
	if isAutoProvisioning(spc) {
		cb.CasPool.MaxPools = *spc.Spec.MaxPools
	}
	return cb
}

func (cb *CasPoolBuilder) withDiskType(diskType string) *CasPoolBuilder {
	cb.CasPool.Type = diskType
	return cb
}

func (cb *CasPoolBuilder) withAnnotations(annotations map[string]string) *CasPoolBuilder {
	cb.CasPool.Annotations = annotations
	return cb
}

// Build returns an instance of cas pool object.
func (cb *CasPoolBuilder) Build() *apis.CasPool {
	return cb.CasPool
}

// poolCreateWorker is a worker function which will create a storagepool.
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

// withBlockDevices builds the CasPool object by filling details like
// blockDeviceList,nodeName etc.
// Some of the fields of the CasPool object is passed to CAS engine.
// CasPool object(type) is the contract on which CAS engine is instantiated for cStor pool creation.
func (pc *PoolCreateConfig) withDisks(casPool *apis.CasPool, spc *apis.StoragePoolClaim) (*apis.CasPool, error) {
	// getDiskList will hold node and block devices attached to it to be used for storagepool provisioning.
	nodeBDs, err := pc.NodeBlockDeviceSelector()
	if err != nil {
		return nil, errors.Wrapf(err, "aborting storagepool create operation as no node qualified")
	}

	claimedNodeBDs, err := pc.ClaimBlockDevice(nodeBDs, spc)
	if err != nil {
		return nil, errors.Wrapf(err, "aborting storagepool create operation as no claimed block devices available")
	}

	// Fill the node name to the CasPool object.
	casPool.NodeName = claimedNodeBDs.NodeName
	//casPool.DiskList = nodeDisks.Disks.Items
	//TODO: Improve Following Code
	if spc.Spec.PoolSpec.PoolType == string(apis.PoolTypeStripedCPV) {
		for _, claimedBD := range claimedNodeBDs.BlockDeviceList {
			var bdList []apis.CspBlockDevice
			var group apis.BlockDeviceGroup
			blockDevice := apis.CspBlockDevice{
				Name:        claimedBD.BDName,
				InUseByPool: true,
				DeviceID:    claimedBD.DeviceID,
			}
			bdList = append(bdList, blockDevice)
			group = apis.BlockDeviceGroup{
				Item: bdList,
			}
			casPool.BlockDeviceList = append(casPool.BlockDeviceList, group)
		}
		return casPool, nil
	}
	count := blockdevice.DefaultDiskCount[spc.Spec.PoolSpec.PoolType]
	for i := 0; i < len(claimedNodeBDs.BlockDeviceList); i = i + count {
		var bdList []apis.CspBlockDevice
		var group apis.BlockDeviceGroup
		for j := 0; j < count; j++ {

			blockDevice := apis.CspBlockDevice{
				Name:        claimedNodeBDs.BlockDeviceList[i+j].BDName,
				InUseByPool: true,
				DeviceID:    claimedNodeBDs.BlockDeviceList[i+j].DeviceID,
			}
			bdList = append(bdList, blockDevice)
		}
		group = apis.BlockDeviceGroup{
			Item: bdList,
		}
		casPool.BlockDeviceList = append(casPool.BlockDeviceList, group)
	}
	return casPool, nil
}

// TODO: Move to block device package
func (pc *PoolCreateConfig) getDeviceID(blockDeviceName string) (string, error) {
	var deviceID string
	blockDevice, err := pc.BlockDeviceClient.Get(blockDeviceName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	if len(blockDevice.Spec.DevLinks) != 0 && len(blockDevice.Spec.DevLinks[0].Links) != 0 {
		deviceID = blockDevice.Spec.DevLinks[0].Links[0]
	} else {
		deviceID = blockDevice.Spec.Path
	}
	return deviceID, nil
}
