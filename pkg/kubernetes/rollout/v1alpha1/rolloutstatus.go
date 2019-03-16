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
	api_extn_v1beta1 "k8s.io/api/extensions/v1beta1"
)

// rolloutOutput struct contaons message and boolean value to show rolloutstatus
type rolloutOutput struct {
	IsRolledout bool   `json:"IsRolledout"`
	Message     string `json:"Message"`
}

// ExtnV1Beta1DeployRolloutStatusGenerator generates RolloutOutput from deploymnet object for that deployment
var ExtnV1Beta1DeployRolloutStatusGenerator = func(deploy api_extn_v1beta1.Deployment) (op []byte, err error) {
	r := rolloutOutput{}
	r.IsRolledout = false
	var cond *api_extn_v1beta1.DeploymentCondition
	// list all conditions and and select that condition which type is Progressing.
	for i := range deploy.Status.Conditions {
		c := deploy.Status.Conditions[i]
		if c.Type == api_extn_v1beta1.DeploymentProgressing {
			cond = &c
		}
	}
	// if deploy.Generation <= deploy.Status.ObservedGeneration then deployment spec is not updated yet.
	// it marked IsRolledout as false and update message accordingly
	if deploy.Generation <= deploy.Status.ObservedGeneration {
		// If Progressing condition's reason is ProgressDeadlineExceeded then it is not rolled out.
		if cond != nil && cond.Reason == "ProgressDeadlineExceeded" {
			r.Message = fmt.Sprintf("Deployment exceeded its progress deadline")
			return json.Marshal(r)
		}
		// if deploy.Status.UpdatedReplicas < *deploy.Spec.Replicas then some of the replicas are updated
		// and some of them are not. It marked IsRolledout as false and update message accordingly
		if deploy.Spec.Replicas != nil && deploy.Status.UpdatedReplicas < *deploy.Spec.Replicas {
			r.Message = fmt.Sprintf("Waiting for deployment rollout to finish: %d out of %d new replicas have been updated",
				deploy.Status.UpdatedReplicas, *deploy.Spec.Replicas)
			return json.Marshal(r)
		}
		// if deploy.Status.Replicas > deploy.Status.UpdatedReplicas then some of the older replicas are in running state
		// because newer replicas are not in running state. It waits for newer replica to come into reunning state then terminate.
		// It marked IsRolledout as false and update message accordingly
		if deploy.Status.Replicas > deploy.Status.UpdatedReplicas {
			r.Message = fmt.Sprintf("Waiting for deployment rollout to finish: %d old replicas are pending termination",
				deploy.Status.Replicas-deploy.Status.UpdatedReplicas)
			return json.Marshal(r)
		}
		// if deploy.Status.AvailableReplicas < deploy.Status.UpdatedReplicas then all the replicas are updated but they are
		// not in running state. It marked IsRolledout as false and update message accordingly.
		if deploy.Status.AvailableReplicas < deploy.Status.UpdatedReplicas {
			r.Message = fmt.Sprintf("Waiting for deployment rollout to finish: %d of %d updated replicas are available",
				deploy.Status.AvailableReplicas, deploy.Status.UpdatedReplicas)
		}
		r.IsRolledout = true
		r.Message = fmt.Sprintf("Deployment %q successfully rolled out", deploy.Name)
		return json.Marshal(r)
	}
	r.Message = fmt.Sprintf("Waiting for deployment spec update to be observed")
	return json.Marshal(r)
}

// AppsV1DeployRolloutStatusGenerator generates RolloutOutput from deploymnet object for that deployment
var AppsV1DeployRolloutStatusGenerator = func(deploy api_apps_v1.Deployment) (op []byte, err error) {
	r := rolloutOutput{}
	r.IsRolledout = false
	var cond *api_apps_v1.DeploymentCondition
	// list all conditions and and select that condition which type is Progressing.
	for i := range deploy.Status.Conditions {
		c := deploy.Status.Conditions[i]
		if c.Type == api_apps_v1.DeploymentProgressing {
			cond = &c
		}
	}
	// if deploy.Generation <= deploy.Status.ObservedGeneration then deployment spec is not updated yet.
	// it marked IsRolledout as false and update message accordingly
	if deploy.Generation <= deploy.Status.ObservedGeneration {
		// If Progressing condition's reason is ProgressDeadlineExceeded then it is not rolled out.
		if cond != nil && cond.Reason == "ProgressDeadlineExceeded" {
			r.Message = fmt.Sprintf("Deployment exceeded its progress deadline")
			return json.Marshal(r)
		}
		// if deploy.Status.UpdatedReplicas < *deploy.Spec.Replicas then some of the replicas are updated
		// and some of them are not. It marked IsRolledout as false and update message accordingly
		if deploy.Spec.Replicas != nil && deploy.Status.UpdatedReplicas < *deploy.Spec.Replicas {
			r.Message = fmt.Sprintf("Waiting for deployment rollout to finish: %d out of %d new replicas have been updated",
				deploy.Status.UpdatedReplicas, *deploy.Spec.Replicas)
			return json.Marshal(r)
		}
		// if deploy.Status.Replicas > deploy.Status.UpdatedReplicas then some of the older replicas are in running state
		// because newer replicas are not in running state. It waits for newer replica to come into reunning state then terminate.
		// It marked IsRolledout as false and update message accordingly
		if deploy.Status.Replicas > deploy.Status.UpdatedReplicas {
			r.Message = fmt.Sprintf("Waiting for deployment rollout to finish: %d old replicas are pending termination",
				deploy.Status.Replicas-deploy.Status.UpdatedReplicas)
			return json.Marshal(r)
		}
		// if deploy.Status.AvailableReplicas < deploy.Status.UpdatedReplicas then all the replicas are updated but they are
		// not in running state. It marked IsRolledout as false and update message accordingly.
		if deploy.Status.AvailableReplicas < deploy.Status.UpdatedReplicas {
			r.Message = fmt.Sprintf("Waiting for deployment rollout to finish: %d of %d updated replicas are available",
				deploy.Status.AvailableReplicas, deploy.Status.UpdatedReplicas)
		}
		r.IsRolledout = true
		r.Message = fmt.Sprintf("Deployment %q successfully rolled out", deploy.Name)
		return json.Marshal(r)
	}
	r.Message = fmt.Sprintf("Waiting for deployment spec update to be observed")
	return json.Marshal(r)
}

