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
	blockdevice "github.com/openebs/maya/pkg/blockdevice/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NodeBlockDeviceSelector selects a node and disks attached to it.
func (ac *Config) NodeBlockDeviceSelector() (*nodeBlockDevice, error) {
	listBD, err := ac.getBlockDevice()
	if err != nil {
		return nil, err
	}
	if listBD == nil || len(listBD.Items) == 0 {
		return nil, errors.New("no block device object found")
	}
	nodeBlockDeviceMap, err := ac.getCandidateNode(listBD)
	if err != nil {
		return nil, err
	}
	selectedBD := ac.selectNode(nodeBlockDeviceMap)

	return selectedBD, nil
}

// getUsedBlockDeviceMap gives list of disks that has already been used for pool provisioning.
func (ac *Config) getUsedBlockDeviceMap() (map[string]int, error) {
	// Get the list of block devices that has been used already for pool provisioning
	cspList, err := ac.CspClient.List(metav1.ListOptions{})

	if err != nil {
		return nil, errors.Wrapf(err, "unable to get the list of storagepools")
	}
	usedBDMap := make(map[string]int)
	for _, csp := range cspList.CStorPoolList.Items {
		for _, group := range csp.Spec.Group {
			for _, disk := range group.Item {
				usedBDMap[disk.Name]++
			}
		}
	}

	return usedBDMap, nil
}

// getUsedNodeMap form a used node map to keep a track of nodes on the top of which storagepool cannot be provisioned
// for a given storagepoolclaim.
func (ac *Config) getUsedNodeMap() (map[string]int, error) {
	cspList, err := ac.CspClient.List(metav1.ListOptions{LabelSelector: string(apis.StoragePoolClaimCPK) + "=" + ac.Spc.Name})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get the list of storagepools for stragepoolclaim %s", ac.Spc.Name)
	}
	usedNodeMap := make(map[string]int)
	for _, sp := range cspList.CStorPoolList.Items {
		usedNodeMap[sp.Labels[string(apis.HostNameCPK)]]++
	}
	return usedNodeMap, nil
}

func (ac *Config) getCandidateNode(listBlockDevice *ndmapis.BlockDeviceList) (map[string]*blockDeviceList, error) {
	usedDiskMap, err := ac.getUsedBlockDeviceMap()
	if err != nil {
		return nil, err
	}
	usedNodeMap, err := ac.getUsedNodeMap()
	if err != nil {
		return nil, err
	}
	// nodeBlockDeviceMap is the data structure holding host name as key
	// and nodeBlockDevice struct as value.
	nodeBlockDeviceMap := make(map[string]*blockDeviceList)
	for _, value := range listBlockDevice.Items {
		// If the disk is already being used, do not consider this as a part for provisioning pool.
		if usedDiskMap[value.Name] == 1 {
			continue
		}
		// If the node is already being used for a given spc, do not consider this as a part for provisioning pool.
		if usedNodeMap[value.Labels[string(apis.HostNameCPK)]] == 1 {
			continue
		}
		if nodeBlockDeviceMap[value.Labels[string(apis.HostNameCPK)]] == nil {
			// Entry to this block means first time the hostname will be mapped for the first time.
			// Obviously, this entry of hostname(node) is for a usable disk and initialize diskCount to 1.
			nodeBlockDeviceMap[value.Labels[string(apis.HostNameCPK)]] = &blockDeviceList{Items: []string{value.Name}}
		} else {
			// Entry to this block means the hostname was already mapped and it has more than one disk and at least two disks.
			nodeDisk := nodeBlockDeviceMap[value.Labels[string(apis.HostNameCPK)]]
			// Add the current disk to the diskList for this node.
			nodeDisk.Items = append(nodeDisk.Items, value.Name)
		}
	}
	return nodeBlockDeviceMap, nil
}

func (ac *Config) selectNode(nodeBlockDeviceMap map[string]*blockDeviceList) *nodeBlockDevice {
	// selectedBlockDevice will hold a node capable to form pool and list of disks attached to it.
	selectedBlockDevice := &nodeBlockDevice{
		NodeName: "",
		BlockDevices: blockDeviceList{
			Items: []string{},
		},
	}
	// diskCount will hold the number of disk that will be selected from a qualified
	// node for specific pool type
	var diskCount int
	// minRequiredDiskCount will hold the required number of disk that should be selected from a qualified
	// node for specific pool type
	minRequiredDiskCount := blockdevice.DefaultDiskCount[ac.poolType()]
	for node, val := range nodeBlockDeviceMap {
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
			selectedBlockDevice.BlockDevices.Items = append(selectedBlockDevice.BlockDevices.Items, val.Items[i])
		}
		selectedBlockDevice.NodeName = node
		break
	}
	return selectedBlockDevice
}

// diskFilterConstraint takes a value for key "ndm.io/disk-type" and form a label.
func diskFilterConstraint(diskType string) string {
	var label string
	if diskType == string(apis.TypeSparseCPV) {
		label = string(apis.NdmDiskTypeCPK) + "=" + string(apis.TypeSparseCPV)
	} else {
		label = string(apis.NdmBlockDeviceTypeCPK) + "=" + string(apis.TypeBlockDeviceCPV)
	}
	return label
}

// ProvisioningType returns the way pool should be provisioned e.g. auto or manual.
func ProvisioningType(spc *apis.StoragePoolClaim) string {
	if len(spc.Spec.BlockDevices.BlockDeviceList) == 0 {
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

// getBlockDevice return the all disks of a certain type(e.g. sparse, blockdevice) which is specified in spc.
func (ac *Config) getBlockDevice() (*ndmapis.BlockDeviceList, error) {
	diskFilterLabel := diskFilterConstraint(ac.Spc.Spec.Type)
	bdL, err := ac.BlockDeviceClient.List(metav1.ListOptions{LabelSelector: diskFilterLabel})
	if err != nil {
		return nil, err
	}
	bdl := bdL.Filter(blockdevice.FilterClaimedDevices, blockdevice.FilterNonInactive)
	return bdl.BlockDeviceList, nil
}
