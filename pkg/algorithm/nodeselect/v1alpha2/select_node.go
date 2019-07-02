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
	"github.com/pkg/errors"
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

func (ac *Config) GetBDListForNode(pool *apis.PoolSpec) []string {
	var BDList []string
	for _, group := range pool.RaidGroups {
		for _, bd := range group.BlockDevices {
			BDList = append(BDList, bd.BlockDeviceName)
		}
	}
	return BDList
}

func (ac *Config) ClaimBDListForNode(BDList []string) error {
	if len(BDList) == 0 {
		return errors.New("No block devices to claim")
	}
	return nil
}

// TODO: Fix following function -- (Current is mock only )
func ValidatePoolSpec(pool *apis.PoolSpec) bool {
	return true
}
