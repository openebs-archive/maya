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
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AdmissionResponse embeds K8S admission response API.
type AdmissionResponse struct {
	AR *v1beta1.AdmissionResponse
}

// NewAdmissionResponse returns an empty instance of AdmissionResponse.
func NewAdmissionResponse() *AdmissionResponse {
	return &AdmissionResponse{AR: &v1beta1.AdmissionResponse{}}
}

// WithResultAsFailure sets failure result.
func (ar *AdmissionResponse) WithResultAsFailure(err error, code int32) *AdmissionResponse {
	ar.AR.Result = &metav1.Status{
		Status:  metav1.StatusFailure,
		Code:    code,
		Reason:  metav1.StatusReasonBadRequest,
		Message: err.Error(),
	}
	return ar
}

// WithResultAsSuccess sets success result.
func (ar *AdmissionResponse) WithResultAsSuccess(code int32) *AdmissionResponse {
	ar.AR.Result = &metav1.Status{
		Status: metav1.StatusSuccess,
		Code:   code,
	}
	return ar
}

// SetAllowed sets allowed to true.
func (ar *AdmissionResponse) SetAllowed() *AdmissionResponse {
	ar.AR.Allowed = true
	return ar
}

// UnSetAllowed sets allowed to false.
func (ar *AdmissionResponse) UnSetAllowed() *AdmissionResponse {
	ar.AR.Allowed = false
	return ar
}

// BuildForAPIObject builds for api admission response object.
func BuildForAPIObject(ar *v1beta1.AdmissionResponse) *AdmissionResponse {
	return &AdmissionResponse{AR: ar}
}
