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
	kclient "github.com/openebs/maya/pkg/kubernetes/client/v1alpha1"
)

// getClientsetFunc is a typed function that
// abstracts fetching internal clientset
type getClientsetFunc func() (cs *clientset.Clientset, err error)

// listFunc is a typed function that abstracts
// listing upgrade result instances
type listFunc func(cs *clientset.Clientset, namespace string, opts metav1.ListOptions) (*apis.UpgradeResultList, error)

// getFunc is a typed function that abstracts
// getting upgrade result instances
type getFunc func(cs *clientset.Clientset, name string, opts metav1.GetOptions) (*apis.UpgradeResult, error)

// kubeclient enables kubernetes API operations
// on upgrade result instance
type kubeclient struct {
	*kclient.Client
	// clientset refers to upgrade's
	// clientset that will be responsible to
	// make kubernetes API calls
	clientset *clientset.Clientset
	// handle to get kubernetes config
	getConfig kclient.GetConfigFunc
	// functions useful during mocking
	getClientset getClientsetFunc
	list         listFunc
	get          getFunc
}

// kubeclientBuildOption defines the abstraction
// to build a kubeclient instance
type kubeclientBuildOption func(*kubeclient)

// withDefaults sets the default options
// of kubeclient instance
func (k *kubeclient) withDefaults() {
	if k.getClientset == nil {
		k.getClientset = func() (clients *clientset.Clientset, err error) {
			config, err := kclient.GetConfig(k.Client)
			if err != nil {
				return nil, err
			}
			return clientset.NewForConfig(config)
		}
	}
	if k.list == nil {
		k.list = func(cs *clientset.Clientset, namespace string, opts metav1.ListOptions) (*apis.UpgradeResultList, error) {
			return cs.OpenebsV1alpha1().UpgradeResults(namespace).List(opts)
		}
	}
}

// WithClientset sets the kubernetes client against
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
func (k *kubeclient) List(namespace string, opts metav1.ListOptions) (*apis.UpgradeResultList, error) {
	if strings.TrimSpace(namespace) == "" {
		return nil, errors.New("failed to get namespace for upgrade result: missing upgradeResult namespace")
	}
	cs, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.list(cs, namespace, opts)
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
