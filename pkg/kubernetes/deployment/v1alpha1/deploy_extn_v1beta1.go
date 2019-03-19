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
	"encoding/json"
	"fmt"

	api_extn_v1beta1 "k8s.io/api/extensions/v1beta1"
)

// predicateExtnV1Beta1 abstracts conditional logic w.r.t the deployment instance
//
// NOTE:
// predicateExtnV1Beta1 is a functional approach versus traditional approach to mix
// conditions such as *if-else* within blocks of business logic
//
// NOTE:
// predicateExtnV1Beta1 approach enables clear separation of conditionals from
// imperatives i.e. actions that form the business logic
type predicateExtnV1Beta1 func(*deployExtnV1Beta1) (nameOrMsg string, ok bool)

// deployExtnV1Beta1BuildOption is a typed function that abstracts anykind of operation
// against the provided deployment instance
//
// This is the basic building block to create functional operations
// against the deployment instance
type deployExtnV1Beta1BuildOption func(*deployExtnV1Beta1)

// deployExtnV1Beta1 is a wrapper over api_extn_v1beta1.Deployment
type deployExtnV1Beta1 struct {
	object *api_extn_v1beta1.Deployment // kubernetes deployment instance
	checks []predicateExtnV1Beta1       // predicate list for deployExtnV1Beta1
}

// predicateKey is wrapper over string It represent status msg key type for predicate
type predicateKey string

// rolloutMessages contains a group of status message for each predicate checks.
// It useses predicateKey as key.
var rolloutMessages = map[predicateKey]string{
	// ProgressDeadlineExceededPK refer to status msg for ProgressDeadlineExceeded in
	// this case changes is not done successfully.
	ProgressDeadlineExceededPK: "Deployment exceeded its progress deadline",
	// OlderReplicaActivePK refer to status message for older replicas availble
	OlderReplicaActivePK: "Waiting for deployment rollout to finish: %d out of %d new replicas have been updated",
	// TerminationInProgressPK refer to status message for older replica's termination
	TerminationInProgressPK: "Waiting for deployment rollout to finish: %d old replicas are pending termination",
	// UpdationInProgressPK refer to status message for updated repicas status
	UpdationInProgressPK: "Waiting for deployment rollout to finish: %d of %d updated replicas are available",
	// SyncExceededPK refer to status message for deployment spec sync.
	SyncExceededPK: "Waiting for deployment spec update to be observed",
}

const (
	// ProgressDeadlineExceededPK refer to status msg for ProgressDeadlineExceeded in
	// this case changes is not done successfully.
	ProgressDeadlineExceededPK predicateKey = "ProgressDeadlineExceeded"
	// SyncExceededPK refer to status message for deployment spec sync.
	SyncExceededPK predicateKey = "SyncExceededPK"
	// OlderReplicaActivePK refer to status message for older replicas availble
	OlderReplicaActivePK predicateKey = "OlderReplicaActive"
	// TerminationInProgressPK refer to status message for older replica's termination
	TerminationInProgressPK predicateKey = "TerminationInProgress"
	// UpdationInProgressPK refer to status message for updated repicas status
	UpdationInProgressPK predicateKey = "UpdationInProgress"
)

// rolloutMessage takes predicateKey and get status message from rolloutMessages map
func rolloutMessage(p predicateKey) string {
	return rolloutMessages[p]
}

// KubeClient returns a new instance of kubeclient meant for deployment
func DeployExtnV1Beta1(opts ...deployExtnV1Beta1BuildOption) *deployExtnV1Beta1 {
	k := &deployExtnV1Beta1{}
	for _, o := range opts {
		o(k)
	}
	return k
}

// WithExtnV1Beta1Deployment is a deployExtnV1Beta1BuildOption caller can pass deployment schema
// with this function to create deployExtnV1Beta1 object
func WithExtnV1Beta1Deployment(deploy *api_extn_v1beta1.Deployment) deployExtnV1Beta1BuildOption {
	return func(d *deployExtnV1Beta1) {
		d.object = deploy
	}
}

// RollOutStatus runs checks against deployment instance and generates rollout status with msg
func (b *deployExtnV1Beta1) RollOutStatus() (op []byte, err error) {
	r := rolloutOutput{}
	r.IsRolledout = true

	for _, opt := range b.checks {
		msg, check := opt(b)
		if check {
			r.IsRolledout = false
			r.Message = msg
			return json.Marshal(r)
		}
	}

	// if no checks fails it is success
	r.Message = "Deployment successfully rolled out"
	return json.Marshal(r)
}

// AddCheck adds the predicate as a condition to be validated against the
// daemonset instance
func (d *deployExtnV1Beta1) AddCheck(p predicateExtnV1Beta1) *deployExtnV1Beta1 {
	d.checks = append(d.checks, p)
	return d
}

