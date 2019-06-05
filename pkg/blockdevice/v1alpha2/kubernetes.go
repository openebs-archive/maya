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

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kclient "github.com/openebs/maya/pkg/kubernetes/client/v1alpha1"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/generated/openebs.io/ndm/v1alpha1/clientset/internalclientset"
)

//TODO: While using these packages UnitTest must be written to corresponding function

// getClientsetFn is a typed function that
// abstracts fetching of internal clientset
type getClientsetFn func() (clientset *clientset.Clientset, err error)

// getClientsetFromPathFn is a typed function that
// abstracts fetching of clientset from kubeConfigPath
type getClientsetForPathFn func(kubeConfigPath string) (clientset *clientset.Clientset, err error)

// listFn is a typed function that abstracts
// listing of block device
type listFn func(cli *clientset.Clientset, namespace string, opts metav1.ListOptions) (*apis.BlockDeviceList, error)

// getFn is a typed function that
// abstracts fetching of block deivce
type getFn func(cli *clientset.Clientset, namespace, name string, opts metav1.GetOptions) (*apis.BlockDevice, error)

// Kubeclient enables kubernetes API operations
// on block device instance
type Kubeclient struct {
	// clientset refers to block device
	// clientset that will be responsible to
	// make kubernetes API calls
	clientset *clientset.Clientset
	// kubeconfig path to get kubernetes clientset
	kubeConfigPath string
	namespace      string
	// functions useful during mocking
	getClientset        getClientsetFn
	getClientsetForPath getClientsetForPathFn
	list                listFn
	get                 getFn
}

// KubeclientBuildOption defines the abstraction
// to build a kubeclient instance
type KubeclientBuildOption func(*Kubeclient)

// WithDefaults sets the default options
// of kubeclient instance
func (k *Kubeclient) WithDefaults() {
	if k.getClientset == nil {
		k.getClientset = func() (clients *clientset.Clientset, err error) {
			config, err := kclient.New().Config()
			if err != nil {
				return nil, err
			}
			return clientset.NewForConfig(config)
		}
	}
	if k.getClientsetForPath == nil {
		k.getClientsetForPath = func(kubeConfigPath string) (clients *clientset.Clientset, err error) {
			config, err := kclient.New(kclient.WithKubeConfigPath(kubeConfigPath)).Config()
			if err != nil {
				return nil, err
			}
			return clientset.NewForConfig(config)
		}
	}
	if k.list == nil {
		k.list = func(cli *clientset.Clientset, namespace string, opts metav1.ListOptions) (*apis.BlockDeviceList, error) {
			return cli.OpenebsV1alpha1().BlockDevices(namespace).List(opts)
		}
	}

	if k.get == nil {
		k.get = func(cli *clientset.Clientset, namespace, name string, opts metav1.GetOptions) (*apis.BlockDevice, error) {
			return cli.OpenebsV1alpha1().BlockDevices(namespace).Get(name, opts)
		}
	}
}

// WithKubeClient sets the kubernetes client against
// the kubeclient instance
func WithKubeClient(c *clientset.Clientset) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.clientset = c
	}
}

// WithKubeConfigPath sets the kubeConfig path
// against client instance
func WithKubeConfigPath(kubeConfigPath string) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.kubeConfigPath = kubeConfigPath
	}
}

// NewKubeClient returns a new instance of kubeclient meant for
// block device operations
func NewKubeClient(opts ...KubeclientBuildOption) *Kubeclient {
	k := &Kubeclient{}
	for _, o := range opts {
		o(k)
	}
	k.WithDefaults()
	return k
}

func (k *Kubeclient) getClientsetForPathOrDirect() (*clientset.Clientset, error) {
	if k.kubeConfigPath != "" {
		return k.getClientsetForPath(k.kubeConfigPath)
	}
	return k.getClientset()
}

// WithNamespace sets the kubernetes namespace against
// the provided namespace
func (k *Kubeclient) WithNamespace(namespace string) *Kubeclient {
	k.namespace = namespace
	return k
}

// getClientOrCached returns either a new instance
// of kubernetes client or its cached copy
func (k *Kubeclient) getClientOrCached() (*clientset.Clientset, error) {
	if k.clientset != nil {
		return k.clientset, nil
	}
	c, err := k.getClientsetForPathOrDirect()
	if err != nil {
		return nil, err
	}
	k.clientset = c
	return k.clientset, nil
}

// List returns a list of disk
// instances present in kubernetes cluster
func (k *Kubeclient) List(opts metav1.ListOptions) (*apis.BlockDeviceList, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.list(cli, k.namespace, opts)
}

// Get returns a disk object
func (k *Kubeclient) Get(name string, opts metav1.GetOptions) (*apis.BlockDevice, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.get(cli, k.namespace, name, opts)
}
