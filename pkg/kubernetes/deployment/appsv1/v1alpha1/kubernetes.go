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
	client "github.com/openebs/maya/pkg/kubernetes/client/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// getClientsetFn is a typed function
// that abstracts fetching of internal clientset
type getClientsetFn func() (clientset *kubernetes.Clientset, err error)

// getFn is a typed function that abstracts get of deployment instances
type getFn func(cli *kubernetes.Clientset, name, namespace string,
	opts *metav1.GetOptions) (*appsv1.Deployment, error)

// rolloutStatusFn is a typed function that abstracts
// rollout status of deployment instances
type rolloutStatusFn func(d *appsv1.Deployment) (*rolloutOutput, error)

// rolloutStatusfFn is a typed function that abstracts
// rollout status of deployment instances
type rolloutStatusfFn func(d *appsv1.Deployment) ([]byte, error)

// kubeclient enables kubernetes API operations on deployment instance
type kubeclient struct {
	// clientset refers to kubernetes clientset. It is responsible to
	// make kubernetes API calls for crud op
	clientset *kubernetes.Clientset
	namespace string

	// functions useful during mocking
	getClientset   getClientsetFn
	get            getFn
	rolloutStatus  rolloutStatusFn
	rolloutStatusf rolloutStatusfFn
}

// kubeclientBuildOption defines the abstraction to build a kubeclient instance
type kubeclientBuildOption func(*kubeclient)

// withDefaults sets the default options of kubeclient instance
func (k *kubeclient) withDefaults() {
	if k.getClientset == nil {
		k.getClientset = func() (
			clients *kubernetes.Clientset, err error) {
			config, err := client.GetConfig(client.New())
			if err != nil {
				return nil, err
			}
			return kubernetes.NewForConfig(config)
		}
	}

	if k.get == nil {
		k.get = func(cli *kubernetes.Clientset, name,
			namespace string, opts *metav1.GetOptions) (
			d *appsv1.Deployment, err error) {
			d, err = cli.AppsV1().
				Deployments(namespace).
				Get(name, *opts)
			return
		}
	}

	if k.rolloutStatus == nil {
		k.rolloutStatus = func(d *appsv1.Deployment) (
			*rolloutOutput, error) {
			status, err := New(
				WithAPIObject(d)).
				RolloutStatus()
			if err != nil {
				return nil, err
			}
			return NewRollout(
				withOutputObject(status)).
				AsRolloutOutput()
		}
	}

	if k.rolloutStatusf == nil {
		k.rolloutStatusf = func(d *appsv1.Deployment) (
			[]byte, error) {
			status, err := New(
				WithAPIObject(d)).
				RolloutStatus()
			if err != nil {
				return nil, err
			}
			return NewRollout(
				withOutputObject(status)).
				Raw()
		}
	}

}

// WithClientset sets the kubernetes client against the kubeclient instance
func WithClientset(c *kubernetes.Clientset) kubeclientBuildOption {
	return func(k *kubeclient) {
		k.clientset = c
	}
}

// WithNamespace set namespace in kubeclient object
func WithNamespace(namespace string) kubeclientBuildOption {
	return func(k *kubeclient) {
		k.namespace = namespace
	}
}

// KubeClient returns a new instance of kubeclient meant for deployment.
// caller can configure it with different kubeclientBuildOption
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
func (k *kubeclient) getClientOrCached() (*kubernetes.Clientset, error) {
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

// Get returns deployment object for given name
func (k *kubeclient) Get(name string) (*appsv1.Deployment, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.get(cli, name, k.namespace, &metav1.GetOptions{})
}

// RolloutStatusf returns deployment's rollout status for given name
// in raw bytes
func (k *kubeclient) RolloutStatusf(name string) (op []byte, err error) {
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
func (k *kubeclient) RolloutStatus(name string) (*rolloutOutput, error) {
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
