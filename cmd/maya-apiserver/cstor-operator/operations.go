/*
Copyright 2019 The OpenEBS Authors

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
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/hash/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// PoolConfig is the config to carry out disk operations.
type PoolConfig struct {
	spc        *apis.StoragePoolClaim
	cspList    *apis.CStorPoolList
	controller *Controller
}

// handlerfunc is typed predicates to handle disk operations performed on spc.
type handlerfunc func(*PoolConfig) (*apis.StoragePoolClaim, error)

// HandlerPredicates contains a list of predicates that should be executed in order so as to
// reach desired state in response to any change in disk list on spc.
var HandlerPredicates = []handlerfunc{
	HandleDiskRemoval,
	HandleDiskAddition,
}

// NewPoolConfig is the constructor for PoolConfig struct.
func (c *Controller) NewPoolConfig(spc *apis.StoragePoolClaim) (*PoolConfig, error) {
	cspList, err := c.getCsp(spc)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not list CSP for SPC %s", spc.Name)
	}
	newPoolConfig := &PoolConfig{
		spc:        spc,
		cspList:    cspList,
		controller: c,
	}
	return newPoolConfig, nil
}

// handleDiskHashChange is called if the hash of disk list changes on SPC.
func (c *Controller) handleDiskHashChange(spc *apis.StoragePoolClaim) (*apis.StoragePoolClaim, error) {
	err := c.executeHandlerPredicates(spc)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to execute handler predicates for disk operations for spc %s", spc.Name)
	}
	// patch spc with the new disk list hash once the execution is successful.
	spc, err = c.patchSpcWithDiskHash(spc)
	if err != nil {
		return nil, errors.Wrapf(err, "could not patch spc %s with newer disk hash for disk operations", spc.Name)
	}
	return spc, nil
}

// executeHandlerPredicates executes all the handler predicates in order.
func (c *Controller) executeHandlerPredicates(spc *apis.StoragePoolClaim) error {
	for _, p := range HandlerPredicates {
		poolConfig, err := c.NewPoolConfig(spc)
		if err != nil {
			return errors.Wrapf(err, "could not initialize the pool config for spc %s", spc.Name)
		}
		_, err = p(poolConfig)
		if err != nil {
			return errors.Wrapf(err, "disk operation was not successful for spc %s", spc.Name)
		}
	}
	return nil
}

// HandleDiskRemoval executes a set of disk operations predicates that could happen as a result of removal
// of disks on SPC.
func HandleDiskRemoval(pc *PoolConfig) (*apis.StoragePoolClaim, error) {
	for _, p := range removeDiskOpsPredicates {
		p()(pc)
		err := pc.updateCspList()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to update disk operations for spc %s", pc.spc.Name)
		}
	}
	return pc.spc, nil
}

// HandleDiskAddition executes a set of disk operations predicates that could happen as a result of addition
// of new disks on SPC.
func HandleDiskAddition(pc *PoolConfig) (*apis.StoragePoolClaim, error) {
	for _, p := range addDiskOpsPredicates {
		p()(pc)
		err := pc.updateCspList()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to update disk operations for spc %s", pc.spc.Name)
		}
	}
	return pc.spc, nil
}

// DiskOps is typed function for disk operations predicate functions.
type DiskOps func(*PoolConfig)

// opsPredicate is a typed function for function returning disk operations predicate functions.
type opsPredicate func() DiskOps

// removeDiskOpsPredicates contains a list of predicates that carry out disk removal operations.
var removeDiskOpsPredicates = []opsPredicate{
	RemoveDisk,
	DeletePool,
}

// addDiskOpsPredicates contains a list of predicates that carry out disk addition operations.
var addDiskOpsPredicates = []opsPredicate{
	ReattachDisk,
	ReplaceDisk,
	ExpandPool,
}

// RemoveDisk removes a disk from cstor pool.
func RemoveDisk() DiskOps {
	return func(pc *PoolConfig) {
		removedDisks := pc.getRemovedDisks()
		for _, disk := range removedDisks {
			pc.setInUseByPool(disk, false)
			// TODO: Enqueue operation for disk removal.
		}
	}
}

// DeletePool deletes a pool if top level vdev is lost as a result of disk removals.
func DeletePool() DiskOps {
	return func(pc *PoolConfig) {
		for _, csp := range pc.cspList.Items {
			cspCopy := csp
			if isTopVdevLost(&cspCopy) {
				enqueueDeleteOperation(&cspCopy)
				pc.updatePoolConfig(cspCopy)
			}
		}
	}
}

// ReattachDisk re-attaches a disk on the cstor pool.
func ReattachDisk() DiskOps {
	return func(pc *PoolConfig) {
		dettachedDisk := pc.getDettachedCspDisks()
		for disk := range dettachedDisk {
			pc.setInUseByPool(disk, true)
			// TOdO: Enqueue operation for reattaching disk.
		}
	}
}

// ReplaceDisk replaces a disk on the cstor pool.
func ReplaceDisk() DiskOps {
	return func(pc *PoolConfig) {
		replacementDisks := pc.getAddedDisks()
		for _, disk := range replacementDisks {
			pc.replaceDisk(disk)
			// TODO: Enqueue operation for disk replacment.
		}
	}
}

// ExpandPool expands the cstor pool vertically.
func ExpandPool() DiskOps {
	return func(pc *PoolConfig) {
		nodeCspMap := pc.getCspNodeMap()
		newDisks := pc.getAddedDisks()
		nodeDisk := pc.getnodeDiskMap(newDisks)
		for node, disks := range nodeDisk {
			csp := nodeCspMap[node]
			if csp == nil {
				continue
			}
			pc.expandCsp(csp, disks)
		}
	}
}

// expandCsp adds disk to the CSP.
func (pc *PoolConfig) expandCsp(csp *apis.CStorPool, disks []DiskDetails) {
	var newGroup apis.DiskGroup
	var cspdisk []apis.CspDisk
	var deviceIDs []string
	defaultDiskCount := nodeselect.DefaultDiskCount[pc.spc.Spec.PoolSpec.PoolType]
	diskCount := 0
	if len(disks) >= defaultDiskCount {
		diskCount = (len(disks) / defaultDiskCount) * defaultDiskCount
		for i := 0; i < defaultDiskCount; i = i + defaultDiskCount {
			for j := 0; j < diskCount; j++ {
				var item apis.CspDisk
				item.Name = disks[j].DiskName
				item.DeviceID = disks[j].DeviceID
				item.InUseByPool = true
				deviceIDs = append(deviceIDs, disks[j].DeviceID)
				cspdisk = append(cspdisk, item)
			}
			newGroup.Item = cspdisk
			csp.Spec.Group = append(csp.Spec.Group, newGroup)
		}
		enqueueAddOperation(csp, deviceIDs)
		pc.updatePoolConfig(*csp)
	}
}

// updatePoolConfig updates the csp in poolconfig object, with the updated csp object(passed as an argument).
func (pc *PoolConfig) updatePoolConfig(csp apis.CStorPool) {
	for i, cspGot := range pc.cspList.Items {
		if cspGot.Name == csp.Name {
			pc.cspList.Items[i] = csp
		}
	}
}

// replaceDisk takes disk name as an input and put this disk in place of a removed disk on CSP.
func (pc *PoolConfig) replaceDisk(diskName string) {
	for i, csp := range pc.cspList.Items {
		for j, group := range csp.Spec.Group {
			for k, disk := range group.Item {
				if disk.InUseByPool == false {
					pc.cspList.Items[i].Spec.Group[j].Item[k].Name = diskName
					pc.cspList.Items[i].Spec.Group[j].Item[k].InUseByPool = true
				}
			}
		}
	}
}

// setInUseByPool takes disk name as an input and mark it true/false based on passed inUseTruthy argument.
func (pc *PoolConfig) setInUseByPool(diskName string, inUseTruthy bool) {
	for i, csp := range pc.cspList.Items {
		for j, group := range csp.Spec.Group {
			for k, disk := range group.Item {
				if disk.Name == diskName {
					pc.cspList.Items[i].Spec.Group[j].Item[k].InUseByPool = inUseTruthy
				}
			}
		}
	}
}

// getnodeDiskMap returns a map with key as hostname and value as details of disk attached to the node.
func (pc *PoolConfig) getnodeDiskMap(disks []string) map[string][]DiskDetails {
	nodeDiskMap := make(map[string][]DiskDetails)
	for _, disk := range disks {
		gotDisk, err := pc.controller.clientset.OpenebsV1alpha1().Disks().Get(disk, metav1.GetOptions{})
		if err != nil {
			return nil
		}
		if gotDisk == nil {
			return nil
		}
		devID := getDeviceID(gotDisk)
		disk := &DiskDetails{
			DiskName: gotDisk.Name,
			DeviceID: devID,
		}
		nodeDiskMap[gotDisk.Labels[string(apis.HostNameCPK)]] = append(nodeDiskMap[gotDisk.Labels[string(apis.HostNameCPK)]], *disk)
	}
	return nodeDiskMap
}

// getCspNodeMap returns a map with key as hostname and value as the csp object on that host.
func (pc *PoolConfig) getCspNodeMap() map[string]*apis.CStorPool {
	cspNodeMap := make(map[string]*apis.CStorPool)
	for _, csp := range pc.cspList.Items {
		// Pinning the variable to avoid scope lint issues.
		cspCopy := csp
		cspNodeMap[csp.Labels[string(apis.HostNameCPK)]] = &cspCopy
	}
	return cspNodeMap
}

// getSpcDisks returns map of spc disks present on SPC.
func (pc *PoolConfig) getSpcDisks() map[string]bool {
	// Make a map containing all the disks present in spc.
	spcDisks := make(map[string]bool)
	for _, disk := range pc.spc.Spec.Disks.DiskList {
		spcDisks[disk] = true
	}
	return spcDisks
}

// getCspDisks returns a map of disks present on CSPs.
func (pc *PoolConfig) getCspDisks() map[string]bool {
	// Make a map containing all the disks present in csp
	cspDisks := make(map[string]bool)
	for _, csp := range pc.cspList.Items {
		for _, group := range csp.Spec.Group {
			for _, disk := range group.Item {
				cspDisks[disk.Name] = true
			}
		}
	}
	return cspDisks
}

// getDettachedCspDisks returns a map which tells whether a disk is detached but present on SPC.
func (pc *PoolConfig) getDettachedCspDisks() map[string]bool {
	// Make a map containing all the disks present in csp which in not present in SPC.
	spcDisks := pc.getSpcDisks()
	cspDisks := make(map[string]bool)
	for _, csp := range pc.cspList.Items {
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
func (pc *PoolConfig) getRemovedDisks() []string {
	var removedDisk []string
	// Get the disks present on CSPs
	cspDisks := pc.getCspDisks()

	// Get the disk present on SPC
	spcDisks := pc.getSpcDisks()
	for disk := range cspDisks {
		if spcDisks[disk] == false {
			removedDisk = append(removedDisk, disk)
		}
	}
	return removedDisk
}

// getAddedDisks returns a list of disks that is added to SPC.
func (pc *PoolConfig) getAddedDisks() []string {
	var addedDisk []string
	// get the disks present on CSPs
	cspDisks := pc.getCspDisks()
	// get the disk present on SPC
	spcDisks := pc.getSpcDisks()
	for disk := range spcDisks {
		if cspDisks[disk] == false {
			addedDisk = append(addedDisk, disk)
		}
	}
	return addedDisk
}

// updateCspList updates the modified csp in poolconfig to upstream csp (k8s etcd) and puts the new csp(with changed RV)
// back into poolconfig.
func (pc *PoolConfig) updateCspList() error {
	for i, csp := range pc.cspList.Items {
		cspCopy := csp
		cspGot, err := pc.controller.updateCsp(&cspCopy)
		if err != nil {
			return errors.Wrapf(err, "failed to update csp %s for disk replacement operations for spc %s", csp.Name, pc.spc.Name)
		}
		pc.cspList.Items[i] = *cspGot
	}
	return nil
}

// TODO: Move to CSP Package
func (c *Controller) patchSpcWithDiskHash(spc *apis.StoragePoolClaim) (*apis.StoragePoolClaim, error) {

	diskHash, _ := hash.Hash(spc.Spec.Disks)
	spcPatch := make([]Patch, 1)
	spcPatch[0].Op = PatchOperationReplace
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
	if err != nil {
		return nil, errors.Wrapf(err, "failed to marshal spc patch data")
	}
	spcGot, err := c.clientset.OpenebsV1alpha1().StoragePoolClaims().Patch(spc.Name, types.JSONPatchType, spcPatchJSON)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to patch spc %s with the new disk list hash", spc.Name)
	}

	return spcGot, nil
}

// TODO: Move to CSP Package
func (c *Controller) getCsp(spc *apis.StoragePoolClaim) (*apis.CStorPoolList, error) {
	cspList, err := c.clientset.OpenebsV1alpha1().CStorPools().List(metav1.ListOptions{LabelSelector: string(apis.StoragePoolClaimCPK) + "=" + spc.Name})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get csp objects for spc %s", spc.Name)
	}
	return cspList, nil

}

// TODO: Move to CSP package
func (c *Controller) updateCsp(csp *apis.CStorPool) (*apis.CStorPool, error) {
	csp, err := c.clientset.OpenebsV1alpha1().CStorPools().Update(csp)
	return csp, err
}

// TODO: Move to CSP package
func (c *Controller) deleteCsp(csp *apis.CStorPool) error {
	err := c.clientset.OpenebsV1alpha1().CStorPools().Delete(csp.Name, &metav1.DeleteOptions{})
	return err
}

// TODO: Logic for other pool topologies
func isTopVdevLost(csp *apis.CStorPool) bool {
	for _, group := range csp.Spec.Group {
		removedDiskCount := removedDiskCountInGroup(group)
		if isPoolTopVdevLost(csp, removedDiskCount) {
			return true
		}
	}
	return false
}

func removedDiskCountInGroup(group apis.DiskGroup) int {
	count := 0
	for _, disk := range group.Item {
		if disk.InUseByPool == false {
			count++
		}
	}
	return count
}

func isPoolTopVdevLost(csp *apis.CStorPool, removedDiskCountInGroup int) bool {
	if removedDiskCountInGroup >= nodeselect.DefaultDiskCount[csp.Spec.PoolSpec.PoolType] {
		return true
	}
	return false
}

// TODO: Move to disk package
// getDeviceID returns the device ID of disk in case deviceID is present else device path.
func getDeviceID(disk *apis.Disk) string {
	var deviceID string
	if len(disk.Spec.DevLinks) != 0 && len(disk.Spec.DevLinks[0].Links) != 0 {
		deviceID = disk.Spec.DevLinks[0].Links[0]
	} else {
		deviceID = disk.Spec.Path
	}
	return deviceID
}

// enqueueAddOperation inserts a add operation  into csp object.
func enqueueAddOperation(csp *apis.CStorPool, deviceIDs []string) *apis.CStorPool {
	newAdpcperation := &apis.CstorOperation{
		Action:   apis.PoolExpandAction,
		Status:   apis.PoolOperationStatusInit,
		NewDisks: deviceIDs,
	}
	csp.Operations = append(csp.Operations, *newAdpcperation)
	return csp
}

// enqueueDeleteOperation inserts a delete operation  into csp object.
func enqueueDeleteOperation(csp *apis.CStorPool) *apis.CStorPool {
	newDeleteOperation := &apis.CstorOperation{
		Action: apis.PoolDeleteAction,
		Status: apis.PoolOperationStatusInit,
	}
	csp.Operations = append(csp.Operations, *newDeleteOperation)
	return csp
}
