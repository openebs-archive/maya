// Copyright © 2017-2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package k8s

import (
	magentyptesv1 "github.com/openebs/maya/cmd/maya-agent/types/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// #############################################################################

// Note: The following code is boilerplate code needed to satisfy the
// StorageBackendAdaptor as a resource in the cluster in terms of how it
// expects CRD's to be created, operate and used.

// #############################################################################

// StorageBackendAdaptorSpec is a specification for a storage backend adaptor resource
type StorageBackendAdaptorSpec struct {
	metav1.TypeMeta `json:",inline"`
	Metadata        metav1.ObjectMeta `json:"metadata"`

	Spec magentyptesv1.StorageBackendAdaptor `json:"spec"`
}

// StorageBackendAdaptorList struct is a list of storage backend adaptor
type StorageBackendAdaptorList struct {
	metav1.TypeMeta `json:",inline"`
	Metadata        metav1.ListMeta `json:"metadata"`

	Items []StorageBackendAdaptorSpec `json:"items"`
}
