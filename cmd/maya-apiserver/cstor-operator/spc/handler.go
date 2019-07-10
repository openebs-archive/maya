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
	klabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/kubernetes/pkg/util/slice"
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
	// spc finalizers should be removed only after deletion of corresponding cspc
	if isSPCDeletetionCandidate(spc) {
		return c.removeFinalizer(spc)
	}
	err = c.createOrUpdate(spc)
	if err != nil {
		return err
	}
	return nil
}

// createOrUpdate is a wrapper function that calls the actual function to create
// or update cstorpoolcluster based on the availability of cspc
func (c *Controller) createOrUpdate(spc *apis.StoragePoolClaim) error {
	var newSpcLease Leaser
	//Check whether create or update is required or not
	if !c.isPoolSpecPending(spc) {
		return nil
	}
	newSpcLease = &Lease{spc, SpcLeaseKey, c.clientset, c.kubeclientset}
	err := newSpcLease.Hold()
	if err != nil {
		return errors.Wrapf(err, "Could not acquire lease on spc object")
	}
	glog.V(4).Infof("Lease acquired successfully on storagepoolclaim %s ", spc.Name)
	defer newSpcLease.Release()
	err = c.CreateOrUpdateCStorPoolCluster(spc)
	if err != nil {
		runtime.HandleError(
			errors.Wrapf(
				err,
				"failed to create or update cstorpoolcluster from spc %s",
				spc.Name,
			),
		)
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
		return errors.Errorf(
			"aborting storagepool create operation as specified poolType is '%s' which is invalid",
			poolType,
		)
	}
	return nil
}

// validateDiskType validates the disk types in spc.
func validateDiskType(spc *apis.StoragePoolClaim) error {
	diskType := spc.Spec.Type
	if !spcv1alpha1.SupportedDiskTypes[apis.CasPoolValString(diskType)] {
		return errors.Errorf(
			"aborting storagepool create operation as specified type is %s which is invalid",
			diskType,
		)
	}
	return nil
}

// validateAutoSpcMaxPool validates the max pool count in auto spc
func validateAutoSpcMaxPool(spc *apis.StoragePoolClaim) error {
	if isAutoProvisioning(spc) {
		maxPools := spc.Spec.MaxPools
		if maxPools == nil {
			return errors.Errorf(
				"validation of spc object failed as no maxpool value is present on spc %s",
				spc.Name,
			)
		}
		if *maxPools < 0 {
			return errors.Errorf(
				"aborting storagepool create operation for %s as invalid maxPool value %d",
				spc.Name,
				maxPools,
			)
		}
		return nil
	}
	return errors.Errorf("validation of spc object failed manual provisioning is not supported")
}

