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
	ndmapis "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	disk "github.com/openebs/maya/pkg/disk/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NodeDiskSelector selects a node and disks attached to it.
func (ac *Config) NodeDiskSelector() (*nodeDisk, error) {
	listDisk, err := ac.getDisk()
	if err != nil {
		return nil, err
	}
	if listDisk == nil || len(listDisk.Items) == 0 {
		return nil, errors.New("no disk object found")
	}
	nodeDiskMap, err := ac.getCandidateNode(listDisk)
	if err != nil {
		return nil, err
	}
	selectedDisk := ac.selectNode(nodeDiskMap)

	return selectedDisk, nil
}

// getUsedDiskMap gives list of disks that has already been used for pool provisioning.
func (ac *Config) getUsedDiskMap() (map[string]int, error) {
	// Get the list of disk that has been used already for pool provisioning
	cspList, err := ac.CspClient.List(v1.ListOptions{})

	if err != nil {
		return nil, errors.Wrapf(err, "unable to get the list of storagepools")
	}
	usedDiskMap := make(map[string]int)
	for _, csp := range cspList.CStorPoolList.Items {
		for _, group := range csp.Spec.Group {
			for _, disk := range group.Item {
				usedDiskMap[disk.Name]++
			}
		}
	}

	return usedDiskMap, nil
}

// getUsedNodeMap form a used node map to keep a track of nodes on the top of which storagepool cannot be provisioned
// for a given storagepoolclaim.
func (ac *Config) getUsedNodeMap() (map[string]int, error) {
	cspList, err := ac.CspClient.List(v1.ListOptions{LabelSelector: string(apis.StoragePoolClaimCPK) + "=" + ac.Spc.Name})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get the list of storagepools for stragepoolclaim %s", ac.Spc.Name)
	}
	usedNodeMap := make(map[string]int)
	for _, sp := range cspList.CStorPoolList.Items {
		usedNodeMap[sp.Labels[string(apis.HostNameCPK)]]++
	}
	return usedNodeMap, nil
}

func (ac *Config) getCandidateNode(listDisk *ndmapis.DiskList) (map[string]*diskList, error) {
	usedDiskMap, err := ac.getUsedDiskMap()
	if err != nil {
		return nil, err
	}
	usedNodeMap, err := ac.getUsedNodeMap()
	if err != nil {
		return nil, err
	}
	// nodeDiskMap is the data structure holding host name as key
	// and nodeDisk struct as value.
	nodeDiskMap := make(map[string]*diskList)
	for _, value := range listDisk.Items {
		// If the disk is already being used, do not consider this as a part for provisioning pool.
		if usedDiskMap[value.Name] == 1 {
			continue
		}
		// If the node is already being used for a given spc, do not consider this as a part for provisioning pool.
		if usedNodeMap[value.Labels[string(apis.HostNameCPK)]] == 1 {
			continue
		}
		if nodeDiskMap[value.Labels[string(apis.HostNameCPK)]] == nil {
			// Entry to this block means first time the hostname will be mapped for the first time.
			// Obviously, this entry of hostname(node) is for a usable disk and initialize diskCount to 1.
			nodeDiskMap[value.Labels[string(apis.HostNameCPK)]] = &diskList{Items: []string{value.Name}}
		} else {
			// Entry to this block means the hostname was already mapped and it has more than one disk and at least two disks.
			nodeDisk := nodeDiskMap[value.Labels[string(apis.HostNameCPK)]]
			// Add the current disk to the diskList for this node.
			nodeDisk.Items = append(nodeDisk.Items, value.Name)
		}
	}
	return nodeDiskMap, nil
}

func (ac *Config) selectNode(nodeDiskMap map[string]*diskList) *nodeDisk {
	// selectedDisk will hold a node capable to form pool and list of disks attached to it.
	selectedDisk := &nodeDisk{
		NodeName: "",
		Disks: diskList{
			Items: []string{},
		},
	}
	// diskCount will hold the number of disk that will be selected from a qualified
	// node for specific pool type
	var diskCount int
	// minRequiredDiskCount will hold the required number of disk that should be selected from a qualified
	// node for specific pool type
	minRequiredDiskCount := DefaultDiskCount[ac.poolType()]
	for node, val := range nodeDiskMap {
		// If the current disk count on the node is less than the required disks
		// then this is a dirty node and it will not qualify.
		if len(val.Items) < minRequiredDiskCount {
			continue
		}
		diskCount = minRequiredDiskCount
		if ProvisioningType(ac.Spc) == ProvisioningTypeManual {
			diskCount = (len(val.Items) / minRequiredDiskCount) * minRequiredDiskCount
		}
		for i := 0; i < diskCount; i++ {
			selectedDisk.Disks.Items = append(selectedDisk.Disks.Items, val.Items[i])
		}
		selectedDisk.NodeName = node
		break
	}
	return selectedDisk
}

// diskFilterConstraint takes a value for key "ndm.io/disk-type" and form a label.
func diskFilterConstraint(diskType string) string {
	var label string
	if diskType == string(apis.TypeSparseCPV) {
		label = string(apis.NdmDiskTypeCPK) + "=" + string(apis.TypeSparseCPV)
	} else {
		label = string(apis.NdmDiskTypeCPK) + "=" + string(apis.TypeDiskCPV)
	}
	return label
}

// ProvisioningType returns the way pool should be provisioned e.g. auto or manual.
func ProvisioningType(spc *apis.StoragePoolClaim) string {
	if len(spc.Spec.Disks.DiskList) == 0 {
		return ProvisioningTypeAuto
	}
	return ProvisioningTypeManual
}

// poolType returns the type of pool provisioning e.g. mirrored or striped for a given spc.
func (ac *Config) poolType() string {
	if ac.Spc.Spec.PoolSpec.PoolType == string(apis.PoolTypeMirroredCPV) {
		return string(apis.PoolTypeMirroredCPV)
	}
	if ac.Spc.Spec.PoolSpec.PoolType == string(apis.PoolTypeStripedCPV) {
		return string(apis.PoolTypeStripedCPV)
	}
	if ac.Spc.Spec.PoolSpec.PoolType == string(apis.PoolTypeRaidzCPV) {
		return string(apis.PoolTypeRaidzCPV)
	}
	if ac.Spc.Spec.PoolSpec.PoolType == string(apis.PoolTypeRaidz2CPV) {
		return string(apis.PoolTypeRaidz2CPV)
	}
	return ""
}

// getDisk return the all disks of a certain type(e.g. sparse, disk) which is specified in spc.
func (ac *Config) getDisk() (*ndmapis.DiskList, error) {
	diskFilterLabel := diskFilterConstraint(ac.Spc.Spec.Type)
	dL, err := ac.DiskClient.List(v1.ListOptions{LabelSelector: diskFilterLabel})
	if err != nil {
		return nil, err
	}
	dl := dL.Filter(disk.FilterInactiveReverse)
	return dl.DiskList, nil
}
