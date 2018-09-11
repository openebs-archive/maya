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
	mach_apis_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//openebs "github.com/openebs/maya/pkg/client/clientset/versioned"
	"errors"
	"fmt"
	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	openebs "github.com/openebs/maya/pkg/client/generated/clientset/internalclientset"
)

// clientset struct holds the interface of internalclientset
// i.e. openebs.
// This struct will be binded to method ListDisk and is useful in mocking
// and unit testing.
type clientSet struct {
	oecs openebs.Interface
}

// nodeDisk struct will be used as a value for a map nodeDiskMap (map defined in ListDisk function)
// The struct will be useful in forming the data structure nodeDiskMap which will be manipulated
// to efficiently select the nodes and disk for dynamic pool provisioning.

// The struct can incorporate several other constraints(that might come in future)
// related to disk that will be useful in selecting disks
type nodeDisk struct {
	//diskList is the list of usable disks that can be used in storagepool provisioning.
	diskList []string
}

// nodeDiskAlloter will try to allot nodes for pool creation as specified in
// maxPool field of the storagepoolclaim and return a list of selected disks from
// those selected nodes.

// For exapmle, if maxPool=5 and minPool=3, it will try to search for 5 nodes that will qualify for
// pool provisioning. At least 3 node should qualify else pool will not be provisioned and pool creation
// will be aborted gracefully with proper log messages.

// If no minPool field is present,at least one node must qualify for pool provisioning.

// nodeDiskAlloter can be made more intelligent as per the required pool constraints for alloting nodes.
func (k *clientSet) nodeDiskAlloter(cp *v1alpha1.CasPool) ([]string, error) {

	// pendingAllotment holds the value for the remaining node allotments
	// for pool provisioning.
	var pendingAllotment int

	// assign maxPools to pendingAllotment as right now maxPool is the number of allotments
	// that needs to be done.
	pendingAllotment = cp.MaxPools
	// get the labels on the basis of which disk list will be filtered
	diskFilterLabel := diskFilterConstraint(cp.Type)

	// Request kube-apiserver for the list of disk (powered by NDM)
	// Currently, all the disks are returned,but the disk that is already a part of pool
	// should not be returned.
	listDisk, err := k.oecs.OpenebsV1alpha1().Disks().List(mach_apis_meta_v1.ListOptions{LabelSelector: diskFilterLabel})
	if err != nil {
		return nil, fmt.Errorf("error in getting the disk list:%v", err)
	}
	if len(listDisk.Items) == 0 {
		return nil, errors.New("no disk object found")
	}

	// pendingAllotment holds the number of pools that will be pending to be provisioned.
	err, nodeDiskMap, pendingAllotment := k.nodeSelector(listDisk, cp.PoolType, cp.StoragePoolClaim, pendingAllotment)
	if err != nil {
		return nil, err
	}
	// gotAllotment is the count of nodes where storagepool can be provisioned
	gotAllotment := cp.MaxPools - pendingAllotment
	if gotAllotment < cp.MinPools {
		return nil, fmt.Errorf("not enough nodes qualified for pool:only %d node could be alloted but required is %d", gotAllotment, cp.MinPools)
	}
	// if alloted node was less than the maxPool that means partial allotment is done and
	// some allotment is still pending.
	if gotAllotment < cp.MaxPools {
		glog.Warning("partial node allotment done:pending node allotment:", pendingAllotment)
	}

	// diskSelector will get the list of disks from nodeDiskMap by selecting disks from
	// qualified nodes only.
	// diskSelector rejects a dirty node which will not qualify for pool creation.
	selectedDisk := diskSelector(nodeDiskMap, cp.PoolType)
	return selectedDisk, nil
}

// nodeSelector function will select candidate nodes that will qualify for storagepool provisioning in accordance
// with the pool constraints.

// nodeSelector will basically try to form a map by iterating over the list.
// During the formation of map, if the pendingAllotment can be done with the current map data,
// it will stop iterating over disks.
// During this map formation some nodes can enter into map which is not capable of forming the pool
// this node is dirty node and will be rejected by diskSelector while selecting disk.

// NOTE: Not all the selected nodes may qualify. These nodes are dirty nodes

// Finally diskSelector function will vote for qualified nodes.

