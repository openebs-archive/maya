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
	apisv1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
	kclient "github.com/openebs/maya/pkg/kubernetes/client/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// getClientsetFn is a typed function that
// abstracts fetching of internal clientset
type getClientsetFn func() (clientset *clientset.Clientset, err error)

// getClientsetFromPathFn is a typed function that
// abstracts fetching of clientset from kubeConfigPath
type getClientsetForPathFn func(kubeConfigPath string) (clientset *clientset.Clientset, err error)

// listFn is a typed function that abstracts
// listing of cstor pool
type listFn func(cli *clientset.Clientset, opts metav1.ListOptions) (*apisv1alpha1.CStorPoolClusterList, error)

type getFn func(cli *clientset.Clientset, name string, opts metav1.GetOptions) (*apisv1alpha1.CStorPoolCluster, error)

type createFn func(cli *clientset.Clientset, spc *apisv1alpha1.CStorPoolCluster) (*apisv1alpha1.CStorPoolCluster, error)

type deleteFn func(cli *clientset.Clientset, name string, opts *metav1.DeleteOptions) (*apisv1alpha1.CStorPoolCluster, error)

type updateFn func(cli *clientset.Clientset, spc *apisv1alpha1.CStorPoolCluster) (*apisv1alpha1.CStorPoolCluster, error)

// Kubeclient enables kubernetes API operations
// on cstor storage pool instance
type Kubeclient struct {
	// clientset refers to cstor storage pool's
	// clientset that will be responsible to
	// make kubernetes API calls
	clientset *clientset.Clientset
	// kubeconfig path to get kubernetes clientset
	kubeConfigPath string
	// functions useful during mocking
	getClientset        getClientsetFn
	getClientsetForPath getClientsetForPathFn
	list                listFn
	get                 getFn
	create              createFn
	del                 deleteFn
	update              updateFn
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
		k.list = func(cli *clientset.Clientset, opts metav1.ListOptions) (*apisv1alpha1.CStorPoolClusterList, error) {
			return cli.OpenebsV1alpha1().CStorPoolClusters().List(opts)
		}
	}

	if k.get == nil {
		k.get = func(cli *clientset.Clientset, name string, opts metav1.GetOptions) (*apisv1alpha1.CStorPoolCluster, error) {
			return cli.OpenebsV1alpha1().CStorPoolClusters().Get(name, opts)
		}
	}

	if k.create == nil {
		k.create = func(cli *clientset.Clientset, spc *apisv1alpha1.CStorPoolCluster) (*apisv1alpha1.CStorPoolCluster, error) {
			return cli.OpenebsV1alpha1().CStorPoolClusters().Create(spc)
		}
	}

	if k.update == nil {
		k.update = func(cli *clientset.Clientset, spc *apisv1alpha1.CStorPoolCluster) (*apisv1alpha1.CStorPoolCluster, error) {
			return cli.OpenebsV1alpha1().CStorPoolClusters().Update(spc)
		}
	}

	if k.del == nil {
		k.del = func(cli *clientset.Clientset, name string, opts *metav1.DeleteOptions) (*apisv1alpha1.CStorPoolCluster, error) {
			return nil, cli.OpenebsV1alpha1().CStorPoolClusters().Delete(name, opts)
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
func WithKubeConfigPath(path string) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.kubeConfigPath = path
	}
}

// NewKubeClient returns a new instance of kubeclient meant for
// cstor volume replica operations
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

// List returns a list of cstor pool
// instances present in kubernetes cluster
func (k *Kubeclient) List(opts metav1.ListOptions) (*apisv1alpha1.CStorPoolClusterList, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.list(cli, opts)
}

// Get returns a spc object
func (k *Kubeclient) Get(name string, opts metav1.GetOptions) (*apisv1alpha1.CStorPoolCluster, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.get(cli, name, opts)
}

// Create creates a spc object
func (k *Kubeclient) Create(spc *apisv1alpha1.CStorPoolCluster) (*apisv1alpha1.CStorPoolCluster, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.create(cli, spc)
}

// Update updates a spc object
func (k *Kubeclient) Update(spc *apisv1alpha1.CStorPoolCluster) (*apisv1alpha1.CStorPoolCluster, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.update(cli, spc)
}

// Delete deletes a spc object
func (k *Kubeclient) Delete(name string, opts *metav1.DeleteOptions) (*apisv1alpha1.CStorPoolCluster, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.del(cli, name, opts)
}