// AddChecks adds the provided predicates as conditions to be validated against
// the daemonset instance
func (b *deployExtnV1Beta1) AddChecks(p []predicateExtnV1Beta1) *deployExtnV1Beta1 {
	for _, check := range p {
		b.AddCheck(check)
	}
	return b
}

// IsProgressDeadlineExceededV1B1 is used to check updation is timed out or not. If
// `Progressing` condition's reason is `ProgressDeadlineExceeded` then it is not rolled out.
func IsProgressDeadlineExceededV1B1() predicateExtnV1Beta1 {
	return func(d *deployExtnV1Beta1) (string, bool) {
		for _, cond := range d.object.Status.Conditions {
			if cond.Type == api_extn_v1beta1.DeploymentProgressing &&
				cond.Reason == "ProgressDeadlineExceeded" {
				return rolloutMessage(ProgressDeadlineExceededPK), true
			}
		}
		return "", false
	}
}

// IsProgressDeadlineExceededV1B1 is used to check updation is timed out or not. If
// `Progressing` condition's reason is `ProgressDeadlineExceeded` then it is not rolled out.
func (d *deployExtnV1Beta1) IsProgressDeadlineExceededExceededV1B1() (string, bool) {
	return IsProgressDeadlineExceededV1B1()(d)
}

// IsOlderReplicaActiveV1B1 check if older replica's are stil active or not if Status.UpdatedReplicas
// < *Spec.Replicas then some of the replicas are updated and some of them are not.
func IsOlderReplicaActiveV1B1() predicateExtnV1Beta1 {
	return func(d *deployExtnV1Beta1) (string, bool) {
		return fmt.Sprintf(rolloutMessage(OlderReplicaActivePK), d.object.Status.UpdatedReplicas, *d.object.Spec.Replicas),
			d.object.Spec.Replicas != nil && d.object.Status.UpdatedReplicas < *d.object.Spec.Replicas
	}
}

// IsOlderReplicaActiveV1B1 check if older replica's are stil active or not if Status.UpdatedReplicas
// < *Spec.Replicas then some of the replicas are updated and some of them are not.
func (d *deployExtnV1Beta1) IsOlderReplicaActiveV1B1() (string, bool) {
	return IsOlderReplicaActiveV1B1()(d)
}

// IsTerminationInProgressV1B1 checks for older replicas are waiting to terminate or not.
// if Status.Replicas > Status.UpdatedReplicas then some of the older replicas are in running state because newer
// replicas are not in running state. It waits for newer replica to come into reunning state then terminate.
func IsTerminationInProgressV1B1() predicateExtnV1Beta1 {
	return func(d *deployExtnV1Beta1) (string, bool) {
		return fmt.Sprintf(rolloutMessage(TerminationInProgressPK), d.object.Status.Replicas-
			d.object.Status.UpdatedReplicas), d.object.Status.Replicas > d.object.Status.UpdatedReplicas
	}
}

// IsTerminationInProgressV1B1 checks for older replicas are waiting to terminate or not.
// if Status.Replicas > Status.UpdatedReplicas then some of the older replicas are in running state because newer
// replicas are not in running state. It waits for newer replica to come into reunning state then terminate.
func (d *deployExtnV1Beta1) IsTerminationInProgressV1B1() (string, bool) {
	return IsTerminationInProgressV1B1()(d)
}

// IsUpdationInProgressV1B1 Checks if all the replicas are updated or not. If Status.AvailableReplicas < Status.UpdatedReplicas
// then all the older replicas are not there but there are less number of availableReplicas
func IsUpdationInProgressV1B1() predicateExtnV1Beta1 {
	return func(d *deployExtnV1Beta1) (string, bool) {
		return fmt.Sprintf(rolloutMessage(UpdationInProgressPK), d.object.Status.AvailableReplicas,
			d.object.Status.UpdatedReplicas), d.object.Status.AvailableReplicas < d.object.Status.UpdatedReplicas
	}
}

// IsUpdationInProgressV1B1 Checks if all the replicas are updated or not. If Status.AvailableReplicas < Status.UpdatedReplicas
// then all the older replicas are not there but there are less number of availableReplicas
func (d *deployExtnV1Beta1) IsUpdationInProgressV1B1() (string, bool) {
	return IsUpdationInProgressV1B1()(d)
}

// IsSyncSpecV1B1 compare generation in status and spec and check if deployment spec is synced or not.
// If Generation <= Status.ObservedGeneration then deployment spec is not updated yet.
func IsSyncSpecV1B1() predicateExtnV1Beta1 {
	return func(d *deployExtnV1Beta1) (string, bool) {
		return rolloutMessage(SyncExceededPK),
			d.object.Generation > d.object.Status.ObservedGeneration
	}
}

// IsSyncSpecV1B1 compare generation in status and spec and check if deployment spec is synced or not.
// If Generation <= Status.ObservedGeneration then deployment spec is not updated yet.
func (d *deployExtnV1Beta1) IsSyncSpecV1B1() (string, bool) {
	return IsSyncSpecV1B1()(d)
}
