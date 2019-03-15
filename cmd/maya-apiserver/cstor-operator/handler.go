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
	"encoding/json"
	nodeselect "github.com/openebs/maya/pkg/algorithm/nodeselect/v1alpha1"
	"github.com/openebs/maya/pkg/hash/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"strings"

	"github.com/golang/glog"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/pkg/errors"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
)

const (
	// DEFAULTPOOlCOUNT will be set to maxPool field of spc if maxPool field is not provided.
	DEFAULTPOOlCOUNT = 3
)

type DiskOperations struct {
	spc        *apis.StoragePoolClaim
	cspList    *apis.CStorPoolList
	controller *Controller
}

var (
	supportedPool = map[apis.CasPoolValString]bool{
		apis.PoolTypeStripedCPV:  true,
		apis.PoolTypeMirroredCPV: true,
		apis.PoolTypeRaidzCPV:    true,
		apis.PoolTypeRaidz2CPV:   true,
	}
)

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the spcPoolUpdated resource
// with the current status of the resource.
func (c *Controller) syncHandler(key, operation string, object interface{}) error {
	// getSpcResource will take a key as argument which contains the namespace/name or simply name
	// of the object and will fetch the object.
	spcGot, err := c.getSpcResource(key)
	if err != nil {
		return err
	}
	// Call the spcEventHandler which will take spc object , key(namespace/name of object) and type of operation we need to to for storage pool
	// Type of operation for storage pool e.g. create, delete etc.
	events, err := c.spcEventHandler(operation, spcGot)
	if events == ignoreEvent {
		glog.Warning("None of the SPC handler was executed")
		return nil
	}
	if err != nil {
		return err
	}
	// If this function returns a error then the object will be requeued.
	// No need to error out even if it occurs,
	return nil
}

// spcPoolEventHandler is to handle SPC related events.
func (c *Controller) spcEventHandler(operation string, spcGot *apis.StoragePoolClaim) (string, error) {
	switch operation {
	case addEvent:
		// Call addEventHandler in case of add event.
		err := c.addEventHandler(spcGot, false)
		return addEvent, err

	case updateEvent:
		// TO-DO : Handle Business Logic
		// Hook Update Business Logic Here
		return updateEvent, nil
	case syncEvent:
		err := c.syncSpc(spcGot)
		if err != nil {
			glog.Errorf("Storagepool %s could not be synced:%v", spcGot.Name, err)
		}
		return syncEvent, nil
	case diskops:
		_, err := c.handleDiskHashChange(spcGot)
		return diskops, err
	default:
		// operation with tag other than add,update and sync are ignored.
		return ignoreEvent, nil
	}
}

// enqueueSpc takes a SPC resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than SPC.
func (c *Controller) enqueueSpc(queueLoad *QueueLoad) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(queueLoad.Object); err != nil {
		runtime.HandleError(err)
		return
	}
	queueLoad.Key = key
	c.workqueue.AddRateLimited(queueLoad)
}

// getSpcResource returns object corresponding to the resource key
func (c *Controller) getSpcResource(key string) (*apis.StoragePoolClaim, error) {
	// Convert the key(namespace/name) string into a distinct name
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(errors.Wrapf(err, "Invalid resource key: %s", key))
		return nil, err
	}
	spcGot, err := c.clientset.OpenebsV1alpha1().StoragePoolClaims().Get(name, metav1.GetOptions{})
	if err != nil {
		// The SPC resource may no longer exist, in which case we stop
		// processing.
		if k8serror.IsNotFound(err) {
			runtime.HandleError(errors.Wrapf(err, "spcGot '%s' in work queue no longer exists", key))
			// No need to return error to caller as we still want to fire the delete handler
			// using the spc key(name)
			// If error is returned the caller function will return without calling the spcEventHandler
			// function that invokes business logic for pool deletion
			return nil, nil
		}
		return nil, err
	}
	return spcGot, nil
}

// synSpc is the function which tries to converge to a desired state for the spc.
func (c *Controller) syncSpc(spcGot *apis.StoragePoolClaim) error {
	glog.V(1).Infof("Reconciling storagepoolclaim %s", spcGot.Name)
	err := c.addEventHandler(spcGot, true)
	return err
}

