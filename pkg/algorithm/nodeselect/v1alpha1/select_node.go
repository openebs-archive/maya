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
	"github.com/golang/glog"
	ndmapis "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	blockdevice "github.com/openebs/maya/pkg/blockdevice/v1alpha1"
	bdc "github.com/openebs/maya/pkg/blockdeviceclaim/v1alpha1"
	env "github.com/openebs/maya/pkg/env/v1alpha1"
	spcv1alpha1 "github.com/openebs/maya/pkg/storagepoolclaim/v1alpha1"
	volume "github.com/openebs/maya/pkg/volume"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BDDetails holds the claimed block device details
type BDDetails struct {
	DeviceID string
	BDName   string
}

// ClaimedBDDetails holds the node name and
// claimed block device deatils corresponding to node
type ClaimedBDDetails struct {
	NodeName        string
	BlockDeviceList []BDDetails
}

// NodeBlockDeviceSelector selects a node and block devices attached to it.
func (ac *Config) NodeBlockDeviceSelector() (*nodeBlockDevice, error) {
	var filteredNodeBDs map[string]*blockDeviceList
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

	filteredNodeBDs = nodeBlockDeviceMap
	if ProvisioningType(ac.Spc) == ProvisioningTypeAuto {
		filteredNodeBDs, err = ac.getFilteredNodeBlockDevices(nodeBlockDeviceMap)
		if err != nil {
			return nil, err
		}
	}

	selectedBD := ac.selectNode(filteredNodeBDs)

	return selectedBD, nil
}

// getFilteredNodeBlockDevices returns the map of node name and block device
// list if no claims are present on list of block devices. If claims are present
// then it retunrns those block devices on which claims are created
func (ac *Config) getFilteredNodeBlockDevices(nodeBDs map[string]*blockDeviceList) (map[string]*blockDeviceList, error) {
	namespace := env.Get(env.OpenEBSNamespace)
	bdcKubeclient := bdc.NewKubeClient().
		WithNamespace(namespace)
	bdcList, err := bdcKubeclient.List(metav1.ListOptions{LabelSelector: string(apis.StoragePoolClaimCPK) + "=" + ac.Spc.Name})
	if err != nil {
		return nil, err
	}

	customBDCList := &bdc.BlockDeviceClaimList{
		ObjectList: bdcList,
	}

	nodeClaimedBDs := customBDCList.GetBlockDeviceNamesByNode()

	newNodeBDMap := make(map[string]*blockDeviceList)
	for node, bdList := range nodeBDs {
		if len(nodeClaimedBDs[node]) == 0 {
			newNodeBDMap[node] = &blockDeviceList{
				Items: bdList.Items,
			}
		} else {
			newNodeBDMap[node] = &blockDeviceList{
				Items: nodeClaimedBDs[node],
			}
		}
	}
	return newNodeBDMap, nil
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
	// bdCount will hold the number of block devices that will be selected from a qualified
	// node for specific pool type
	var bdCount int
	// minRequiredDiskCount will hold the required number of disk that should be selected from a qualified
	// node for specific pool type
	minRequiredDiskCount := blockdevice.DefaultDiskCount[ac.poolType()]
	for node, val := range nodeBlockDeviceMap {
		// If the current block device count on the node is less than the required disks
		// then this is a dirty node and it will not qualify.
		if len(val.Items) < minRequiredDiskCount {
			continue
		}
		bdCount = minRequiredDiskCount
		if ProvisioningType(ac.Spc) == ProvisioningTypeManual {
			bdCount = (len(val.Items) / minRequiredDiskCount) * minRequiredDiskCount
		}
		for i := 0; i < bdCount; i++ {
			selectedBlockDevice.BlockDevices.Items = append(selectedBlockDevice.BlockDevices.Items, val.Items[i])
		}
		selectedBlockDevice.NodeName = node
		break
	}
	return selectedBlockDevice
}