func (k *clientSet) nodeSelector(listDisk *v1alpha1.DiskList, poolType string, spc string, pendingAllotment int) (error, map[string]*nodeDisk, int) {

	err, usedDiskMap := k.getUsedDiskMap()
	if err != nil {
		return err, nil, pendingAllotment
	}
	err, usedNodeMap := k.getUsedNodeMap(spc)
	if err != nil {
		return err, nil, pendingAllotment
	}
	// nodeDiskMap is the data structure holding host name as key
	// and nodeDisk struct as value
	nodeDiskMap := make(map[string]*nodeDisk)
	for _, value := range listDisk.Items {

		// If the disk is already being used, do not consider this as a part for provisioning pool
		if usedDiskMap[value.Name] == 1 {
			continue
		}
		if usedNodeMap[value.Labels[string(v1alpha1.HostNameCPK)]] == 1 {
			continue
		}
		// if no more allotment is required, stop processing
		if pendingAllotment == 0 {
			glog.Info("Required pool allotment done")
			break
		}
		if nodeDiskMap[value.Labels[string(v1alpha1.HostNameCPK)]] == nil {
			// Entry to this block means first time the hostname will be mapped for the first time.
			// Obviously, this entry of hostname(node) is for a usable disk and initialize diskCount to 1.
			nodeDiskMap[value.Labels[string(v1alpha1.HostNameCPK)]] = &nodeDisk{diskList: []string{value.Name}}
			// If pool type is striped the node qualifies for pool creation hence pendingAllotment decremented.
			if poolType == string(v1alpha1.PoolTypeStripedCPV) {
				pendingAllotment--
			}
		} else {
			// Entry to this block means the hostname was already mapped and it has more than one disk and at least two disks.
			nodeDisk := nodeDiskMap[value.Labels[string(v1alpha1.HostNameCPK)]]
			// Add the current disk to the diskList for this node.
			nodeDisk.diskList = append(nodeDisk.diskList, value.Name)
			// If pool type is mirrored the node qualifies for pool creation hence pendingAllotment decremented.
			if poolType == string(v1alpha1.PoolTypeMirroredCPV) {
				if len(nodeDisk.diskList) == int(v1alpha1.MirroredDiskCountCPV) {
					pendingAllotment--
				}
			}
		}

	}
	return nil, nodeDiskMap, pendingAllotment
}

// diskSelector is the function that will select the required number of disks from qualified nodes
// so as to provision storagepool

func diskSelector(nodeDiskMap map[string]*nodeDisk, poolType string) []string {

	// selectedDisk will hold a list of disk that will be used to provision storage pool, after a
	// minimum number of node qualifies
	var selectedDisk []string

	// requiredDiskCount will hold the required number of disk that should be selcted from a qualified
	// node for specific pool type
	var requiredDiskCount int
	// If pool type is striped, 1 disk should be selected
	if poolType == string(v1alpha1.PoolTypeStripedCPV) {
		requiredDiskCount = int(v1alpha1.StripedDiskCountCPV)
	}
	// If pool type is mirrored, 2 disks should be selected
	if poolType == string(v1alpha1.PoolTypeMirroredCPV) {
		requiredDiskCount = int(v1alpha1.MirroredDiskCountCPV)
	}
	// Range over the nodeDiskMap map to get the list of disks
	for _, val := range nodeDiskMap {

		// If the current disk count on the node is less than the required disks
		// then this is a dirty node and it will not qualify.
		if len(val.diskList) < requiredDiskCount {
			continue
		}
		// Select the required disk from qualified nodes.
		for i := 0; i < requiredDiskCount; i++ {
			selectedDisk = append(selectedDisk, val.diskList[i])
		}
	}
	return selectedDisk
}

// diskFilterConstraint will form labels that will be used to filter disks
// It will return the labels in which filtering can be done.
// e.g.
// "ndm.io/disk-type=disk" , "ndm.io/disk-type=sparse"
func diskFilterConstraint(diskType string) string {
	var label string
	if diskType == string(v1alpha1.TypeSparseCPV) {
		label = string(v1alpha1.NdmDiskTypeCPK) + "=" + string(v1alpha1.TypeSparseCPV)
	} else {
		label = string(v1alpha1.NdmDiskTypeCPK) + "=" + string(v1alpha1.TypeDiskCPV)
	}
	return label
}

// form usedDisk map that will hold the list of all used disks

func (k *clientSet) getUsedDiskMap() (error, map[string]int) {
	// Get the list of disk that has been used already for pool provisioning
	spList, err := k.oecs.OpenebsV1alpha1().StoragePools().List(mach_apis_meta_v1.ListOptions{})
	if err != nil {
		return fmt.Errorf("unable to get the list of storagepools:%v", err), nil
	}
	// Form a map that will hold all the used disk
	usedDiskMap := make(map[string]int)
	for _, sp := range spList.Items {
		for _, usedDisk := range sp.Spec.Disks.DiskList {
			usedDiskMap[usedDisk]++
		}

	}
	return nil, usedDiskMap
}

// form usedNode map to keep a track of nodes on the top of which storagepool cannot be provisioned for a
// given storagepoolcalim

func (k *clientSet) getUsedNodeMap(spc string) (error, map[string]int) {
	// Get the list of storagepool
	spList, err := k.oecs.OpenebsV1alpha1().StoragePools().List(mach_apis_meta_v1.ListOptions{LabelSelector: string(v1alpha1.StoragePoolClaimCPK) + "=" + spc})
	if err != nil {
		return fmt.Errorf("unable to get the list of storagepools for stragepoolclaim %s:%v", spc, err), nil
	}
	// Form a map that will hold all the nodes where storagepool for the spc has been already created.
	usedNodeMap := make(map[string]int)
	for _, sp := range spList.Items {
		usedNodeMap[sp.Labels[string(v1alpha1.HostNameCPK)]]++
	}
	return nil, usedNodeMap
}
