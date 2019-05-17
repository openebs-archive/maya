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
	"fmt"
	"os"

	kube "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// HomeDir return the Home Directory of the environement
func HomeDir() (string, error) {
	if h := os.Getenv("HOME"); h != "" { // linux
		return h, nil
	} else if h := os.Getenv("USERPROFILE"); h != "" { // windows
		return h, nil
	}

	return "", fmt.Errorf("Not able to locate home directory")
}

// GetConfigPath returns the path of kubeconfig
func GetConfigPath() (kubeConfigPath string, err error) {
	home, err := HomeDir()
	if err != nil {
		return
	}
	kubeConfigPath = os.Getenv("KUBECONFIG")
	if kubeConfigPath == "" {
		// Parse the kube config path
		kubeConfigPath = home + "/.kube/config"
	}
	return kubeConfigPath, err
}

// GetClientSet returns the clientset for interacting the kubernetes cluster
func GetClientSet() (cl *kube.Clientset, err error) {
	kubeConfigPath, err := GetConfigPath()
	if err != nil {
		return
	}

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return
	}

	// create the clientset
	return kube.NewForConfig(config)
}
