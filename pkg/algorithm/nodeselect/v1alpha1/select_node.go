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
	bdc "github.com/openebs/maya/pkg/blockdeviceclaim/v1alpha1"
	env "github.com/openebs/maya/pkg/env/v1alpha1"
	spcv1alpha1 "github.com/openebs/maya/pkg/storagepoolclaim/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	"github.com/openebs/maya/pkg/volume"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"strings"
)

const (
	// CStorBDTagAnnotationKey is the annotation key for SPC allowed BD tags
	CStorBDTagAnnotationKey = "openebs.io/allowed-bd-tags"
	// BlockDeviceTagLabelKey is the key to fetch tag of a block
	// device.
	// For more info : https://github.com/openebs/node-disk-manager/pull/400
	BlockDeviceTagLabelKey = "openebs.io/block-device-tag"
	// StoragePoolKind holds the value of StoragePoolClaim
	StoragePoolKind = "StoragePoolClaim"
	// APIVersion holds the value of OpenEBS version
	OpenEBS_APIVersion = "openebs.io/v1alpha1"
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
	//var filteredNodeBDs map[string]*blockDeviceList
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

	selectedBD, err := ac.selectNode(nodeBlockDeviceMap)
	if err != nil {
		return nil, err
	}

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
	// get used blockdevice from csp present in cluster
	usedDiskMap, err := ac.getUsedBlockDeviceMap()
	if err != nil {
		return nil, err
	}
	// get used node names from csp related to spc
	usedNodeMap, err := ac.getUsedNodeMap()
	if err != nil {
		return nil, err
	}
	// nodeBlockDeviceMap is the data structure holding host name as key
	// and nodeBlockDevice struct as value.
	nodeBlockDeviceMap := make(map[string]*blockDeviceList)
	for _, value := range listBlockDevice.Items {
		nodeName := value.Labels[string(apis.HostNameCPK)]
		// If the disk is already being used, do not consider this as a part for provisioning pool.
		if usedDiskMap[value.Name] == 1 {
			continue
		}
		// If the node is already being used for a given spc, do not consider this as a part for provisioning pool.
		if usedNodeMap[nodeName] == 1 {
			continue
		}
		//If node is already visited in this reconciliation then continue with
		//other nodes
		if ac.VisitedNodes[nodeName] {
			continue
		}
		if nodeBlockDeviceMap[nodeName] == nil {
			// Entry to this block means first time the hostname will be mapped for the first time.
			// Obviously, this entry of hostname(node) is for a usable disk and initialize diskCount to 1.
			nodeBlockDeviceMap[nodeName] = &blockDeviceList{Items: []string{value.Name}}
		} else {
			// Entry to this block means the hostname was already mapped and it has more than one disk and at least two disks.
			nodeDisk := nodeBlockDeviceMap[nodeName]
			// Add the current disk to the diskList for this node.
			nodeDisk.Items = append(nodeDisk.Items, value.Name)
		}
	}
	return nodeBlockDeviceMap, nil
}

