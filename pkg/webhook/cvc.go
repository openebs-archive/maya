/*
Copyright 2020 The OpenEBS Authors.

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
	}
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

// validatePoolListChanges returns error if user modified existing pool names with new
// pool name(s) or if user performed more than one replica scale down at a time
func validatePoolListChanges(cvcOldObj, cvcNewObj *apis.CStorVolumeClaim) error {
	// Check the new CVC spec changes with old CVC status(Comparing with status
	// is more appropriate than comparing with spec)
	oldCurrentPoolNames := cvcOldObj.Status.PoolInfo
	newDesiredPoolNames := cvc.GetDesiredReplicaPoolNames(cvcNewObj)
	modifiedPoolNames := util.ListDiff(oldCurrentPoolNames, newDesiredPoolNames)
	if len(newDesiredPoolNames) >= len(oldCurrentPoolNames) {
		// If no.of pools on new spec >= no.of pools in old status(scaleup as well
		// as migration case then all the pools in old status must present in new
		// spec)
		if len(modifiedPoolNames) > 0 {
			return errors.Errorf(
				"volume replica migration directly by modifying pool names %v is not yet supported",
				modifiedPoolNames,
			)
		}
	} else {
		// If no.of pools in new spec < no.of pools in old status(scale down
		// volume replica case) then there should at most one change in
		// oldSpec.PoolInfo - newSpec.PoolInfo
		if len(modifiedPoolNames) > 1 {
			return errors.Errorf(
				"Can't perform more than one replica scale down requested scale down count %d",
				len(modifiedPoolNames),
			)
		}
	}
	// Reject the request if someone perform scaling when CVC is not in Bound
	// state
	// NOTE: We should not reject the controller request which Updates status as
	// Bound as well as pool info in status and spec
	// TODO: Make below check as cvcOldObj.ISBound()
	// If CVC Status is not bound then reject
	if cvcOldObj.Status.Phase != apis.CStorVolumeClaimPhaseBound {
		// If controller is updating pool info then new CVC will be in bound state
		if cvcNewObj.Status.Phase != apis.CStorVolumeClaimPhaseBound &&
			// Performed scaling operation on CVC
			len(oldCurrentPoolNames) != len(newDesiredPoolNames) {
			return errors.Errorf(
				"Can't perform scaling of volume replicas when CVC is not in %s state",
				apis.CStorVolumeClaimPhaseBound,
			)
		}
	}
	return nil
}

// validateReplicaScaling returns error if user updated pool list when scaling is
// already in progress.
// Note: User can perform scaleup of multiple replicas by adding multiple pool
//       names at time but not by updating CVC pool names with multiple edits.
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

// validatePoolNames returns error if there is repeatition of pool names either
// under spec or status of cvc
func validatePoolNames(cvcObj *apis.CStorVolumeClaim) error {
	// TODO: Change cvcObj.GetDesiredReplicaPoolNames()
	replicaPoolNames := cvc.GetDesiredReplicaPoolNames(cvcObj)
	// Check repeatition of pool names under Spec of CVC Object
	if !util.IsUniqueList(replicaPoolNames) {
		return errors.Errorf(
			"duplicate pool names %v found under spec of cvc %s",
			replicaPoolNames,
			cvcObj.Name,
		)
	}
	// Check repeatition of pool names under Status of CVC Object
	if !util.IsUniqueList(cvcObj.Status.PoolInfo) {
		return errors.Errorf(
			"duplicate pool names %v found under status of cvc %s",
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
