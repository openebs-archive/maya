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
	"strings"

	stringer "github.com/openebs/maya/pkg/apis/stringer/v1alpha1"
	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	client "github.com/openebs/maya/pkg/client/k8s/v1alpha1"
	"k8s.io/client-go/kubernetes"
)

// getClientsetFn is a typed function that
// abstracts fetching of clientset
type getClientsetFn func() (clientset *kubernetes.Clientset, err error)

// getpvcFn is a typed function that
// abstracts fetching of pvc
type getFn func(cli *kubernetes.Clientset, name string, namespace string, opts metav1.GetOptions) (*corev1.PersistentVolumeClaim, error)

// listFn is a typed function that abstracts
// listing of pvcs
type listFn func(cli *kubernetes.Clientset, namespace string, opts metav1.ListOptions) (*corev1.PersistentVolumeClaimList, error)

// deleteFn is a typed function that abstracts
// deletion of pvcs
type deleteFn func(cli *kubernetes.Clientset, namespace string, name string, deleteOpts *metav1.DeleteOptions) error

// deleteFn is a typed function that abstracts
// deletion of pvc's collection
type deleteCollectionFn func(cli *kubernetes.Clientset, namespace string, listOpts metav1.ListOptions, deleteOpts *metav1.DeleteOptions) error

// createFn is a typed function that abstracts
// creation of pvc
type createFn func(cli *kubernetes.Clientset, namespace string, pvc *corev1.PersistentVolumeClaim) (*corev1.PersistentVolumeClaim, error)

// Kubeclient enables kubernetes API operations
// on pvc instance
type Kubeclient struct {
	// clientset refers to pvc clientset
	// that will be responsible to
	// make kubernetes API calls
	clientset *kubernetes.Clientset

	// namespace holds the namespace on which
	// kubeclient has to operate
	namespace string

	// functions useful during mocking
	getClientset  getClientsetFn
	list          listFn
	get           getFn
	create        createFn
	del           deleteFn
	delCollection deleteCollectionFn
}

// KubeclientBuildOption abstracts creating an
// instance of kubeclient
type KubeclientBuildOption func(*Kubeclient)

// withDefaults sets the default options
// of kubeclient instance
func (k *Kubeclient) withDefaults() {
	if k.getClientset == nil {
		k.getClientset = func() (clients *kubernetes.Clientset, err error) {
			config, err := client.Config().Get()
			if err != nil {
				return nil, err
			}
			return kubernetes.NewForConfig(config)
		}
	}
	if k.get == nil {
		k.get = func(cli *kubernetes.Clientset, name string, namespace string, opts metav1.GetOptions) (*corev1.PersistentVolumeClaim, error) {
			return cli.CoreV1().PersistentVolumeClaims(namespace).Get(name, opts)
		}
	}
	if k.list == nil {
		k.list = func(cli *kubernetes.Clientset, namespace string, opts metav1.ListOptions) (*corev1.PersistentVolumeClaimList, error) {
			return cli.CoreV1().PersistentVolumeClaims(namespace).List(opts)
		}
	}
	if k.del == nil {
		k.del = func(cli *kubernetes.Clientset, namespace string, name string, deleteOpts *metav1.DeleteOptions) error {
			return cli.CoreV1().PersistentVolumeClaims(namespace).Delete(name, deleteOpts)
		}
	}
	if k.delCollection == nil {
		k.delCollection = func(cli *kubernetes.Clientset, namespace string, listOpts metav1.ListOptions, deleteOpts *metav1.DeleteOptions) error {
			return cli.CoreV1().PersistentVolumeClaims(namespace).DeleteCollection(deleteOpts, listOpts)
		}
	}
	if k.create == nil {
		k.create = func(cli *kubernetes.Clientset, namespace string, pvc *corev1.PersistentVolumeClaim) (*corev1.PersistentVolumeClaim, error) {
			return cli.CoreV1().PersistentVolumeClaims(namespace).Create(pvc)
		}
	}
}

// WithNamespace sets the kubernetes client against
// the provided namespace
func WithNamespace(namespace string) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.namespace = namespace
	}
}

// WithClientSet sets the kubernetes client against
// the kubeclient instance
func WithClientSet(c *kubernetes.Clientset) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.clientset = c
	}
}

// KubeClient returns a new instance of kubeclient meant for
// cstor volume replica operations
func NewKubeClient(opts ...KubeclientBuildOption) *Kubeclient {
	k := &Kubeclient{}
	for _, o := range opts {
		o(k)
	}
	k.withDefaults()
	return k
}

// getClientSetOrCached returns either a new instance
// of kubernetes client or its cached copy
func (k *Kubeclient) getClientSetOrCached() (*kubernetes.Clientset, error) {
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

// Get returns a pvc resource
// instances present in kubernetes cluster
func (k *Kubeclient) Get(name string, opts metav1.GetOptions) (*corev1.PersistentVolumeClaim, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("failed to get pvc: missing pvc name")
	}
	cli, err := k.getClientSetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get pvc {%s}", name)
	}
	return k.get(cli, name, k.namespace, opts)
}

// List returns a list of pvc
// instances present in kubernetes cluster
func (k *Kubeclient) List(opts metav1.ListOptions) (*corev1.PersistentVolumeClaimList, error) {
	cli, err := k.getClientSetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list pvc listoptions: '%v'", opts)
	}
	return k.list(cli, k.namespace, opts)
}

// Delete deletes a pvc instance from the
// kubecrnetes cluster
func (k *Kubeclient) Delete(name string, deleteOpts *metav1.DeleteOptions) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("failed to delete pvc: missing pvc name")
	}
	cli, err := k.getClientSetOrCached()
	if err != nil {
		return errors.Wrapf(err, "failed to delete pvc {%s}", name)
	}
	return k.del(cli, k.namespace, name, deleteOpts)
}

// Create creates a pvc in specified namespace in kubernetes cluster
func (k *Kubeclient) Create(pvc *corev1.PersistentVolumeClaim) (*corev1.PersistentVolumeClaim, error) {
	cli, err := k.getClientSetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create pvc: %s", stringer.Yaml("persistent volume claim", pvc))
	}
	return k.create(cli, k.namespace, pvc)
}

// DeleteCollection deletes a collection of pvc objects.
func (k *Kubeclient) DeleteCollection(listOpts metav1.ListOptions, deleteOpts *metav1.DeleteOptions) error {
	cli, err := k.getClientSetOrCached()
	if err != nil {
		return errors.Wrapf(err, "failed to delete the collection of pvcs")
	}
	return k.delCollection(cli, k.namespace, listOpts, deleteOpts)
}
