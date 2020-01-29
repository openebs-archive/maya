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
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	nodeselect "github.com/openebs/maya/pkg/algorithm/nodeselect/v1alpha2"
	ndmapis "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	blockdevice "github.com/openebs/maya/pkg/blockdevice/v1alpha2"
	blockdeviceclaim "github.com/openebs/maya/pkg/blockdeviceclaim/v1alpha1"
	cspcv1alpha1 "github.com/openebs/maya/pkg/cstor/poolcluster/v1alpha1"
	cspi "github.com/openebs/maya/pkg/cstor/poolinstance/v1alpha3"
	cvr "github.com/openebs/maya/pkg/cstor/volumereplica/v1alpha1"
	env "github.com/openebs/maya/pkg/env/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

// TODO: Make better naming conventions from review comments

// PoolValidator is build to validate pool spec, raid groups and blockdevices
type PoolValidator struct {
	poolSpec  *apis.PoolSpec
	namespace string
	nodeName  string
	cspcName  string
}

// Builder is the builder object for Builder
type Builder struct {
	object *PoolValidator
}

// NewPoolSpecValidator returns new instance of poolValidator
func NewPoolSpecValidator() *PoolValidator {
	return &PoolValidator{}
}

// NewBuilder returns new instance of builder
func NewBuilder() *Builder {
	return &Builder{object: NewPoolSpecValidator()}
}

// build returns built instance of PoolValidator
func (b *Builder) build() *PoolValidator {
	return b.object
}

// withPoolSpec sets the poolSpec field of PoolValidator with provided values
func (b *Builder) withPoolSpec(poolSpec apis.PoolSpec) *Builder {
	b.object.poolSpec = &poolSpec
	return b
}

// withPoolNamespace sets the namespace field of poolValidator with provided
// values
func (b *Builder) withPoolNamespace() *Builder {
	b.object.namespace = env.Get(env.OpenEBSNamespace)
	return b
}

// withPoolNodeName sets the node name field of poolValidator with provided
// values
func (b *Builder) withPoolNodeName(nodeName string) *Builder {
	b.object.nodeName = nodeName
	return b
}

// withCSPCName sets the cspc name field of poolValidator with provided argument
func (b *Builder) withCSPCName(cspcName string) *Builder {
	b.object.cspcName = cspcName
	return b
}

// validateCSPC validates CSPC spec for Create, Update and Delete operation of the object.
func (wh *webhook) validateCSPC(ar *v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	req := ar.Request
	response := &v1beta1.AdmissionResponse{}
	// validates only if requested operation is CREATE or UPDATE
	if req.Operation == v1beta1.Update {
		klog.V(5).Infof("Admission webhook update request for type %s", req.Kind.Kind)
		return wh.validateCSPCUpdateRequest(req)
	} else if req.Operation == v1beta1.Create {
		klog.V(5).Infof("Admission webhook create request for type %s", req.Kind.Kind)
		return wh.validateCSPCCreateRequest(req)
	} else if req.Operation == v1beta1.Delete {
		klog.V(5).Infof("Admission webhook delete request for type %s", req.Kind.Kind)
		return wh.validateCSPCDeleteRequest(req)
	}

	return response
}

// validateCSPCCreateRequest validates CSPC create request
func (wh *webhook) validateCSPCCreateRequest(req *v1beta1.AdmissionRequest) *v1beta1.AdmissionResponse {
	response := NewAdmissionResponse().SetAllowed().WithResultAsSuccess(http.StatusAccepted).AR
	var cspc apis.CStorPoolCluster
	err := json.Unmarshal(req.Object.Raw, &cspc)
	if err != nil {
		klog.Errorf("Could not unmarshal raw object: %v, %v", err, req.Object.Raw)
		response = BuildForAPIObject(response).UnSetAllowed().WithResultAsFailure(err, http.StatusBadRequest).AR
		return response
	}
	if ok, msg := cspcValidation(&cspc); !ok {
		err := errors.Errorf("invalid cspc specification: %s", msg)
		response = BuildForAPIObject(response).UnSetAllowed().WithResultAsFailure(err, http.StatusUnprocessableEntity).AR
		return response
	}
	return response
}

