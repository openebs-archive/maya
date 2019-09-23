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
	"fmt"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"

	clientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
	client "github.com/openebs/maya/pkg/kubernetes/client/v1alpha1"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
)

// getClientsetFn is a typed function that
// abstracts fetching of internal clientset
type getClientsetFn func() (clientset *clientset.Clientset, err error)

// getClientsetFromPathFn is a typed function that
// abstracts fetching of clientset from kubeConfigPath
type getClientsetForPathFn func(
	kubeConfigPath string) (clientset *clientset.Clientset, err error)

// createFn is a typed function that abstracts
// creating csi volume instance
type createFn func(
	cli *clientset.Clientset,
	upgradeResultObj *apis.CStorVolumeClaim,
	namespace string) (*apis.CStorVolumeClaim, error)

// getFn is a typed function that abstracts
// fetching a csi volume instance
type getFn func(cli *clientset.Clientset, name, namespace string,
	opts metav1.GetOptions) (*apis.CStorVolumeClaim, error)

// listFn is a typed function that abstracts
// listing of csi volume instances
type listFn func(
	cli *clientset.Clientset,
	namespace string,
	opts metav1.ListOptions) (*apis.CStorVolumeClaimList, error)

// delFn is a typed function that abstracts
// deleting a csi volume instance
type delFn func(
	cli *clientset.Clientset,
	name,
	namespace string,
	opts *metav1.DeleteOptions) error

// updateFn is a typed function that abstracts
// updating csi volume instance
type updateFn func(
	cli *clientset.Clientset,
	cvc *apis.CStorVolumeClaim,
	namespace string) (*apis.CStorVolumeClaim, error)

// patchFn is a typed function that abstracts
// patching csi volume instance
type patchFn func(
	cli *clientset.Clientset,
	oldCVC *apis.CStorVolumeClaim,
	newCVC *apis.CStorVolumeClaim,
	subresources ...string,
) (*apis.CStorVolumeClaim, error)

// Kubeclient enables kubernetes API operations
// on csi volume instance
type Kubeclient struct {
	// clientset refers to csi volume's
	// clientset that will be responsible to
	// make kubernetes API calls
	clientset *clientset.Clientset

	kubeConfigPath string

	// namespace holds the namespace on which
	// kubeclient has to operate
	namespace string

	// functions useful during mocking
	getClientset        getClientsetFn
	getClientsetForPath getClientsetForPathFn
	get                 getFn
	list                listFn
	del                 delFn
	create              createFn
	update              updateFn
	patch               patchFn
}

// KubeclientBuildOption defines the abstraction
// to build a kubeclient instance
type KubeclientBuildOption func(*Kubeclient)

// defaultGetClientset is the default implementation to
// get kubernetes clientset instance
func defaultGetClientset() (clients *clientset.Clientset, err error) {

	config, err := client.GetConfig(client.New())
	if err != nil {
		return nil, err
	}

	return clientset.NewForConfig(config)

}

// defaultGetClientsetForPath is the default implementation to
// get kubernetes clientset instance based on the given
// kubeconfig path
func defaultGetClientsetForPath(
	kubeConfigPath string,
) (clients *clientset.Clientset, err error) {
	config, err := client.GetConfig(
		client.New(client.WithKubeConfigPath(kubeConfigPath)))
	if err != nil {
		return nil, err
	}

	return clientset.NewForConfig(config)
}

// defaultGet is the default implementation to get
// a cstorvolumeclaim instance in kubernetes cluster
func defaultGet(
	cli *clientset.Clientset,
	name, namespace string,
	opts metav1.GetOptions,
) (*apis.CStorVolumeClaim, error) {
	return cli.OpenebsV1alpha1().
		CStorVolumeClaims(namespace).
		Get(name, opts)
}

// defaultList is the default implementation to list
// CstorVolumeClaim instances in kubernetes cluster
func defaultList(
	cli *clientset.Clientset,
	namespace string,
	opts metav1.ListOptions,
) (*apis.CStorVolumeClaimList, error) {
	return cli.OpenebsV1alpha1().
		CStorVolumeClaims(namespace).
		List(opts)
}

