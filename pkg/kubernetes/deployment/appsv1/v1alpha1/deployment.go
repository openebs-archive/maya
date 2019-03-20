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

	api_apps_v1 "k8s.io/api/apps/v1"
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

// rolloutStatus  is a typed function that abstracts status message formation logic
type rolloutStatus func(*deploy) string

// deploy is a wrapper over aapi_apps_v1.Deployment
type deploy struct {
	object *api_apps_v1.Deployment // kubernetes deployment instance
	checks []predicate             // predicate list for deploy
}

// predicateKey is wrapper over string. It is used to get predicate function and status msg.
type predicateKey string

// rolloutStatuses contains a group of status message for each predicate checks.
// It useses predicateKey as key.
var rolloutStatuses = map[predicateKey]rolloutStatus{
	// ProgressDeadlineExceededPK refer to rolloutStatus for predicate IsProgressDeadlineExceeded.
	ProgressDeadlineExceededPK: func(d *deploy) string {
		return "Deployment exceeded its progress deadline"
	},
	// OlderReplicaActivePK refer to rolloutStatus for predicate IsOlderReplicaActive.
	OlderReplicaActivePK: func(d *deploy) string {
		if d.object.Spec.Replicas != nil {
			return fmt.Sprintf("Waiting for deployment rollout to finish: %d out of %d new replicas have been updated",
				d.object.Status.UpdatedReplicas, *d.object.Spec.Replicas)
		}
		return "Waiting for deployment rollout to finish: some older replicas have been updated"
	},
	// TerminationInProgressPK refer rolloutStatus for predicate IsTerminationInProgress.
	TerminationInProgressPK: func(d *deploy) string {
		return fmt.Sprintf("Waiting for deployment rollout to finish: %d old replicas are pending termination",
			d.object.Status.Replicas-d.object.Status.UpdatedReplicas)
	},
	// UpdationInProgressPK refer to rolloutStatus for predicate IsUpdationInProgress.
	UpdationInProgressPK: func(d *deploy) string {
		return fmt.Sprintf("Waiting for deployment rollout to finish: %d of %d updated replicas are available",
			d.object.Status.AvailableReplicas, d.object.Status.UpdatedReplicas)
	},
	// NotSyncSpecPK refer to status rolloutStatus for predicate IsNotSyncSpec.
	NotSyncSpecPK: func(d *deploy) string {
		return "Waiting for deployment spec update to be observed"
	},
}

// rolloutChecks contains a group of predicate it useses predicateKey as key.
var rolloutChecks = map[predicateKey]predicate{
	// ProgressDeadlineExceededPK refer to predicate IsProgressDeadlineExceeded.
	ProgressDeadlineExceededPK: IsProgressDeadlineExceeded(),
	// OlderReplicaActivePK refer to predicate IsOlderReplicaActive.
	OlderReplicaActivePK: IsOlderReplicaActive(),
	// TerminationInProgressPK refer to predicate IsTerminationInProgress.
	TerminationInProgressPK: IsTerminationInProgress(),
	// UpdationInProgressPK refer to predicate IsUpdationInProgress.
	UpdationInProgressPK: IsUpdationInProgress(),
	// NotSyncSpecPK refer to predicate IsSyncSpec.
	NotSyncSpecPK: IsNotSyncSpec(),
}