// validateCSPCDeleteRequest validates CSPC delete request
// if any cvrs exist on the cspc pools then deletion is invalid
func (wh *webhook) validateCSPCDeleteRequest(req *v1beta1.AdmissionRequest) *v1beta1.AdmissionResponse {
	response := NewAdmissionResponse().SetAllowed().WithResultAsSuccess(http.StatusAccepted).AR
	cspiList, err := cspi.NewKubeClient().WithNamespace(req.Namespace).List(
		metav1.ListOptions{
			LabelSelector: string(apis.CStorPoolClusterCPK) + "=" + req.Name,
		})
	if err != nil {
		klog.Errorf("Could not list cspi for cspc %s: %s", req.Name, err.Error())
		response = BuildForAPIObject(response).UnSetAllowed().WithResultAsFailure(err, http.StatusBadRequest).AR
		return response
	}
	for _, cspiObj := range cspiList.Items {
		// list cvrs in all namespaces
		cvrList, err := cvr.NewKubeclient().WithNamespace("").List(metav1.ListOptions{
			LabelSelector: "cstorpoolinstance.openebs.io/name=" + cspiObj.Name,
		})
		if err != nil {
			klog.Errorf("Could not list cvr for cspi %s: %s", cspiObj.Name, err.Error())
			response = BuildForAPIObject(response).UnSetAllowed().WithResultAsFailure(err, http.StatusBadRequest).AR
			return response
		}
		if len(cvrList.Items) != 0 {
			err := errors.Errorf("invalid cspc %s deletion: volume still exists on pool %s", req.Name, cspiObj.Name)
			response = BuildForAPIObject(response).UnSetAllowed().WithResultAsFailure(err, http.StatusUnprocessableEntity).AR
			return response
		}
	}
	return response
}

func cspcValidation(cspc *apis.CStorPoolCluster) (bool, string) {
	usedNodes := map[string]bool{}
	if len(cspc.Spec.Pools) == 0 {
		return false, fmt.Sprintf("pools in cspc should have at least one item")
	}

	repeatedBlockDevices := getDuplicateBlockDeviceList(cspc)
	if len(repeatedBlockDevices) > 0 {
		return false, fmt.Sprintf("invalid cspc: cspc {%s} has duplicate blockdevices entries %v",
			cspc.Name,
			repeatedBlockDevices)
	}

	buildPoolValidator := NewBuilder().
		withPoolNamespace().
		withCSPCName(cspc.Name)
	for _, pool := range cspc.Spec.Pools {
		pool := pool // pin it
		nodeName, err := nodeselect.GetNodeFromLabelSelector(pool.NodeSelector)
		if err != nil {
			return false, fmt.Sprintf(
				"failed to get node from pool nodeSelector: {%v} error: {%v}",
				pool.NodeSelector,
				err,
			)
		}
		if usedNodes[nodeName] {
			return false, fmt.Sprintf("invalid cspc: duplicate node %s entry", nodeName)
		}
		usedNodes[nodeName] = true
		pValidate := buildPoolValidator.withPoolSpec(pool).
			withPoolNodeName(nodeName).build()
		ok, msg := pValidate.poolSpecValidation()
		if !ok {
			return false, fmt.Sprintf("invalid pool spec: %s", msg)
		}
	}
	return true, ""
}

