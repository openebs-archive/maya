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
	"errors"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/upgrade/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/generated/openebs.io/upgrade/v1alpha1/clientset/internalclientset"
	client "github.com/openebs/maya/pkg/kubernetes/client/v1alpha1"
)

// getClientFunc is a typed function that
// abstracts fetching kubernetes client
type getClientFunc func() (cli *client.Client)

// getClientsetFunc is a typed function that
// abstracts fetching internal clientset
type getClientsetFunc func() (cs *clientset.Clientset, err error)

// listFunc is a typed function that abstracts
// listing upgrade result instances
type listFunc func(cs *clientset.Clientset, opts metav1.ListOptions) (*apis.UpgradeResultList, error)

// getFunc is a typed function that abstracts
// getting upgrade result instances
type getFunc func(cs *clientset.Clientset, name string, opts metav1.GetOptions) (*apis.UpgradeResult, error)

// kubeclient enables kubernetes API operations
// on upgrade result instance
type kubeclient struct {
	client    *client.Client
	getClient getClientFunc
	// clientset refers to upgrade's
	// clientset that will be responsible to
	// make kubernetes API calls
	clientset *clientset.Clientset
	// handle to get kubernetes config
	getConfig client.GetConfigFunc
	// functions useful during mocking
	getClientset getClientsetFunc
	namespace    string
	list         listFunc
	get          getFunc
}

// kubeclientBuildOption defines the abstraction
// to build a kubeclient instance
type kubeclientBuildOption func(*kubeclient)

// withDefaults sets the default options
// of kubeclient instance
func (k *kubeclient) withDefaults() {
	if k.getClient == nil {
		k.getClient = func() (cli *client.Client) {
			return client.New()
		}
	}
	if k.getClientset == nil {
		k.getClientset = func() (cs *clientset.Clientset, err error) {
			config, err := client.GetConfig(k.getClient())
			if err != nil {
				return nil, err
			}
			return clientset.NewForConfig(config)
		}
	}
	if k.list == nil {
		k.list = func(cs *clientset.Clientset, opts metav1.ListOptions) (*apis.UpgradeResultList, error) {
			return cs.OpenebsV1alpha1().UpgradeResults(k.namespace).List(opts)
		}
	}
	if k.get == nil {
		k.get = func(cs *clientset.Clientset, name string, opts metav1.GetOptions) (*apis.UpgradeResult, error) {
			return cs.OpenebsV1alpha1().UpgradeResults(k.namespace).Get(name, opts)
		}
	}
}

// WithClient sets the kubernetes client against
// the kubeclient instance
func WithClient(c *client.Client) kubeclientBuildOption {
	return func(k *kubeclient) {
		k.client = c
	}
}

// WithClientset sets the kubernetes clientset against
// the kubeclient instance
func WithClientset(c *clientset.Clientset) kubeclientBuildOption {
	return func(k *kubeclient) {
		k.clientset = c
	}
}

// KubeClient returns a new instance of kubeclient meant for
// upgrade result related operations
func KubeClient(opts ...kubeclientBuildOption) *kubeclient {
	k := &kubeclient{}
	for _, o := range opts {
		o(k)
	}
	k.withDefaults()
	return k
}

// WithNamespace sets namespace that should be used during
// kuberenets API calls against upgradeResult resource
func WithNamespace(namespace string) kubeclientBuildOption {
	return func(k *kubeclient) {
		k.namespace = namespace
	}
}

// getClientOrCached returns either a new instance
// of kubernetes client or its cached copy
func (k *kubeclient) getClientOrCached() (*clientset.Clientset, error) {
	if k.clientset != nil {
		return k.clientset, nil
	}
	c, err := k.getClientset()
	if err != nil {
		return nil, err
	}
	k.clientset = c
	return k.clientset, nil
}

// List returns a list of upgrade result
// instances present in kubernetes cluster
func (k *kubeclient) List(opts metav1.ListOptions) (*apis.UpgradeResultList, error) {
	cs, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.list(cs, opts)
}

// Get returns an upgrade result instance from kubernetes cluster
func (k *kubeclient) Get(name string, opts metav1.GetOptions) (*apis.UpgradeResult, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("failed to get upgrade result: missing upgradeResult name")
	}
	cs, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.get(cs, name, opts)
}
