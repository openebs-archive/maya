/*
Copyright 2019 The OpenEBS Authors

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
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=kubeassert

// KubeAssert contains the desired assertions of one or
// more resources
type KubeAssert struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubeAssertSpec   `json:"spec"`
	Status KubeAssertStatus `json:"status"`
}

// KubeAssertSpec provides the specifications of a KubeAssert
type KubeAssertSpec struct {
	// List of checks that gets verified
	Checks []Check `json:"checks"`
}

// Check describes one verification
type Check struct {
	// Description of this check
	Desc string `json:"desc"`
	// Given represents the current resource(s) available
	// in the cluster
	Given Given `json:"given"`
	// Then specifies what is expected from the given
	// resource(s)
	Then Then `json:"then"`
}

// Given represents the current resource(s) available
// in the cluster
type Given struct {
	// resource kind to be verified
	Kind string `json:"kind"`
	// name of the resource to be verified
	Name string `json:"name"`
	// namespace of the resource(s) to be verified
	Namespace string `json:"namespace"`
	// api version of the resource(s) to be verified
	APIVersion string `json:"apiVersion"`
	// labels of resource(s) to be verified
	LabelSelector string `json:"labelSelector"`
	// annotations of resource(s) to be verified
	AnnotationSelector string `json:"annotationSelector"`
}

// Then specifies what is expected from the given
// resource(s)
type Then struct {
	// Expect represents the verification conditions
	Expect Expect `json:"expect"`
	// Options to be used during verification
	Options Options `json:"options"`
}

// Expect represents the expectation specifications
type Expect struct {
	// Match is a list of expected matches
	Match []string `json:"match"`
}

// Options represent the tunables that may be
// used while verifying the expectations
type Options struct {
	// Number of seconds before expectation is initiated.
	InitialDelaySeconds int32 `json:"initialDelaySeconds"`
	// Number of seconds after which the handler times out.
	TimeoutSeconds int32 `json:"timeoutSeconds"`
	// How often (in seconds) to perform the check.
	PeriodSeconds int32 `json:"periodSeconds"`
	// Minimum consecutive successes for the probe to be considered
	// successful after having failed.
	SuccessThreshold int32 `json:"successThreshold"`
	// Minimum consecutive failures for the probe to be considered
	// failed after having succeeded.
	FailureThreshold int32 `json:"failureThreshold"`
}

// KubeAssertStatus represents the current state of KubeAssert
type KubeAssertStatus struct {
	Phase string `json:"phase"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=kubeasserts

// KubeAssertList is a list of KubeAsserts
type KubeAssertList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []KubeAssert `json:"items"`
}