// getDuplicateBlockDeviceList returns list of block devices that are
// duplicated in CSPC
func getDuplicateBlockDeviceList(cspc *apis.CStorPoolCluster) []string {
	duplicateBlockDeviceList := []string{}
	blockDeviceMap := map[string]bool{}
	addedBlockDevices := map[string]bool{}
	for _, poolSpec := range cspc.Spec.Pools {
		for _, raidGroup := range poolSpec.RaidGroups {
			for _, bd := range raidGroup.BlockDevices {
				// update duplicateBlockDeviceList only if block device is
				// repeated in CSPC and doesn't exist in duplicate block device
				// list.
				if blockDeviceMap[bd.BlockDeviceName] &&
					!addedBlockDevices[bd.BlockDeviceName] {
					duplicateBlockDeviceList = append(
						duplicateBlockDeviceList,
						bd.BlockDeviceName)
					addedBlockDevices[bd.BlockDeviceName] = true
				} else if !blockDeviceMap[bd.BlockDeviceName] {
					blockDeviceMap[bd.BlockDeviceName] = true
				}
			}
		}
	}
	return duplicateBlockDeviceList
}

func (poolValidator *PoolValidator) poolSpecValidation() (bool, string) {
	if len(poolValidator.poolSpec.RaidGroups) == 0 {
		return false, "at least one raid group should be present on pool spec"
	}
	// TODO : Add validation for pool config
	// Pool config will require mutating webhooks also.
	for _, raidGroup := range poolValidator.poolSpec.RaidGroups {
		raidGroup := raidGroup // pin it
		ok, msg := poolValidator.raidGroupValidation(&raidGroup)
		if !ok {
			return false, msg
		}
	}

	return true, ""
}

func (poolValidator *PoolValidator) raidGroupValidation(
	raidGroup *apis.RaidGroup) (bool, string) {
	if raidGroup.Type == "" &&
		poolValidator.poolSpec.PoolConfig.DefaultRaidGroupType == "" {
		return false, fmt.Sprintf("any one type at raid group or default raid group type be specified ")
	}
	if _, ok := apis.SupportedPRaidType[apis.PoolType(raidGroup.Type)]; !ok {
		return false, fmt.Sprintf("unsupported raid type '%s' specified", apis.PoolType(raidGroup.Type))
	}

	if len(raidGroup.BlockDevices) == 0 {
		return false, fmt.Sprintf("number of block devices honouring raid type should be specified")
	}

	if raidGroup.Type != string(apis.PoolStriped) {
		if len(raidGroup.BlockDevices) != apis.SupportedPRaidType[apis.PoolType(raidGroup.Type)] {
			return false, fmt.Sprintf("number of block devices honouring raid type should be specified")
		}
	} else {
		if len(raidGroup.BlockDevices) < apis.SupportedPRaidType[apis.PoolType(raidGroup.Type)] {
			return false, fmt.Sprintf("number of block devices honouring raid type should be specified")
		}
	}

	for _, bd := range raidGroup.BlockDevices {
		bd := bd
		ok, msg := poolValidator.blockDeviceValidation(&bd)
		if !ok {
			return false, msg
		}
	}
	return true, ""
}

// blockDeviceValidation validates following steps:
// 1. block device name shouldn't be empty.
// 2. If block device has claim it verifies whether claim is created by this CSPC
func (poolValidator *PoolValidator) blockDeviceValidation(
	bd *apis.CStorPoolClusterBlockDevice) (bool, string) {
	if bd.BlockDeviceName == "" {
		return false, fmt.Sprint("block device name cannot be empty")
	}
	bdObj, err := blockdevice.NewKubeClient().
		WithNamespace(poolValidator.namespace).
		Get(bd.BlockDeviceName, metav1.GetOptions{})
	if err != nil {
		return false, fmt.Sprintf(
			"failed to get block device: {%s} details error: %v",
			bd.BlockDeviceName,
			err,
		)
	}
	err = blockdevice.
		BuilderForAPIObject(bdObj).
		BlockDevice.
		ValidateBlockDevice(
			blockdevice.CheckIfBDIsActive(),
			blockdevice.CheckIfBDIsNonFsType(),
			blockdevice.CheckIfBDBelongsToNode(poolValidator.nodeName))
	if err != nil {
		return false, fmt.Sprintf("%v", err)
	}
	if bdObj.Spec.NodeAttributes.NodeName != poolValidator.nodeName {
		return false, fmt.Sprintf(
			"pool validation failed: block device %s doesn't belongs to pool node %s",
			bd.BlockDeviceName,
			poolValidator.nodeName,
		)
	}
	if bdObj.Status.ClaimState == ndmapis.BlockDeviceClaimed {
		// TODO: Need to check how NDM
		if bdObj.Spec.ClaimRef != nil {
			bdcName := bdObj.Spec.ClaimRef.Name
			if err := poolValidator.blockDeviceClaimValidation(bdcName, bdObj.Name); err != nil {
				return false, fmt.Sprintf("error: %v", err)
			}
		}
	}
	return true, ""
}

