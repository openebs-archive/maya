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

// rolloutStatus  is a typed function that abstracts status message formation logic
type rolloutStatus func(*deploy) string

// rolloutStatuses contains a group of status message for each predicate checks.
// It useses predicateName as key.
var rolloutStatuses = map[predicateName]rolloutStatus{
	// PredicateProgressDeadlineExceeded refer to rolloutStatus for predicate IsProgressDeadlineExceeded.
	PredicateProgressDeadlineExceeded: func(d *deploy) string {
		return "Deployment exceeded its progress deadline"
	},
	// PredicateOlderReplicaActive refer to rolloutStatus for predicate IsOlderReplicaActive.
	PredicateOlderReplicaActive: func(d *deploy) string {
		if d.object.Spec.Replicas == nil {
			return "Replica update in progress : some older replicas have been updated"
		}
		return fmt.Sprintf("Replica update in progress : %d out of %d new replicas have been updated",
			d.object.Status.UpdatedReplicas, *d.object.Spec.Replicas)
	},
	// PredicateTerminationInProgress refer rolloutStatus for predicate IsTerminationInProgress.
	PredicateTerminationInProgress: func(d *deploy) string {
		return fmt.Sprintf("Replica termination in progress : %d old replicas are pending termination",
			d.object.Status.Replicas-d.object.Status.UpdatedReplicas)
	},
	// PredicateUpdateInProgress refer to rolloutStatus for predicate IsUpdateInProgress.
	PredicateUpdateInProgress: func(d *deploy) string {
		return fmt.Sprintf("Replica update in progress : %d of %d updated replicas are available",
			d.object.Status.AvailableReplicas, d.object.Status.UpdatedReplicas)
	},
	// PredicateNotSpecSynced refer to status rolloutStatus for predicate IsNotSyncSpec.
	PredicateNotSpecSynced: func(d *deploy) string {
		return "Deployment rollout in-progress : deployment rollout in-progress : waiting for deployment spec update to be observed"
	},
}

// rolloutChecks contains a group of predicate it useses predicateName as key.
var rolloutChecks = map[predicateName]predicate{
	// PredicateProgressDeadlineExceeded refer to predicate IsProgressDeadlineExceeded.
	PredicateProgressDeadlineExceeded: IsProgressDeadlineExceeded(),
	// PredicateOlderReplicaActive refer to predicate IsOlderReplicaActive.
	PredicateOlderReplicaActive: IsOlderReplicaActive(),
	// PredicateTerminationInProgress refer to predicate IsTerminationInProgress.
	PredicateTerminationInProgress: IsTerminationInProgress(),
	// PredicateUpdateInProgress refer to predicate IsUpdationInProgress.
	PredicateUpdateInProgress: IsUpdateInProgress(),
	// PredicateNotSpecSynced refer to predicate IsSyncSpec.
	PredicateNotSpecSynced: IsNotSyncSpec(),
}

// rolloutOutput struct contaons message and boolean value to show rolloutstatus
type rolloutOutput struct {
	IsRolledout bool   `json:"isRolledout"`
	Message     string `json:"message"`
}

// rawFn is a typed function that abstracts conversion of rolloutOutput struct to raw byte
type rawFn func(r *rolloutOutput) ([]byte, error)

// rolloutOutputBuilder enables getting various output format of rolloutOutput
type rolloutOutputBuilder struct {
	output *rolloutOutput
	raw    rawFn
}

// rolloutOutputBuildOption defines the abstraction to build a rolloutOutputBuilder instance
type rolloutOutputBuildOption func(*rolloutOutputBuilder)

// isRollout range over rolloutChecks map and check status of each predicate
// also it generates status message from rolloutStatuses using predicate key
func (d *deploy) isRollout() (string, bool) {
	msg := ""
	ok := false
	for pk, p := range rolloutChecks {
		if ok = p(d); ok {
			msg = rolloutStatuses[pk](d)
			return msg, !ok
		}
	}
	return msg, !ok
}

// RolloutStatus runs checks against deployment instance
// and generates rollout status as rolloutOutput
func (d *deploy) RolloutStatus() (op *rolloutOutput, err error) {
	op = &rolloutOutput{}
	msg, ok := d.isRollout()
	op.IsRolledout = ok
	if !ok {
		op.Message = msg
		return
	}
	op.Message = "Deployment successfully rolled out"
	return
}

// rolloutStatusf returns a new instance of rolloutOutputBuilder meant for rolloutOutput.
// caller can configure it with different rolloutOutputBuildOption
func rolloutStatusf(opts ...rolloutOutputBuildOption) *rolloutOutputBuilder {
	r := &rolloutOutputBuilder{}
	for _, o := range opts {
		o(r)
	}
	r.withDefaults()
	return r
}

// withOutputObject sets rolloutOutput in rolloutOutputBuilder instance
func withOutputObject(o *rolloutOutput) rolloutOutputBuildOption {
	return func(r *rolloutOutputBuilder) {
		r.output = o
	}
}

// withDefaults sets the default options of rolloutOutputBuilder instance
func (r *rolloutOutputBuilder) withDefaults() {
	if r.raw == nil {
		r.raw = func(o *rolloutOutput) ([]byte, error) {
			return json.Marshal(o)
		}
	}
}

// Raw returns raw bytes outpot of rolloutOutput
func (r *rolloutOutputBuilder) Raw() ([]byte, error) {
	if r.output == nil {
		return nil, fmt.Errorf("Unable to get rollout status output")
	}
	return r.raw(r.output)
}
