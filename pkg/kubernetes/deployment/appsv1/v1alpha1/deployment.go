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
	appsv1 "k8s.io/api/apps/v1"
)

// predicate abstracts conditional logic w.r.t the deployment instance
//
// NOTE:
// predicate is a functional approach versus traditional approach to mix
// conditions such as *if-else* within blocks of business logic
//
// NOTE:
// predicate approach enables clear separation of conditionals from
// imperatives i.e. actions that form the business logic
type predicate func(*deploy) bool

// deployBuildOption is a typed function that abstracts anykind of operation
// against the provided deployment instance
//
// This is the basic building block to create functional operations
// against the deployment instance
type deployBuildOption func(*deploy)

// deploy is a wrapper over appsv1.Deployment
type deploy struct {
	object *appsv1.Deployment // kubernetes deployment instance
	checks []predicate        // predicate list for deploy
}

// predicateName type is wrapper over string.
// It is used to refer predicate and status msg.
type predicateName string

const (
	// PredicateProgressDeadlineExceeded refer to predicate IsProgressDeadlineExceeded.
	PredicateProgressDeadlineExceeded predicateName = "ProgressDeadlineExceeded"
	// PredicateNotSpecSynced refer to predicate IsNotSpecSynced
	PredicateNotSpecSynced predicateName = "NotSpecSynced"
	// PredicateOlderReplicaActive refer to predicate IsOlderReplicaActive
	PredicateOlderReplicaActive predicateName = "OlderReplicaActive"
	// PredicateTerminationInProgress refer to predicate IsTerminationInProgress
	PredicateTerminationInProgress predicateName = "TerminationInProgress"
	// PredicateUpdateInProgress refer to predicate IsUpdateInProgress.
	PredicateUpdateInProgress predicateName = "UpdateInProgress"
)

// New returns a new instance of deploy meant for deployment
func New(opts ...deployBuildOption) *deploy {
	k := &deploy{}
	for _, o := range opts {
		o(k)
	}
	return k
}

// WithAPIObject is a deployBuildOption caller can pass deployment schema
// with this function to create deploy object
func WithAPIObject(deployment *appsv1.Deployment) deployBuildOption {
	return func(d *deploy) {
		d.object = deployment
	}
}

// isRollout range over rolloutChecks map and check status of each predicate
// also it generates status message from rolloutStatuses using predicate key
func (d *deploy) isRollout() (predicateName, bool) {
	for pk, p := range rolloutChecks {
		if p(d) {
			return pk, false
		}
	}
	return "", true
}

// failedRollout returns rollout status message for fail condition
func (d *deploy) failedRollout(name predicateName) *rolloutOutput {
	return &rolloutOutput{
		Message:     rolloutStatuses[name](d),
		IsRolledout: false,
	}
}

// failedRollout returns rollout status message for success condition
func (d *deploy) successRollout(name predicateName) *rolloutOutput {
	return &rolloutOutput{
		Message:     "Deployment successfully rolled out",
		IsRolledout: false,
	}
}

// RolloutStatus returns rollout message of deployment instance
func (d *deploy) RolloutStatus() (op *rolloutOutput, err error) {
	pk, ok := d.isRollout()
	if ok {
		return d.successRollout(pk), nil
	}
	return d.failedRollout(pk), nil
}

// RolloutStatus returns rollout message of deployment instance
// in byte format
func (d *deploy) RolloutStatusRaw() (op []byte, err error) {
	message, err := d.RolloutStatus()
	if err != nil {
		return nil, err
	}
	return NewRollout(
		withOutputObject(message)).
		Raw()
}

// AddCheck adds the predicate as a condition to be validated
// against the deployment instance
func (d *deploy) AddCheck(p predicate) *deploy {
	d.checks = append(d.checks, p)
	return d
}

// AddChecks adds the provided predicates as conditions to be
// validated against the deployment instance
func (d *deploy) AddChecks(p []predicate) *deploy {
	for _, check := range p {
		d.AddCheck(check)
	}
	return d
}

// IsProgressDeadlineExceeded is used to check update is timed out or not.
// If `Progressing` condition's reason is `ProgressDeadlineExceeded` then
// it is not rolled out.
func IsProgressDeadlineExceeded() predicate {
	return func(d *deploy) bool {
		return d.IsProgressDeadlineExceeded()
	}
}

// IsProgressDeadlineExceeded is used to check update is timed out or not.
// If `Progressing` condition's reason is `ProgressDeadlineExceeded` then
// it is not rolled out.
func (d *deploy) IsProgressDeadlineExceeded() bool {
	for _, cond := range d.object.Status.Conditions {
		if cond.Type == appsv1.DeploymentProgressing &&
			cond.Reason == "ProgressDeadlineExceeded" {
			return true
		}
	}
	return false
}

// IsOlderReplicaActive check if older replica's are still active or not if
// Status.UpdatedReplicas < *Spec.Replicas then some of the replicas are
// updated and some of them are not.
func IsOlderReplicaActive() predicate {
	return func(d *deploy) bool {
		return d.IsOlderReplicaActive()
	}
}

// IsOlderReplicaActive check if older replica's are still active or not if
// Status.UpdatedReplicas < *Spec.Replicas then some of the replicas are
// updated and some of them are not.
func (d *deploy) IsOlderReplicaActive() bool {
	return d.object.Spec.Replicas != nil &&
		d.object.Status.UpdatedReplicas < *d.object.Spec.Replicas
}

// IsTerminationInProgress checks for older replicas are waiting to
// terminate or not. If Status.Replicas > Status.UpdatedReplicas then
// some of the older replicas are in running state because newer
// replicas are not in running state. It waits for newer replica to
// come into running state then terminate.
func IsTerminationInProgress() predicate {
	return func(d *deploy) bool {
		return d.IsTerminationInProgress()
	}
}

// IsTerminationInProgress checks for older replicas are waiting to
// terminate or not. If Status.Replicas > Status.UpdatedReplicas then
// some of the older replicas are in running state because newer
// replicas are not in running state. It waits for newer replica to
// come into running state then terminate.
func (d *deploy) IsTerminationInProgress() bool {
	return d.object.Status.Replicas > d.object.Status.UpdatedReplicas
}

// IsUpdateInProgress Checks if all the replicas are updated or not.
// If Status.AvailableReplicas < Status.UpdatedReplicas then all the
//older replicas are not there but there are less number of availableReplicas
func IsUpdateInProgress() predicate {
	return func(d *deploy) bool {
		return d.IsUpdateInProgress()
	}
}

// IsUpdateInProgress Checks if all the replicas are updated or not.
// If Status.AvailableReplicas < Status.UpdatedReplicas then all the
// older replicas are not there but there are less number of availableReplicas
func (d *deploy) IsUpdateInProgress() bool {
	return d.object.Status.AvailableReplicas < d.object.Status.UpdatedReplicas
}

// IsNotSyncSpec compare generation in status and spec and check if
// deployment spec is synced or not. If Generation <= Status.ObservedGeneration
// then deployment spec is not updated yet.
func IsNotSyncSpec() predicate {
	return func(d *deploy) bool {
		return d.IsNotSyncSpec()
	}
}

// IsNotSyncSpec compare generation in status and spec and check if
// deployment spec is synced or not. If Generation <= Status.ObservedGeneration
// then deployment spec is not updated yet.
func (d *deploy) IsNotSyncSpec() bool {
	return d.object.Generation > d.object.Status.ObservedGeneration
}
