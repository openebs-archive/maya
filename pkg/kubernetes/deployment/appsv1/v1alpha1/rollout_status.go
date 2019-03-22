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
)

// rolloutStatus  is a typed function that
// abstracts status message formation logic
type rolloutStatus func(*deploy) string

// rolloutStatuses contains a group of status message for
// each predicate checks. It uses predicateName as key.
var rolloutStatuses = map[predicateName]rolloutStatus{
	// PredicateProgressDeadlineExceeded refer to rolloutStatus
	// for predicate IsProgressDeadlineExceeded.
	PredicateProgressDeadlineExceeded: func(d *deploy) string {
		return "Deployment exceeded its progress deadline"
	},
	// PredicateOlderReplicaActive refer to rolloutStatus for
	// predicate IsOlderReplicaActive.
	PredicateOlderReplicaActive: func(d *deploy) string {
		if d.object.Spec.Replicas == nil {
			return "Replica update in progress : some older replicas have been updated"
		}
		return fmt.Sprintf(
			"Replica update in progress : %d out of %d new replicas have been updated",
			d.object.Status.UpdatedReplicas, *d.object.Spec.Replicas)
	},
	// PredicateTerminationInProgress refer rolloutStatus
	// for predicate IsTerminationInProgress.
	PredicateTerminationInProgress: func(d *deploy) string {
		return fmt.Sprintf(
			"Replica termination in progress : %d old replicas are pending termination",
			d.object.Status.Replicas-d.object.Status.UpdatedReplicas)
	},
	// PredicateUpdateInProgress refer to rolloutStatus for predicate IsUpdateInProgress.
	PredicateUpdateInProgress: func(d *deploy) string {
		return fmt.Sprintf(
			"Replica update in progress : %d of %d updated replicas are available",
			d.object.Status.AvailableReplicas, d.object.Status.UpdatedReplicas)
	},
	// PredicateNotSpecSynced refer to status rolloutStatus for predicate IsNotSyncSpec.
	PredicateNotSpecSynced: func(d *deploy) string {
		return "Deployment rollout in-progress : deployment rollout in-progress : waiting for deployment spec update to be observed"
	},
}

// rolloutChecks contains a group of predicate it uses predicateName as key.
var rolloutChecks = map[predicateName]predicate{
	// PredicateProgressDeadlineExceeded refer to predicate IsProgressDeadlineExceeded.
	PredicateProgressDeadlineExceeded: IsProgressDeadlineExceeded(),
	// PredicateOlderReplicaActive refer to predicate IsOlderReplicaActive.
	PredicateOlderReplicaActive: IsOlderReplicaActive(),
	// PredicateTerminationInProgress refer to predicate IsTerminationInProgress.
	PredicateTerminationInProgress: IsTerminationInProgress(),
	// PredicateUpdateInProgress refer to predicate IsUpdateInProgress.
	PredicateUpdateInProgress: IsUpdateInProgress(),
	// PredicateNotSpecSynced refer to predicate IsSyncSpec.
	PredicateNotSpecSynced: IsNotSyncSpec(),
}

// rolloutOutput struct contains message and boolean value to show rolloutstatus
type rolloutOutput struct {
	IsRolledout bool   `json:"isRolledout"`
	Message     string `json:"message"`
}

// rawFn is a typed function that abstracts
// conversion of rolloutOutput struct to raw byte
type rawFn func(r *rolloutOutput) ([]byte, error)

// rawFn is a typed function that abstracts
// conversion of rolloutOutput struct
type asRolloutOutputFn func(r *rolloutOutput) (*rolloutOutput, error)

// rollout enables getting various output format of rolloutOutput
type rollout struct {
	output          *rolloutOutput
	raw             rawFn
	asRolloutOutput asRolloutOutputFn
}

// rolloutBuildOption defines the
// abstraction to build a rollout instance
type rolloutBuildOption func(*rollout)

// rolloutStatusf returns new instance of rollout
// meant for rolloutOutput. caller can configure it with different
// rolloutOutputBuildOption
func rolloutStatusf(opts ...rolloutBuildOption) *rollout {
	r := &rollout{}
	for _, o := range opts {
		o(r)
	}
	r.withDefaults()
	return r
}

// withOutputObject sets rolloutOutput in rollout instance
func withOutputObject(o *rolloutOutput) rolloutBuildOption {
	return func(r *rollout) {
		r.output = o
	}
}

// withDefaults sets the default options of rolloutBuilder instance
func (r *rollout) withDefaults() {
	if r.raw == nil {
		r.raw = func(o *rolloutOutput) ([]byte, error) {
			return json.Marshal(o)
		}
	}

	if r.asRolloutOutput == nil {
		r.asRolloutOutput = func(o *rolloutOutput) (*rolloutOutput, error) {
			return o, nil
		}
	}
}

// Raw returns raw bytes outpot of rollout
func (r *rollout) Raw() ([]byte, error) {
	if r.output == nil {
		return nil, fmt.Errorf("Unable to get rollout status output")
	}
	return r.raw(r.output)
}

// AsRolloutOutput returns rolloutOutput struct as output
func (r *rollout) AsRolloutOutput() (*rolloutOutput, error) {
	if r.output == nil {
		return nil, fmt.Errorf("Unable to get rollout status output")
	}
	return r.asRolloutOutput(r.output)
}
