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

// predicateAppsv1 abstracts conditional logic w.r.t the deployment instance
//
// NOTE:
// predicateAppsv1 is a functional approach versus traditional approach to mix
// conditions such as *if-else* within blocks of business logic
//
// NOTE:
// predicateAppsv1 approach enables clear separation of conditionals from
// imperatives i.e. actions that form the business logic
type predicateAppsv1 func(*deployAppsv1) (nameOrMsg string, ok bool)

// deployAppsv1BuildOption is a typed function that abstracts anykind of operation
// against the provided deployment instance
//
// This is the basic building block to create functional operations
// against the deployment instance
type deployAppsv1BuildOption func(*deployAppsv1)

// deployAppsv1 is a wrapper over api_extn_v1beta1.Deployment
type deployAppsv1 struct {
	object *api_apps_v1.Deployment // kubernetes deployment instance
	checks []predicateAppsv1       // predicate list for deployAppsv1
}

// KubeClient returns a new instance of kubeclient meant for deployment
func DeployAppsv1(opts ...deployAppsv1BuildOption) *deployAppsv1 {
	k := &deployAppsv1{}
	for _, o := range opts {
		o(k)
	}
	return k
}

// WithAppsv1Deployment is a deployAppsv1BuildOption caller can pass deployment schema
// with this function to create deployAppsv1 object
func WithAppsv1Deployment(deploy *api_apps_v1.Deployment) deployAppsv1BuildOption {
	return func(d *deployAppsv1) {
		d.object = deploy
	}
}

// RollOutStatus runs checks against deployment instance and generates rollout status with msg
func (b *deployAppsv1) RollOutStatus() (op []byte, err error) {
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
func (d *deployAppsv1) AddCheck(p predicateAppsv1) *deployAppsv1 {
	d.checks = append(d.checks, p)
	return d
}

// AddChecks adds the provided predicates as conditions to be validated against
// the daemonset instance
func (b *deployAppsv1) AddChecks(p []predicateAppsv1) *deployAppsv1 {
	for _, check := range p {
		b.AddCheck(check)
	}
	return b
}

// IsProgressDeadlineExceededV1 is used to check updation is timed out or not. If
// `Progressing` condition's reason is `ProgressDeadlineExceeded` then it is not rolled out.
func IsProgressDeadlineExceededV1() predicateAppsv1 {
	return func(d *deployAppsv1) (string, bool) {
		for _, cond := range d.object.Status.Conditions {
			if cond.Type == api_apps_v1.DeploymentProgressing &&
				cond.Reason == "ProgressDeadlineExceeded" {
				return rolloutMessage(ProgressDeadlineExceededPK), true
			}
		}
		return "", false
	}
}

// IsProgressDeadlineExceededV1 is used to check updation is timed out or not. If
// `Progressing` condition's reason is `ProgressDeadlineExceeded` then it is not rolled out.
func (d *deployAppsv1) IsProgressDeadlineExceededExceededV1() (string, bool) {
	return IsProgressDeadlineExceededV1()(d)
}

// IsOlderReplicaActiveV1 check if older replica's are stil active or not if Status.UpdatedReplicas
// < *Spec.Replicas then some of the replicas are updated and some of them are not.
func IsOlderReplicaActiveV1() predicateAppsv1 {
	return func(d *deployAppsv1) (string, bool) {
		return fmt.Sprintf(rolloutMessage(OlderReplicaActivePK), d.object.Status.UpdatedReplicas, *d.object.Spec.Replicas),
			d.object.Spec.Replicas != nil && d.object.Status.UpdatedReplicas < *d.object.Spec.Replicas
	}
}

// IsOlderReplicaActiveV1B1 check if older replica's are stil active or not if Status.UpdatedReplicas
// < *Spec.Replicas then some of the replicas are updated and some of them are not.
func (d *deployAppsv1) IsOlderReplicaActiveV1() (string, bool) {
	return IsOlderReplicaActiveV1()(d)
}

// IsTerminationInProgressV1 checks for older replicas are waiting to terminate or not.
// if Status.Replicas > Status.UpdatedReplicas then some of the older replicas are in running state because newer
// replicas are not in running state. It waits for newer replica to come into reunning state then terminate.
func IsTerminationInProgressV1() predicateAppsv1 {
	return func(d *deployAppsv1) (string, bool) {
		return fmt.Sprintf(rolloutMessage(TerminationInProgressPK), d.object.Status.Replicas-
			d.object.Status.UpdatedReplicas), d.object.Status.Replicas > d.object.Status.UpdatedReplicas
	}
}

// IsTerminationInProgressV1 checks for older replicas are waiting to terminate or not.
// if Status.Replicas > Status.UpdatedReplicas then some of the older replicas are in running state because newer
// replicas are not in running state. It waits for newer replica to come into reunning state then terminate.
func (d *deployAppsv1) IsTerminationInProgressV1() (string, bool) {
	return IsTerminationInProgressV1()(d)
}

// IsUpdationInProgressV1 Checks if all the replicas are updated or not. If Status.AvailableReplicas < Status.UpdatedReplicas
// then all the older replicas are not there but there are less number of availableReplicas
func IsUpdationInProgressV1() predicateAppsv1 {
	return func(d *deployAppsv1) (string, bool) {
		return fmt.Sprintf(rolloutMessage(UpdationInProgressPK), d.object.Status.AvailableReplicas,
			d.object.Status.UpdatedReplicas), d.object.Status.AvailableReplicas < d.object.Status.UpdatedReplicas
	}
}

// IsUpdationInProgressV1 Checks if all the replicas are updated or not. If Status.AvailableReplicas < Status.UpdatedReplicas
// then all the older replicas are not there but there are less number of availableReplicas
func (d *deployAppsv1) IsUpdationInProgressV1() (string, bool) {
	return IsUpdationInProgressV1()(d)
}

// IsSyncSpecV1 compare generation in status and spec and check if deployment spec is synced or not.
// If Generation <= Status.ObservedGeneration then deployment spec is not updated yet.
func IsSyncSpecV1() predicateAppsv1 {
	return func(d *deployAppsv1) (string, bool) {
		return rolloutMessage(SyncExceededPK),
			d.object.Generation > d.object.Status.ObservedGeneration
	}
}

// IsSyncSpecV1 compare generation in status and spec and check if deployment spec is synced or not.
// If Generation <= Status.ObservedGeneration then deployment spec is not updated yet.
func (d *deployAppsv1) IsSyncSpecV1() (string, bool) {
	return IsSyncSpecV1()(d)
}
