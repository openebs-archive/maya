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
	"strings"

	client "github.com/openebs/maya/pkg/kubernetes/client/v1alpha1"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// getClientsetFn is a typed function
// that abstracts fetching of internal clientset
type getClientsetFn func() (clientset *kubernetes.Clientset, err error)

// getClientsetFromPathFn is a typed function that
// abstracts fetching of clientset from kubeConfigPath
type getClientsetForPathFn func(kubeConfigPath string) (clientset *kubernetes.Clientset, err error)

// getFn is a typed function that abstracts get of deployment instances
type getFn func(cli *kubernetes.Clientset, name, namespace string,
	opts *metav1.GetOptions) (*appsv1.Deployment, error)

// createFn is a typed function that abstracts
// creation of deployment
type createFn func(cli *kubernetes.Clientset, namespace string, deploy *appsv1.Deployment) (*appsv1.Deployment, error)

// deleteFn is a typed function that abstracts
// deletion of deployments
type deleteFn func(cli *kubernetes.Clientset, namespace string, name string, opts *metav1.DeleteOptions) error

// rolloutStatusFn is a typed function that abstracts
// rollout status of deployment instances
type rolloutStatusFn func(d *appsv1.Deployment) (*RolloutOutput, error)

// rolloutStatusfFn is a typed function that abstracts
// rollout status of deployment instances
type rolloutStatusfFn func(d *appsv1.Deployment) ([]byte, error)

// defaultGet is default implementation of get function
func defaultGet(cli *kubernetes.Clientset, name,
	namespace string, opts *metav1.GetOptions) (
	d *appsv1.Deployment, err error) {
	d, err = cli.AppsV1().
		Deployments(namespace).
		Get(name, *opts)
	return
}

// defaultCreate is default implementation of create function
func defaultCreate(cli *kubernetes.Clientset,
	namespace string, deploy *appsv1.Deployment) (
	d *appsv1.Deployment, err error) {
	d, err = cli.AppsV1().
		Deployments(namespace).
		Create(deploy)
	return
}

// defaultDel is default implementation of del function
func defaultDel(cli *kubernetes.Clientset, namespace,
	name string, opts *metav1.DeleteOptions) (err error) {
	err = cli.AppsV1().
		Deployments(namespace).
		Delete(name, opts)
	return
}

// defaultRolloutStatus is default implementation of rolloutStatus function
func defaultRolloutStatus(d *appsv1.Deployment) (
	*RolloutOutput, error) {
	b, err := NewBuilderForAPIObject(d).
		Build()
	if err != nil {
		return nil, err
	}
	return b.RolloutStatus()
}

// deafultRolloutStatusf is default implementation of rolloutStatusf function
func deafultRolloutStatusf(d *appsv1.Deployment) (
	[]byte, error) {
	b, err := NewBuilderForAPIObject(d).
		Build()
	if err != nil {
		return nil, err
	}
	return b.RolloutStatusRaw()
}

// Kubeclient enables kubernetes API operations on deployment instance
type Kubeclient struct {
	// clientset refers to kubernetes clientset. It is responsible to
	// make kubernetes API calls for crud op
	clientset *kubernetes.Clientset
	namespace string

	// kubeconfig path to get kubernetes clientset
	kubeConfigPath string

	// functions useful during mocking
	getClientset        getClientsetFn
	getClientsetForPath getClientsetForPathFn
	get                 getFn
	create              createFn
	del                 deleteFn
	rolloutStatus       rolloutStatusFn
	rolloutStatusf      rolloutStatusfFn
}

// KubeclientBuildOption defines the abstraction to build a kubeclient instance
type KubeclientBuildOption func(*Kubeclient)

