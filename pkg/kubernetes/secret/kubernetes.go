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

package secret

import (
	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	client "github.com/openebs/maya/pkg/kubernetes/client/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// getClientsetFn is a typed function that abstracts
// fetching an instance of kubernetes clientset
type getClientsetFn func() (clientset *kubernetes.Clientset, err error)

// getClientsetFromPathFn is a typed function that
// abstracts fetching of clientset from kubeConfigPath
type getClientsetForPathFn func(kubeConfigPath string) (clientset *kubernetes.Clientset, err error)

// getFn is a typed function that abstracts
// fetching an instance of secret
type getFn func(cli *kubernetes.Clientset, namespace, name string, opts metav1.GetOptions) (*corev1.Secret, error)

// createFn is a typed function that abstracts
// to create secret
type createFn func(cli *kubernetes.Clientset, namespace string, secret *corev1.Secret) (*corev1.Secret, error)

// deleteFn is a typed function that abstracts
// to delete secret
type deleteFn func(cli *kubernetes.Clientset, namespace, name string, opts *metav1.DeleteOptions) error

// Kubeclient enables kubernetes API operations on storageclass instance
type Kubeclient struct {
	// clientset refers to storageclass clientset
	// that will be responsible to
	// make kubernetes API calls
	clientset *kubernetes.Clientset

	// namespace holds the namespace on which
	// kubeclient has to operate
	namespace string

	// kubeconfig path to get kubernetes clientset
	kubeConfigPath string

	// functions useful during mocking
	getClientset        getClientsetFn
	getClientsetForPath getClientsetForPathFn
	get                 getFn
	create              createFn
	del                 deleteFn
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
	if k.get == nil {
		k.get = func(cli *kubernetes.Clientset, namespace, name string, opts metav1.GetOptions) (*corev1.Secret, error) {
			return cli.CoreV1().Secrets(namespace).Get(name, opts)
		}
	}
	if k.create == nil {
		k.create = func(cli *kubernetes.Clientset, namespace string, secret *corev1.Secret) (*corev1.Secret, error) {
			return cli.CoreV1().Secrets(namespace).Create(secret)
		}
	}
	if k.del == nil {
		k.del = func(cli *kubernetes.Clientset, namespace, name string, opts *metav1.DeleteOptions) error {
			return cli.CoreV1().Secrets(namespace).Delete(name, opts)
		}
	}
}

// NewKubeClient returns a new instance of kubeclient meant for storageclass
func NewKubeClient(opts ...KubeClientBuildOption) *Kubeclient {
	k := &Kubeclient{}
	for _, o := range opts {
		o(k)
	}
	k.withDefaults()
	return k
}

// WithClientSet sets the kubernetes client against
// the kubeclient instance
func WithClientSet(c *kubernetes.Clientset) KubeClientBuildOption {
	return func(k *Kubeclient) {
		k.clientset = c
	}
}

// WithKubeConfigPath sets the kubeConfig path
// against client instance
func WithKubeConfigPath(path string) KubeClientBuildOption {
	return func(k *Kubeclient) {
		k.kubeConfigPath = path
	}
}

func (k *Kubeclient) getClientsetForPathOrDirect() (*kubernetes.Clientset, error) {
	if k.kubeConfigPath != "" {
		return k.getClientsetForPath(k.kubeConfigPath)
	}
	return k.getClientset()
}

// getClientsetOrCached returns either a new
// instance of kubernetes clientset or its
// cached copy cached copy
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

// Get return a secret instance present in kubernetes cluster
func (k *Kubeclient) Get(name string, opts metav1.GetOptions) (*corev1.Secret, error) {
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get secret {%s}", name)
	}
	return k.get(cli, k.namespace, name, opts)
}

// Create creates and returns a secret instance
func (k *Kubeclient) Create(secret *corev1.Secret) (*corev1.Secret, error) {
	if secret == nil {
		return nil, errors.New("failed to create secret: nil secret object")
	}
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create secret")
	}
	return k.create(cli, k.namespace, secret)
}

// Delete deletes the secret if present in kubernetes cluster
func (k *Kubeclient) Delete(name string, opts *metav1.DeleteOptions) error {
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return errors.Wrapf(err, "failed to delete secret: {%s}", name)
	}
	return k.del(cli, k.namespace, name, opts)
}