// selectNode returns node name and list of blockdevice(nodeBlockDevice) can be used to
// provision cstor pools
func (ac *Config) selectNode(nodeBlockDeviceMap map[string]*blockDeviceList) (*nodeBlockDevice, error) {
	// selectedBlockDevice will hold a node capable to form pool and list of disks attached to it.
	selectedBlockDevice := &nodeBlockDevice{
		NodeName: "",
		BlockDevices: blockDeviceList{
			Items: []string{},
		},
	}
	var bdcList *ndmapis.BlockDeviceClaimList
	var err error
	spcObj := spcv1alpha1.SPC{Object: ac.Spc}
	// bdCount will hold the number of block devices that will be selected from a qualified
	// node for specific pool type
	var bdCount int
	// minRequiredDiskCount will hold the required number of disk that should be selected from a qualified
	// node for specific pool type
	minRequiredDiskCount := spcv1alpha1.DefaultDiskCount[ac.poolType()]

	// Below snippet is required to get the node and blockdevice topology from
	// blockdeviceclaims which were created during previous reconciliation
	if spcObj.IsAutoProvisioning() {
		// List all blockdeviceclaims related to spc
		// NOTE: This list contains blockdevices on which csps' are already
		// created filter and use it
		bdcList, err = bdc.NewKubeClient().
			WithNamespace(ac.Namespace).
			List(metav1.ListOptions{LabelSelector: string(apis.StoragePoolClaimCPK) + "=" + ac.Spc.Name})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to select node and blockdevices")
		}
		// customBDCList is a wrapper over blockdeviceclaim list
		customBDCList := bdc.BlockDeviceClaimList{
			ObjectList: bdcList,
		}
		// Form node and blockdevice topology for blockdevices which has bdc
		claimedDevicesOnNode := customBDCList.GetBlockDeviceNamesByNode()
		for claimedNodeName, bdNameList := range claimedDevicesOnNode {
			filteredBlockDevices := nodeBlockDeviceMap[claimedNodeName]

			//Check is this node is usable or not
			if filteredBlockDevices != nil {
				qualifiedBDList := util.ListIntersection(filteredBlockDevices.Items, bdNameList)
				// check whether bdNameList contains partially created claimed BDs
				// i.e BD count is less than required BDs
				if len(qualifiedBDList) < minRequiredDiskCount {
					count := minRequiredDiskCount - len(qualifiedBDList)
					availableBDOnNode := util.ListDiff(filteredBlockDevices.Items, qualifiedBDList)
					if len(availableBDOnNode) < count {
						continue
					}
					qualifiedBDList = append(qualifiedBDList, availableBDOnNode[:count]...)
				}
				selectedBlockDevice.BlockDevices.Items = append(
					selectedBlockDevice.BlockDevices.Items,
					qualifiedBDList[:minRequiredDiskCount]...,
				)
				selectedBlockDevice.NodeName = claimedNodeName
				ac.VisitedNodes[selectedBlockDevice.NodeName] = true
				return selectedBlockDevice, nil
			}
		}
	}

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
	if selectedBlockDevice.NodeName != "" {
		ac.VisitedNodes[selectedBlockDevice.NodeName] = true
	}
	return selectedBlockDevice, nil
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
	spcObj := &spcv1alpha1.SPC{Object: ac.Spc}
	diskFilterLabel := diskFilterConstraint(diskType)
	bdList, err := ac.BlockDeviceClient.List(metav1.ListOptions{LabelSelector: diskFilterLabel})
	if err != nil {
		return nil, err
	}
	filterList := []string{blockdevice.FilterNonInactive, blockdevice.FilterNotAllowedBDTag}

	if diskType == string(apis.TypeSparseCPV) {
		filterList = append(filterList, blockdevice.FilterSparseDevices)
	} else {
		filterList = append(filterList, blockdevice.FilterNonSparseDevices)
	}

	if spcObj.IsAutoProvisioning() {
		filterList = append(filterList, blockdevice.FilterNonFSType, blockdevice.FilterNonRelesedDevices)
	}

	allowedBDTags := getAllowedTagMap(ac.Spc.GetAnnotations())
	filterOptions := &blockdevice.FilterOptions{
		AllowedBDTags: allowedBDTags,
	}

	bdl = bdList.Filter(filterOptions, filterList...)
	if len(bdl.Items) == 0 {
		return nil, errors.Errorf("type {%s} devices are not available to provision pools in %s mode", diskType, ProvisioningType(ac.Spc))
	}

	// Below snippet is required when BDs are claimed but CSPs are not created.
	// This might be useful in a case where multiple spc's are created at a time
	if spcObj.IsAutoProvisioning() {
		newBDL, err := bdl.GetUsableBlockDevices(ac.Spc.Name, ac.Namespace)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get list of usable blockdevice list")
		}
		bdl = newBDL
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

			// Setting OwnerReference
			if spc == nil {
				return nil, errors.Errorf("failed to build BDC object: spc object is nil")
			}
			// WithOwnerReference method accepts an OwnerReference object
			// and uses it to set the owner for the BlockDeviceClaim
			trueVal := true
			owner := metav1.OwnerReference{
				APIVersion:         OpenEBS_APIVersion,
				Kind:               StoragePoolKind,
				UID:                spc.ObjectMeta.UID,
				Name:               spc.ObjectMeta.Name,
				BlockOwnerDeletion: &trueVal,
				Controller:         &trueVal,
			}

			//TODO: Move below code to some function
			newBDCObj, err := bdc.NewBuilder().
				WithName(bdcName).
				WithNamespace(namespace).
				WithLabels(labels).
				WithBlockDeviceName(bdName).
				WithHostName(hostName).
				WithCapacity(capacity).
				WithOwnerReference(owner).
				WithFinalizer(spcv1alpha1.SPCFinalizer).
				Build()

			value, ok := bdObj.Labels[BlockDeviceTagLabelKey]

			if ok {
				ls := &metav1.LabelSelector{
					MatchLabels: map[string]string{BlockDeviceTagLabelKey: value},
				}
				newBDCObj.Object.Spec.Selector = ls
			}

			if err != nil {
				return nil, errors.Wrapf(err, "failed to build block device claim for bd {%s}", bdName)
			}
			_, err = bdcKubeclient.Create(newBDCObj.Object)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to create block device claim for bdc {%s}", bdcName)
			}
			klog.Infof("successfully created block device claim {%s} for block device {%s}", bdcName, bdName)
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

// getAllowedTagMap returns a map of the allowed BD tags
// Example :
// If the CSPC annotation is passed and following is the BD tag annotation
//
// cstor.openebs.io/allowed-bd-tags:fast,slow
//
// Then, a map {"fast":true,"slow":true} is returned.
func getAllowedTagMap(cspcAnnotation map[string]string) map[string]bool {
	allowedTagsMap := make(map[string]bool)
	allowedTags := cspcAnnotation[CStorBDTagAnnotationKey]
	if strings.TrimSpace(allowedTags) == "" {
		return allowedTagsMap
	}
	allowedTagsList := strings.Split(allowedTags, ",")
	for _, v := range allowedTagsList {
		if strings.TrimSpace(v) == "" {
			continue
		}
		allowedTagsMap[v] = true
	}
	return allowedTagsMap
}