// withDefaults sets the default options of kubeclient instance
func (k *Kubeclient) withDefaults() {

	if k.getClientset == nil {
		k.getClientset = func() (
			clients *kubernetes.Clientset, err error) {
			return client.New().
				Clientset()
		}
	}
	if k.getClientsetForPath == nil {
		k.getClientsetForPath = func(kubeConfigPath string) (
			clients *kubernetes.Clientset, err error) {
			return client.New(client.WithKubeConfigPath(kubeConfigPath)).
				Clientset()
		}
	}

	if k.get == nil {
		k.get = defaultGet
	}

	if k.create == nil {
		k.create = defaultCreate
	}

	if k.del == nil {
		k.del = defaultDel
	}

	if k.create == nil {
		k.create = func(cli *kubernetes.Clientset,
			namespace string, deploy *appsv1.Deployment) (
			d *appsv1.Deployment, err error) {
			d, err = cli.AppsV1().
				Deployments(namespace).
				Create(deploy)
			return
		}
	}

	if k.del == nil {
		k.del = func(cli *kubernetes.Clientset, namespace,
			name string, opts *metav1.DeleteOptions) (err error) {
			err = cli.AppsV1().
				Deployments(namespace).
				Delete(name, opts)
			return
		}
	}

	if k.rolloutStatus == nil {
		k.rolloutStatus = defaultRolloutStatus
	}

	if k.rolloutStatusf == nil {
		k.rolloutStatusf = deafultRolloutStatusf
	}

}

// WithClientset sets the kubernetes client against the kubeclient instance
func WithClientset(c *kubernetes.Clientset) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.clientset = c
	}
}

// WithNamespace sets the kubernetes client against
// the provided namespace
func (k *Kubeclient) WithNamespace(namespace string) *Kubeclient {
	k.namespace = namespace
	return k
}

// WithKubeConfigPath sets the kubeConfig path
// against client instance
func WithKubeConfigPath(path string) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.kubeConfigPath = path
	}
}

// WithNamespace set namespace in kubeclient object
func WithNamespace(namespace string) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.namespace = namespace
	}
}

// NewKubeClient returns a new instance of kubeclient meant for deployment.
// caller can configure it with different kubeclientBuildOption
func NewKubeClient(opts ...KubeclientBuildOption) *Kubeclient {
	k := &Kubeclient{}
	for _, o := range opts {
		o(k)
	}
	k.withDefaults()
	return k
}

func (k *Kubeclient) getClientsetForPathOrDirect() (*kubernetes.Clientset, error) {
	if k.kubeConfigPath != "" {
		return k.getClientsetForPath(k.kubeConfigPath)
	}
	return k.getClientset()
}

// getClientOrCached returns either a new instance
// of kubernetes client or its cached copy
func (k *Kubeclient) getClientOrCached() (*kubernetes.Clientset, error) {
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

// Get returns deployment object for given name
func (k *Kubeclient) Get(name string) (*appsv1.Deployment, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.get(cli, name, k.namespace, &metav1.GetOptions{})
}

// GetRaw returns deployment object for given name
func (k *Kubeclient) GetRaw(name string) ([]byte, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	d, err := k.get(cli, name, k.namespace, &metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return json.Marshal(d)
}

// Delete deletes a deployment instance from the
// kubernetes cluster
func (k *Kubeclient) Delete(name string, opts *metav1.DeleteOptions) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("failed to delete deployment: missing deployment name")
	}
	cli, err := k.getClientOrCached()
	if err != nil {
		return errors.Wrapf(err, "failed to delete deployment {%s}", name)
	}
	return k.del(cli, k.namespace, name, opts)
}

// Create creates a deployment in specified namespace in kubernetes cluster
func (k *Kubeclient) Create(deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	if deployment == nil {
		return nil, errors.New("failed to create deployment: nil deployment object")
	}
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create deployment {%s} in namespace {%s}", deployment.Name, deployment.Namespace)
	}
	return k.create(cli, k.namespace, deployment)
}

// RolloutStatusf returns deployment's rollout status for given name
// in raw bytes
func (k *Kubeclient) RolloutStatusf(name string) (op []byte, err error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	d, err := k.get(cli, name, k.namespace, &metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return k.rolloutStatusf(d)
}

// RolloutStatus returns deployment's rollout status for given name
func (k *Kubeclient) RolloutStatus(name string) (*RolloutOutput, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	d, err := k.get(cli, name, k.namespace, &metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return k.rolloutStatus(d)
}
