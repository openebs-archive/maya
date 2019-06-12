/*
Copyright 2018 The OpenEBS Authors

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
	"fmt"
	"strings"
	"time"

	"github.com/golang/glog"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	openebs "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
	env "github.com/openebs/maya/pkg/env/v1alpha1"
	spcv1alpha1 "github.com/openebs/maya/pkg/storagepoolclaim/v1alpha1"
	"github.com/pkg/errors"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
)

var (
	// supportedPool is a map holding the supported raid configurations.
	supportedPool = map[apis.CasPoolValString]bool{
		apis.PoolTypeStripedCPV:  true,
		apis.PoolTypeMirroredCPV: true,
		apis.PoolTypeRaidzCPV:    true,
		apis.PoolTypeRaidz2CPV:   true,
	}
)

const (
	// DiskStateActive is the active state of the disk
	DiskStateActive = "Active"
	// ProvisioningTypeManual is the manual type of provisioning pool
	ProvisioningTypeManual = "manual"
)

type clientSet struct {
	oecs openebs.Interface
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the spcPoolUpdated resource
// with the current status of the resource.
func (c *Controller) syncHandler(key string) error {
	startTime := time.Now()
	glog.V(4).Infof("Started syncing storagepoolclaim %q (%v)", key, startTime)
	defer func() {
		glog.V(4).Infof("Finished syncing storagepoolclaim %q (%v)", key, time.Since(startTime))
	}()

	// Convert the namespace/name string into a distinct namespace and name
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the spc resource with this namespace/name
	spc, err := c.spcLister.Get(name)
	if k8serror.IsNotFound(err) {
		runtime.HandleError(fmt.Errorf("spc '%s' has been deleted", key))
		return nil
	}
	if err != nil {
		return err
	}

	// Deep-copy otherwise we are mutating our cache.
	// TODO: Deep-copy only when needed.
	spcGot := spc.DeepCopy()
	err = c.syncSpc(spcGot)
	return err
}

// enqueueSpc takes a SPC resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than SPC.
func (c *Controller) enqueueSpc(spc interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(spc); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

// synSpc is the function which tries to converge to a desired state for the spc.
func (c *Controller) syncSpc(spc *apis.StoragePoolClaim) error {
	err := validate(spc)
	if err != nil {
		glog.Errorf("Validation of spc failed:%s", err)
		return nil
	}
	pendingPoolCount, err := c.getPendingPoolCount(spc)
	if err != nil {
		return err
	}
	if pendingPoolCount > 0 {
		err = c.create(pendingPoolCount, spc)
		if err != nil {
			return err
		}
	}
	return nil
}

// create is a wrapper function that calls the actual function to create pool as many time
// as the number of pools need to be created.
func (c *Controller) create(pendingPoolCount int, spc *apis.StoragePoolClaim) error {
	var newSpcLease Leaser
	newSpcLease = &Lease{spc, SpcLeaseKey, c.clientset, c.kubeclientset}
	err := newSpcLease.Hold()
	if err != nil {
		return errors.Wrapf(err, "Could not acquire lease on spc object")
	}
	glog.V(4).Infof("Lease acquired successfully on storagepoolclaim %s ", spc.Name)
	defer newSpcLease.Release()
	for poolCount := 1; poolCount <= pendingPoolCount; poolCount++ {
		glog.Infof("Provisioning pool %d/%d for storagepoolclaim %s", poolCount, pendingPoolCount, spc.Name)
		err = c.CreateStoragePool(spc)
		if err != nil {
			runtime.HandleError(errors.Wrapf(err, "Pool provisioning failed for %d/%d for storagepoolclaim %s", poolCount, pendingPoolCount, spc.Name))
		}
	}
	return nil
}

// validate validates the spc configuration before creation of pool.
func validate(spc *apis.StoragePoolClaim) error {
	for _, v := range validateFuncList {
		err := v(spc)
		if err != nil {
			return err
		}
	}
	return nil
}

// validateFunc is typed function for spc validation functions.
type validateFunc func(*apis.StoragePoolClaim) error

// validateFuncList holds a list of validate functions for spc
var validateFuncList = []validateFunc{
	validatePoolType,
	validateDiskType,
	validateAutoSpcMaxPool,
}

// validatePoolType validates pool type in spc.
func validatePoolType(spc *apis.StoragePoolClaim) error {
	poolType := spc.Spec.PoolSpec.PoolType
	ok := supportedPool[apis.CasPoolValString(poolType)]
	if !ok {
		return errors.Errorf("aborting storagepool create operation as specified poolType is '%s' which is invalid", poolType)
	}
	return nil
}

// validateDiskType validates the disk types in spc.
func validateDiskType(spc *apis.StoragePoolClaim) error {
	diskType := spc.Spec.Type
	if !spcv1alpha1.SupportedDiskTypes[apis.CasPoolValString(diskType)] {
		return errors.Errorf("aborting storagepool create operation as specified type is %s which is invalid", diskType)
	}
	return nil
}

// validateAutoSpcMaxPool validates the max pool count in auto spc
func validateAutoSpcMaxPool(spc *apis.StoragePoolClaim) error {
	if isAutoProvisioning(spc) {
		maxPools := spc.Spec.MaxPools
		if maxPools == nil {
			return errors.Errorf("validation of spc object is failed as no max pool field present in spc %s", spc.Name)
		}
		if *maxPools < 0 {
			return errors.Errorf("aborting storagepool create operation for %s as invalid maxPool value %d", spc.Name, maxPools)
		}
	}
	return nil
}

// getCurrentPoolCount give the current pool count for the given auto provisioned spc.
func (c *Controller) getCurrentPoolCount(spc *apis.StoragePoolClaim) (int, error) {
	// Get the current count of provisioned pool for the storagepool claim
	cspList, err := c.clientset.OpenebsV1alpha1().CStorPools().List(metav1.ListOptions{LabelSelector: string(apis.StoragePoolClaimCPK) + "=" + spc.Name})
	if err != nil {
		return 0, errors.Errorf("unable to get current pool count:unable to list cstor pools: %v", err)
	}
	return len(cspList.Items), nil
}

// isPoolPending tells whether some pool is pending to be created.
func (c *Controller) isPoolPending(spc *apis.StoragePoolClaim) bool {
	pCount, err := c.getPendingPoolCount(spc)
	if err != nil {
		glog.Errorf("Unable to get pending pool count for spc %s:%s", spc.Name, err)
		return false
	}
	if pCount > 0 {
		return true
	}
	return false
}

// getPendingPoolCount gives the count of pool that needs to be provisioned for a given spc.
func (c *Controller) getPendingPoolCount(spc *apis.StoragePoolClaim) (int, error) {
	var err error
	var pendingPoolCount int
	if isAutoProvisioning(spc) {
		pendingPoolCount, err = c.getAutoSpcPendingPoolCount(spc)
	} else {
		pendingPoolCount, err = c.getManualSpcPendingPoolCount(spc)
	}
	if err != nil {
		return 0, errors.Wrapf(err, "failed to get pending pool count for spc %s", spc.Name)
	}
	if isValidPendingPoolCount(pendingPoolCount) {
		return pendingPoolCount, nil
	}
	return 0, nil
}

// getAutoSpcPendingPoolCount get the pending pool count for auto provisioned spc.
func (c *Controller) getAutoSpcPendingPoolCount(spc *apis.StoragePoolClaim) (int, error) {
	// Getting pending pool count in case of auto provisioned spc.
	err := validateAutoSpcMaxPool(spc)
	if err != nil {
		return 0, errors.Wrapf(err, "error in max pool value in spc %s", spc.Name)
	}
	currentPoolCount, err := c.getCurrentPoolCount(spc)
	if err != nil {
		return 0, err
	}
	maxPoolCount := *(spc.Spec.MaxPools)
	pendingPoolCount := maxPoolCount - currentPoolCount
	return pendingPoolCount, nil
}

// getManualSpcPendingPoolCount gets the pending pool count for manual provisioned spc.
func (c *Controller) getManualSpcPendingPoolCount(spc *apis.StoragePoolClaim) (int, error) {
	usableNodeCount, err := c.getUsableNodeCount(spc)
	if err != nil {
		return 0, err
	}
	pendingPoolCount := len(usableNodeCount)
	return pendingPoolCount, nil
}

// getFreeDiskNodeMap forms a map that holds block device names which can be used to create a pool.
func (c *Controller) getFreeDiskNodeMap() (map[string]string, error) {
	freeNodeDiskMap := make(map[string]string)

	//TODO: Update below snippet tomake use of builder and blockdevice kubeclient
	//package
	// Get all block device from kube-apiserver
	namespace := env.Get(env.OpenEBSNamespace)
	blockDeviceList, err := c.ndmclientset.OpenebsV1alpha1().BlockDevices(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	usedBlockDeviceMap, err := c.getUsedBlockDeviceMap()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get the used block device map ")
	}
	for _, blockDevice := range blockDeviceList.Items {
		if usedBlockDeviceMap[blockDevice.Name] == 1 {
			continue
		}
		freeNodeDiskMap[blockDevice.Name] = blockDevice.Labels[string(apis.HostNameCPK)]
	}
	return freeNodeDiskMap, nil
}

// getUsableNodeCount forms a map that holds node which can be used to provision pool.
func (c *Controller) getUsableNodeCount(spc *apis.StoragePoolClaim) (map[string]int, error) {
	nodeCountMap := make(map[string]int)
	freeNodeDiskMap, err := c.getFreeDiskNodeMap()
	if err != nil {
		return nil, err
	}
	for _, spcBlockDevice := range spc.Spec.BlockDevices.BlockDeviceList {
		if !(len(strings.TrimSpace(freeNodeDiskMap[spcBlockDevice])) == 0) {
			nodeCountMap[freeNodeDiskMap[spcBlockDevice]]++
		}
	}
	return nodeCountMap, nil
}

// getUsedBlockDeviceMap form usedDisk map that will hold the list of all used
// block device
// TODO: Move to blockDevice package
func (c *Controller) getUsedBlockDeviceMap() (map[string]int, error) {
	// Get the list of block devices that has been used already for pool provisioning
	cspList, err := c.clientset.OpenebsV1alpha1().CStorPools().List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get the list of cstor pool")
	}
	// Form a map that will hold all the used block device
	usedBlockDeviceMap := make(map[string]int)
	for _, csp := range cspList.Items {
		for _, group := range csp.Spec.Group {
			for _, bd := range group.Item {
				usedBlockDeviceMap[bd.Name]++
			}
		}
	}
	return usedBlockDeviceMap, nil
}

// isValidPendingPoolCount tells whether the pending pool count is valid or not.
func isValidPendingPoolCount(pendingPoolCout int) bool {
	if pendingPoolCout < 0 {
		return false
	}
	return true
}

// isAutoProvisioning returns true the spc is auto provisioning type.
func isAutoProvisioning(spc *apis.StoragePoolClaim) bool {
	return spc.Spec.BlockDevices.BlockDeviceList == nil
}

// isManualProvisioning returns true if the spc is auto provisioning type.
func isManualProvisioning(spc *apis.StoragePoolClaim) bool {
	return spc.Spec.BlockDevices.BlockDeviceList != nil
}
