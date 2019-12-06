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

package app

import (
	nodeselect "github.com/openebs/maya/pkg/algorithm/nodeselect/v1alpha2"
	ndmapis "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	apisbd "github.com/openebs/maya/pkg/blockdevice/v1alpha2"
	apiscsp "github.com/openebs/maya/pkg/cstor/poolinstance/v1alpha3"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

func (pc *PoolConfig) handleOperations() {
	// TODO: Disable pool-mgmt reconciliation
	// before carrying out ant day 2 ops.
	pc.expandPool()
	pc.replaceBlockDevice()
	// Once all operations are executed enable pool-mgmt
	// reconciliation
}

// replaceBlockDevice replaces block devices in cStor pools as specified in CSPC.
func (pc *PoolConfig) replaceBlockDevice() error {
	for _, pool := range pc.AlgorithmConfig.CSPC.Spec.Pools {
		pool := pool
		var cspiObj *apis.CStorPoolInstance
		nodeName, err := nodeselect.GetNodeFromLabelSelector(pool.NodeSelector)
		if err != nil {
			return errors.Wrapf(err,
				"could not get node name for node selector {%v} "+
					"from cspc %s", pool.NodeSelector, pc.AlgorithmConfig.CSPC.Name)
		}

		cspiObj, err = pc.getCSPIWithNodeName(nodeName)
		if err != nil {
			return errors.Wrapf(err, "failed to get cspi with node name %s", nodeName)
		}

		if isPoolSpecBlockDevicesGotReplaced(&pool, cspiObj) {
			pc.updateExistingCSPI(&pool, cspiObj)
			_, err = apiscsp.NewKubeClient().
				WithNamespace(pc.AlgorithmConfig.Namespace).
				Update(cspiObj)
			if err != nil {
				klog.Errorf("could not replace block device in cspi %s: %s", cspiObj.Name, err.Error())
			}
		}
	}
	return nil
}

// expandPool expands the required cStor pools as specified in CSPC
func (pc *PoolConfig) expandPool() error {
	for _, pool := range pc.AlgorithmConfig.CSPC.Spec.Pools {
		pool := pool
		var cspiObj *apis.CStorPoolInstance
		nodeName, err := nodeselect.GetNodeFromLabelSelector(pool.NodeSelector)
		if err != nil {
			return errors.Wrapf(err,
				"could not get node name for node selector {%v} "+
					"from cspc %s", pool.NodeSelector, pc.AlgorithmConfig.CSPC.Name)
		}

		cspiObj, err = pc.getCSPIWithNodeName(nodeName)
		if err != nil {
			return errors.Wrapf(err, "failed to get cspi with node name %s", nodeName)
		}

		// Pool expansion for raid group types other than striped
		if len(pool.RaidGroups) > len(cspiObj.Spec.RaidGroups) {
			cspiObj = pc.addGroupToPool(&pool, cspiObj)
		}

		// Pool expansion for striped raid group
		pc.expandExistingStripedGroup(&pool, cspiObj)

		_, err = apiscsp.NewKubeClient().WithNamespace(pc.AlgorithmConfig.Namespace).Update(cspiObj)
		if err != nil {
			klog.Errorf("could not update cspi %s: %s", cspiObj.Name, err.Error())
		}
	}
	return nil
}

// addGroupToPool adds a raid group to the cspi
func (pc *PoolConfig) addGroupToPool(cspcPoolSpec *apis.PoolSpec, cspi *apis.CStorPoolInstance) *apis.CStorPoolInstance {
	for _, cspcRaidGroup := range cspcPoolSpec.RaidGroups {
		validGroup := true
		cspcRaidGroup := cspcRaidGroup
		if !isRaidGroupPresentOnCSPI(&cspcRaidGroup, cspi) {
			if cspcRaidGroup.Type == "" {
				cspcRaidGroup.Type = cspcPoolSpec.PoolConfig.DefaultRaidGroupType
			}

			for _, bd := range cspcRaidGroup.BlockDevices {
				err := pc.ClaimBD(bd.BlockDeviceName)
				if err != nil {
					klog.Errorf("failed to created bdc for bd %s:%s", bd.BlockDeviceName, err.Error())
				}
			}
			for _, bd := range cspcRaidGroup.BlockDevices {
				err := pc.isBDUsable(bd.BlockDeviceName)
				if err != nil {
					klog.Errorf("could not use bd %s for expanding pool "+
						"%s:%s", bd.BlockDeviceName, cspi.Name, err.Error())
					validGroup = false
					break
				}
			}
			if validGroup {
				cspi.Spec.RaidGroups = append(cspi.Spec.RaidGroups, cspcRaidGroup)
			}
		}
	}
	return cspi
}

// updateExistingCSPI updates the CSPI object with new block devices only when
// there are changes in block devices between raid groups of CSPI and CSPC
func (pc *PoolConfig) updateExistingCSPI(
	cspcPoolSpec *apis.PoolSpec, cspi *apis.CStorPoolInstance) *apis.CStorPoolInstance {
	var err error
	for _, cspcRaidGroup := range cspcPoolSpec.RaidGroups {
		cspcRaidGroup := cspcRaidGroup
		cspiRaidGroup := getReplacedCSPIRaidGroup(&cspcRaidGroup, cspi)
		if cspiRaidGroup != nil {
			err = pc.replaceExistingBlockDevice(cspcRaidGroup, cspiRaidGroup)
			if err != nil {
				klog.Infof(
					"failed to replace block device in raid group type: {%v} error: %v",
					cspcRaidGroup.Type,
					err,
				)
			}
		}
	}
	return cspi
}

// expandExistingStripedGroup adds newly added block devices to the existing striped
// groups present on CSPI
func (pc *PoolConfig) expandExistingStripedGroup(cspcPoolSpec *apis.PoolSpec, cspi *apis.CStorPoolInstance) {
	for _, cspcGroup := range cspcPoolSpec.RaidGroups {
		cspcGroup := cspcGroup
		if getRaidGroupType(cspcGroup, cspcPoolSpec) != string(apis.PoolStriped) || !isRaidGroupPresentOnCSPI(&cspcGroup, cspi) {
			continue
		}
		pc.addBlockDeviceToGroup(&cspcGroup, cspi)
	}
}

// getRaidGroupType returns the raid type for the provided group
func getRaidGroupType(group apis.RaidGroup, poolSpec *apis.PoolSpec) string {
	if group.Type != "" {
		return group.Type
	}
	return poolSpec.PoolConfig.DefaultRaidGroupType
}

// addBlockDeviceToGroup adds block devices to the provided raid group on CSPI
func (pc *PoolConfig) addBlockDeviceToGroup(group *apis.RaidGroup, cspi *apis.CStorPoolInstance) *apis.CStorPoolInstance {
	for i, groupOnCSPI := range cspi.Spec.RaidGroups {
		groupOnCSPI := groupOnCSPI
		if isRaidGroupPresentOnCSPI(group, cspi) {
			if len(group.BlockDevices) > len(groupOnCSPI.BlockDevices) {
				newBDs := getAddedBlockDevicesInGroup(group, &groupOnCSPI)
				if len(newBDs) == 0 {
					klog.V(2).Infof("No new block devices "+
						"added for group {%+v} on cspi %s", groupOnCSPI, cspi.Name)
				}
				pc.ClaimBDList(newBDs)
				for _, bdName := range newBDs {
					err := pc.isBDUsable(bdName)
					if err != nil {
						klog.Errorf("could not use bd %s for "+
							"expanding pool %s:%s", bdName, cspi.Name, err.Error())
						break
					}
					cspi.Spec.RaidGroups[i].BlockDevices =
						append(cspi.Spec.RaidGroups[i].BlockDevices,
							apis.CStorPoolClusterBlockDevice{BlockDeviceName: bdName})
				}
			}
		}
	}
	return cspi
}

// isRaidGroupPresentOnCSPI returns true if the provided
// raid group is already present on CSPI
// TODO: Validation webhook should ensure that in striped group type
// the blockdevices are only added and existing block device are not
// removed.
func isRaidGroupPresentOnCSPI(group *apis.RaidGroup, cspi *apis.CStorPoolInstance) bool {
	blockDeviceMap := make(map[string]bool)
	for _, bd := range group.BlockDevices {
		blockDeviceMap[bd.BlockDeviceName] = true
	}
	for _, cspiRaidGroup := range cspi.Spec.RaidGroups {
		for _, cspiBDs := range cspiRaidGroup.BlockDevices {
			if blockDeviceMap[cspiBDs.BlockDeviceName] {
				return true
			}
		}
	}
	return false
}

func (pc *PoolConfig) replaceExistingBlockDevice(
	cspcRaidGroup apis.RaidGroup,
	cspiRaidGroup *apis.RaidGroup) error {
	cspcBlockDeviceMap := make(map[string]bool)
	cspiBlockDeviceMap := make(map[string]bool)
	var oldBlockDeviceName string
	var newBlockDeviceName string

	// Form CSPI Block Device Map
	for _, bd := range cspiRaidGroup.BlockDevices {
		cspiBlockDeviceMap[bd.BlockDeviceName] = true
	}
	// Form CSPC Block Device Map
	for _, bd := range cspcRaidGroup.BlockDevices {
		cspcBlockDeviceMap[bd.BlockDeviceName] = true
	}
	// Find Old Block Device Name
	for bdName := range cspiBlockDeviceMap {
		if !cspcBlockDeviceMap[bdName] {
			oldBlockDeviceName = bdName
			break
		}
	}
	// Find New Block Device Name
	for bdName := range cspcBlockDeviceMap {
		if !cspiBlockDeviceMap[bdName] {
			newBlockDeviceName = bdName
			break
		}
	}

	if oldBlockDeviceName == "" || newBlockDeviceName == "" {
		return errors.Errorf(
			"failed to find new block device {%s} or old block device {%s}",
			oldBlockDeviceName,
			newBlockDeviceName,
		)
	}

	// Verify is that new block device is usable
	err := pc.isBDUsable(newBlockDeviceName)
	if err != nil {
		return errors.Wrapf(
			err,
			"could not use bd %s for replacement",
			newBlockDeviceName)
	}
	//Replace old block device with new block device in CSPI
	for index, bd := range cspiRaidGroup.BlockDevices {
		if bd.BlockDeviceName == oldBlockDeviceName {
			cspiRaidGroup.BlockDevices[index].BlockDeviceName = newBlockDeviceName
			return nil
		}
	}
	return nil
}

// getReplacedCSPIRaidGroup returns the corresponding CSPI raid group for provided CSPC
// raid group only if there is one block device replacement or else it will return
// nil
func getReplacedCSPIRaidGroup(
	cspcRaidGroup *apis.RaidGroup,
	cspi *apis.CStorPoolInstance) *apis.RaidGroup {
	blockDeviceMap := make(map[string]bool)
	for _, bd := range cspcRaidGroup.BlockDevices {
		blockDeviceMap[bd.BlockDeviceName] = true
	}
	for _, cspiRaidGroup := range cspi.Spec.RaidGroups {
		cspiRaidGroup := cspiRaidGroup
		misMatchedBDCount := 0
		for _, cspiBD := range cspiRaidGroup.BlockDevices {
			if !blockDeviceMap[cspiBD.BlockDeviceName] {
				misMatchedBDCount++
			}
		}
		if misMatchedBDCount == 1 {
			return &cspiRaidGroup
		}
	}
	return nil
}

// getAddedBlockDevicesInGroup returns the added block device list
func getAddedBlockDevicesInGroup(groupOnCSPC, groupOnCSPI *apis.RaidGroup) []string {
	var addedBlockDevices []string

	// bdPresentOnCSPI is a map whose key is block devices
	// name present on the CSPI and corresponding value for
	// the key is true.
	bdPresentOnCSPI := make(map[string]bool)
	for _, bdCSPI := range groupOnCSPI.BlockDevices {
		bdPresentOnCSPI[bdCSPI.BlockDeviceName] = true
	}

	for _, bdCSPC := range groupOnCSPC.BlockDevices {
		if !bdPresentOnCSPI[bdCSPC.BlockDeviceName] {
			addedBlockDevices = append(addedBlockDevices, bdCSPC.BlockDeviceName)
		}
	}
	return addedBlockDevices
}

// getCSPIWithNodeName returns a cspi object with provided node name
// TODO: Move to CSPI package
func (pc *PoolConfig) getCSPIWithNodeName(nodeName string) (*apis.CStorPoolInstance, error) {
	cspiList, _ := apiscsp.
		NewKubeClient().
		WithNamespace(pc.AlgorithmConfig.Namespace).
		List(metav1.ListOptions{LabelSelector: string(apis.CStorPoolClusterCPK) + "=" + pc.AlgorithmConfig.CSPC.Name})

	cspiListBuilder := apiscsp.ListBuilderFromAPIList(cspiList).WithFilter(apiscsp.HasNodeName(nodeName)).List()
	if len(cspiListBuilder.ObjectList.Items) == 1 {
		return &cspiListBuilder.ObjectList.Items[0], nil
	}
	return nil, errors.New("No CSPI(s) found")
}

// isBDUsable returns no error if BD can be used.
// If BD has no BDC -- it is created
// TODO: Move to algorithm package
func (pc *PoolConfig) isBDUsable(bdName string) error {
	bdObj, err := apisbd.
		NewKubeClient().
		WithNamespace(pc.AlgorithmConfig.Namespace).
		Get(bdName, metav1.GetOptions{})
	isBDUsable, err := pc.AlgorithmConfig.IsClaimedBDUsable(bdObj)
	if err != nil {
		return errors.Wrapf(err, "bd %s cannot be used as could not get claim status", bdName)
	}

	if !isBDUsable {
		return errors.Errorf("BD %s cannot be used as it is already claimed but not by cspc", bdName)
	}
	return nil
}

// ClaimBDList claims a list of block device
func (pc *PoolConfig) ClaimBDList(bdList []string) {
	for _, bdName := range bdList {
		err := pc.ClaimBD(bdName)
		if err != nil {
			klog.Errorf("failed to create bdc for bd %s: %s", bdName, err.Error())
		}
	}
}

// ClaimBD calims a block device
func (pc *PoolConfig) ClaimBD(bdName string) error {
	bdObj, err := apisbd.
		NewKubeClient().
		WithNamespace(pc.AlgorithmConfig.Namespace).
		Get(bdName, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "could not get bd object %s", bdName)
	}
	// If blockdevice is already claimed no need of creating claim
	if bdObj.Status.ClaimState == ndmapis.BlockDeviceClaimed {
		return nil
	}
	err = pc.AlgorithmConfig.ClaimBD(bdObj)
	if err != nil {
		return errors.Wrapf(err, "failed to claim bd %s", bdName)

	}
	return nil
}

// isPoolSpecBlockDevicesGotReplaced return true if any block device in CSPC pool
// spec got replaced. If no block device changes are detected then it will
// return false
func isPoolSpecBlockDevicesGotReplaced(
	cspcPoolSpec *apis.PoolSpec, cspi *apis.CStorPoolInstance) bool {
	cspcBlockDeviceMap := getBlockDeviceMapFromRaidGroups(cspcPoolSpec.RaidGroups)
	for _, rg := range cspi.Spec.RaidGroups {
		for _, bd := range rg.BlockDevices {
			if !cspcBlockDeviceMap[bd.BlockDeviceName] {
				return true
			}
		}
	}
	return false
}

// getBlockDeviceMapFromRaidGroups will return map of block devices that are in
// use by raid groups
func getBlockDeviceMapFromRaidGroups(
	raidGroups []apis.RaidGroup) map[string]bool {
	blockDeviceMap := make(map[string]bool)
	for _, rg := range raidGroups {
		for _, bd := range rg.BlockDevices {
			blockDeviceMap[bd.BlockDeviceName] = true
		}
	}
	return blockDeviceMap
}
