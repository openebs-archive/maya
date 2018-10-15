/*
Copyright 2018 The OpenEBS Authors.
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

package k8s

import (
	api_core_v1 "k8s.io/api/core/v1"
	// Install special auth plugins like GCP Plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
)

// IsNSinGoodPhase checks if supplied namespace is in good phase or not
// by matching phase of supplied namespace with pre-identified Good phase list (NsGoodPhases)
func (k8s K8S) IsNSinGoodPhase(namespace api_core_v1.Namespace) bool {
	for _, phase := range NsGoodPhases {
		if phase == namespace.Status.Phase {
			return true
		}
	}

	return false
}

// IsPodStateWait checks if supplied pod state is wait state or not
// by matching state of supplied pod state with pre-identified Wait states list (PodWaitStates)
func (k8s K8S) IsPodStateWait(podState string) bool {
	for _, state := range PodWaitStates {
		if state == podState {
			return true
		}
	}

	return false
}

// IsPodStateGood checks if supplied pod state is good or not
// by matching state of supplied pod state with pre-identified Good states list (PodGoodStates)
func (k8s K8S) IsPodStateGood(podState string) bool {
	for _, state := range PodGoodStates {
		if state == podState {
			return true
		}
	}

	return false
}

// IsPodRunning returns whether all the containers in a given pod is running
// TODO: Check if it works for the Pod where one container has successfully terminated but other is running
func (k8s K8S) IsPodRunning(pod *api_core_v1.Pod) bool {
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if !containerStatus.Ready {
			return false
		}
	}
	return true
}