// diskFilterConstraint takes a value for key "ndm.io/disk-type" and form a label.
func diskFilterConstraint(diskType string) string {
	if spcv1alpha1.SupportedDiskTypes[apis.CasPoolValString(diskType)] {
		return string(apis.NdmBlockDeviceTypeCPK) + "=" + string(apis.TypeBlockDeviceCPV)
	}
	return ""
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
	var bdl *blockdevice.BlockDeviceList
	diskType := ac.Spc.Spec.Type
	diskFilterLabel := diskFilterConstraint(diskType)
	bdList, err := ac.BlockDeviceClient.List(metav1.ListOptions{LabelSelector: diskFilterLabel})
	if err != nil {
		return nil, err
	}
	filterList := []string{blockdevice.FilterNonInactive}

	if diskType == string(apis.TypeSparseCPV) {
		filterList = append(filterList, blockdevice.FilterSparseDevices)
	} else {
		filterList = append(filterList, blockdevice.FilterNonSparseDevices)
	}

	if ProvisioningType(ac.Spc) == ProvisioningTypeAuto {
		filterList = append(filterList, blockdevice.FilterNonFSType, blockdevice.FilterNonRelesedDevices)
	}

	bdl = bdList.Filter(filterList...)
	if len(bdl.Items) == 0 {
		return nil, errors.Errorf("type {%s} devices are not available to provision pools in %s mode", diskType, ProvisioningType(ac.Spc))
	}
	return bdl.BlockDeviceList, nil
}

//TODO: Make changes in below code in refactor PR
// 1) Use Builder Pattern or some approach to present below code

// ClaimBlockDevice will create BDC for corresponding BD
func (ac *Config) ClaimBlockDevice(nodeBDs *nodeBlockDevice, spc *apis.StoragePoolClaim) (*ClaimedBDDetails, error) {
	nodeClaimedBDs := &ClaimedBDDetails{
		NodeName:        "",
		BlockDeviceList: []BDDetails{},
	}

	if nodeBDs == nil || len(nodeBDs.BlockDevices.Items) == 0 {
		return nil, errors.New("No valid block devices are available to claim")
	}

	namespace := env.Get(env.OpenEBSNamespace)
	bdcKubeclient := bdc.NewKubeClient().
		WithNamespace(namespace)
	labels := map[string]string{string(apis.StoragePoolClaimCPK): spc.Name}
	lselector := string(apis.StoragePoolClaimCPK) + "=" + spc.Name

	nodeClaimedBDs.NodeName = nodeBDs.NodeName
	pendingBDCCount := 0

	bdcObjList, err := bdcKubeclient.List(metav1.ListOptions{LabelSelector: lselector})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list block device claims for {%s}", spc.Name)
	}
	customBDCObjList := bdc.ListBuilderFromAPIList(bdcObjList)

	for _, bdName := range nodeBDs.BlockDevices.Items {
		var hostName, bdcName string
		claimedBD := BDDetails{}

		bdObj, err := ac.BlockDeviceClient.Get(bdName, metav1.GetOptions{})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get block device {%s}", bdName)
		}
		hostName = bdObj.Labels[string(apis.HostNameCPK)]

		if bdObj.IsClaimed() {
			bdcName = bdObj.Spec.ClaimRef.Name
			bdcObj := customBDCObjList.GetBlockDeviceClaim(bdcName)
			if bdcObj == nil {
				return nil, errors.Errorf("block device {%s} is already in use", bdcName)
			}
		} else {
			bdcName = "bdc-" + string(bdObj.UID)
			bdcObj := customBDCObjList.GetBlockDeviceClaim(bdcName)
			if bdcObj != nil {
				pendingBDCCount++
				continue
			}
			capacity := volume.ByteCount(bdObj.Spec.Capacity.Storage)
			//TODO: Move below code to some function
			newBDCObj, err := bdc.NewBuilder().
				WithName(bdcName).
				WithNamespace(namespace).
				WithLabels(labels).
				WithBlockDeviceName(bdName).
				WithHostName(hostName).
				WithCapacity(capacity).
				WithOwnerReference(spc).
				Build()
			if err != nil {
				return nil, errors.Wrapf(err, "failed to build block device claim for bd {%s}", bdName)
			}

			_, err = bdcKubeclient.Create(newBDCObj.Object)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to create block device claim for bdc {%s}", bdcName)
			}
			glog.Infof("successfully created block device claim {%s} for block device {%s}", bdcName, bdName)
			// As a part of reconcilation we will create pool if all the block
			// devices are claimed
			pendingBDCCount++
			continue
		}
		claimedBD.DeviceID = bdObj.GetDeviceID()
		claimedBD.BDName = bdObj.Name
		nodeClaimedBDs.BlockDeviceList = append(nodeClaimedBDs.BlockDeviceList, claimedBD)
	}
	if pendingBDCCount != 0 {
		return nil, errors.Errorf("pending block device claim count %d on node {%s}", pendingBDCCount, nodeClaimedBDs.NodeName)
	}
	return nodeClaimedBDs, nil
}
