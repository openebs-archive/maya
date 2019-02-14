package kubernetes

import (
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
