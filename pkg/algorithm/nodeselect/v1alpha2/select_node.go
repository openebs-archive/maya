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
	"github.com/golang/glog"
	ndmapis "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	bd "github.com/openebs/maya/pkg/blockdevice/v1alpha2"
	bdc "github.com/openebs/maya/pkg/blockdeviceclaim/v1alpha1"
	csp "github.com/openebs/maya/pkg/cstor/newpool/v1alpha3"
	nodeapis "github.com/openebs/maya/pkg/kubernetes/node/v1alpha1"
	"github.com/openebs/maya/pkg/volume"
	"github.com/pkg/errors"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SelectNode returns a node where pool should be created.
func (ac *Config) SelectNode() (*apis.PoolSpec, string, error) {
	usedNodes, err := ac.GetUsedNode()
	if err != nil {
		return nil, "", errors.Wrapf(err, "could not get used nodes list for pool creation")
	}
	for _, pool := range ac.CSPC.Spec.Pools {
		// pin it
		pool := pool
		nodeName, err := GetNodeFromLabelSelector(pool.NodeSelector)
		if err != nil || nodeName == "" {
			glog.Errorf("could not use node for selectors {%v}", pool.NodeSelector)
			continue
		}
		if ac.VisitedNodes[nodeName] {
			continue
		} else {
			ac.VisitedNodes[nodeName] = true

			if !usedNodes[nodeName] {
				return &pool, nodeName, nil
			}
		}

	}
	return nil, "", errors.New("no node qualified for pool creation")
}

// GetNodeFromLabelSelector returns the node name selected by provided labels
// TODO : Move it to node package
func GetNodeFromLabelSelector(labels map[string]string) (string, error) {
	nodeList, err := nodeapis.NewKubeClient().List(metav1.ListOptions{LabelSelector: getLabelSelectorString(labels)})
	if err != nil {
		return "", errors.Wrap(err, "failed to get node list from the node selector")
	}
	if len(nodeList.Items) != 1 {
		return "", errors.Errorf("could not get a unique node from the given node selectors")
	}
	return nodeList.Items[0].Name, nil
}

// getLabelSelectorString returns a string of label selector form label map to be used in
// list options.
// TODO : Move it to node package
func getLabelSelectorString(selector map[string]string) string {
	var selectorString string
	for key, value := range selector {
		selectorString = selectorString + key + "=" + value + ","
	}
	selectorString = selectorString[:len(selectorString)-len(",")]
	return selectorString
}

// GetUsedNode returns a map of node for which pool has already been created.
// Note : Filter function is not used from node builder package as it needs
// CSP builder package which cam cause import loops.
func (ac *Config) GetUsedNode() (map[string]bool, error) {
	usedNode := make(map[string]bool)
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
		usedNode[cspObj.Labels[string(apis.HostNameCPK)]] = true
	}
	return usedNode, nil
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
	pendingClaim := 0
	for _, bdName := range BD {
		bdAPIObj, err := bd.NewKubeClient().WithNamespace(ac.Namespace).Get(bdName, metav1.GetOptions{})
		if err != nil {
			return errors.Wrapf(err, "error in getting details for BD {%s} whether it is claimed", bdName)
		}
		if bd.BuilderForAPIObject(bdAPIObj).BlockDevice.IsClaimed() {
			IsClaimedBDUsable, errBD := ac.IsClaimedBDUsable(bdAPIObj)
			if errBD != nil {
				return errors.Wrapf(err, "error in getting details for BD {%s} for usability", bdName)
			}
			if !IsClaimedBDUsable {
				return errors.Errorf("BD {%s} already in use", bdName)
			}
			continue
		}

		err = ac.ClaimBD(bdAPIObj)
		if err != nil {
			return errors.Wrapf(err, "Failed to claim BD {%s}", bdName)
		}
		pendingClaim++
	}

	if pendingClaim > 0 {
		return errors.Errorf("%d block device claims are pending", pendingClaim)
	}
	return nil
}

// ClaimBD claims a given BlockDevice
func (ac *Config) ClaimBD(bdObj *ndmapis.BlockDevice) error {
	newBDCObj, err := bdc.NewBuilder().
		WithName("bdc-cstor-" + string(bdObj.UID)).
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
	if k8serror.IsAlreadyExists(err) {
		glog.Infof("BDC for BD {%s} already created", bdObj.Name)
		return nil
	}
	if err != nil {
		return errors.Wrapf(err, "failed to create block device claim for bd {%s}", bdObj.Name)
	}
	return nil
}

// IsClaimedBDUsable returns true if the passed BD is already claimed and can be
// used for provisioning
func (ac *Config) IsClaimedBDUsable(bdAPIObj *ndmapis.BlockDevice) (bool, error) {
	bdObj := bd.BuilderForAPIObject(bdAPIObj)
	if bdObj.BlockDevice.IsClaimed() {
		bdcName := bdObj.BlockDevice.Object.Spec.ClaimRef.Name
		bdcAPIObject, err := bdc.NewKubeClient().WithNamespace(ac.Namespace).Get(bdcName, metav1.GetOptions{})
		if err != nil {
			return false, errors.Wrapf(err, "could not get block device claim for block device {%s}", bdAPIObj.Name)
		}
		bdcObj := bdc.BuilderForAPIObject(bdcAPIObject)
		if bdcObj.BDC.HasLabel(string(apis.CStorPoolClusterCPK), ac.CSPC.Name) {
			return true, nil
		}
	} else {
		return false, errors.Errorf("block device {%s} is not claimed", bdAPIObj.Name)
	}
	return false, nil
}

// ValidatePoolSpec validates the pool spec.
// TODO: Fix following function -- (Current is mock only )
func ValidatePoolSpec(pool *apis.PoolSpec) bool {
	return true
}
