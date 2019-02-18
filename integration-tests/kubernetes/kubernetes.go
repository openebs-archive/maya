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

// CheckForNamespace returns true if target namespace exists in v1.NamespaceList
func CheckForNamespace(namespaces v1.NamespaceList, targetNamespace string) bool {
	for _, namespace := range namespaces.Items {
		if namespace.GetName() == targetNamespace {
			return false
		}
	}
	return true
}
