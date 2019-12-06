/*
Copyright 2019 The OpenEBS Authors.

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

package webhook

import (
	"fmt"
	nodeselect "github.com/openebs/maya/pkg/algorithm/nodeselect/v1alpha2"
	ndmapis "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	bd "github.com/openebs/maya/pkg/blockdevice/v1alpha2"
	bdc "github.com/openebs/maya/pkg/blockdeviceclaim/v1alpha1"
	cspcv1alpha1 "github.com/openebs/maya/pkg/cstor/poolcluster/v1alpha1"
	"github.com/openebs/maya/pkg/volume"
	"github.com/pkg/errors"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BlockDeviceReplacement contains old and new CSPC to validate for block device replacement
type BlockDeviceReplacement struct {
	// OldCSPC is the persisted CSPC in etcd.
	OldCSPC *apis.CStorPoolCluster
	// NewCSPC is the CSPC after it has been modified but yet not persisted to etcd.
	NewCSPC *apis.CStorPoolCluster
}

// NewBlockDeviceReplacement returns an empty BlockDeviceReplacement object.
func NewBlockDeviceReplacement() *BlockDeviceReplacement {
	return &BlockDeviceReplacement{
		OldCSPC: &apis.CStorPoolCluster{},
		NewCSPC: &apis.CStorPoolCluster{},
	}
}

// WithOldCSPC sets the old persisted CSPC into the BlockDeviceReplacement object.
func (bdr *BlockDeviceReplacement) WithOldCSPC(oldCSPC *apis.CStorPoolCluster) *BlockDeviceReplacement {
	bdr.OldCSPC = oldCSPC
	return bdr
}

// WithNewCSPC sets the new CSPC as a result of CSPC modification which is not yet persisted,
// into the BlockDeviceReplacement object
func (bdr *BlockDeviceReplacement) WithNewCSPC(newCSPC *apis.CStorPoolCluster) *BlockDeviceReplacement {
	bdr.NewCSPC = newCSPC
	return bdr
}

type poolspecs struct {
	oldSpec []apis.PoolSpec
	newSpec []apis.PoolSpec
}

// ValidateForBDReplacementCase validates the changes in CSPC for block device replacement in a raid group only if the
// update/edit of CSPC can trigger a block device replacement.
func ValidateForBDReplacementCase(commonPoolSpecs *poolspecs, bdr *BlockDeviceReplacement) (bool, string) {
	for i, oldPoolSpec := range commonPoolSpecs.oldSpec {
		oldPoolSpec := oldPoolSpec
		if ok, msg := bdr.IsReplacementValid(&oldPoolSpec, &commonPoolSpecs.newSpec[i]); !ok {
			return false, msg
		}
	}
	return true, ""
}

// getCommonPoolSpecs get the same pool specs from old persisted CSPC and the new CSPC after modification
// which is not persisted yet.
func getCommonPoolSpecs(cspcNew, cspcOld *apis.CStorPoolCluster) (*poolspecs, error) {
	commonPoolSpecs := &poolspecs{
		oldSpec: []apis.PoolSpec{},
		newSpec: []apis.PoolSpec{},
	}
	for _, oldPool := range cspcOld.Spec.Pools {
		oldNodeName, err := nodeselect.GetNodeFromLabelSelector(oldPool.NodeSelector)
		if err != nil {
			return nil, err
		}

		for _, newPool := range cspcNew.Spec.Pools {
			newNodeName, err := nodeselect.GetNodeFromLabelSelector(oldPool.NodeSelector)
			if err != nil {
				return nil, err
			}
			if oldNodeName == newNodeName {
				commonPoolSpecs.oldSpec = append(commonPoolSpecs.oldSpec, oldPool)
				commonPoolSpecs.newSpec = append(commonPoolSpecs.newSpec, newPool)
				break
			}
		}
	}
	return commonPoolSpecs, nil
}

// IsReplacementValid validates the pool specs on CSPC for block device replacement case.
func (bdr *BlockDeviceReplacement) IsReplacementValid(oldPoolSpec, newPoolSpec *apis.PoolSpec) (bool, string) {
	newToOldBd := make(map[string]string)
	for _, oldRg := range oldPoolSpec.RaidGroups {
		oldRg := oldRg // pin it
		if oldRg.Type == "" {
			oldRg.Type = oldPoolSpec.PoolConfig.DefaultRaidGroupType
		}
		for _, newRg := range newPoolSpec.RaidGroups {
			newRg := newRg // pin it
			if IsRaidGroupCommon(oldRg, newRg) {
				if IsBlockDeviceReplacementCase(&oldRg, &newRg) {
					if ok, msg := bdr.IsBDReplacementValid(&newRg, &oldRg); !ok {
						return false, msg
					}
					newBD := GetNewBDFromRaidGroups(&newRg, &oldRg)
					for k, v := range newBD {
						newToOldBd[k] = v
					}
				}
				break
			}
		}
	}

	for newBD, oldBD := range newToOldBd {
		err := createBDC(newBD, oldBD, bdr.OldCSPC)
		if err != nil {
			return false, err.Error()
		}
	}
	return true, ""
}

// IsRaidGroupCommon returns true if the provided raid groups are the same raid groups.
func IsRaidGroupCommon(rgOld, rgNew apis.RaidGroup) bool {
	oldBdMap := make(map[string]bool)
	for _, oldBD := range rgOld.BlockDevices {
		oldBdMap[oldBD.BlockDeviceName] = true
	}

	for _, newBD := range rgNew.BlockDevices {
		if oldBdMap[newBD.BlockDeviceName] {
			return true
		}
	}
	return false
}

// IsBlockDeviceReplacementCase returns true if the edit/update of CSPC can trigger a blockdevice
// replacement.
func IsBlockDeviceReplacementCase(newRaidGroup, oldRaidGroup *apis.RaidGroup) bool {
	count := GetNumberOfDiskReplaced(newRaidGroup, oldRaidGroup)
	return count >= 1
}

// GetNumberOfDiskReplaced returns the nuber of disk replaced in raid group.
func GetNumberOfDiskReplaced(newRG, oldRG *apis.RaidGroup) int {
	var count int
	oldBlockDevicesMap := make(map[string]bool)
	for _, bdOld := range oldRG.BlockDevices {
		oldBlockDevicesMap[bdOld.BlockDeviceName] = true
	}
	for _, newBD := range newRG.BlockDevices {
		if !oldBlockDevicesMap[newBD.BlockDeviceName] {
			count++
		}
	}
	return count
}

// IsBDReplacementValid validates for BD replacement.
func (bdr *BlockDeviceReplacement) IsBDReplacementValid(newRG, oldRG *apis.RaidGroup) (bool, string) {

	if oldRG.Type == string(apis.PoolStriped) {
		return false, "cannot replace  blockdevice in stripe raid group"
	}

	// Not more than 1 bd should be replaced in a raid group.
	if IsMoreThanOneDiskReplaced(newRG, oldRG) {
		return false, "cannot replace more than one blockdevice in a raid group"
	}

	// The incoming BD for replacement should not be present in the current CSPC.
	if bdr.IsNewBDPresentOnCurrentCSPC(newRG, oldRG) {
		return false, "the new blockdevice intended to use for replacement is already a part of the current cspc"
	}

	// No background replacement should be going on in the raid group undergoing replacement.
	if ok, err := bdr.IsExistingReplacmentInProgress(oldRG); ok {
		return false, fmt.Sprintf("cannot replace blockdevice as a "+
			"background replacement may be in progress in the raid group: %s", err.Error())
	}

	// The incoming BD should be a valid entry if
	// 1. The BD does not have a BDC.
	// 2. The BD has a BDC with the current CSPC label and there is no successor of this BD
	//    present in the CSPC.
	if !bdr.AreNewBDsValid(newRG, oldRG, bdr.OldCSPC) {
		return false, "the new blockdevice intended to use for replacement in invalid"
	}

	return true, ""
}

// IsMoreThanOneDiskReplaced returns true if more than one disk is replaced in the same raid group.
func IsMoreThanOneDiskReplaced(newRG, oldRG *apis.RaidGroup) bool {
	count := GetNumberOfDiskReplaced(newRG, oldRG)
	return count > 1
}

// IsNewBDPresentOnCurrentCSPC returns true if the new/incoming BD that will be used for replacement
// is already present in CSPC.
func (bdr *BlockDeviceReplacement) IsNewBDPresentOnCurrentCSPC(newRG, oldRG *apis.RaidGroup) bool {
	newBDs := GetNewBDFromRaidGroups(newRG, oldRG)
	for _, pool := range bdr.OldCSPC.Spec.Pools {
		for _, rg := range pool.RaidGroups {
			for _, bd := range rg.BlockDevices {
				if _, ok := newBDs[bd.BlockDeviceName]; ok {
					return true
				}
			}
		}
	}
	return false
}

// IsExistingReplacmentInProgress returns true if a block device in raid group is under active replacement.
func (bdr *BlockDeviceReplacement) IsExistingReplacmentInProgress(oldRG *apis.RaidGroup) (bool, error) {
	for _, v := range oldRG.BlockDevices {
		bdcObject, err := bdr.GetBDCOfBD(v.BlockDeviceName)
		if err != nil {
			return true, errors.Errorf("failed to query for any existing replacement in the raid group : %s", err.Error())
		}
		if bdcObject.HasAnnotationKey(apis.PredecessorBDKey) {
			return true, errors.Errorf("replacement is still in progress for bd %s", v.BlockDeviceName)
		}
	}
	return false, nil
}

// AreNewBDsValid returns true if the new BDs are valid BDs for replacement.
func (bdr *BlockDeviceReplacement) AreNewBDsValid(newRG, oldRG *apis.RaidGroup, oldcspc *apis.CStorPoolCluster) bool {
	newBDs := GetNewBDFromRaidGroups(newRG, oldRG)
	for bd := range newBDs {
		bdc, err := bdr.GetBDCOfBD(bd)
		if err != nil {
			return false
		}
		if !bdr.IsBDValid(bd, bdc, oldcspc) {
			return false
		}
	}
	return true
}

// IsBDValid returns true if the new BD is a valid BD for replacement.
func (bdr *BlockDeviceReplacement) IsBDValid(bd string, bdc *bdc.BlockDeviceClaim, oldcspc *apis.CStorPoolCluster) bool {
	if bdc != nil && !bdc.HasLabel(string(apis.CStorPoolClusterCPK), oldcspc.Name) {
		return false
	}
	predecessorMap, err := bdr.GetPredecessorBDIfAny(oldcspc)
	if err != nil {
		return false
	}
	if predecessorMap[bd] {
		return false
	}
	return true
}

// GetPredecessorBDIfAny returns a map of predecessor BDs if any in the current CSPC
// Note: Predecessor BDs in a CSPC are those BD for which a new BD has appeared in the CSPC and
//       replacement is still in progress
//
// For example,
// (b1,b2) is a group in cspc
// which has been changed to ( b3,b2 )  [Notice that b1 got replaced by b3],
// now b1 is not present in CSPC but the replacement is still in progress in background.
// In this case b1 is a predecessor BD.
func (bdr *BlockDeviceReplacement) GetPredecessorBDIfAny(cspcOld *apis.CStorPoolCluster) (map[string]bool, error) {
	predecessorBDMap := make(map[string]bool)
	for _, pool := range cspcOld.Spec.Pools {
		for _, rg := range pool.RaidGroups {
			for _, bd := range rg.BlockDevices {
				bdc, err := bdr.GetBDCOfBD(bd.BlockDeviceName)
				if err != nil {
					return nil, err
				}
				if bdc == nil {
					continue
				}
				predecessorBDMap[bdc.Object.GetAnnotations()[apis.PredecessorBDKey]] = true
			}
		}
	}
	return predecessorBDMap, nil
}

// GetBDCOfBD returns the BDC object for corresponding BD.
func (bdr *BlockDeviceReplacement) GetBDCOfBD(bdName string) (*bdc.BlockDeviceClaim, error) {
	bdcList, err := bdc.NewKubeClient().WithNamespace(bdr.OldCSPC.Namespace).List(v1.ListOptions{})
	if err != nil {
		return nil, errors.Errorf("failed to list bdc: %s", err.Error())
	}
	list := bdc.ListBuilderFromAPIList(bdcList).WithFilter(bdc.HasBD(bdName)).List()

	// If there is not BDC for a BD -- this means it an acceptable situation for BD replacement
	// The incoming BD finally will have a BDC created, hence no error is returned.
	if list.Len() == 0 {
		return nil, nil
	}

	if list.Len() != 1 {
		return nil, errors.Errorf("did not get exact one bdc for the bd %s", bdName)
	}
	return bdc.BuilderForAPIObject(&list.ObjectList.Items[0]).BDC, nil
}

func createBDC(newBD, oldBD string, cspcOld *apis.CStorPoolCluster) error {
	bdObj, err := bd.NewKubeClient().WithNamespace(cspcOld.Namespace).Get(newBD, v1.GetOptions{})
	if err != nil {
		return err
	}
	err = ClaimBD(bdObj, oldBD, cspcOld)
	if err != nil {
		return err
	}
	return nil
}

// ClaimBD claims a given BlockDevice
func ClaimBD(newBdObj *ndmapis.BlockDevice, oldBD string, cspcOld *apis.CStorPoolCluster) error {
	newBDCObj, err := bdc.NewBuilder().
		WithName("bdc-cstor-" + string(newBdObj.UID)).
		WithNamespace(newBdObj.Namespace).
		WithLabels(map[string]string{string(apis.CStorPoolClusterCPK): cspcOld.Name}).
		WithAnnotations(map[string]string{apis.PredecessorBDKey: oldBD}).
		WithBlockDeviceName(newBdObj.Name).
		WithHostName(newBdObj.Labels[string(apis.HostNameCPK)]).
		WithCapacity(volume.ByteCount(newBdObj.Spec.Capacity.Storage)).
		WithCSPCOwnerReference(cspcOld).
		WithFinalizer(cspcv1alpha1.CSPCFinalizer).
		Build()

	if err != nil {
		return errors.Wrapf(err, "failed to build block device claim for bd {%s}", newBdObj.Name)
	}

	bdcClient := bdc.NewKubeClient().WithNamespace(newBdObj.Namespace)
	bdcObj, err := bdcClient.Get(newBDCObj.Object.Name, v1.GetOptions{})
	if k8serror.IsNotFound(err) {
		_, err = bdcClient.Create(newBDCObj.Object)
		if err != nil {
			return errors.Wrapf(err, "failed to create block device claim for bd {%s}", newBdObj.Name)
		}
		return nil
	} else if err != nil {
		return errors.Wrapf(err, "failed to get block device claim for bd {%s}", newBdObj.Name)
	}

	updatedBDCObj, err := bdc.BuilderForAPIObject(bdcObj).
		WithAnnotations(map[string]string{apis.PredecessorBDKey: oldBD}).
		Build()
	if err != nil {
		return errors.Wrapf(err, "failed to add annotation on block device claim {%s}", bdcObj.Name)
	}

	_, err = bdcClient.
		Update(updatedBDCObj.Object)
	return err
}

// GetNewBDFromRaidGroups returns a map of new successor bd to old bd for replacement in a raid group
func GetNewBDFromRaidGroups(newRG, oldRG *apis.RaidGroup) map[string]string {
	newToOldBlockDeviceMap := make(map[string]string)
	oldBlockDevicesMap := make(map[string]bool)
	newBlockDevicesMap := make(map[string]bool)

	for _, bdOld := range oldRG.BlockDevices {
		oldBlockDevicesMap[bdOld.BlockDeviceName] = true
	}

	for _, bdNew := range newRG.BlockDevices {
		newBlockDevicesMap[bdNew.BlockDeviceName] = true
	}
	var newBD, oldBD string

	for _, newRG := range newRG.BlockDevices {
		if !oldBlockDevicesMap[newRG.BlockDeviceName] {
			newBD = newRG.BlockDeviceName
			break
		}
	}

	for _, oldRG := range oldRG.BlockDevices {
		if !newBlockDevicesMap[oldRG.BlockDeviceName] {
			oldBD = oldRG.BlockDeviceName
			break
		}
	}
	newToOldBlockDeviceMap[newBD] = oldBD
	return newToOldBlockDeviceMap
}