func (poolValidator *PoolValidator) blockDeviceClaimValidation(bdcName, bdName string) error {
	bdcObject, err := blockdeviceclaim.NewKubeClient().
		WithNamespace(poolValidator.namespace).
		Get(bdcName, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err,
			"could not get block device claim for block device {%s}", bdName)
	}
	cspcName := bdcObject.
		GetAnnotations()[string(apis.CStorPoolClusterCPK)]
	if cspcName != poolValidator.cspcName {
		return errors.Wrapf(err,
			"cann't use claimed blockdevice %s",
			bdName,
		)
	}
	return nil
}

// validateCSPCUpdateRequest validates CSPC update request
// ToDo: Remove repetitive code.
func (wh *webhook) validateCSPCUpdateRequest(req *v1beta1.AdmissionRequest) *v1beta1.AdmissionResponse {
	response := NewAdmissionResponse().SetAllowed().WithResultAsSuccess(http.StatusAccepted).AR
	var cspcNew apis.CStorPoolCluster
	err := json.Unmarshal(req.Object.Raw, &cspcNew)
	if err != nil {
		klog.Errorf("Could not unmarshal raw object: %v, %v", err, req.Object.Raw)
		response = BuildForAPIObject(response).UnSetAllowed().WithResultAsFailure(err, http.StatusBadRequest).AR
		return response
	}
	// Get CSPC old object
	cspcOld, err := cspcv1alpha1.NewKubeClient().WithNamespace(cspcNew.Namespace).Get(cspcNew.Name, v1.GetOptions{})
	if err != nil {
		err = errors.Errorf("could not fetch existing cspc for validation: %s", err.Error())
		response = BuildForAPIObject(response).UnSetAllowed().WithResultAsFailure(err, http.StatusInternalServerError).AR
		return response
	}

	// return success from here when there is no change in old and new spec
	if reflect.DeepEqual(cspcNew.Spec, cspcOld.Spec) {
		return response
	}

	if ok, msg := cspcValidation(&cspcNew); !ok {
		err = errors.Errorf("invalid cspc specification: %s", msg)
		response = BuildForAPIObject(response).UnSetAllowed().WithResultAsFailure(err, http.StatusUnprocessableEntity).AR
		return response
	}

	bdr := NewBlockDeviceReplacement().WithNewCSPC(&cspcNew).WithOldCSPC(cspcOld)
	commonPoolSpec, err := getCommonPoolSpecs(&cspcNew, cspcOld)

	if err != nil {
		err = errors.Errorf("could not find common pool specs for validation: %s", err.Error())
		response = BuildForAPIObject(response).UnSetAllowed().WithResultAsFailure(err, http.StatusInternalServerError).AR
		return response
	}

	if ok, msg := ValidateSpecChanges(commonPoolSpec, bdr); !ok {
		err = errors.Errorf("invalid cspc specification: %s", msg)
		response = BuildForAPIObject(response).UnSetAllowed().WithResultAsFailure(err, http.StatusUnprocessableEntity).AR
		return response
	}

	return response
}
