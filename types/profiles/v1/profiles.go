/*
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
// functions related to the profiles
package v1

import (
	"fmt"
	"strings"
)

// ProvisionPolicyKey type deals with the keys associated with
// volume provisioning. It is expected that these keys will have
// corresponding values.
//
// USAGE:
//  When OpenEBS storage is managed with Kubernetes (K8s) as the underlying
// container orchestrator, then provision policy key can be used as following:
//
// apiVersion: storage.k8s.io/v1
// kind: StorageClass
// metadata:
//    name: openebs-jupyter
//    labels:
//     provision/replica-count: 2
//     provision/controller-count: 1
// provisioner: openebs.io/provisioner-iscsi
// parameters:
//   pool: hostdir-var
//   replica: "2"
//   size: 5G
type ProvisionPolicyKey string

const (
	ReplicaCountPPK         ProvisionPolicyKey = "provision/replica-count"
	ControllerCountPPK      ProvisionPolicyKey = "provision/controller-count"
	ReplicaTolerationPPK    ProvisionPolicyKey = "provision/replica-toleration"
	ControllerTolerationPPK ProvisionPolicyKey = "provision/controller-toleration"
)

// PlacementPolicyKey is a constant which will be used as a key. This key can be
// associated with one or more placement policies as value(s).
//
// USAGE:
//  When OpenEBS storage is managed with Kubernetes (K8s) as the underlying
// container orchestrator; then placement policy key can be used as following:
//
// apiVersion: storage.k8s.io/v1
// kind: StorageClass
// metadata:
//    name: openebs-jupyter
//    labels:
//     placement: "spread-replica, sticky-replica"
//     provision/replica-count: 2
// provisioner: openebs.io/provisioner-iscsi
// parameters:
//   pool: hostdir-var
//   replica: "2"
//   size: 5G
const (
	PlacementPolicyKey string = "placement"
)

// PlacementPolicyValue type deals with various volume placement related
// policy value(s). These can be thought of as flags. Presence of these policy
// flags indicate enabling of corresponding placement.
//
// NOTE to developer:
//  Care should be exercised to filter/prioritize policies if some of them
// are conflicting. Particular examples will be mentioned when the developer
// have them.
//
// USAGE:
//  Already mentioned in PlacementPolicyKey
type PlacementPolicyValue string

const (
	SpreadReplicaPPV PlacementPolicyValue = "spread-replica"
	StickyReplicaPPV PlacementPolicyValue = "sticky-replica"
)

// IsStickyReplica indicates if sticking replicas to their placement
// Nodes is mentioned in provided policies.
func IsStickyReplica(policies map[string]string) (bool, error) {
	// TODO logic
	return false, fmt.Errorf("Not implemented")
}

// IsSpreadReplica indicates if spreading the replicas across Nodes
// is mentioned in provided policies.
func IsSpreadReplica(policies map[string]string) (bool, error) {
	// TODO logic
	return false, fmt.Errorf("Not implemented")
}

// GetReplicaCount gets the value of replica count if available
func GetReplicaCount(policies map[string]string) (string, bool) {
	for k, v := range policies {
		if strings.TrimSpace(k) == string(ReplicaCountPPK) {
			return strings.TrimSpace(v), true
		}
	}

	return "", false
}