// addEventHandler is the event handler for the add event of spc.
func (c *Controller) addEventHandler(spc *apis.StoragePoolClaim, resync bool) error {
	// TODO: Move these to admission webhook plugins or CRD validations
	mutateSpc, err := validate(spc)
	if err != nil {
		return err
	}
	// validate can mutate spc object -- for example if maxPool field is not present in case
	// of auto provisioning, maxPool will default to 3.
	// We need to immediately patch SPC object here.
	if mutateSpc && !resync {
		spc, err = c.clientset.OpenebsV1alpha1().StoragePoolClaims().Update(spc)
		if err != nil {
			return errors.Wrap(err, "spc patch for defaulting the field(s) failed")
		}
	}
	pendingPoolCount, err := c.getPendingPoolCount(spc)
	if err != nil {
		return err
	}
	err = c.create(pendingPoolCount, spc)
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
	// TODO: Think of putting this info log as levelled info log.
	glog.Infof("Lease acquired successfully on storagepoolclaim %s ", spc.Name)
	defer newSpcLease.Release()
	for poolCount := 1; poolCount <= pendingPoolCount; poolCount++ {
		glog.Infof("Provisioning pool %d/%d for storagepoolclaim %s", poolCount, pendingPoolCount, spc.Name)
		err = CreateStoragePool(spc)
		if err != nil {
			runtime.HandleError(errors.Wrapf(err, "Pool provisioning failed for %d/%d for storagepoolclaim %s", poolCount, pendingPoolCount, spc.Name))
			break
		}
	}

	_, err = c.patchSpcWithDiskHash(spc)
	if err != nil {
		return errors.Wrapf(err, "failed to patch disk hash for pool create event for spc %s", spc.Name)
	}
	return nil
}

// validate validates the spc configuration before creation of pool.
func validate(spc *apis.StoragePoolClaim) (bool, error) {
	// TODO: Move these to admission webhook plugins or CRD validations
	// Validations for poolType
	var mutateSpc bool
	// If maxPool field is skipped or entered a 0 value in SPC then for the add event it will default to 3.
	// In case of resync maxPool field will not be mutated.
	if len(spc.Spec.Disks.DiskList) == 0 {
		maxPools := spc.Spec.MaxPools
		if maxPools == 0 {
			spc.Spec.MaxPools = DEFAULTPOOlCOUNT
			mutateSpc = true
		}
		if maxPools < 0 {
			return mutateSpc, errors.Errorf("aborting storagepool create operation for %s as invalid maxPool value %d", spc.Name, maxPools)
		}
	}
	poolType := spc.Spec.PoolSpec.PoolType
	if poolType == "" {
		return mutateSpc, errors.Errorf("aborting storagepool create operation for %s as no poolType is specified", spc.Name)
	}

	ok := supportedPool[apis.CasPoolValString(poolType)]
	if !ok {
		return mutateSpc, errors.Errorf("aborting storagepool create operation as specified poolType is %s which is invalid", poolType)
	}

	diskType := spc.Spec.Type
	if !(diskType == string(apis.TypeSparseCPV) || diskType == string(apis.TypeDiskCPV)) {
		return mutateSpc, errors.Errorf("aborting storagepool create operation as specified type is %s which is invalid", diskType)
	}
	return mutateSpc, nil
}

// getCurrentPoolCount give the current pool count for the given auto provisioned spc.
func (c *Controller) getCurrentPoolCount(spc *apis.StoragePoolClaim) (int, error) {
	// Get the current count of provisioned pool for the storagepool claim
	cspList, err := c.clientset.OpenebsV1alpha1().CStorPools().List(metav1.ListOptions{LabelSelector: string(apis.StoragePoolClaimCPK) + "=" + spc.Name})
	if err != nil {
		return 0, errors.Errorf("unable to get current pool count:unable to list storagepools: %v", err)
	}
	return len(cspList.Items), nil
}