// getCurrentPoolCount give the current pool count for the given auto provisioned spc.
func (c *Controller) getCurrentPoolCount(spc *apis.StoragePoolClaim) (int, error) {
	// Get the current count of provisioned pool for the storagepool claim
	cspList, err := c.clientset.
		OpenebsV1alpha1().
		CStorPools().
		List(metav1.ListOptions{
			LabelSelector: string(apis.StoragePoolClaimCPK) + "=" + spc.Name,
		},
		)
	if err != nil {
		return 0, errors.Errorf(
			"unable to get current pool count:unable to list cstor pools: %v",
			err,
		)
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

// isPoolSpecPending returns true when create or update operation is required
func (c *Controller) isPoolSpecPending(spc *apis.StoragePoolClaim) bool {
	// SPC and CSPC name will be same in case of auto provisioning of cstor
	// pools creations
	if !isAutoProvisioning(spc) {
		glog.Errorf("manual provisioning of pools using %s is not supported", spc.Name)
		return false
	}
	namespace := env.Get(env.OpenEBSNamespace)
	customSPCObj := &spcv1alpha1.SPC{Object: spc}
	selector := klabels.SelectorFromSet(customSPCObj.GetDefaultSPCLabels())
	cspcObjList, err := c.cspcLister.
		CStorPoolClusters(namespace).
		List(selector)
	if err != nil {
		glog.Errorf("failed to get cspc %s in namespace %s", spc.Name, namespace)
		return false
	}
	if len(cspcObjList) == 0 {
		return true
	}
	if len(cspcObjList) > 1 {
		glog.Errorf("got more than one cspc using spc %s", spc.Name, namespace)
		return false
	}
	//TODO: should we support down scale of pools
	if len(cspcObjList[0].Spec.Pools) >= *spc.Spec.MaxPools {
		return false
	}
	return true
}

// getPendingPoolCount gives the count of pool that needs to be provisioned for a given spc.
func (c *Controller) getPendingPoolCount(spc *apis.StoragePoolClaim) (int, error) {
	if !isAutoProvisioning(spc) {
		return 0, errors.Errorf("manual provisioning of spc is no more supported")
	}
	return 1, nil
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

// removeFinalizer will delete cspc and remove finalizers on spc
func (c *Controller) removeFinalizer(spc *apis.StoragePoolClaim) error {
	cspcName := spc.Name
	namespace := env.Get(env.OpenEBSNamespace)
	err := c.deleteCSPCResource(spc.Name, namespace)
	if err != nil {
		return errors.Wrapf(
			err,
			"failed to delete cspc %s in namespace %s",
			cspcName,
			namespace,
		)
	}
	return c.removeSPCFinalizer(spc)
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

// deleteCSPCResource removes the finalizer from cspc and delete cspc resource
func (c *Controller) deleteCSPCResource(cspcName, namespace string) error {
	cspcObj, err := c.cspcLister.
		CStorPoolClusters(namespace).
		Get(cspcName)
	if k8serror.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return errors.Wrapf(
			err,
			"failed to get cspc %s in namespace %s",
			cspcName,
			namespace,
		)
	}
	newCSPCObj, err := c.removeCSPCFinalizer(cspcObj)
	if err != nil {
		return nil
	}
	err = c.clientset.
		OpenebsV1alpha1().
		CStorPoolClusters(newCSPCObj.Namespace).
		Delete(newCSPCObj.Name, &metav1.DeleteOptions{})
	if err != nil {
		return errors.Wrapf(
			err,
			"failed to delete cspc %s in namespace %s",
			newCSPCObj.Name,
			newCSPCObj.Namespace,
		)
	}
	return nil
}

//TODO: Club removeCSPCFinalizer and removeSPCFinalizer by using interface

// removeCSPCFinalizer will remove finalizer on cspc
func (c *Controller) removeCSPCFinalizer(
	cspcObj *apis.CStorPoolCluster) (*apis.CStorPoolCluster, error) {
	if len(cspcObj.Finalizers) == 0 {
		return cspcObj, nil
	}
	dupCSPCObj := cspcObj.DeepCopy()
	dupCSPCObj.Finalizers = []string{}
	patchBytes, err := getPatchData(cspcObj, dupCSPCObj)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to get patch bytes for cspc %s in namespace %s",
			dupCSPCObj.Name,
			dupCSPCObj.Namespace,
		)
	}
	newCSPCObj, err := c.clientset.
		OpenebsV1alpha1().
		CStorPoolClusters(cspcObj.Namespace).
		Patch(dupCSPCObj.Name, types.MergePatchType, patchBytes)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to remove finalizers from cspc %s in namespace %s",
			cspcObj.Name,
			cspcObj.Namespace,
		)
	}
	return newCSPCObj, nil
}

// removeSPCFinalizer will remove finalizer on spc
func (c *Controller) removeSPCFinalizer(
	spcObj *apis.StoragePoolClaim) error {
	if len(spcObj.Finalizers) == 0 {
		return nil
	}
	dupSPCObj := spcObj.DeepCopy()
	dupSPCObj.Finalizers = []string{}
	patchBytes, err := getPatchData(spcObj, dupSPCObj)
	if err != nil {
		return errors.Wrapf(
			err,
			"failed to get patch bytes for spc %s",
			dupSPCObj.Name,
		)
	}
	_, err = c.clientset.
		OpenebsV1alpha1().
		StoragePoolClaims().
		Patch(dupSPCObj.Name, types.MergePatchType, patchBytes)
	if err != nil {
		return errors.Wrapf(
			err,
			"failed to remove finalizers from spc %s",
			spcObj.Name,
		)
	}
	return nil
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

// isSPCDeletionCandidate return true when deletion timestamp & finalizer is
// available on spc
func isSPCDeletetionCandidate(spc *apis.StoragePoolClaim) bool {
	return spc.ObjectMeta.DeletionTimestamp != nil &&
		slice.ContainsString(spc.ObjectMeta.Finalizers, spcv1alpha1.SPCFinalizer, nil)
}
