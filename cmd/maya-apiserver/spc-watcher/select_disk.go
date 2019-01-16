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
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	openebs "github.com/openebs/maya/pkg/client/generated/clientset/internalclientset"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
)

const (
	// DiskStateActive is the active state of the disks.
	DiskStateActive        = "Active"
	ProvisioningTypeManual = "manual"
	ProvisioningTypeAuto   = "auto"
)

// clientset struct holds the interface of internalclientset
// i.e. openebs.
// This struct will be binded to method ListDisk and is useful in mocking
// and unit testing.
type clientSet struct {
	oecs openebs.Interface
}

type diskList struct {
	//diskList is the list of usable disks that can be used in storagepool provisioning.
	items []string
}

type nodeDisk struct {
	nodeName string
	disks    diskList
}

func (k *clientSet) nodeDiskAlloter(cp *v1alpha1.StoragePoolClaim) (*nodeDisk, error) {
	// Request kube-apiserver for the list of disk (powered by NDM)
	// Currently, all the disks are returned,but the disk that is already a part of pool
	// should not be returned.
	listDisk, err := k.getDisk(cp)
	if err != nil {
		return nil, errors.Errorf("error in getting the disk list:%v", err)
	}
	if len(listDisk.Items) == 0 {
		return nil, errors.New("no disk object found")
	}
	var provisioningType string
	if len(cp.Spec.Disks.DiskList) == 0 {
		provisioningType = ProvisioningTypeAuto
	} else {
		provisioningType = ProvisioningTypeManual
	}
	// pendingAllotment holds the number of pools that will be pending to be provisioned.
	nodeDiskMap, err := k.nodeSelector(listDisk, cp.Spec.PoolSpec.PoolType, cp.Name)
	if err != nil {
		return nil, err
	}
	selectedDisk := diskSelector(nodeDiskMap, cp.Spec.PoolSpec.PoolType, provisioningType)
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

func (k *clientSet) nodeSelector(listDisk *v1alpha1.DiskList, poolType string, spc string) (map[string]*diskList, error) {

	usedDiskMap, err := k.getUsedDiskMap()
	if err != nil {
		return nil, err
	}
	usedNodeMap, err := k.getUsedNodeMap(spc)
	if err != nil {
		return nil, err
	}
	// nodeDiskMap is the data structure holding host name as key
	// and nodeDisk struct as value
	nodeDiskMap := make(map[string]*diskList)
	for _, value := range listDisk.Items {

		// If the disk is already being used, do not consider this as a part for provisioning pool
		if usedDiskMap[value.Name] == 1 {
			continue
		}
		// If the disk not Active, do not consider this as a part for provisioning pool
		if value.Status.State != DiskStateActive {
			continue
		}
		if usedNodeMap[value.Labels[string(v1alpha1.HostNameCPK)]] == 1 {
			continue
		}

		if nodeDiskMap[value.Labels[string(v1alpha1.HostNameCPK)]] == nil {
			// Entry to this block means first time the hostname will be mapped for the first time.
			// Obviously, this entry of hostname(node) is for a usable disk and initialize diskCount to 1.
			nodeDiskMap[value.Labels[string(v1alpha1.HostNameCPK)]] = &diskList{items: []string{value.Name}}
		} else {
			// Entry to this block means the hostname was already mapped and it has more than one disk and at least two disks.
			nodeDisk := nodeDiskMap[value.Labels[string(v1alpha1.HostNameCPK)]]
			// Add the current disk to the diskList for this node.
			nodeDisk.items = append(nodeDisk.items, value.Name)
		}

	}
	return nodeDiskMap, nil
}

// diskSelector is the function that will select the required number of disks from qualified nodes
// so as to provision storagepool.
func diskSelector(nodeDiskMap map[string]*diskList, poolType, provisioningType string) *nodeDisk {

	// selectedDisk will hold a list of disk that will be used to provision storage pool, after a
	// minimum number of node qualifies
	selectedDisk := &nodeDisk{
		nodeName: "",
		disks: diskList{
			items: []string{},
		},
	}

	// diskCount will hold the number of disk that will be selected from a qualified
	// node for specific pool type
	var diskCount int
	// minRequiredDiskCount will hold the required number of disk that should be selected from a qualified
	// node for specific pool type
	var minRequiredDiskCount int
	// If pool type is striped, at least 1 disk should be selected
	if poolType == string(v1alpha1.PoolTypeStripedCPV) {
		minRequiredDiskCount = int(v1alpha1.StripedDiskCountCPV)
	}
	// If pool type is mirrored, at least 2 disks should be selected
	if poolType == string(v1alpha1.PoolTypeMirroredCPV) {
		minRequiredDiskCount = int(v1alpha1.MirroredDiskCountCPV)
	}
	// Range over the nodeDiskMap map to get the list of disks
	for node, val := range nodeDiskMap {

		// If the current disk count on the node is less than the required disks
		// then this is a dirty node and it will not qualify.
		if len(val.items) < minRequiredDiskCount {
			continue
		}
		if provisioningType == ProvisioningTypeManual {
			diskCount = len(val.items)
		}
		if provisioningType == ProvisioningTypeAuto {
			diskCount = minRequiredDiskCount
		}
		// Select the required disk from qualified nodes.
		if poolType == string(v1alpha1.PoolTypeStripedCPV) {
			for i := 0; i < diskCount; i++ {
				selectedDisk.disks.items = append(selectedDisk.disks.items, val.items[i])
			}
		}
		if poolType == string(v1alpha1.PoolTypeMirroredCPV) {
			for i := 0; i < diskCount/2*2; i = i + 2 {
				selectedDisk.disks.items = append(selectedDisk.disks.items, val.items[i])
				selectedDisk.disks.items = append(selectedDisk.disks.items, val.items[i+1])
			}
		}
		selectedDisk.nodeName = node
		break
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

// Form usedDisk map that will hold the list of all used disks
func (k *clientSet) getUsedDiskMap() (map[string]int, error) {
	// Get the list of disk that has been used already for pool provisioning
	spList, err := k.oecs.OpenebsV1alpha1().StoragePools().List(mach_apis_meta_v1.ListOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get the list of storagepools")
	}
	// Form a map that will hold all the used disk
	usedDiskMap := make(map[string]int)
	for _, sp := range spList.Items {
		for _, usedDisk := range sp.Spec.Disks.DiskList {
			usedDiskMap[usedDisk]++
		}

	}
	return usedDiskMap, nil
}

// Form usedNode map to keep a track of nodes on the top of which storagepool cannot be provisioned for a
// given storagepoolcalim
func (k *clientSet) getUsedNodeMap(spc string) (map[string]int, error) {
	// Get the list of storagepool
	spList, err := k.oecs.OpenebsV1alpha1().StoragePools().List(mach_apis_meta_v1.ListOptions{LabelSelector: string(v1alpha1.StoragePoolClaimCPK) + "=" + spc})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get the list of storagepools for stragepoolclaim %s", spc)
	}
	// Form a map that will hold all the nodes where storagepool for the spc has been already created.
	usedNodeMap := make(map[string]int)
	for _, sp := range spList.Items {
		usedNodeMap[sp.Labels[string(v1alpha1.HostNameCPK)]]++
	}
	return usedNodeMap, nil
}

func (k *clientSet) getDisk(cp *v1alpha1.StoragePoolClaim) (*v1alpha1.DiskList, error) {
	diskFilterLabel := diskFilterConstraint(cp.Spec.Type)
	// Request kube-apiserver for the list of disk (powered by NDM)
	// Currently, all the disks are returned,but the disk that is already a part of pool
	// should not be returned.
	listDisk, err := k.oecs.OpenebsV1alpha1().Disks().List(mach_apis_meta_v1.ListOptions{LabelSelector: diskFilterLabel})
	if len(cp.Spec.Disks.DiskList) == 0 {
		return listDisk, err
	}
	listDisk = &v1alpha1.DiskList{
		Items: []v1alpha1.Disk{},
	}
	spcDisks := cp.Spec.Disks.DiskList
	for _, v := range spcDisks {
		getDisk, err := k.oecs.OpenebsV1alpha1().Disks().Get(v, mach_apis_meta_v1.GetOptions{})
		if err != nil {
			runtime.HandleError(errors.Wrapf(err, "Error in fetching disk"))
		} else {
			// Deep-copy not required unless the object internal fields of objects are pointer referenced.
			listDisk.Items = append(listDisk.Items, *getDisk)
		}
	}
	return listDisk, nil
}
