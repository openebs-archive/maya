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
	"encoding/json"
	"errors"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/upgrade/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/generated/openebs.io/upgrade/v1alpha1/clientset/internalclientset"
	client "github.com/openebs/maya/pkg/kubernetes/client/v1alpha1"
)

// getClientsetFunc is a typed function that
// abstracts fetching internal clientset
type getClientsetFunc func() (cs *clientset.Clientset, err error)

// listFunc is a typed function that abstracts
// listing upgrade result instances
type listFunc func(cs *clientset.Clientset, namespace string, opts metav1.ListOptions) (*apis.UpgradeResultList, error)

// getFunc is a typed function that abstracts
// getting upgrade result instances
type getFunc func(cs *clientset.Clientset, name string, namespace string, opts metav1.GetOptions) (*apis.UpgradeResult, error)

// createFunc is a typed function that abstracts
// creating upgrade result instances
type createFunc func(cs *clientset.Clientset, upgradeResultObj *apis.UpgradeResult,
	namespace string) (*apis.UpgradeResult, error)

// patchFunc is a typed function that abstracts
// patching upgrade result instances
type patchFunc func(cs *clientset.Clientset, name string, pt types.PatchType, patchObj []byte,
	namespace string) (*apis.UpgradeResult, error)

// kubeclient enables kubernetes API operations
// on upgrade result instance
type kubeclient struct {
	// clientset refers to upgrade's
	// clientset that will be responsible to
	// make kubernetes API calls
	clientset *clientset.Clientset
	namespace string
	// functions useful during mocking
	getClientset getClientsetFunc
	list         listFunc
	get          getFunc
	create       createFunc
	patch        patchFunc
}

// kubeclientBuildOption defines the abstraction
// to build a kubeclient instance
type kubeclientBuildOption func(*kubeclient)

// withDefaults sets the default options
// of kubeclient instance
func (k *kubeclient) withDefaults() {
	if k.getClientset == nil {
		k.getClientset = func() (cs *clientset.Clientset, err error) {
			config, err := client.GetConfig(client.New())
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
	if k.get == nil {
		k.get = func(cs *clientset.Clientset, name string, namespace string, opts metav1.GetOptions) (*apis.UpgradeResult, error) {
			return cs.OpenebsV1alpha1().UpgradeResults(namespace).Get(name, opts)
		}
	}
	if k.create == nil {
		k.create = func(cs *clientset.Clientset, upgradeResultObj *apis.UpgradeResult,
			namespace string) (*apis.UpgradeResult, error) {
			return cs.OpenebsV1alpha1().
				UpgradeResults(namespace).
				Create(upgradeResultObj)
		}
	}
	if k.patch == nil {
		k.patch = func(cs *clientset.Clientset, name string,
			pt types.PatchType, patchObj []byte,
			namespace string) (*apis.UpgradeResult, error) {
			return cs.OpenebsV1alpha1().
				UpgradeResults(namespace).
				Patch(name, pt, patchObj)
		}
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
	return k.list(cs, k.namespace, opts)
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
	return k.get(cs, name, k.namespace, opts)
}

// Create creates an upgrade result instance
// and returns raw upgradeResult instance
func (k *kubeclient) CreateAsRaw(upgradeResultObj *apis.UpgradeResult) ([]byte, error) {
	ur, err := k.Create(upgradeResultObj)
	if err != nil {
		return nil, err
	}
	return json.Marshal(ur)
}

// Create creates an upgrade result instance in kubernetes cluster
func (k *kubeclient) Create(upgradeResultObj *apis.UpgradeResult) (*apis.UpgradeResult, error) {
	cs, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.create(cs, upgradeResultObj, k.namespace)
}

// Patch returns the patched upgrade result instance
func (k *kubeclient) Patch(name string, pt types.PatchType,
	patchObj []byte) (*apis.UpgradeResult, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("failed to patch upgrade result: missing upgradeResult name")
	}
	cs, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.patch(cs, name, pt, patchObj, k.namespace)
}
