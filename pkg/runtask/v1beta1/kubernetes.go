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

package v1beta1

import (
	"strings"

	"github.com/pkg/errors"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
	client "github.com/openebs/maya/pkg/kubernetes/client/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// getClientsetFunc is a typed function that
// abstracts fetching internal clientset
type getClientsetFunc func() (cs *clientset.Clientset, err error)

// getFunc is a typed function that abstracts
// getting runtask instances
type getFunc func(cs *clientset.Clientset, name, namespace string, opts metav1.GetOptions) (*apis.RunTask, error)

// Kubeclient enables kubernetes API operationson runtask instance
type Kubeclient struct {
	// clientset refers to openebs clientset that will be
	// responsible to make kubernetes API calls
	clientset *clientset.Clientset
	namespace string
	// functions useful during mocking
	getClientset getClientsetFunc
	get          getFunc
}

// KubeclientBuildOption defines the abstraction
// to build a Kubeclient instance
type KubeclientBuildOption func(*Kubeclient)

// withDefaults sets the default options of Kubeclient instance
func (k *Kubeclient) withDefaults() {
	if k.getClientset == nil {
		k.getClientset = func() (cs *clientset.Clientset, err error) {
			config, err := client.GetConfig(client.New())
			if err != nil {
				return nil, errors.Wrapf(err, "failed to get clientset")
			}
			return clientset.NewForConfig(config)
		}
	}

	if k.get == nil {
		k.get = func(cs *clientset.Clientset, name, namespace string,
			opts metav1.GetOptions) (*apis.RunTask, error) {
			return cs.OpenebsV1alpha1().
				RunTasks(namespace).
				Get(name, opts)
		}
	}
}

// WithClientset sets the kubernetes clientset against
// the kubeclient instance
func WithClientset(c *clientset.Clientset) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.clientset = c
	}
}

// WithNamespace sets the namespace against the kubeclient instance
func WithNamespace(namespace string) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.namespace = namespace
	}
}

// NewKubeClient returns a new instance of kubeclient meant for
// runtask related operations
func NewKubeClient(opts ...KubeclientBuildOption) *Kubeclient {
	k := &Kubeclient{}
	for _, o := range opts {
		o(k)
	}
	k.withDefaults()
	return k
}

// getClientOrCached returns either a new instance
// of kubernetes client or its cached copy
func (k *Kubeclient) getClientOrCached() (*clientset.Clientset, error) {
	if k.clientset != nil {
		return k.clientset, nil
	}
	c, err := k.getClientset()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get clientset")
	}
	k.clientset = c
	return k.clientset, nil
}

// Get returns a runtask instance for given name
func (k *Kubeclient) Get(name string, opts metav1.GetOptions) (*apis.RunTask, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("failed to get runtask: missing runtask name")
	}
	cs, err := k.getClientOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get runtask '%s'", name)
	}
	rt, err := k.get(cs, name, k.namespace, opts)
	if err != nil {
		return nil, errors.Wrapf(err,
			"failed to get runtask '%s' with namespace '%s'", name, k.namespace)
	}
	return rt, nil
}