// defaultCreate is the default implementation to delete
// a cstorvolumeclaim instance in kubernetes cluster
func defaultDel(
	cli *clientset.Clientset,
	name, namespace string,
	opts *metav1.DeleteOptions,
) error {
	deletePropagation := metav1.DeletePropagationForeground
	opts.PropagationPolicy = &deletePropagation
	err := cli.OpenebsV1alpha1().
		CStorVolumeClaims(namespace).
		Delete(name, opts)
	return err
}

// defaultCreate is the default implementation to create
// a cstorvolumeclaim instance in kubernetes cluster
func defaultCreate(
	cli *clientset.Clientset,
	cvc *apis.CStorVolumeClaim,
	namespace string,
) (*apis.CStorVolumeClaim, error) {
	return cli.OpenebsV1alpha1().
		CStorVolumeClaims(namespace).
		Create(cvc)
}

// defaultPatch is the default implementation to patch
// a cstorvolumeclaim instance in kubernetes cluster
func defaultPatch(
	cli *clientset.Clientset,
	oldCVC *apis.CStorVolumeClaim,
	newCVC *apis.CStorVolumeClaim,
	subresources ...string,
) (*apis.CStorVolumeClaim, error) {
	patchBytes, err := getPatchData(oldCVC, newCVC)
	if err != nil {
		return nil,
			fmt.Errorf(
				"can't patch CVC %s as generate path data failed: %v",
				CVCKey(oldCVC), err,
			)
	}

	updatedCVC, updateErr := cli.OpenebsV1alpha1().
		CStorVolumeClaims(oldCVC.Namespace).
		Patch(
			oldCVC.Name, types.MergePatchType,
			patchBytes, subresources...,
		)
	if updateErr != nil {
		return nil,
			fmt.Errorf("can't patch status of  CVC %s with %v",
				CVCKey(oldCVC), updateErr,
			)
	}
	return updatedCVC, nil
}

// defaultUpdate is the default implementation to update
// a cstorvolumeclaim instance in kubernetes cluster
func defaultUpdate(
	cli *clientset.Clientset,
	cvc *apis.CStorVolumeClaim,
	namespace string,
) (*apis.CStorVolumeClaim, error) {
	return cli.OpenebsV1alpha1().
		CStorVolumeClaims(namespace).
		Update(cvc)
}

// withDefaults sets the default options
// of kubeclient instance
func (k *Kubeclient) withDefaults() {
	if k.getClientset == nil {
		k.getClientset = defaultGetClientset
	}
	if k.getClientsetForPath == nil {
		k.getClientsetForPath = defaultGetClientsetForPath
	}
	if k.get == nil {
		k.get = defaultGet
	}
	if k.list == nil {
		k.list = defaultList
	}
	if k.del == nil {
		k.del = defaultDel
	}
	if k.create == nil {
		k.create = defaultCreate
	}
	if k.update == nil {
		k.update = defaultUpdate
	}
	if k.patch == nil {
		k.patch = defaultPatch
	}
}

// WithClientSet sets the kubernetes client against
// the kubeclient instance
func WithClientSet(c *clientset.Clientset) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.clientset = c
	}
}

// WithNamespace sets the kubernetes client against
// the provided namespace
func WithNamespace(namespace string) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.namespace = namespace
	}
}

// WithNamespace sets the provided namespace
// against this Kubeclient instance
func (k *Kubeclient) WithNamespace(namespace string) *Kubeclient {
	k.namespace = namespace
	return k
}

// WithKubeConfigPath sets the kubernetes client
// against the provided path
func WithKubeConfigPath(path string) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.kubeConfigPath = path
	}
}

// NewKubeclient returns a new instance of
// kubeclient meant for csi volume operations
func NewKubeclient(opts ...KubeclientBuildOption) *Kubeclient {
	k := &Kubeclient{}
	for _, o := range opts {
		o(k)
	}

	k.withDefaults()
	return k
}

func (k *Kubeclient) getClientsetForPathOrDirect() (
	*clientset.Clientset, error) {
	if k.kubeConfigPath != "" {
		return k.getClientsetForPath(k.kubeConfigPath)
	}

	return k.getClientset()
}

