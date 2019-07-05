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
	cspcbd "github.com/openebs/maya/pkg/cstor/poolcluster/v1alpha1/cstorpoolblockdevice"
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

type bdCapacityRepetition struct {
	sizeRepetitions map[uint64]int
}

// NodeBlockDeviceSelector selects required node and block devices attached to
// the node
func (ac *Config) NodeBlockDeviceSelector() (map[string]*cspcbd.ListBuilder, error) {
	listBD, err := ac.getBlockDevice()
	if err != nil {
		return nil, err
	}
	if listBD == nil || len(listBD.Items) == 0 {
		return nil, errors.New("no blockdevices available to create cstorpoolcluster")
	}
	selectedCapacity, err := ac.getCommonBDCapacity(listBD)
	if err != nil {
		return nil, err
	}
	nodeBlockDeviceList, err := ac.getCandidateNodeBlockDevices(listBD, selectedCapacity)
	if err != nil {
		return nil, err
	}

	selectNodeBDs, err := ac.selectQualifiedNodes(nodeBlockDeviceList)
	if err != nil {
		return nil, err
	}
	return selectNodeBDs, nil
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

// getCandidateNode make the map of node and blockdevices topology
// For example it forms map of
// N1 -> bd1, bd2, bd3
// N2 -> bd4, bd5
func (ac *Config) getCandidateNodeBlockDevices(
	listBlockDevice *ndmapis.BlockDeviceList,
	selectedCapacity uint64,
) (map[string]*cspcbd.ListBuilder, error) {
	minBDCount := blockdevice.DefaultDiskCount[ac.poolType()]
	// nodeBlockDeviceMap is the data structure holding host name as key
	// and cstorpoolcluster blockdevicelist as a value
	nodeBlockDeviceMap := make(map[string]*cspcbd.ListBuilder)
	for _, blockDevice := range listBlockDevice.Items {
		bd := blockdevice.BlockDevice{
			BlockDevice: &blockDevice,
		}
		hostName := blockDevice.Labels[string(apis.HostNameCPK)]
		listBuilder := nodeBlockDeviceMap[hostName]
		capacity := blockDevice.Spec.Capacity.Storage
		if capacity == selectedCapacity &&
			(listBuilder == nil || listBuilder.Len() < minBDCount) {
			capacityStr := volume.ByteCount(capacity)
			devID := bd.GetDeviceID()
			cspcBDObj, err := cspcbd.NewBuilder().
				WithBlockDeviceName(blockDevice.Name).
				WithCapacity(capacityStr).
				WithDevLink(devID).Build()
			if err != nil {
				glog.Errorf(
					"failed to build cspc blockdevice %s error: %v",
					blockDevice.Name,
					err,
				)
				continue
			}

			if listBuilder == nil {
				// Entry to this block means first time the hostname will be mapped for the first time.
				// Obviously, this entry of hostname(node) is for a usable
				// blockdevices
				nodeBlockDeviceMap[hostName] = cspcbd.ListBuilderForObjectNew(cspcBDObj)
			} else {
				// Add the current blockdevice to the existing listbuilder of
				// blockdevice
				listBuilder = listBuilder.ListBuilderForObject(cspcBDObj)
			}
		}
	}
	return nodeBlockDeviceMap, nil
}

// getCommonBDCapacity returns common blockdevice capacity accross the nodes in
// cluster
func (ac *Config) getCommonBDCapacity(listBlockDevice *ndmapis.BlockDeviceList) (uint64, error) {
	nodeBDsCapacities := make(map[string]*bdCapacityRepetition)
	minBDCount := blockdevice.DefaultDiskCount[ac.poolType()]
	mapCapacityRepetitions := map[uint64]int{}
	if ac.Spc.Spec.MaxPools == nil {
		return 0, errors.Errorf(
			"failed to get common capacity accross the nodes: maxPool field in spc is nil",
		)
	}
	minReqNodeCount := *ac.Spc.Spec.MaxPools
	totalRequiredBD := minBDCount * minReqNodeCount

	for _, bd := range listBlockDevice.Items {
		hostName := bd.Labels[string(apis.HostNameCPK)]
		capacity := bd.Spec.Capacity.Storage
		if nodeBDsCapacities[hostName] == nil {
			tmpSizeRep := &bdCapacityRepetition{map[uint64]int{}}
			tmpSizeRep.sizeRepetitions[capacity]++
			nodeBDsCapacities[hostName] = tmpSizeRep
		} else {
			tmpSizeRep := nodeBDsCapacities[hostName]
			tmpSizeRep.sizeRepetitions[capacity]++
		}
		mapCapacityRepetitions[capacity]++
	}
	for capacity, count := range mapCapacityRepetitions {
		qualifiedNodes := 0
		if count >= totalRequiredBD {
			for _, bdCapacityRep := range nodeBDsCapacities {
				if bdCapacityRep.sizeRepetitions[capacity] >= minBDCount {
					qualifiedNodes++
				}
				if qualifiedNodes == minReqNodeCount {
					return capacity, nil
				}
			}
		}
	}
	return 0, errors.Errorf(
		"no common blockdevice capacity is available accross the nodes to create cstorpoolcluster",
	)
}

// selectQualifiedNodes returns map of qualified nodes and required count of
// blockdevices attached to that node
func (ac *Config) selectQualifiedNodes(nodeCSPCBlockDeviceList map[string]*cspcbd.ListBuilder) (map[string]*cspcbd.ListBuilder, error) {
	qualifiedNodes := map[string]*cspcbd.ListBuilder{}
	minBDCount := blockdevice.DefaultDiskCount[ac.poolType()]
	minReqNodeCount := *ac.Spc.Spec.MaxPools
	qualifiedNodeCount := 0

	for key, cspcbdList := range nodeCSPCBlockDeviceList {
		cspcbdList := cspcbdList
		if cspcbdList.Len() >= minBDCount {
			qualifiedNodes[key] = cspcbdList
			qualifiedNodeCount++
		}
		if qualifiedNodeCount == minReqNodeCount {
			return qualifiedNodes, nil
		}
	}
	if len(qualifiedNodes) == 0 {
		return nil, errors.Errorf("nodes doesn't have enough blockdevices to create cstorpoolcluster")
	}
	return nil, errors.Errorf(
		"cluster doesn't have %d nodes with %d blockdevices to create cstorpoolcluster",
		minReqNodeCount,
		minBDCount,
	)
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
	glog.Infof("Length of blockdevice %d", len(bdList.Items))
	filterList := []string{blockdevice.FilterNonInactive, blockdevice.FilterUnclaimedDevices}

	if diskType == string(apis.TypeSparseCPV) {
		filterList = append(filterList, blockdevice.FilterSparseDevices)
	} else {
		filterList = append(filterList, blockdevice.FilterNonSparseDevices)
	}

	if ProvisioningType(ac.Spc) == ProvisioningTypeAuto {
		filterList = append(filterList, blockdevice.FilterNonFSType)
	} else {
		// Only auto spc pool provision is supported
		return nil, errors.Errorf(
			"creation of cstorpoolcluster via manual spc %s is not supported",
			ac.Spc.Name,
		)
	}

	bdl = bdList.Filter(filterList...)
	if len(bdl.Items) == 0 {
		return nil, errors.Errorf(
			"type {%s} blockdevices are not available to create cstorpoolcluster in %s mode",
			diskType,
			ProvisioningType(ac.Spc),
		)
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
