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

	client "github.com/openebs/maya/pkg/kubernetes/client/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// getClientsetFn is a typed function that
// abstracts fetching of internal clientset
type getClientsetFn func() (clientset *kubernetes.Clientset, err error)

// getFn is a typed function that abstracts get of service instances
type getFn func(cli *kubernetes.Clientset, name, namespace string,
	opts metav1.GetOptions) (*corev1.Service, error)

// listFn is a typed function that abstracts list of service instances
type listFn func(cli *kubernetes.Clientset, namespace string,
	opts metav1.ListOptions) (*corev1.ServiceList, error)

// delFn is a typed function that abstracts delete of service instances
type delFn func(cli *kubernetes.Clientset, name, namespace string,
	opts *metav1.DeleteOptions) error

// Kubeclient enables kubernetes API operations on service instance
type Kubeclient struct {
	// clientset refers to kubernetes clientset. It is responsible to
	// make kubernetes API calls for crud op
	clientset *kubernetes.Clientset
	namespace string

	// functions useful during mocking
	getClientset getClientsetFn
	list         listFn
	get          getFn
	del          delFn
}

// KubeclientBuildOption defines the abstraction to build a kubeclient instance
type KubeclientBuildOption func(*Kubeclient)

// withDefaults sets the default options of kubeclient instance
func (k *Kubeclient) withDefaults() {
	if k.getClientset == nil {
		k.getClientset = func() (clients *kubernetes.Clientset, err error) {
			config, err := client.New().Config()
			if err != nil {
				return nil, err
			}
			return kubernetes.NewForConfig(config)
		}
	}
	if k.get == nil {
		k.get = func(cli *kubernetes.Clientset, name,
			namespace string, opts metav1.GetOptions) (
			r *corev1.Service, err error) {
			r, err = cli.CoreV1().
				Services(namespace).
				Get(name, opts)
			return
		}
	}
	if k.list == nil {
		k.list = func(cli *kubernetes.Clientset,
			namespace string, opts metav1.ListOptions) (
			rl *corev1.ServiceList, err error) {
			rl, err = cli.CoreV1().
				Services(namespace).
				List(opts)
			return
		}
	}
	if k.del == nil {
		k.del = func(cli *kubernetes.Clientset, name,
			namespace string, opts *metav1.DeleteOptions) (err error) {
			deletePropagation := metav1.DeletePropagationForeground
			opts.PropagationPolicy = &deletePropagation
			err = cli.CoreV1().
				Services(namespace).
				Delete(name, opts)
			return
		}
	}
}

// WithClientset sets the kubernetes client against the kubeclient instance
func WithClientset(c *kubernetes.Clientset) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.clientset = c
	}
}

// WithNamespace set namespace in kubeclient object
func WithNamespace(namespace string) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.namespace = namespace
	}
}

// KubeClient returns a new instance of kubeclient meant for service,
// caller can configure it with different kubeclientBuildOption
func KubeClient(opts ...KubeclientBuildOption) *Kubeclient {
	k := &Kubeclient{}
	for _, o := range opts {
		o(k)
	}
	k.withDefaults()
	return k
}

// getClientOrCached returns either a new
// instance of kubernetes client or its cached copy
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

// Get returns service object for given name
func (k *Kubeclient) Get(name string, opts metav1.GetOptions) (*corev1.Service, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.get(cli, name, k.namespace, opts)
}

// GetRaw returns service object for given name in byte format
func (k *Kubeclient) GetRaw(name string, opts metav1.GetOptions) ([]byte, error) {
	svc, err := k.Get(name, opts)
	if err != nil {
		return nil, err
	}
	return json.Marshal(svc)
}

// List returns list of services
func (k *Kubeclient) List(opts metav1.ListOptions) (*corev1.ServiceList, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.list(cli, k.namespace, opts)
}

// ListRaw returns list of services in byte format
func (k *Kubeclient) ListRaw(opts metav1.ListOptions) ([]byte, error) {
	svcList, err := k.List(opts)
	if err != nil {
		return nil, err
	}
	return json.Marshal(svcList)
}

// Delete returns service object for given name
func (k *Kubeclient) Delete(name string, options *metav1.DeleteOptions) error {
	cli, err := k.getClientOrCached()
	if err != nil {
		return err
	}
	return k.del(cli, name, k.namespace, options)
}
