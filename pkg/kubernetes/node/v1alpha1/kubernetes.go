// Copyright © 2018-2019 The OpenEBS Authors
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

package v1alpha1

import (
	client "github.com/openebs/maya/pkg/kubernetes/client/v1alpha1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// getClientsetFn is a typed function that
// abstracts fetching of clientset
type getClientsetFn func() (clientset *kubernetes.Clientset, err error)

// getClientsetFromPathFn is a typed function that
// abstracts fetching of clientset from kubeConfigPath
type getClientsetForPathFn func(kubeConfigPath string) (clientset *kubernetes.Clientset, err error)

// listFn is a typed function that abstracts
// listing of nodes
type listFn func(cli *kubernetes.Clientset, opts metav1.ListOptions) (*corev1.NodeList, error)

// Kubeclient enables kubernetes API operations on node instance
type Kubeclient struct {
	// clientset refers to node clientset
	// that will be responsible to
	// make kubernetes API calls
	clientset *kubernetes.Clientset

	// kubeconfig path to get kubernetes clientset
	kubeConfigPath string

	// functions useful during mocking
	getClientset        getClientsetFn
	getClientsetForPath getClientsetForPathFn
	list                listFn
}

// KubeClientBuildOption defines the abstraction
// to build a kubeclient instance
type KubeClientBuildOption func(*Kubeclient)

func (k *Kubeclient) withDefaults() {
	if k.getClientset == nil {
		k.getClientset = func() (clients *kubernetes.Clientset, err error) {
			return client.New().Clientset()
		}
	}
	if k.getClientsetForPath == nil {
		k.getClientsetForPath = func(kubeConfigPath string) (clients *kubernetes.Clientset, err error) {
			return client.New(client.WithKubeConfigPath(kubeConfigPath)).Clientset()
		}
	}
	if k.list == nil {
		k.list = func(cli *kubernetes.Clientset, opts metav1.ListOptions) (*corev1.NodeList, error) {
			return cli.CoreV1().Nodes().List(opts)
		}
	}
}

// WithKubeConfigPath sets the kubeConfig path
// against client instance
func WithKubeConfigPath(path string) KubeClientBuildOption {
	return func(k *Kubeclient) {
		k.kubeConfigPath = path
	}
}

// NewKubeClient returns a new instance of kubeclient meant for node
func NewKubeClient(opts ...KubeClientBuildOption) *Kubeclient {
	k := &Kubeclient{}
	for _, o := range opts {
		o(k)
	}
	k.withDefaults()
	return k
}

func (k *Kubeclient) getClientsetForPathOrDirect() (*kubernetes.Clientset, error) {
	if k.kubeConfigPath != "" {
		return k.getClientsetForPath(k.kubeConfigPath)
	}
	return k.getClientset()
}

// getClientsetOrCached returns either a new instance
// of kubernetes client or its cached copy
func (k *Kubeclient) getClientsetOrCached() (*kubernetes.Clientset, error) {
	if k.clientset != nil {
		return k.clientset, nil
	}

	cs, err := k.getClientsetForPathOrDirect()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get clientset")
	}
	k.clientset = cs
	return k.clientset, nil
}

// List returns a list of nodes instances present in kubernetes cluster
func (k *Kubeclient) List(opts metav1.ListOptions) (*corev1.NodeList, error) {
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list")
	}
	return k.list(cli, opts)
}
