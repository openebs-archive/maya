/*
Copyright 2018 The OpenEBS Authors

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

package internalk8s

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
)

type k8sClient struct {
	clientSet *kubernetes.Clientset
}

// NewK8sClient returns k8sClient, it is used to do all the internal agerated
// kubernetes operations
func NewK8sClient() (*k8sClient, error) {
	configPath, err := getExternalConfigPath()
	if err != nil {
		return nil, err
	}
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		return nil, err
	}
	clientSet, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}
	return &k8sClient{
		clientSet: clientSet,
	}, nil
}

// getEBSPersistantVolumeClaims returns claimNames, which are running on OpenEBS
func (k *k8sClient) getEBSPersistantVolumeClaims(namespace string) (claimNames []string, err error) {
	volumeClaimList, err := k.clientSet.Core().PersistentVolumeClaims(namespace).List(metav1.ListOptions{})
	if err != nil {
		return
	}
	for _, volumeClaim := range volumeClaimList.Items {
		if volumeClaim.Annotations["volume.beta.kubernetes.io/storage-provisioner"] == "openebs.io/provisioner-iscsi" {
			claimNames = append(claimNames, volumeClaim.GetName())
		}
	}
	return
}

// GetPodWithEBSVolume returns pod running on openebs volumes
func (k *k8sClient) GetPodWithEBSVolume(namespace string) (pods []v1.Pod, err error) {
	// get claim Name from all namespaces
	claimNames, err := k.getEBSPersistantVolumeClaims("")
	if err != nil {
		return
	}
	// don't do further pod request
	if len(claimNames) == 0 {
		return
	}
	podList, err := k.clientSet.Core().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		return
	}
	for _, pod := range podList.Items {
		if isEBSPod(claimNames, pod) {
			pods = append(pods, pod)
		}
	}
	return
}
