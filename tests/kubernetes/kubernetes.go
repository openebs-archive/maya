// Copyright © 2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kubernetes

import (
	"strings"

	v1 "k8s.io/api/core/v1"
)

// CheckPodsRunning returns true if the number of pods is equal to expected pods and all pods are in running state
func CheckPodsRunning(pods v1.PodList, expectedPods int) bool {
	if len(pods.Items) < expectedPods {
		return false
	}

	for _, pod := range pods.Items {
		if pod.Status.Phase != "Running" {
			return false
		}
	}
	return true
}

// CheckForNamespace returns true if target namespace exists in v1.NamespaceList
func CheckForNamespace(namespaces v1.NamespaceList, targetNamespace string) bool {
	for _, namespace := range namespaces.Items {
		if namespace.GetName() == targetNamespace {
			return false
		}
	}
	return true
}

// CheckForPod returns true if pods is in Running state
func CheckForPod(pods v1.PodList, targetRegex string) bool {
	for _, pod := range pods.Items {
		if strings.Contains(pod.GetName(), targetRegex) && pod.Status.Phase == "Running" {
			return true
		}
	}

	return false
}
