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
	"context"
	"net/http"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/pkg/errors"
	v1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

func (wh *webhook) validateSPCDeleteRequest(req *v1.AdmissionRequest) *v1.AdmissionResponse {
	response := NewAdmissionResponse().
		SetAllowed().
		WithResultAsSuccess(http.StatusAccepted).AR

	spcObj, err := wh.clientset.OpenebsV1alpha1().StoragePoolClaims().
		Get(context.TODO(), req.Name, metav1.GetOptions{})
	if err != nil {
		err = errors.Wrapf(err, "failed to get spc %s", req.Name)
		response = BuildForAPIObject(response).UnSetAllowed().WithResultAsFailure(err, http.StatusBadRequest).AR
		return response
	}

	if value := spcObj.GetAnnotations()[skipValidation]; value == "true" {
		klog.Infof("Skipping validations for %s due to SPC has skip validation", spcObj.Name)
		return response
	}

	cspList, err := wh.clientset.OpenebsV1alpha1().CStorPools().List(context.TODO(),
		metav1.ListOptions{
			LabelSelector: string(apis.StoragePoolClaimCPK) + "=" + req.Name,
		})
	if err != nil {
		err = errors.Wrapf(err, "could not list csp for spc %s", req.Name)
		klog.Error(err)
		response = BuildForAPIObject(response).UnSetAllowed().WithResultAsFailure(err, http.StatusBadRequest).AR
		return response
	}
	for _, cspObj := range cspList.Items {
		// list cvrs in all namespaces
		cvrList, err := wh.clientset.OpenebsV1alpha1().CStorVolumeReplicas("").
			List(context.TODO(), metav1.ListOptions{
				LabelSelector: string(apis.CStorPoolKey) + "=" + cspObj.Name,
			})
		if err != nil {
			err = errors.Wrapf(err, "Could not list cvr for csp %s", cspObj.Name)
			klog.Error(err)
			response = BuildForAPIObject(response).UnSetAllowed().WithResultAsFailure(err, http.StatusBadRequest).AR
			return response
		}
		if len(cvrList.Items) != 0 {
			err := errors.Errorf("invalid spc %s deletion: volumereplicas still exists on pool %s", req.Name, cspObj.Name)
			klog.Error(err)
			response = BuildForAPIObject(response).UnSetAllowed().WithResultAsFailure(err, http.StatusUnprocessableEntity).AR
			return response
		}
	}

	return response
}
