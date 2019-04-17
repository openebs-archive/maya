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
	"strings"

	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	client "github.com/openebs/maya/pkg/kubernetes/client/v1alpha1"
	admission "k8s.io/api/admissionregistration/v1beta1"

	"k8s.io/client-go/kubernetes"
)

// getClientsetFunc is a typed function that
// abstracts fetching internal clientset
type getClientsetFunc func() (cs *kubernetes.Clientset, err error)

// listFunc is a typed function that abstracts
// listing validatingWebhookConfiguration instances
type listFunc func(cs *kubernetes.Clientset, opts metav1.ListOptions) (*admission.ValidatingWebhookConfigurationList, error)

// getFunc is a typed function that abstracts
// getting validatingWebhookConfiguration instances
type getFunc func(cs *kubernetes.Clientset, name string, opts metav1.GetOptions) (*admission.ValidatingWebhookConfiguration, error)

// Kubeclient enables kubernetes API operations
// on upgrade result instance
type Kubeclient struct {
	// clientset refers to upgrade's
	// clientset that will be responsible to
	// make kubernetes API calls
	clientset *kubernetes.Clientset
	namespace string
	// functions useful during mocking
	getClientset getClientsetFunc
	list         listFunc
	get          getFunc
}

// KubeclientBuildOption defines the abstraction
// to build a kubeclient instance
type KubeclientBuildOption func(*Kubeclient)

// withDefaults sets the default options
// of kubeclient instance
func (k *Kubeclient) withDefaults() {
	if k.getClientset == nil {
		k.getClientset = func() (cs *kubernetes.Clientset, err error) {
			config, err := client.GetConfig(client.New())
			if err != nil {
				return nil, err
			}
			return kubernetes.NewForConfig(config)
		}
	}
	if k.list == nil {
		k.list = func(cs *kubernetes.Clientset, opts metav1.ListOptions) (*admission.ValidatingWebhookConfigurationList, error) {
			return cs.Admissionregistration().ValidatingWebhookConfigurations().List(opts)
		}
	}
	if k.get == nil {
		k.get = func(cs *kubernetes.Clientset, name string, opts metav1.GetOptions) (*admission.ValidatingWebhookConfiguration, error) {
			return cs.Admissionregistration().ValidatingWebhookConfigurations().Get(name, opts)
		}
	}
}

// WithClientset sets the kubernetes clientset against
// the kubeclient instance
func WithClientset(c *kubernetes.Clientset) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.clientset = c
	}
}

// KubeClient returns a new instance of kubeclient meant for
// admission webhook related operations
func KubeClient(opts ...KubeclientBuildOption) *Kubeclient {
	k := &Kubeclient{}
	for _, o := range opts {
		o(k)
	}
	k.withDefaults()
	return k
}

// WithNamespace sets namespace that should be used during
// kuberenets API calls against namespaced resource
func WithNamespace(namespace string) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.namespace = namespace
	}
}

// getClientOrCached returns either a new instance
// of kubernetes client or its cached copy
func (k *Kubeclient) getClientOrCached() (*kubernetes.Clientset, error) {
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

// List takes label and field selectors, and returns the list of ValidatingWebhookConfigurations
// that match those selectors.
func (k *Kubeclient) List(opts metav1.ListOptions) (*admission.ValidatingWebhookConfigurationList, error) {
	cs, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.list(cs, opts)
}

// Get takes name of the validatingWebhookConfiguration, and returns the
// corresponding validatingWebhookConfiguration object, and an error if there is any.
func (k *Kubeclient) Get(name string, opts metav1.GetOptions) (*admission.ValidatingWebhookConfiguration, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("failed to get ValidatingWebhookConfiguration: missing configuration name")
	}
	cs, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.get(cs, name, opts)
}