// getPendingPoolCount gives the count of pool that needs to be provisioned for a given spc.
func (c *Controller) getPendingPoolCount(spc *apis.StoragePoolClaim) (int, error) {
	var pendingPoolCount int
	if len(spc.Spec.Disks.DiskList) == 0 {
		// Getting pending pool count in case of auto provisioned spc.
		currentPoolCount, err := c.getCurrentPoolCount(spc)
		if err != nil {
			return 0, err
		}
		maxPoolCount := int(spc.Spec.MaxPools)
		pendingPoolCount = maxPoolCount - currentPoolCount
	} else {
		// TODO: -- Refactor using disk refernces to find the used disks
		// Getting pending pool count in case of manual provisioned spc.
		// allNodeDiskMap holds the map : diskName --> hostName ; for all the disks.
		allNodeDiskMap := make(map[string]string)
		// Get all disk in one shot from kube-apiserver
		diskList, err := c.clientset.OpenebsV1alpha1().Disks().List(metav1.ListOptions{})
		if err != nil {
			return 0, err
		}
		newClientSet := &clientSet{
			c.clientset,
		}
		// Get used disk for the SPC
		// usedDiskMap holds the disk which are already used.
		usedDiskMap, err := newClientSet.getUsedDiskMap()
		for _, disk := range diskList.Items {
			if usedDiskMap[disk.Name] == 1 {
				continue
			}
			allNodeDiskMap[disk.Name] = disk.Labels[string(apis.HostNameCPK)]
		}
		// nodeCountMap holds the node names as the key over which pool should be provisioned.
		nodeCountMap := make(map[string]int)
		for _, spcDisk := range spc.Spec.Disks.DiskList {
			if !(len(strings.TrimSpace(allNodeDiskMap[spcDisk])) == 0) {
				nodeCountMap[allNodeDiskMap[spcDisk]]++
			}
		}
		pendingPoolCount = len(nodeCountMap)
	}
	if pendingPoolCount < 0 {
		return 0, errors.Errorf("Got invalid pending pool count %d", pendingPoolCount)
	}
	return pendingPoolCount, nil
}

// TODO: Add unit test for following functions

func (c *Controller) handleDiskHashChange(spc *apis.StoragePoolClaim) (*apis.StoragePoolClaim, error) {
	removalDiskOperations, err := c.NewDiskOperations(spc)
	if err != nil {
		return nil, err
	}
	spcGot, err := removalDiskOperations.handleDiskRemovalHashChange()
	if err != nil {
		return nil, errors.Wrapf(err, "could not remove/replace disks for spc %s", spc.Name)
	}

	additionDiskOperations, err := c.NewDiskOperations(spc)
	if err != nil {
		return nil, err
	}
	err = additionDiskOperations.handleDiskAdditionHashChange()
	if err != nil {
		return nil, errors.Wrapf(err, "could not remove/replace disks for spc %s", spc.Name)
	}

	spc, err = c.patchSpcWithDiskHash(spcGot)
	if err != nil {
		return nil, errors.Wrapf(err, "could not patch spc %s with newer disk hash for disk operations", spcGot.Name)
	}
	return spc, nil
}

func (c *Controller) NewDiskOperations(spc *apis.StoragePoolClaim) (*DiskOperations, error) {
	cspList, err := c.getCsp(spc)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not list CSP for SPC %s", spc.Name)
	}
	newDiskOperations := &DiskOperations{
		spc:        spc,
		cspList:    cspList,
		controller: c,
	}
	return newDiskOperations, nil
}

func (do *DiskOperations) handleDiskRemovalHashChange() (*apis.StoragePoolClaim, error) {
	removedDisks, err := do.getRemovedDisks()
	if err != nil {
		return nil, errors.Wrapf(err, "Could not get list of removed disks for SPC %s", do.spc.Name)
	}
	for _, disk := range removedDisks {
		do.removeDisk(disk)
	}
	for _, csp := range do.cspList.Items {
		csp, err := do.controller.updateCsp(&csp)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to update csp %s for disk operations for spc %s", csp.Name, do.spc.Name)
		}
		if isTopVdevLost(csp) {
			err := do.controller.deleteCsp(csp)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to delete csp %s for disk operations for spc %s", csp.Name, do.spc.Name)
			}

		}
	}
	return do.spc, nil
}

func (do *DiskOperations) handleDiskAdditionHashChange() error {
	dettachedDisk := do.getDettachedCspDisks()
	for disk, _ := range dettachedDisk {
		do.reAttachDisk(disk)
	}

	replacementDisks, err := do.getAddedDisks()
	if err != nil {
		return errors.Wrapf(err, "could not get newely added disks for replacement for spc %s", do.spc.Name)
	}

	for _, disk := range replacementDisks {
		do.replaceDisk(disk)
	}

	for i, csp := range do.cspList.Items {
		csp, err := do.controller.updateCsp(&csp)
		if err != nil {
			return errors.Wrapf(err, "failed to update csp %s for disk replacement operations for spc %s", csp.Name, do.spc.Name)
		}
		do.cspList.Items[i] = *csp
	}

	newDisks, err := do.getAddedDisks()
	nodeDisk := do.getnodeDiskMap(newDisks)
	err = do.addDisk(nodeDisk)
	if err != nil {
		return errors.Wrapf(err, "failed to add disk for spc %s", do.spc.Name)
	}
	return nil
}

