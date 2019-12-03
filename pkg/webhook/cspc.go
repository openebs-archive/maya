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
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	cspcv1alpha1 "github.com/openebs/maya/pkg/cstor/poolcluster/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"net/http"
)

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
	}

	klog.V(2).Info("Admission wehbook for PVC not " +
		"configured for operations other than UPDATE and CREATE")
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

func cspcValidation(cspc *apis.CStorPoolCluster) (bool, string) {
	if len(cspc.Spec.Pools) == 0 {
		return false, fmt.Sprintf("pools in cspc should have at least one item")
	}

	for _, pool := range cspc.Spec.Pools {
		pool := pool // pin it
		ok, msg := poolSpecValidation(&pool)
		if !ok {
			return false, fmt.Sprintf("invalid pool spec: %s", msg)
		}
	}
	return true, ""
}

func poolSpecValidation(pool *apis.PoolSpec) (bool, string) {
	if pool.NodeSelector == nil || len(pool.NodeSelector) == 0 {
		return false, "nodeselector should not be empty"
	}
	if len(pool.RaidGroups) == 0 {
		return false, "at least one raid group should be present on pool spec"
	}
	// TODO : Add validation for pool config
	// Pool config will require mutating webhooks also.
	for _, raidGroup := range pool.RaidGroups {
		raidGroup := raidGroup // pin it
		ok, msg := raidGroupValidation(&raidGroup, &pool.PoolConfig)
		if !ok {
			return false, msg
		}
	}

	return true, ""
}

func raidGroupValidation(raidGroup *apis.RaidGroup, pool *apis.PoolConfig) (bool, string) {
	if raidGroup.Type == "" && pool.DefaultRaidGroupType == "" {
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
		ok, msg := blockDeviceValidation(&bd)
		if !ok {
			return false, msg
		}
	}
	return true, ""
}

func blockDeviceValidation(bd *apis.CStorPoolClusterBlockDevice) (bool, string) {
	if bd.BlockDeviceName == "" {
		return false, fmt.Sprint("block device name cannot be empty")
	}
	return true, ""
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
	if ok, msg := cspcValidation(&cspcNew); !ok {
		err = errors.Errorf("invalid cspc specification: %s", msg)
		response = BuildForAPIObject(response).UnSetAllowed().WithResultAsFailure(err, http.StatusUnprocessableEntity).AR
		return response
	}

	cspcOld, err := cspcv1alpha1.NewKubeClient().WithNamespace(cspcNew.Namespace).Get(cspcNew.Name, v1.GetOptions{})
	if err != nil {
		err = errors.Errorf("could not fetch existing cspc for validation: %s", err.Error())
		response = BuildForAPIObject(response).UnSetAllowed().WithResultAsFailure(err, http.StatusInternalServerError).AR
		return response
	}

	bdr := NewBlockDeviceReplacement().WithNewCSPC(&cspcNew).WithOldCSPC(cspcOld)
	commonPoolSpec, err := getCommonPoolSpecs(&cspcNew, cspcOld)

	if err != nil {
		err = errors.Errorf("could not find common pool specs for validation: %s", err.Error())
		response = BuildForAPIObject(response).UnSetAllowed().WithResultAsFailure(err, http.StatusInternalServerError).AR
		return response
	}

	if ok, msg := ValidateForBDReplacementCase(commonPoolSpec, bdr); !ok {
		err = errors.Errorf("invalid cspc specification: %s", msg)
		response = BuildForAPIObject(response).UnSetAllowed().WithResultAsFailure(err, http.StatusUnprocessableEntity).AR
		return response
	}

	return response
}
