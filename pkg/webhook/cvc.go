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
	"net/http"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
	cvc "github.com/openebs/maya/pkg/cstorvolumeclaim/v1alpha1"
	util "github.com/openebs/maya/pkg/util"
	"github.com/pkg/errors"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

type validateFunc func(cvcOldObj, cvcNewObj *apis.CStorVolumeClaim) error

type getCVC func(name, namespace string, clientset clientset.Interface) (*apis.CStorVolumeClaim, error)

func (wh *webhook) validateCVCUpdateRequest(req *v1beta1.AdmissionRequest, getCVC getCVC) *v1beta1.AdmissionResponse {
	response := NewAdmissionResponse().
		SetAllowed().
		WithResultAsSuccess(http.StatusAccepted).AR
	var cvcNewObj apis.CStorVolumeClaim
	err := json.Unmarshal(req.Object.Raw, &cvcNewObj)
	if err != nil {
		klog.Errorf("Couldn't unmarshal raw object: %v to cvc error: %v", req.Object.Raw, err)
		response = BuildForAPIObject(response).UnSetAllowed().WithResultAsFailure(err, http.StatusBadRequest).AR
		return response
	}

	// Get old CVC object by making call to etcd
	cvcOldObj, err := getCVC(cvcNewObj.Name, cvcNewObj.Namespace, wh.clientset)
	if err != nil {
		klog.Errorf("Failed to get CVC %s in namespace %s from etcd error: %v", cvcNewObj.Name, cvcNewObj.Namespace, err)
		response = BuildForAPIObject(response).UnSetAllowed().WithResultAsFailure(err, http.StatusBadRequest).AR
		return response
	}
	err = validateCVCSpecChanges(cvcOldObj, &cvcNewObj)
	if err != nil {
		klog.Errorf("invalid cvc changes: %s error: %s", cvcOldObj.Name, err.Error())
		response = BuildForAPIObject(response).UnSetAllowed().WithResultAsFailure(err, http.StatusBadRequest).AR
		return response
	}
	return response
}

func validateCVCSpecChanges(cvcOldObj, cvcNewObj *apis.CStorVolumeClaim) error {
	validateFuncList := []validateFunc{validateReplicaCount,
		validatePoolListChanges,
		validateReplicaScaling,
		validateStatusPoolList}
	for _, f := range validateFuncList {
		err := f(cvcOldObj, cvcNewObj)
		if err != nil {
			return err
		}
	}

	// Below validations should be done only with new CVC object
	err := validatePoolNames(cvcNewObj)
	if err != nil {
		return err
	}
	return nil
}

// TODO: isScalingInProgress(cvcObj *apis.CStorVolumeClaim) signature need to be
// updated to cvcObj.IsScaleingInProgress()
func isScalingInProgress(cvcObj *apis.CStorVolumeClaim) bool {
	return len(cvcObj.Spec.Policy.ReplicaPoolInfo) != len(cvcObj.Status.PoolInfo)
}

// validateReplicaCount returns error if user modified the replica count after
// provisioning the volume else return nil
func validateReplicaCount(cvcOldObj, cvcNewObj *apis.CStorVolumeClaim) error {
	if cvcOldObj.Spec.ReplicaCount != cvcNewObj.Spec.ReplicaCount {
		return errors.Errorf(
			"cvc %s replicaCount got modified from %d to %d",
			cvcOldObj.Name,
			cvcOldObj.Spec.ReplicaCount,
			cvcNewObj.Spec.ReplicaCount,
		)
	}
	return nil
}

// validatePoolListChanges returns error if user modified only the pool names
func validatePoolListChanges(cvcOldObj, cvcNewObj *apis.CStorVolumeClaim) error {
	oldDesiredPoolNames := cvc.GetDesiredReplicaPoolNames(cvcOldObj)
	newDesiredPoolNames := cvc.GetDesiredReplicaPoolNames(cvcNewObj)
	modifiedPoolNames := util.ListDiff(oldDesiredPoolNames, newDesiredPoolNames)
	if len(newDesiredPoolNames) >= len(oldDesiredPoolNames) {
		// If no.of pools on new spec >= no.of pools in old spec(scaleup as well
		// as migration then there all the pools in old spec must present in new
		// spec)
		if len(modifiedPoolNames) > 0 {
			return errors.Errorf(
				"volume replica migration directly by modifying pool names %v is not yet supported",
				modifiedPoolNames,
			)
		}
	} else {
		// If no.of pools in new spec < no.of pools in old spec(scale down
		// volume replica case) then there should at most one change in
		// oldSpec.PoolInfo - newSpec.PoolInfo
		if len(modifiedPoolNames) > 1 {
			return errors.Errorf(
				"Can't perform more than one replica scale down requested scale down count %d",
				len(modifiedPoolNames),
			)
		}
	}
	return nil
}

// validateReplicaScaling returns error if user updated pool list when scaling is
// already in progress
func validateReplicaScaling(cvcOldObj, cvcNewObj *apis.CStorVolumeClaim) error {
	if isScalingInProgress(cvcOldObj) {
		// if old and new CVC has same count of pools then return true else
		// return false
		if len(cvcOldObj.Spec.Policy.ReplicaPoolInfo) != len(cvcNewObj.Spec.Policy.ReplicaPoolInfo) {
			return errors.Errorf("scaling of CVC %s is already in progress", cvcOldObj.Name)
		}
	}
	return nil
}

// validateStatusPoolList return error if pools under status doesn't exist under
// existing spec else return nil
func validateStatusPoolList(cvcOldObj, cvcNewObj *apis.CStorVolumeClaim) error {
	replicaPoolNames := []string{}
	if len(cvcOldObj.Spec.Policy.ReplicaPoolInfo) == 0 &&
		len(cvcOldObj.Status.PoolInfo) == 0 {
		// Might be a case where controller updating Bound status along with
		// spec and status pool names
		replicaPoolNames = cvc.GetDesiredReplicaPoolNames(cvcNewObj)
	} else {
		replicaPoolNames = cvc.GetDesiredReplicaPoolNames(cvcOldObj)
	}
	// get pool names which are in status but not under spec
	invalidStatusPoolNames := util.ListDiff(cvcNewObj.Status.PoolInfo, replicaPoolNames)
	if len(invalidStatusPoolNames) > 0 {
		return errors.Errorf(
			"replica status pool names %v doesn't exist under spec pool list",
			invalidStatusPoolNames,
		)
	}
	return nil
}

// validatePoolNames returns error if there is repeatition of pool names either
// under spec or status of cvc
func validatePoolNames(cvcObj *apis.CStorVolumeClaim) error {
	// TODO: Change cvcObj.GetDesiredReplicaPoolNames()
	replicaPoolNames := cvc.GetDesiredReplicaPoolNames(cvcObj)
	// Check repeatition of pool names under Spec of CVC Object
	if !util.IsUniqueList(replicaPoolNames) {
		return errors.Errorf(
			"repeatition of pool names %v under spec of cvc %s",
			replicaPoolNames,
			cvcObj.Name,
		)
	}
	// Check repeatition of pool names under Status of CVC Object
	if !util.IsUniqueList(cvcObj.Status.PoolInfo) {
		return errors.Errorf(
			"repeatition of pool names %v under status of cvc %s",
			cvcObj.Status.PoolInfo,
			cvcObj.Name,
		)
	}
	return nil
}

func getCVCObject(name, namespace string,
	clientset clientset.Interface) (*apis.CStorVolumeClaim, error) {
	return clientset.OpenebsV1alpha1().
		CStorVolumeClaims(namespace).
		Get(name, metav1.GetOptions{})
}
