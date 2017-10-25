/*
Copyright 2017 The Kubernetes Authors.
Copyright 2017 The OpenEBS Authors.

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

// This file deals with the labels, constants, defaults & utility
// functions related to the structures of K8s StorageClass Kind
package v1

import (
	"fmt"
	"strings"
)

// SCProvision type deals with the keys associated with
// volume provisioning. It is expected that these keys will have
// corresponding values.
type SCProvisionKey string

const (
	SCReplicaCountProKey         SCProvisionKey = "provision/replica-count"
	SCControllerCountProKey      SCProvisionKey = "provision/controller-count"
	SCReplicaTolerationProKey    SCProvisionKey = "provision/replica-toleration"
	SCControllerTolerationProKey SCProvisionKey = "provision/controller-toleration"
)

// SCPolicyKey constant is the key associated with one or more volume
// policies. It is expected that this key will have corresponding
// policies.
const (
	SCPolicyKey string = "policy"
)

// SCPolicy type deals with various volume policies. These can
// be thought of as flags. Presence of these policy flags indicate
// enabling of corresponding policy.
type SCPolicy string

const (
	SCSpreadReplicaPolicy SCPolicy = "spread-replica"
	SCStickyReplicaPolicy SCPolicy = "sticky-replica"
)

// IsStickyReplicaInSCPol indicates if replicas need to stick
// to their placement Nodes.
func IsStickyReplicaInSCPol(policies map[string]string) (bool, error) {
	// TODO logic
	return false, fmt.Errorf("Not implemented")
}

// IsSpreadReplicaInSCPol indicates if spreading the replicas across Nodes
// is required ?
func IsSpreadReplicaInSCPol(policies map[string]string) (bool, error) {
	// TODO logic
	return false, fmt.Errorf("Not implemented")
}

// GetReplicaCountInSCProv gets the value of replica count if available
func GetReplicaCountInSCProv(provisionings map[string]string) (string, bool) {
	for k, v := range provisionings {
		if strings.TrimSpace(k) == string(SCReplicaCountProKey) {
			return strings.TrimSpace(v), true
		}
	}

	return "", false
}