const (
	// ProgressDeadlineExceededPK refer to predicate IsProgressDeadlineExceeded.
	ProgressDeadlineExceededPK predicateKey = "ProgressDeadlineExceeded"
	// NotSyncSpecPK refer to predicate IsSyncSpec.
	NotSyncSpecPK predicateKey = "NotSyncSpec"
	// OlderReplicaActivePK refer to predicate IsOlderReplicaActive.
	OlderReplicaActivePK predicateKey = "OlderReplicaActive"
	// TerminationInProgressPK refer to predicate IsTerminationInProgress.
	TerminationInProgressPK predicateKey = "TerminationInProgress"
	// UpdationInProgressPK refer to predicate IsUpdationInProgress.
	UpdationInProgressPK predicateKey = "UpdationInProgress"
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
func WithAPIObject(deployment *api_apps_v1.Deployment) deployBuildOption {
	return func(d *deploy) {
		d.object = deployment
	}
}

// isRollout range over rolloutChecks map and check status of each predicate
// also it generates status message from rolloutStatuses using predicate key
func (d *deploy) isRollout() (msg string, ok bool) {
	for pk, p := range rolloutChecks {
		if ok = p(d); ok {
			msg = rolloutStatuses[pk](d)
			return msg, !ok
		}
	}
	return "", !ok
}

// RolloutStatus runs checks against deployment instance
// and generates rollout status as rolloutOutput
func (d *deploy) RolloutStatus() (op rolloutOutput, err error) {
	msg, ok := d.isRollout()
	op.IsRolledout = ok
	if !ok {
		op.Message = msg
		return
	}
	op.Message = "Deployment successfully rolled out"
	return
}

// RolloutStatus converts rolloutOutput to byte
func (d *deploy) RolloutStatusf() (op []byte, err error) {
	res, err := d.RolloutStatus()
	if err != nil {
		return
	}
	return json.Marshal(res)
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

// IsProgressDeadlineExceeded is used to check updation is timed out or not. If
// `Progressing` condition's reason is `ProgressDeadlineExceeded` then it is not rolled out.
func IsProgressDeadlineExceeded() predicate {
	return func(d *deploy) bool {
		return d.IsProgressDeadlineExceeded()
	}
}

// IsProgressDeadlineExceeded is used to check updation is timed out or not. If
// `Progressing` condition's reason is `ProgressDeadlineExceeded` then it is not rolled out.
func (d *deploy) IsProgressDeadlineExceeded() bool {
	for _, cond := range d.object.Status.Conditions {
		if cond.Type == api_apps_v1.DeploymentProgressing &&
			cond.Reason == "ProgressDeadlineExceeded" {
			return true
		}
	}
	return false
}

// IsOlderReplicaActive check if older replica's are stil active or not if Status.UpdatedReplicas
// < *Spec.Replicas then some of the replicas are updated and some of them are not.
func IsOlderReplicaActive() predicate {
	return func(d *deploy) bool {
		return d.IsOlderReplicaActive()
	}
}

// IsOlderReplicaActive check if older replica's are stil active or not if Status.UpdatedReplicas
// < *Spec.Replicas then some of the replicas are updated and some of them are not.
func (d *deploy) IsOlderReplicaActive() bool {
	return d.object.Spec.Replicas != nil && d.object.Status.UpdatedReplicas < *d.object.Spec.Replicas
}

// IsTerminationInProgress checks for older replicas are waiting to terminate or not.
// if Status.Replicas > Status.UpdatedReplicas then some of the older replicas are in running state because newer
// replicas are not in running state. It waits for newer replica to come into reunning state then terminate.
func IsTerminationInProgress() predicate {
	return func(d *deploy) bool {
		return d.IsTerminationInProgress()
	}
}

// IsTerminationInProgress checks for older replicas are waiting to terminate or not.
// if Status.Replicas > Status.UpdatedReplicas then some of the older replicas are in running state because newer
// replicas are not in running state. It waits for newer replica to come into reunning state then terminate.
func (d *deploy) IsTerminationInProgress() bool {
	return d.object.Status.Replicas > d.object.Status.UpdatedReplicas
}

// IsUpdationInProgress Checks if all the replicas are updated or not. If Status.AvailableReplicas < Status.UpdatedReplicas
// then all the older replicas are not there but there are less number of availableReplicas
func IsUpdationInProgress() predicate {
	return func(d *deploy) bool {
		return d.IsUpdationInProgress()
	}
}

// IsUpdationInProgress Checks if all the replicas are updated or not. If Status.AvailableReplicas < Status.UpdatedReplicas
// then all the older replicas are not there but there are less number of availableReplicas
func (d *deploy) IsUpdationInProgress() bool {
	return d.object.Status.AvailableReplicas < d.object.Status.UpdatedReplicas
}

// IsNotSyncSpec compare generation in status and spec and check if deployment spec is synced or not.
// If Generation <= Status.ObservedGeneration then deployment spec is not updated yet.
func IsNotSyncSpec() predicate {
	return func(d *deploy) bool {
		return d.IsNotSyncSpec()
	}
}

// IsNotSyncSpec compare generation in status and spec and check if deployment spec is synced or not.
// If Generation <= Status.ObservedGeneration then deployment spec is not updated yet.
func (d *deploy) IsNotSyncSpec() bool {
	return d.object.Generation > d.object.Status.ObservedGeneration
}
