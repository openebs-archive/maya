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

package v1alpha1

import (
	apisv1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	caspool "github.com/openebs/maya/pkg/caspool/v1alpha1"
	disk "github.com/openebs/maya/pkg/disk/v1alpha2"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
)

/*

Following is a sample auto CSPC YAML:

kind: CStorPoolCluster
metadata:
  name: cstor-sparse-pool-test
spec:
  maxPools: 1
  poolSpec:
   poolType: striped
  name: cstor-sparse-pool-test
  type: sparse

*/
const (
	// DefaultDiskCountMultiplier is the disk count multiplier
	// For example
	// for striped case, number of disk = 1* DefaultDiskCountMultiplier
	// for mirrored case, number of disk = 2* DefaultDiskCountMultiplier
	// for raidz case, number of disk = 3* DefaultDiskCountMultiplier
	// TODO: Move this in Operations struct, once auto CSPC descriptor supports disk multiplier
	DefaultDiskCountMultiplier = 1
)

// nodeFilterPredicate is the filter predicates for node.
type nodeFilterPredicate func(nodeName string) bool

// nodeFilterPredicateList is the list of nodeFilterPredicate
type nodeFilterPredicateList []nodeFilterPredicate

// getCasPoolForAutoProvisioning returns a CasPool object for auto provisioning of CSPC.
func (op *Operations) getCasPoolForAutoProvisioning() (*apisv1alpha1.CasPool, error) {
	// Validate the CSPC object encapsulated in operations.
	err := op.validateAuto()

	if err != nil {
		return nil, errors.Wrap(err, "validation for auto cspc failed")
	}

	// Get the eligible node and disk associated with this node for  pool creation(rather instead of pool creation,
	// call it pool operation as actions other than creation can also be possible).
	nodeName, diskList, err := op.getEligibleDisks()

	// With the available disk list, for a disk group
	diskGroups := op.getDiskGroups(diskList)
	if err != nil {
		return nil, errors.Wrapf(err, "could not get usable node for spc %s", op.CspcObject.Object.Name)
	}

	// get the device ID for the disk list
	// NOTE: Actually, for disk, we have dev link in case of physical disk and dev path for sparse disk
	// We try to get dev link first and if it is not present we get dev path.
	// Just for naming convention we call it device ID (or disk device ID)which can imply to either dev link or dev path.
	diskDeviceIDMap, err := op.getDiskDeviceIDMapForDiskAPIList(diskList)
	if err != nil {
		return nil, errors.Wrapf(err, "could not form disk device ID map for %s", op.CspcObject.Object.Name)
	}

	spcObject := op.CspcObject.Object
	cp := caspool.NewBuilder().
		WithDiskType(spcObject.Spec.Type).
		WithPoolType(op.CspcObject.Object.Spec.PoolSpec.PoolType).
		WithAnnotations(op.CspcObject.GetAnnotations()).
		WithDiskGroup(diskGroups).
		WithCasTemplateName(op.CspcObject.GetCASTName()).
		WithCspcName(op.CspcObject.Object.Name).
		WithNodeName(nodeName).
		WithDiskDeviceIDMap(diskDeviceIDMap).
		Build().Object
	return cp, nil
}

// validateAuto validates auto CSPC
func (op *Operations) validateAuto() error {
	err := NewOperationsBuilderForObject(op).
		WithCheckf(IsMaxPoolNotNil(), "max pool is nil").
		WithCheckf(IsPoolTypeValid(), "pool type is not valid").
		WithCheckf(IsTypeValid(), "disk type is not valid").
		Validate()
	if err != nil {
		return errors.Wrapf(err, "validation for cstorpoolcluster %s failed", op.CspcObject.Object.Name)
	}
	return nil
}

// getDisks returns a list of disk
// The returned disk is
// 1. Active
// 2. Not used in other cstor-pool
// 3. is not attached to a node where a cstor-pool has already been created for the given CSPC.
func (op *Operations) getDisks() (*apisv1alpha1.DiskList, error) {
	diskAPIList, err := op.DiskClient.List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "could not list disk")
	}

	usedDiskMap, err := op.getUsedDiskMap()
	if err != nil {
		return nil, errors.Wrap(err, "could not get used disk map")
	}

	usedNodeMap, err := op.getUsedNode()
	if err != nil {
		return nil, errors.Wrap(err, "could not get used node map")
	}

	newDiskAPIList := disk.NewListBuilderForAPIList(diskAPIList).DiskList.
		Filter(disk.IsActive(), disk.IsType(op.CspcObject.Object.Spec.Type), disk.IsUsable(usedDiskMap), disk.IsUsableNode(usedNodeMap)).
		ObjectList

	if len(newDiskAPIList.Items) == 0 {
		return nil, errors.New("no disks found")
	}

	return newDiskAPIList, nil
}

