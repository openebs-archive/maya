/*
Copyright 2018 The OpenEBS Authors.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=newtestcstorpool

// NewTestCStorPool describes a cstor pool resource created as custom resource.
type NewTestCStorPool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NewCStorPoolSpec `json:"spec"`
	Status CStorPoolStatus  `json:"status"`
}

// NewCStorPoolSpec is the spec listing fields for a CStorPool resource.
type NewCStorPoolSpec struct {
	// HostName is the name of kubernetes node where the pool
	// should be created.
	HostName string `json:"hostName"`
	// NodeSelector is the labels that will be used to select
	// a node for pool provisioning.
	// Required field
	NodeSelector map[string]string `json:"nodeSelector"`
	// PoolConfig is the default pool config that applies to the
	// pool on node.
	PoolConfig PoolConfig `json:"poolConfig"`
	// RaidGroup is the group containing block devices
	RaidGroup []RaidGroup `json:"raidGroup"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=newtestcstorpool

// NewTestCStorPoolList is a list of CStorPoolList resources
type NewTestCStorPoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []NewTestCStorPool `json:"items"`
}
