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

package cspc

import (
	"github.com/golang/glog"
	nodeselect "github.com/openebs/maya/pkg/algorithm/nodeselect/v1alpha2"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	apiscsp "github.com/openebs/maya/pkg/cstor/newpool/v1alpha3"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
func (pc *PoolConfig) replaceBlockDevice() {
	glog.V(2).Info("Block device replacement is not sup[orted ")
}

// expandPool expands the required cStor pools as specified in CSPC
func (pc *PoolConfig) expandPool() error {
	for _, pool := range pc.AlgorithmConfig.CSPC.Spec.Pools {

		var cspObj *apis.NewTestCStorPool
		nodeName, err := nodeselect.GetNodeFromLabelSelector(pool.NodeSelector)
		if err != nil {
			return errors.Wrapf(err,
				"could not get node name for node selector {%v} "+
					"from cspc %s", pool.NodeSelector, pc.AlgorithmConfig.CSPC.Name)
		}

		cspObj, err = pc.getCSPWithNodeName(nodeName)
		if err != nil {
			return errors.Wrapf(err, "failed to csp with node name %s", nodeName)
		}

		if len(pool.RaidGroups) > len(cspObj.Spec.RaidGroups) {
			cspObj = addGroupToPool(&pool, cspObj)
		}

		for _, group := range pool.RaidGroups {
			group := group
			if (group.Type == "stripe") || (group.Type == "" && pool.PoolConfig.DefaultRaidGroupType == "stripe") {
				cspObj = addBlockDeviceToGroup(&group, cspObj)
			}
		}
		_, err = apiscsp.NewKubeClient().WithNamespace(pc.AlgorithmConfig.Namespace).Update(cspObj)
		if err != nil {
			glog.Errorf("could not update csp %s: %s", cspObj.Name, err.Error())
		}
	}

	return nil
}

// addGroupToPool adds a raid group to the csp
func addGroupToPool(cspcPoolSpec *apis.PoolSpec, csp *apis.NewTestCStorPool) *apis.NewTestCStorPool {
	raidGroupPresentonCSP := make(map[string]bool)
	for _, cspRaidGroup := range csp.Spec.RaidGroups {
		raidGroupPresentonCSP[cspRaidGroup.Name] = true
	}

	for _, cspcRaidGroup := range cspcPoolSpec.RaidGroups {
		if !raidGroupPresentonCSP[cspcRaidGroup.Name] {
			if cspcRaidGroup.Type == "" {
				cspcRaidGroup.Type = cspcPoolSpec.PoolConfig.DefaultRaidGroupType
			}
			csp.Spec.RaidGroups = append(csp.Spec.RaidGroups, cspcRaidGroup)
		}
	}
	return csp
}

// addBlockDeviceToGroup adds block devices to the provided raid group on CSP
func addBlockDeviceToGroup(group *apis.RaidGroup, csp *apis.NewTestCStorPool) *apis.NewTestCStorPool {
	for i, groupOnCSP := range csp.Spec.RaidGroups {
		groupOnCSP := groupOnCSP
		if group.Name == groupOnCSP.Name {
			addedBlockDevices, err := getAddedBlockDevicesInGroup(group, &groupOnCSP)
			if err != nil {
				glog.V(2).Infof("No block device added on group %s: %s", group.Name, err)
			}
			for _, bdName := range addedBlockDevices {
				csp.Spec.RaidGroups[i].BlockDevices = append(csp.Spec.RaidGroups[i].BlockDevices, apis.CStorPoolClusterBlockDevice{BlockDeviceName: bdName})
			}
			break
		}
	}
	return csp
}

// getAddedBlockDevicesInGroup returns the added block device list
func getAddedBlockDevicesInGroup(groupOnCSPC *apis.RaidGroup, groupOnCSP *apis.RaidGroup) ([]string, error) {
	var addedBlockDevices []string

	// bdPresentOnCSP is a map whose key is block devices
	// name present on the CSP and corresponding value for
	// the key is true.
	bdPresentOnCSP := make(map[string]bool)
	for _, bdCSP := range groupOnCSP.BlockDevices {
		bdPresentOnCSP[bdCSP.BlockDeviceName] = true
	}

	for _, bdCSPC := range groupOnCSPC.BlockDevices {
		if !bdPresentOnCSP[bdCSPC.BlockDeviceName] {
			addedBlockDevices = append(addedBlockDevices, bdCSPC.BlockDeviceName)
		}
	}

	if len(addedBlockDevices) == 0 {
		return []string{}, errors.Errorf("no addition of block device in the group %s", groupOnCSPC.Name)
	}

	return addedBlockDevices, nil
}

// getCSPWithNodeName returns a csp object with provided node name
// TODO: Move to CSP package
func (pc *PoolConfig) getCSPWithNodeName(nodeName string) (*apis.NewTestCStorPool, error) {
	cspList, _ := apiscsp.NewKubeClient().WithNamespace(pc.AlgorithmConfig.Namespace).
		List(metav1.ListOptions{LabelSelector: string(apis.CStorPoolClusterCPK) + "=" + pc.AlgorithmConfig.CSPC.Name})

	cspListBuilder := apiscsp.ListBuilderFromAPIList(cspList).WithFilter(apiscsp.HasNodeName(nodeName)).List()
	if len(cspListBuilder.ObjectList.Items) == 1 {
		return &cspListBuilder.ObjectList.Items[0], nil
	}
	return nil, errors.New("No CSP(s) found")
}