// getClientOrCached returns either a new instance
// of kubernetes client or its cached copy
func (k *Kubeclient) getClientOrCached() (*clientset.Clientset, error) {
	if k.clientset != nil {
		return k.clientset, nil
	}

	c, err := k.getClientsetForPathOrDirect()
	if err != nil {
		return nil,
			errors.Wrapf(
				err,
				"failed to get clientset",
			)
	}

	k.clientset = c
	return k.clientset, nil
}

// Create creates a cstorvolumeclaim instance
// in kubernetes cluster
func (k *Kubeclient) Create(
	cvc *apis.CStorVolumeClaim) (*apis.CStorVolumeClaim, error) {
	if cvc == nil {
		return nil,
			errors.New(
				"failed to create cstovolumeclaim: nil cvc object",
			)
	}
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to create cvc {%s} in namespace {%s}",
			cvc.Name,
			k.namespace,
		)
	}

	return k.create(cli, cvc, k.namespace)
}

// Get returns cstorvolumeclaim object for given name
func (k *Kubeclient) Get(
	name string,
	opts metav1.GetOptions,
) (*apis.CStorVolumeClaim, error) {
	if name == "" {
		return nil,
			errors.New(
				"failed to get cstorvolumeclaim: missing cvc name",
			)
	}

	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to get cstorvolumeclaim {%s} in namespace {%s}",
			name,
			k.namespace,
		)
	}
	return k.get(cli, name, k.namespace, opts)
}

// GetRaw returns cstorvolumeclaim instance
// in bytes
func (k *Kubeclient) GetRaw(
	name string,
	opts metav1.GetOptions,
) ([]byte, error) {
	if name == "" {
		return nil, errors.New(
			"failed to get raw cstorvolumeclaim: missing cvc name",
		)
	}
	cvc, err := k.Get(name, opts)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to get cstorvolumeclaim {%s} in namespace {%s}",
			name,
			k.namespace,
		)
	}

	return json.Marshal(cvc)
}

// List returns a list of cstorvolumeclaim
// instances present in kubernetes cluster
func (k *Kubeclient) List(
	opts metav1.ListOptions,
) (*apis.CStorVolumeClaimList, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to list cstorvolumeclaims in namespace {%s}",
			k.namespace,
		)
	}

	return k.list(cli, k.namespace, opts)
}

// Delete deletes the cstorvolumeclaim from
// kubernetes
func (k *Kubeclient) Delete(name string) error {
	if name == "" {
		return errors.New(
			"failed to delete cstorvolumeclaim: missing cvc name",
		)
	}
	cli, err := k.getClientOrCached()
	if err != nil {
		return errors.Wrapf(
			err,
			"failed to delete cstorvolumeclaim {%s} in namespace {%s}",
			name,
			k.namespace,
		)
	}

	return k.del(cli, name, k.namespace, &metav1.DeleteOptions{})
}

// Update updates this cstorvolumeclaim instance
// against kubernetes cluster
func (k *Kubeclient) Update(
	cvc *apis.CStorVolumeClaim,
) (*apis.CStorVolumeClaim, error) {
	if cvc == nil {
		return nil,
			errors.New(
				"failed to update cstorvolumeclaim: nil cvc object",
			)
	}

	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to update cstorvolumeclaim {%s} in namespace {%s}",
			cvc.Name,
			cvc.Namespace,
		)
	}

	return k.update(cli, cvc, k.namespace)
}

// Patch patches this cstorvolumeclaim instance
// against kubernetes cluster
func (k *Kubeclient) Patch(
	oldCVC *apis.CStorVolumeClaim,
	newCVC *apis.CStorVolumeClaim,
	subresources ...string,
) (*apis.CStorVolumeClaim, error) {
	if oldCVC == nil || newCVC == nil {
		return nil,
			errors.New(
				"failed to update cstorvolumeclaim: nil cvc object",
			)
	}

	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to update cstorvolumeclaim {%s} in namespace {%s}",
			newCVC.Name,
			newCVC.Namespace,
		)
	}

	return k.patch(cli, oldCVC, newCVC, subresources...)
}