func (do *DiskOperations) addDisk(nodeDisk map[string][]string) error {
	defaultDiskCount := nodeselect.DefaultDiskCount[do.spc.Spec.PoolSpec.PoolType]
	nodeCspMap := do.getCspNodeMap()
	var newGroup apis.DiskGroup
	var cspdisk []apis.CspDisk
	for node, disks := range nodeDisk {
		csp := nodeCspMap[node]
		if csp == nil {
			continue
		}
		diskCount := 0
		if len(disks) >= defaultDiskCount {
			diskCount = (len(disks) / defaultDiskCount) * defaultDiskCount

			for i := 0; i < defaultDiskCount; i = i + defaultDiskCount {
				for j := 0; j < diskCount; j++ {
					var item apis.CspDisk
					item.Name = disks[j]
					item.InUseByPool = true
					cspdisk = append(cspdisk, item)
				}
				newGroup.Item = cspdisk
				csp.Spec.Group = append(csp.Spec.Group, newGroup)

			}
			_, err := do.controller.updateCsp(csp)
			if err != nil {
				return errors.Wrapf(err, "failed to update csp %s for spc %s for disk addition operations", csp.Name, do.spc.Name)
			}
		}
	}
	return nil
}

func (do *DiskOperations) getnodeDiskMap(disks []string) map[string][]string {
	nodeDiskMap := make(map[string][]string)
	for _, disk := range disks {
		gotDisk, err := do.controller.clientset.OpenebsV1alpha1().Disks().Get(disk, metav1.GetOptions{})
		if err != nil {
			return nil
		}
		if gotDisk == nil {
			return nil
		}
		nodeDiskMap[gotDisk.Labels[string(apis.HostNameCPK)]] = append(nodeDiskMap[gotDisk.Labels[string(apis.HostNameCPK)]], disk)
	}
	return nodeDiskMap
}

func (do *DiskOperations) getCspNodeMap() map[string]*apis.CStorPool {
	cspNodeMap := make(map[string]*apis.CStorPool)
	for _, csp := range do.cspList.Items {
		cspNodeMap[csp.Labels[string(apis.HostNameCPK)]] = &csp
	}
	return cspNodeMap
}

func isTopVdevLost(csp *apis.CStorPool) bool {
	for _, group := range csp.Spec.Group {
		count := 0
		for _, disk := range group.Item {
			if disk.InUseByPool == false {
				count++
			}
		}
		// TODO: Remove hardcoding
		if count >= 1 && csp.Spec.PoolSpec.PoolType == "striped" {
			return true
		}
		if count >= 2 && csp.Spec.PoolSpec.PoolType == "mirrored" {
			return true
		}
	}
	return false
}

func (do *DiskOperations) reAttachDisk(diskName string) {
	for i, csp := range do.cspList.Items {
		for j, group := range csp.Spec.Group {
			for k, disk := range group.Item {
				if disk.Name == diskName {
					do.cspList.Items[i].Spec.Group[j].Item[k].InUseByPool = true
				}
			}
		}
	}
}

func (do *DiskOperations) replaceDisk(diskName string) {
	for i, csp := range do.cspList.Items {
		for j, group := range csp.Spec.Group {
			for k, disk := range group.Item {
				if disk.InUseByPool == false {
					do.cspList.Items[i].Spec.Group[j].Item[k].Name = diskName
					do.cspList.Items[i].Spec.Group[j].Item[k].InUseByPool = true
				}
			}
		}
	}
}

func (do *DiskOperations) removeDisk(diskName string) {
	for i, csp := range do.cspList.Items {
		for j, group := range csp.Spec.Group {
			for k, disk := range group.Item {
				if disk.Name == diskName {
					do.cspList.Items[i].Spec.Group[j].Item[k].InUseByPool = false
				}
			}
		}
	}
}

