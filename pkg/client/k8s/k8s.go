/*
Copyright 2017 The OpenEBS Authors

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
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	api_core_v1 "k8s.io/api/core/v1"
	api_extn_v1beta1 "k8s.io/api/extensions/v1beta1"
	mach_apis_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	typed_core_v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	typed_ext_v1beta "k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
)

// K8sClient provides the necessary utility to operate over
// various K8s Kind objects
type K8sClient struct {
	// ns refers to K8s namespace where the operation
	// will be performed
	ns string

	// cs refers to the ClientSet capable of communicating
	// within/outside the current K8s cluster
	cs *kubernetes.Clientset

	// PVC refers to a K8s PersistentVolumeClaim object
	// NOTE: This property enables unit testing
	PVC *api_core_v1.PersistentVolumeClaim

	// Pod refers to a K8s Pod object
	// NOTE: This property enables unit testing
	Pod *api_core_v1.Pod

	// Service refers to a K8s Service object
	// NOTE: This property enables unit testing
	Service *api_core_v1.Service

	// ConfigMap refers to a K8s Service object
	// NOTE: This property enables unit testing
	ConfigMap *api_core_v1.ConfigMap

	// Deployment refers to a K8s Deployment object
	// NOTE: This property enables unit testing
	Deployment *api_extn_v1beta1.Deployment

	// various cert related to connecting to K8s API
	caCert     string
	caPath     string
	clientCert string
	clientKey  string
	insecure   bool
}

func NewK8sClient(ns string) (*K8sClient, error) {
	// get the appropriate clientset
	cs, err := getInClusterCS()
	if err != nil {
		return nil, err
	}

	return &K8sClient{
		ns: ns,
		cs: cs,
	}, nil
}

// cmOps is a utility function that provides a instance capable of
// executing various K8s ConfigMap related operations.
func (k *K8sClient) cmOps() (typed_core_v1.ConfigMapInterface, error) {
	return k.cs.CoreV1().ConfigMaps(k.ns), nil
}

// GetConfigMap fetches the K8s ConfigMap with the provided name
func (k *K8sClient) GetConfigMap(name string, opts mach_apis_meta_v1.GetOptions) (*api_core_v1.ConfigMap, error) {
	if k.ConfigMap != nil {
		return k.ConfigMap, nil
	}

	cops, err := k.cmOps()
	if err != nil {
		return nil, err
	}

	return cops.Get(name, opts)
}

// pvcOps is a utility function that provides a instance capable of
// executing various K8s PVC related operations.
func (k *K8sClient) pvcOps() (typed_core_v1.PersistentVolumeClaimInterface, error) {
	return k.cs.CoreV1().PersistentVolumeClaims(k.ns), nil
}

// GetPVC fetches the K8s PVC with the provided name
func (k *K8sClient) GetPVC(name string, opts mach_apis_meta_v1.GetOptions) (*api_core_v1.PersistentVolumeClaim, error) {
	if k.PVC != nil {
		return k.PVC, nil
	}

	pops, err := k.pvcOps()
	if err != nil {
		return nil, err
	}

	return pops.Get(name, opts)
}

// podOps is a utility function that provides a instance capable of
// executing various K8s pod related operations.
func (k *K8sClient) podOps() (typed_core_v1.PodInterface, error) {
	return k.cs.CoreV1().Pods(k.ns), nil
}

// GetPod fetches the K8s Pod with the provided name
func (k *K8sClient) GetPod(name string, opts mach_apis_meta_v1.GetOptions) (*api_core_v1.Pod, error) {
	if k.Pod != nil {
		return k.Pod, nil
	}

	pops, err := k.podOps()
	if err != nil {
		return nil, err
	}

	return pops.Get(name, opts)
}

// serviceOps is a utility function that provides a instance capable of
// executing various k8s service related operations.
func (k *K8sClient) serviceOps() (typed_core_v1.ServiceInterface, error) {
	return k.cs.CoreV1().Services(k.ns), nil
}

// GetService fetches the K8s Service with the provided name
func (k *K8sClient) GetService(name string, opts mach_apis_meta_v1.GetOptions) (*api_core_v1.Service, error) {
	if k.Service != nil {
		return k.Service, nil
	}

	sops, err := k.serviceOps()
	if err != nil {
		return nil, err
	}

	return sops.Get(name, opts)
}

// deploymentOps is a utility function that provides a instance capable of
// executing various k8s Deployment related operations.
func (k *K8sClient) deploymentOps() (typed_ext_v1beta.DeploymentInterface, error) {
	return k.cs.ExtensionsV1beta1().Deployments(k.ns), nil
}

// GetDeployment fetches the K8s Deployment with the provided name
func (k *K8sClient) GetDeployment(name string, opts mach_apis_meta_v1.GetOptions) (*api_extn_v1beta1.Deployment, error) {
	if k.Deployment != nil {
		return k.Deployment, nil
	}

	dops, err := k.deploymentOps()
	if err != nil {
		return nil, err
	}

	return dops.Get(name, opts)
}

// getInClusterCS is used to initialize and return a new http client capable
// of invoking K8s APIs within the cluster
func getInClusterCS() (*kubernetes.Clientset, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	// creates the in-cluster clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}