// AppsV1StatefulSetRolloutStatusGenerator generates RolloutOutput from statefulset object for that statefulset
// TODO verify it and update with code comments
var AppsV1StatefulSetRolloutStatusGenerator = func(sts api_apps_v1.StatefulSet) (op []byte, err error) {
	r := rolloutOutput{}
	r.IsRolledout = false
	if sts.Spec.UpdateStrategy.Type != api_apps_v1.RollingUpdateStatefulSetStrategyType {
		err = fmt.Errorf("rollout status is only available for %s strategy type", api_apps_v1.RollingUpdateStatefulSetStrategyType)
		return
	}
	if sts.Status.ObservedGeneration == 0 || sts.Generation > sts.Status.ObservedGeneration {
		r.Message = "Waiting for statefulset spec update to be observed"
		return json.Marshal(r)
	}
	if sts.Spec.Replicas != nil && sts.Status.ReadyReplicas < *sts.Spec.Replicas {
		r.Message = fmt.Sprintf("Waiting for %d pods to be ready ", *sts.Spec.Replicas-sts.Status.ReadyReplicas)
		return json.Marshal(r)
	}
	if sts.Spec.UpdateStrategy.Type == api_apps_v1.RollingUpdateStatefulSetStrategyType && sts.Spec.UpdateStrategy.RollingUpdate != nil {
		if sts.Spec.Replicas != nil && sts.Spec.UpdateStrategy.RollingUpdate.Partition != nil {
			if sts.Status.UpdatedReplicas < (*sts.Spec.Replicas - *sts.Spec.UpdateStrategy.RollingUpdate.Partition) {
				r.Message = fmt.Sprintf("Waiting for partitioned roll out to finish: %d out of %d new pods have been updated",
					sts.Status.UpdatedReplicas, *sts.Spec.Replicas-*sts.Spec.UpdateStrategy.RollingUpdate.Partition)
				return json.Marshal(r)
			}
		}
		r.Message = fmt.Sprintf("Partitioned roll out complete: %d new pods have been updated ", sts.Status.UpdatedReplicas)
		return json.Marshal(r)
	}
	if sts.Status.UpdateRevision != sts.Status.CurrentRevision {
		r.Message = fmt.Sprintf("Waiting for statefulset rolling update to complete %d pods at revision %s",
			sts.Status.UpdatedReplicas, sts.Status.UpdateRevision)
		return json.Marshal(r)
	}
	r.IsRolledout = true
	r.Message = fmt.Sprintf("Statefulset %q rolling update complete successfully", sts.Name)
	return json.Marshal(r)
}

// AppsV1DaemonSetRolloutStatusGenerator generates RolloutOutput from daemonSet object for that daemonSet
// TODO verify it and update with code comments
var AppsV1DaemonSetRolloutStatusGenerator = func(daemon api_apps_v1.DaemonSet) (op []byte, err error) {
	r := rolloutOutput{}
	r.IsRolledout = false
	if daemon.Spec.UpdateStrategy.Type != api_apps_v1.RollingUpdateDaemonSetStrategyType {
		err = fmt.Errorf("Rollout status is only available for %s strategy type", api_apps_v1.RollingUpdateStatefulSetStrategyType)
		return
	}
	if daemon.Generation <= daemon.Status.ObservedGeneration {
		if daemon.Status.UpdatedNumberScheduled < daemon.Status.DesiredNumberScheduled {
			r.Message = fmt.Sprintf("Waiting for daemon set %q rollout to finish: %d out of %d new pods have been updated",
				daemon.Name, daemon.Status.UpdatedNumberScheduled, daemon.Status.DesiredNumberScheduled)
			return json.Marshal(r)
		}
		if daemon.Status.NumberAvailable < daemon.Status.DesiredNumberScheduled {
			r.Message = fmt.Sprintf("Waiting for daemon set %q rollout to finish: %d of %d updated pods are available",
				daemon.Name, daemon.Status.NumberAvailable, daemon.Status.DesiredNumberScheduled)
		}
		r.IsRolledout = true
		r.Message = fmt.Sprintf("Daemon set %q successfully rolled out\n", daemon.Name)
		return json.Marshal(r)
	}
	r.Message = fmt.Sprintf("Waiting for daemon set spec update to be observed")
	return json.Marshal(r)
}