// getSpcDisks returns map of spc disks present on SPC.
func (do *DiskOperations) getSpcDisks() map[string]bool {
	// Make a map containing all the disks present in spc.
	spcDisks := make(map[string]bool)
	for _, disk := range do.spc.Spec.Disks.DiskList {
		spcDisks[disk] = true
	}
	return spcDisks
}

func (do *DiskOperations) getCspDisks() (map[string]bool, error) {
	// Make a map containing all the disks present in csp
	// Get all CSP corresponding to the SPC
	cspDisks := make(map[string]bool)
	for _, csp := range do.cspList.Items {
		for _, group := range csp.Spec.Group {
			for _, disk := range group.Item {
				cspDisks[disk.Name] = true
			}
		}
	}
	return cspDisks, nil
}

func (do *DiskOperations) getDettachedCspDisks() map[string]bool {
	// Make a map containing all the disks present in csp whis in not present in SPC.
	spcDisks := do.getSpcDisks()
	cspDisks := make(map[string]bool)
	for _, csp := range do.cspList.Items {
		for _, group := range csp.Spec.Group {
			for _, disk := range group.Item {
				if spcDisks[disk.Name] == true && disk.InUseByPool == false {
					cspDisks[disk.Name] = true
				}
			}
		}
	}
	return cspDisks
}

// getRemovedDisks return a list of disks present on all CSPs for a given SPC, which has been removed from SPC
func (do *DiskOperations) getRemovedDisks() ([]string, error) {
	var removedDisk []string
	// Get the disks present on CSPs
	cspDisks, err := do.getCspDisks()
	if err != nil {
		return []string{}, errors.Wrapf(err, "Could not get removed disks for SPC %s", do.spc.Name)
	}
	// get the disk present on SPC
	spcDisks := do.getSpcDisks()
	for disk, _ := range cspDisks {
		if spcDisks[disk] == false {
			removedDisk = append(removedDisk, disk)
		}
	}
	return removedDisk, nil
}

// getAddedDisks returns a list of disk that is added to SPC.
func (do *DiskOperations) getAddedDisks() ([]string, error) {
	var addedDisk []string
	// get the disks present on CSPs
	cspDisks, err := do.getCspDisks()
	if err != nil {
		return []string{}, errors.Wrapf(err, "Could not get removed disks for SPC %s", do.spc.Name)
	}
	// get the disk present on SPC
	spcDisks := do.getSpcDisks()
	for disk, _ := range spcDisks {
		if cspDisks[disk] == false {
			addedDisk = append(addedDisk, disk)
		}
	}
	return addedDisk, nil
}

func (c *Controller) getCsp(spc *apis.StoragePoolClaim) (*apis.CStorPoolList, error) {
	cspList, err := c.clientset.OpenebsV1alpha1().CStorPools().List(metav1.ListOptions{LabelSelector: string(apis.StoragePoolClaimCPK) + "=" + spc.Name})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get csp objects for spc %s", spc.Name)
	}
	return cspList, nil

}

// TODO: Patch using patch package.
func (c *Controller) patchSpcWithDiskHash(spc *apis.StoragePoolClaim) (*apis.StoragePoolClaim, error) {

	diskHash, _ := hash.Hash(spc.Spec.Disks)
	spcPatch := make([]Patch, 1)
	spcPatch[0].Op = PatchOperation
	// TODO: If there is no annotaion in SPC -- Create it
	if spc.Annotations == nil {
		return nil, errors.Errorf("No annotation found in spc %s", spc.Name)
	}
	if spc.Annotations[spcDiskHashKey] == "" {
		spcPatch[0].Op = PatchOperationAdd
	}
	spcPatch[0].Path = spcDiskHashKeyPath
	spcPatch[0].Value = diskHash
	spcPatchJSON, err := json.Marshal(spcPatch)
	spcGot, err := c.clientset.OpenebsV1alpha1().StoragePoolClaims().Patch(spc.Name, types.JSONPatchType, spcPatchJSON)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to patch spc %s with the new disk list hash", spc.Name)
	}

	return spcGot, nil
}

func (c *Controller) updateCsp(csp *apis.CStorPool) (*apis.CStorPool, error) {
	csp, err := c.clientset.OpenebsV1alpha1().CStorPools().Update(csp)
	return csp, err
}

func (c *Controller) deleteCsp(csp *apis.CStorPool) error {
	err := c.clientset.OpenebsV1alpha1().CStorPools().Delete(csp.Name, &metav1.DeleteOptions{})
	return err
}
