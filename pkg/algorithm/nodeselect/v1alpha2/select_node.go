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

package v1alpha2

import (
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	bd "github.com/openebs/maya/pkg/blockdevice/v1alpha2"
	bdc "github.com/openebs/maya/pkg/blockdeviceclaim/v1alpha1"
	csp "github.com/openebs/maya/pkg/cstor/newpool/v1alpha3"
	"github.com/openebs/maya/pkg/volume"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetCandidateNodeMap returns a map of all nodes where the pool needs to be created.
func (ac *Config) GetCandidateNodeMap() (map[string]bool, error) {
	// TODO : Do not select a node if it is not ready
	candidateNodesMap := make(map[string]bool)
	usedNodeMap, err := ac.GetUsedNodeMap()
	if err != nil {
		return nil, errors.Wrapf(err, "could not get candidate nodes for pool creation")
	}
	for _, pool := range ac.CSPC.Spec.Pools {
		nodeName := pool.NodeSelector[HostName]
		if usedNodeMap[nodeName] == false {
			candidateNodesMap[nodeName] = true
		}
	}
	return candidateNodesMap, nil
}

// GetUsedNodeMap returns a map of node for which pool has already been created.
func (ac *Config) GetUsedNodeMap() (map[string]bool, error) {
	usedNodeMap := make(map[string]bool)
	cspList, err := csp.
		NewKubeClient().
		WithNamespace(ac.Namespace).
		List(
			metav1.
				ListOptions{LabelSelector: string(apis.CStorPoolClusterCPK) + "=" + ac.CSPC.Name},
		)
	if err != nil {
		return nil, errors.Wrap(err, "could not list already created csp(s)")
	}
	for _, cspObj := range cspList.Items {
		usedNodeMap[cspObj.Labels[string(apis.HostNameCPK)]] = true
	}
	return usedNodeMap, nil
}

// SelectNode selects a node and returns the pool spec from the cspc for pool provisioning.
func (ac *Config) SelectNode() (*apis.PoolSpec, error) {
	candidateNodes, err := ac.GetCandidateNodeMap()
	if err != nil {
		return nil, errors.Wrapf(err, "could not get pool spec for pool creation")
	}
	for _, pool := range ac.CSPC.Spec.Pools {
		pool := pool
		nodeName := pool.NodeSelector[HostName]
		if candidateNodes[nodeName] {
			if ValidatePoolSpec(&pool) {
				return &pool, nil
			}
		}
	}
	return nil, errors.New("no node qualified for pool creation")
}

// GetBDListForNode returns a list of BD from the pool spec.
// TODO : Move it to CStorPoolCluster packgage
func (ac *Config) GetBDListForNode(pool *apis.PoolSpec) []string {
	var BDList []string
	for _, group := range pool.RaidGroups {
		for _, bd := range group.BlockDevices {
			BDList = append(BDList, bd.BlockDeviceName)
		}
	}
	return BDList
}

// ClaimBDsForNode claims a given BlockDevice for node
// If the block device(s) is/are already claimed for any other CSPC it returns error.
// If the block device(s) is/are already calimed for the same CSPC -- it is left as it is and can be used for
// pool provisioning.
// If the block device(s) is/are unclaimed, then those are claimed.
func (ac *Config) ClaimBDsForNode(BD []string) error {
	for _, bd := range BD {
		IsBDClaimed, err := ac.IsBDClaimed(bd)
		if err != nil {
			return errors.Wrapf(err, "error in getting details for BD {%s} whether it is claimed", bd)
		}
		if IsBDClaimed {
			IsClaimedBDUsable, err := ac.IsClaimedBDUsable(bd)
			if err != nil {
				return errors.Wrapf(err, "error in getting details for BD {%s} for usability", bd)
			}
			if !IsClaimedBDUsable {
				return errors.Errorf("BD {%s} already in use", bd)
			}
		}
	}
	return ac.ClaimBDList(BD)
}

// ClaimBDList claims a list of  BlockDevices
// If the block device is already claimed -- no action is performed i.e. it remains claimed
func (ac *Config) ClaimBDList(BDList []string) error {
	if len(BDList) == 0 {
		return errors.New("No block devices to claim")
	}
	for _, bd := range BDList {
		IsBDClaimed, err := ac.IsBDClaimed(bd)
		if err != nil {
			return errors.Wrapf(err, "error in getting details for BD {%s} whether it is claimed", bd)
		}
		if IsBDClaimed {
			continue
		}
		err = ac.ClaimBD(bd)
		if err != nil {
			return errors.Wrapf(err, "could not claim block device {%s}", bd)
		}
	}
	return nil
}

// ClaimBD claims a given BlockDevice
func (ac *Config) ClaimBD(BD string) error {
	// TODO: Do not claim a block device if it is not active
	bdObj, err := bd.NewKubeClient().WithNamespace(ac.Namespace).Get(BD, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to claim BD %s", BD)
	}
	newBDCObj, err := bdc.NewBuilder().
		WithName("bdc-" + string(bdObj.UID)).
		WithNamespace(ac.Namespace).
		WithLabels(map[string]string{string(apis.CStorPoolClusterCPK): ac.CSPC.Name}).
		WithBlockDeviceName(bdObj.Name).
		WithHostName(bdObj.Labels[string(apis.HostNameCPK)]).
		WithCapacity(volume.ByteCount(bdObj.Spec.Capacity.Storage)).
		WithCSPCOwnerReference(ac.CSPC).
		Build()

	if err != nil {
		return errors.Wrapf(err, "failed to build block device claim for bd {%s}", bdObj.Name)
	}

	_, err = bdc.NewKubeClient().WithNamespace(ac.Namespace).Create(newBDCObj.Object)
	if err != nil {
		return errors.Wrapf(err, "failed to create block device claim for bd {%s}", bdObj.Name)
	}
	return nil
}

// IsClaimedBDUsable returns true if the passed BD is already claimed and can be
// used for provisioning
func (ac *Config) IsClaimedBDUsable(BD string) (bool, error) {
	bdAPIObj, err := bd.NewKubeClient().WithNamespace(ac.Namespace).Get(BD, metav1.GetOptions{})
	if err != nil {
		return false, errors.Wrapf(err, "could not get block device object {%s}", BD)
	}
	bdObj := bd.BuilderForAPIObject(bdAPIObj)
	if bdObj.BlockDevice.IsClaimed() {
		bdcName := bdObj.BlockDevice.Object.Spec.ClaimRef.Name
		bdcAPIObject, err := bdc.NewKubeClient().WithNamespace(ac.Namespace).Get(bdcName, metav1.GetOptions{})
		if err != nil {
			return false, errors.Wrapf(err, "could not get block device claim for block device {%s}", BD)
		}
		bdcObj := bdc.BuilderForAPIObject(bdcAPIObject)
		if bdcObj.BDC.HasLabel(string(apis.CStorPoolClusterCPK), ac.CSPC.Name) {
			return true, nil
		}
	} else {
		return false, errors.Wrapf(err, "block device {%s} is not claimed", BD)
	}
	return false, nil
}

// IsBDClaimed returns true if the passed BD is already claimed
func (ac *Config) IsBDClaimed(BD string) (bool, error) {
	bdAPIObj, err := bd.NewKubeClient().WithNamespace(ac.Namespace).Get(BD, metav1.GetOptions{})
	if err != nil {
		return false, errors.Wrapf(err, "could not get block device object {%s}", BD)
	}
	bdObj := bd.BuilderForAPIObject(bdAPIObj)
	return bdObj.BlockDevice.IsClaimed(), nil
}

// TODO: Fix following function -- (Current is mock only )
func ValidatePoolSpec(pool *apis.PoolSpec) bool {
	return true
}
