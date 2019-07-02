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
	"github.com/openebs/maya/pkg/volume"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (ac *Config) GetCandidateNodeMap() map[string]bool {
	candidateNodesMap := make(map[string]bool)
	usedNodeMap := ac.GetUsedNodeMap()
	for _, pool := range ac.CSPC.Spec.Pools {
		nodeName := pool.NodeSelector["kubernetes.io/hostName"]
		if usedNodeMap[nodeName] == false {
			candidateNodesMap[nodeName] = true
		}
	}
	return candidateNodesMap
}

func (ac *Config) GetUsedNodeMap() map[string]bool {
	usedNodeMap := make(map[string]bool)
	return usedNodeMap
}

func (ac *Config) SelectNode() *apis.PoolSpec {
	candidateNodes := ac.GetCandidateNodeMap()
	for _, pool := range ac.CSPC.Spec.Pools {
		pool := pool
		nodeName := pool.NodeSelector["kubernetes.io/hostName"]
		if candidateNodes[nodeName] {
			if ValidatePoolSpec(&pool) {
				return &pool
			}
		}
	}
	return nil
}

// GetBDListForNode returns a list of BD from the pool spec.
func (ac *Config) GetBDListForNode(pool *apis.PoolSpec) []string {
	var BDList []string
	for _, group := range pool.RaidGroups {
		for _, bd := range group.BlockDevices {
			BDList = append(BDList, bd.BlockDeviceName)
		}
	}
	return BDList
}

// ClaimBD claims a given BlockDevice
func (ac *Config) ClaimBDsForNode(BD []string) error {
	for _, bd := range BD {
		if ac.IsBDClaimed(bd) {
			if !ac.IsClaimedBDUsable(bd) {
				return errors.Errorf("BD {%s} already in use", bd)
			}
		}
	}
	return ac.ClaimBDList(BD)
}

// GetClaimedBDs returns BDs which are already claimed
func (ac *Config) GetClaimedBDs(BDList []string) map[string]bool {
	claimedBDs := make(map[string]bool)
	for _, bd := range BDList {
		if ac.IsBDClaimed(bd) {
			claimedBDs[bd] = true
		}
	}
	return claimedBDs
}

// ClaimBDList claims a list of  BlockDevices
// If the block device is already claimed -- no action is performed i.e. it remains claimed
func (ac *Config) ClaimBDList(BDList []string) error {
	if len(BDList) == 0 {
		return errors.New("No block devices to claim")
	}
	for _, bd := range BDList {
		if ac.IsBDClaimed(bd) {
			continue
		}
		ac.ClaimBD(bd)
	}
	return nil
}

// ClaimBD claims a given BlockDevice
func (ac *Config) ClaimBD(BD string) error {
	bdObj, err := bd.NewKubeClient().WithNamespace(ac.Namespace).Get(BD, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to claim BD %s", BD)
	}
	newBDCObj, err := bdc.NewBuilder().
		WithName("bdc-" + string(bdObj.UID)).
		WithNamespace(ac.Namespace).
		WithLabels(map[string]string{string(apis.StoragePoolClaimCPK): ac.CSPC.Name}).
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

// TODO : Fix following function

// IsClaimedBDUsable returns true if the passed BD is already claimed and can be
// used for provisioning
func (ac *Config) IsClaimedBDUsable(BD string) bool {
	return false
}

// IsBDClaimed returns true if the passed BD is already claimed
func (ac *Config) IsBDClaimed(BD string) bool {
	return false
}

// TODO: Fix following function -- (Current is mock only )
func ValidatePoolSpec(pool *apis.PoolSpec) bool {
	return true
}
