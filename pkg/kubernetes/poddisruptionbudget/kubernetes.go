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
	"context"
	"strings"
	"sync"

	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	client "github.com/openebs/maya/pkg/kubernetes/client/v1alpha1"
	policy "k8s.io/api/policy/v1beta1"

	"k8s.io/client-go/kubernetes"
)

// make kubernetes clientset as singleton
var (
	kubeClientInst *kubernetes.Clientset
	once           sync.Once
)

// delFunc is a typed function that abstracts deleting poddisruptionbudget
type delFunc func(cs *kubernetes.Clientset, name, namespace string, opts *metav1.DeleteOptions) error

// getFunc is a typed function that abstracts
// getting poddisruptionbudget instances
type getFunc func(cs *kubernetes.Clientset, name, namespace string, opts metav1.GetOptions) (*policy.PodDisruptionBudget, error)

// createFn is a typed function that abstracts
// creating poddisruptionbudget
type createFunc func(cli *kubernetes.Clientset, namespace string,
	pdb *policy.PodDisruptionBudget) (
	*policy.PodDisruptionBudget,
	error,
)

// listFunc is a typed function that abstracts
// listing poddisruptionbudget instances
type listFunc func(cs *kubernetes.Clientset, namespace string, opts metav1.ListOptions) (*policy.PodDisruptionBudgetList, error)

// getClientsetFunc is a typed function that
// abstracts fetching internal clientset
type getClientsetFunc func() (cs *kubernetes.Clientset, err error)

// KubeclientBuildOption defines the abstraction
// to build a kubeclient instance
type KubeclientBuildOption func(*Kubeclient)

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
	create       createFunc
	get          getFunc
	del          delFunc
}

// withDefaults sets the default options
// of kubeclient instance
func (k *Kubeclient) withDefaults() {
	if k.getClientset == nil {
		k.getClientset = func() (cs *kubernetes.Clientset, err error) {
			if kubeClientInst != nil {
				return kubeClientInst, nil
			}
			config, err := client.GetConfig(client.New())
			if err != nil {
				return nil, err
			}
			kubeCS, err := kubernetes.NewForConfig(config)
			if err != nil {
				return nil, err
			}
			once.Do(func() {
				kubeClientInst = kubeCS
			})
			return kubeCS, nil
		}
	}
	if k.del == nil {
		k.del = func(cs *kubernetes.Clientset, name, namesapce string, opts *metav1.DeleteOptions) error {
			return cs.PolicyV1beta1().PodDisruptionBudgets(namesapce).Delete(context.TODO(), name, *opts)
		}
	}
	if k.create == nil {
		k.create = func(cs *kubernetes.Clientset,
			namesapce string, pdb *policy.PodDisruptionBudget) (*policy.PodDisruptionBudget, error) {
			return cs.PolicyV1beta1().PodDisruptionBudgets(namesapce).Create(context.TODO(), pdb, metav1.CreateOptions{})
		}
	}
	if k.list == nil {
		k.list = func(cs *kubernetes.Clientset,
			namespace string, opts metav1.ListOptions) (*policy.PodDisruptionBudgetList, error) {
			return cs.PolicyV1beta1().PodDisruptionBudgets(namespace).List(context.TODO(), opts)
		}
	}
	if k.get == nil {
		k.get = func(
			cs *kubernetes.Clientset,
			name, namespace string, opts metav1.GetOptions) (*policy.PodDisruptionBudget, error) {
			return cs.PolicyV1beta1().PodDisruptionBudgets(namespace).Get(context.TODO(), name, opts)
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
func (k *Kubeclient) WithNamespace(namespace string) *Kubeclient {
	k.namespace = namespace
	return k
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

// Delete deletes poddisruptionbudget object for given name in corresponding
// namespace
func (k *Kubeclient) Delete(name string, options *metav1.DeleteOptions) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("failed to delete poddisruptionbudget: missing name")
	}

	cli, err := k.getClientOrCached()
	if err != nil {
		return err
	}
	return k.del(cli, name, k.namespace, options)
}

// List takes label and field selectors, and returns the list of
// PodDisruptionBudget instances that match those selectors.
func (k *Kubeclient) List(opts metav1.ListOptions) (*policy.PodDisruptionBudgetList, error) {
	cs, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.list(cs, k.namespace, opts)
}

// Create creates poddisruptionbudget, and returns the
// corresponding poddisruptionbudget object, and an error if there is any.
func (k *Kubeclient) Create(pdb *policy.PodDisruptionBudget) (*policy.PodDisruptionBudget, error) {
	if pdb == nil {
		return nil, errors.New("failed to create poddisruptionbudget: nil pdb")
	}
	cs, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.create(cs, k.namespace, pdb)
}

// Get takes name of the poddisruptionbudget, and returns the
// corresponding poddisruptionbudget object, and an error if there is any.
func (k *Kubeclient) Get(name string, opts metav1.GetOptions) (*policy.PodDisruptionBudget, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("failed to get PodDisruptionBudget: missing poddisruptionbudget name")
	}
	cs, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.get(cs, name, k.namespace, opts)
}
