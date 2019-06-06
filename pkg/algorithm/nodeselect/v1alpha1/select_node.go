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
	"time"

	"github.com/golang/glog"
	ndmapis "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	blockdevice "github.com/openebs/maya/pkg/blockdevice/v1alpha1"
	bdc "github.com/openebs/maya/pkg/blockdeviceclaim/v1alpha1"
	env "github.com/openebs/maya/pkg/env/v1alpha1"
	volume "github.com/openebs/maya/pkg/volume"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	retryCount = 10
	waitTime   = 5 * time.Second
)

type ClaimedBDDetails struct {
	DeviceID string
	BDName   string
	BDCName  string
}

type NodeClaimedBDDetails struct {
	NodeName        string
	BlockDeviceList []ClaimedBDDetails
}

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
	bdl := bdL.Filter(blockdevice.FilterNonInactive)
	return bdl.BlockDeviceList, nil
}

//TODO: Make changes in below code befor PR checked in
// 1) Use Builder Pattern or some approach to present below code
// 2) Refactor the below code before removing WIP
// ClaimBlockDevice will create BDC for corresponding BD
func (ac *Config) ClaimBlockDevice(nodeBDs *nodeBlockDevice, spc *apis.StoragePoolClaim) *NodeClaimedBDDetails {
	nodeClaimedBDs := &NodeClaimedBDDetails{
		NodeName:        "",
		BlockDeviceList: []ClaimedBDDetails{},
	}
	namespace := env.Get(env.OpenEBSNamespace)
	bdcKubeclient := bdc.NewKubeClient().
		WithNamespace(namespace)
	labels := map[string]string{string(apis.StoragePoolClaimCPK): spc.Name}
	for _, bdName := range nodeBDs.BlockDevices.Items {
		var hostName, bdcName string
		claimedBD := ClaimedBDDetails{}
		bdObj, err := ac.BlockDeviceClient.Get(bdName, metav1.GetOptions{})
		if err != nil {
			glog.Errorf("falied to get block device {%s} object ERR: {%v}", bdName, err)
			continue
		}
		hostName = bdObj.Labels[string(apis.HostNameCPK)]
		if !bdObj.IsClaimed() {
			bdcName = "bdc-" + string(bdObj.UID)
			capacity := volume.ByteCount(bdObj.Spec.Capacity.Storage)
			//TODO: Move below code to some function
			bdcObj, err := bdc.NewBuilder().
				WithName(bdcName).
				WithNamespace(namespace).
				WithLabels(labels).
				WithBlockDeviceName(bdName).
				WithHostName(hostName).
				WithCapacity(capacity).Build()
			if err != nil {
				glog.Errorf("failed to build block device claim for bd: {%s} ERR: {%v}", bdName, err)
				continue
			}
			_, err = bdcKubeclient.Create(bdcObj.Object)
			if err != nil {
				glog.Errorf("failed to create block device claim for bdc: {%s} ERR: {%v}", bdName, err)
				continue
			}
			isClaimed, err := ac.waitForClaimedStatus(bdName)
			if err != nil {
				//TODO: Handle these things from review suggesstions
				_ = bdcKubeclient.Delete(bdcName, &metav1.DeleteOptions{})
				glog.Errorf("failed to claim block device bd: {%s} ERR: {%v}", bdName, err)
				continue
			}
			if !isClaimed {
				glog.Errorf("Not able to claim the Block Device: %s", bdName)
				_ = bdcKubeclient.Delete(bdcName, &metav1.DeleteOptions{})
				continue
			}
		} else {
			//TODO: Use HasLabel Predicate for below check
			bdcName = bdObj.Spec.ClaimRef.Name
			if bdObj.Labels[string(apis.StoragePoolClaimCPK)] == "" {
				err = bdcKubeclient.PatchBDCWithLabel(labels, bdcName)
				if err != nil {
					glog.Errorf("failed to patch block device claim {%s}", bdcName)
					continue
				}
			} else if bdObj.Labels[string(apis.StoragePoolClaimCPK)] != spc.Name {
				glog.Errorf("block device %s is already in use", bdName)
				continue
			}
		}
		claimedBD.DeviceID = bdObj.GetDeviceID()
		claimedBD.BDName = bdObj.Name
		claimedBD.BDCName = bdcName
		nodeClaimedBDs.NodeName = hostName
		nodeClaimedBDs.BlockDeviceList = append(nodeClaimedBDs.BlockDeviceList, claimedBD)
	}

	selectedNodeBDs := ac.selectBlockDevices(nodeClaimedBDs)
	totalCount := len(nodeClaimedBDs.BlockDeviceList)
	diffCount := totalCount - len(selectedNodeBDs.BlockDeviceList)
	bdcNameList := []string{}
	for i := 1; i <= diffCount; i++ {
		bdcNameList = append(bdcNameList, nodeClaimedBDs.BlockDeviceList[totalCount-i].BDCName)
	}
	_ = bdcKubeclient.DeleteMultipleBDCs(bdcNameList)
	return selectedNodeBDs
}

// waitForClaimStatus will
func (ac *Config) waitForClaimedStatus(bdName string) (bool, error) {
	for i := 0; i < retryCount; i++ {
		bdObj, err := ac.BlockDeviceClient.Get(bdName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if bdObj.IsClaimed() {
			return true, nil
		}
		time.Sleep(waitTime)
	}
	return false, nil
}

func (ac *Config) selectBlockDevices(nodeClaimedBDs *NodeClaimedBDDetails) *NodeClaimedBDDetails {
	var BDCount int
	qualifiedNodeBDs := &NodeClaimedBDDetails{
		NodeName:        nodeClaimedBDs.NodeName,
		BlockDeviceList: []ClaimedBDDetails{},
	}
	minRequiredBDCount := blockdevice.DefaultDiskCount[ac.poolType()]
	currentBDCount := len(nodeClaimedBDs.BlockDeviceList)
	BDCount = minRequiredBDCount

	if len(nodeClaimedBDs.BlockDeviceList) < minRequiredBDCount {
		return qualifiedNodeBDs
	}
	if ProvisioningType(ac.Spc) == ProvisioningTypeManual {
		BDCount = (currentBDCount / minRequiredBDCount) * minRequiredBDCount
	}
	for i := 0; i < BDCount; i++ {
		BDDetails := ClaimedBDDetails{}
		BDDetails.DeviceID = nodeClaimedBDs.BlockDeviceList[i].DeviceID
		BDDetails.BDName = nodeClaimedBDs.BlockDeviceList[i].BDName
		qualifiedNodeBDs.BlockDeviceList = append(qualifiedNodeBDs.BlockDeviceList, BDDetails)
	}
	return qualifiedNodeBDs
}