// getNodeDisk returns a map of node and attached disk.
// The returned node disk
// 1. Obeys pool topology strictly with no spares
// e.g.
// for striped number of disk >1
// for mirrored number of disk in a group = 2, number of group = k, total no. of disk - 2*k
// Similarly, for other raid groups
func (op *Operations) getNodeDisk() (map[string]*apisv1alpha1.DiskList, error) {
	diskAPIList, err := op.getDisks()

	if err != nil {
		return nil, errors.Wrap(err, "could not get usable disk")
	}
	// TODO: Think of nodeDiskMap as a internal structure and apply creational and filter pattern
	nodeDiskMap := make(map[string]*apisv1alpha1.DiskList)
	for _, diskAPIObject := range diskAPIList.Items {
		// pin it
		diskAPIObject := diskAPIObject
		nodeName := disk.BuilderForAPIObject(&diskAPIObject).Disk.GetNodeName()
		if nodeDiskMap[nodeName] == nil {
			nodeDiskMap[nodeName] = &apisv1alpha1.DiskList{Items: []apisv1alpha1.Disk{diskAPIObject}}
		} else {
			nodeDiskMap[nodeName].Items = append(nodeDiskMap[nodeName].Items, diskAPIObject)
		}
	}
	filteredNodediskMap := filterNodeDiskMap(nodeDiskMap, isTopologyObeyed(nodeDiskMap, op.CspcObject.Object.Spec.PoolSpec.PoolType))

	if len(filteredNodediskMap) == 0 {
		return nil, errors.New("no node found with required number of disks")
	}
	return filteredNodediskMap, nil
}

// getEligibleDisks disk returns the eligible disk for pool creation
// currently, the first encountered eligible node with its disk list is returned.
// TODO: Think of priority functions while choosing for eligible nodes
func (op *Operations) getEligibleDisks() (string, *apisv1alpha1.DiskList, error) {
	nodeDiskMap, err := op.getNodeDisk()
	if err != nil {
		return "", nil, errors.Wrapf(err, "failed to get eligible disk")
	}

	for nodeName, diskList := range nodeDiskMap {
		return nodeName, diskList, nil
	}
	return "", nil, errors.Wrapf(err, "got empty node disk map")
}

// all returns true if all the predicates succeed against the provided disk instance.
func (l nodeFilterPredicateList) allFilterPredicates(nodeName string) bool {
	for _, pred := range l {
		if !pred(nodeName) {
			return false
		}
	}
	return true
}

// filterNodeDiskMap will filter the nodes based on provided predicates.
func filterNodeDiskMap(nodeDiskMap map[string]*apisv1alpha1.DiskList, p ...nodeFilterPredicate) map[string]*apisv1alpha1.DiskList {
	var plist nodeFilterPredicateList
	plist = append(plist, p...)
	if len(plist) == 0 {
		return nodeDiskMap
	}

	filtered := make(map[string]*apisv1alpha1.DiskList)
	for node, diskAPIList := range nodeDiskMap {
		diskAPIList := diskAPIList // pin it
		if plist.allFilterPredicates(node) {
			filtered[node] = diskAPIList
		}
	}
	return filtered
}

// isTopologyObeyed returns a predicate that tells whether topology is obeyed for raid group or not
// For example,
// STRIPED -- Disk Count <1 , topology is not obeyed.
// Mirrored -- Disk Count =2*n , n>0, topology is obeyed.
// Note : Topology obeying decision is cStor specific.
func isTopologyObeyed(nodeDiskMap map[string]*apisv1alpha1.DiskList, poolType string) nodeFilterPredicate {
	return func(nodeName string) bool {
		requiredDiskCount := getRequiredDiskCount(poolType)
		if len(nodeDiskMap[nodeName].Items) >= requiredDiskCount {
			return true
		}
		return false
	}
}

// getDiskGroups takes a list of disk as an argument and forms disk group according to specified raid group.
func (op *Operations) getDiskGroups(diskList *apisv1alpha1.DiskList) []apisv1alpha1.CStorPoolClusterDiskGroups {
	var cspcDiskGroup []apisv1alpha1.CStorPoolClusterDiskGroups
	var newDiskList []apisv1alpha1.CStorPoolClusterDisk
	diskObject := disk.NewListBuilderForAPIList(diskList).List()
	groupCount := op.getGroupCount(diskObject)
	diskCountInGroup := op.getDiskCountInGroup(diskObject)
	diskIndex := 0

	for i := 0; i < groupCount; i++ {
		for j := 0; j < diskCountInGroup; j++ {
			newSpcDisk := apisv1alpha1.CStorPoolClusterDisk{
				Name: diskList.Items[diskIndex].Name,
				// TODO: Generate disk ID : (day2ops)
				ID: "",
			}
			newDiskList = append(newDiskList, newSpcDisk)
			diskIndex++
		}
		newDiskGroup := apisv1alpha1.CStorPoolClusterDiskGroups{
			Name:  "group" + strconv.Itoa(i),
			Disks: newDiskList,
		}
		cspcDiskGroup = append(cspcDiskGroup, newDiskGroup)
	}
	return cspcDiskGroup
}

// getRequiredDiskCount returns the required number of disk for a given raid specification.
func getRequiredDiskCount(poolType string) int {
	return disk.DefaultDiskCount[poolType] * DefaultDiskCountMultiplier
}

func (op *Operations) getDiskCountInGroup(diskList *disk.DiskList) int {
	if op.CspcObject.IsPoolTypeStriped() {
		return DefaultDiskCountMultiplier
	}
	return disk.DefaultDiskCount[op.CspcObject.GetPoolType()]
}

func (op *Operations) getGroupCount(diskList *disk.DiskList) int {
	if op.CspcObject.IsPoolTypeStriped() {
		return 1
	}
	return getRequiredDiskCount(op.CspcObject.GetPoolType()) / disk.DefaultDiskCount[op.CspcObject.GetPoolType()]
}
