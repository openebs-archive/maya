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
	"context"
	"errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
	client "github.com/openebs/maya/pkg/kubernetes/client/v1alpha1"
	types "k8s.io/apimachinery/pkg/types"
)

// getClientsetFn is a typed function that
// abstracts fetching of internal clientset
type getClientsetFn func() (clientset *clientset.Clientset, err error)

// getClientsetFromPathFn is a typed function that
// abstracts fetching of clientset from kubeConfigPath
type getClientsetForPathFn func(
	kubeConfigPath string,
) (*clientset.Clientset, error)

// getFn is a typed function that abstracts get of
// cstorvolume replica instances
type getFn func(
	cli *clientset.Clientset,
	name, namespace string,
	opts metav1.GetOptions,
) (*apis.CStorVolumeReplica, error)

// listFn is a typed function that abstracts
// listing of cstor volume replica instances
type listFn func(
	cli *clientset.Clientset,
	namespace string,
	opts metav1.ListOptions,
) (*apis.CStorVolumeReplicaList, error)

// delFn is a typed function that abstracts delete of
// cstorvolume replica instances
type delFn func(
	cli *clientset.Clientset,
	name, namespace string,
	opts *metav1.DeleteOptions,
) error

// patchFn is a typed function that abstracts patch of
// cstorvolume replica instances
type patchFn func(
	cli *clientset.Clientset,
	name, namespace string,
	pt types.PatchType,
	data []byte,
) (*apis.CStorVolumeReplica, error)

// createFn is a typed function that abstracts create of
// cstorvolume replica instances
type createFn func(
	cli *clientset.Clientset,
	namespace string,
	volr *apis.CStorVolumeReplica,
) (*apis.CStorVolumeReplica, error)

// updateFn is a typed function that abstracts update of
// cstorvolume replica instances
type updateFn func(
	cli *clientset.Clientset,
	namespace string,
	volr *apis.CStorVolumeReplica,
) (*apis.CStorVolumeReplica, error)

// Kubeclient enables kubernetes API operations
// on cstor volume replica instance
type Kubeclient struct {
	// clientset refers to cstor volume replica's
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
	patch               patchFn
	update              updateFn
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

// defaultPatch is the default implementation to
// patch cstorvolume replica instance
func defaultPatch(
	cli *clientset.Clientset,
	name, namespace string,
	pt types.PatchType,
	data []byte,
) (*apis.CStorVolumeReplica, error) {
	return cli.OpenebsV1alpha1().
		CStorVolumeReplicas(namespace).
		Patch(context.TODO(), name, pt, data, metav1.PatchOptions{})
}

// defaultGetClientsetForPath is the default implementation to
// get kubernetes clientset instance based on the given
// kubeconfig path
func defaultGetClientsetForPath(
	kubeConfigPath string,
) (clients *clientset.Clientset, err error) {
	config, err := client.GetConfig(
		client.New(client.WithKubeConfigPath(kubeConfigPath)),
	)
	if err != nil {
		return nil, err
	}
	return clientset.NewForConfig(config)
}

// defaultGet is the default implementation to get a
// cstorvolume replica instance in kubernetes cluster
func defaultGet(
	cli *clientset.Clientset,
	name, namespace string,
	opts metav1.GetOptions,
) (*apis.CStorVolumeReplica, error) {
	return cli.OpenebsV1alpha1().
		CStorVolumeReplicas(namespace).
		Get(context.TODO(), name, opts)
}

// defaultGet is the default implementation to list
// cstorvolume replica instances in kubernetes cluster
func defaultList(
	cli *clientset.Clientset,
	namespace string,
	opts metav1.ListOptions,
) (*apis.CStorVolumeReplicaList, error) {
	return cli.OpenebsV1alpha1().
		CStorVolumeReplicas(namespace).
		List(context.TODO(), opts)
}

// defaultGet is the default implementation to delete a
// cstorvolume replica instance in kubernetes cluster
func defaultDel(
	cli *clientset.Clientset,
	name, namespace string,
	opts *metav1.DeleteOptions,
) error {
	// The object exists in the key-value store until the garbage collector
	// deletes all the dependents whose ownerReference.blockOwnerDeletion=true
	// from the key-value store.  API sever will put the "foregroundDeletion"
	// finalizer on the object, and sets its deletionTimestamp.  This policy is
	// cascading, i.e., the dependents will be deleted with Foreground.
	deletePropagation := metav1.DeletePropagationForeground
	opts.PropagationPolicy = &deletePropagation
	err := cli.OpenebsV1alpha1().
		CStorVolumeReplicas(namespace).
		Delete(context.TODO(), name, *opts)
	return err
}

// defaultGet is the default implementation to create a
// cstorvolume replica instance in kubernetes cluster
func defaultCreate(
	cli *clientset.Clientset,
	namespace string,
	volr *apis.CStorVolumeReplica,
) (*apis.CStorVolumeReplica, error) {
	return cli.OpenebsV1alpha1().
		CStorVolumeReplicas(namespace).
		Create(context.TODO(), volr, metav1.CreateOptions{})
}

// defaultUpdate is the default implementation to update a
// cstorvolume replica instance in kubernetes cluster
func defaultUpdate(
	cli *clientset.Clientset,
	namespace string,
	volr *apis.CStorVolumeReplica,
) (*apis.CStorVolumeReplica, error) {
	return cli.OpenebsV1alpha1().
		CStorVolumeReplicas(namespace).
		Update(context.TODO(), volr, metav1.UpdateOptions{})
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

	if k.patch == nil {
		k.patch = defaultPatch
	}

	if k.update == nil {
		k.update = defaultUpdate
	}

}

// WithKubeClient sets the kubernetes client against
// the kubeclient instance
func WithKubeClient(c *clientset.Clientset) KubeclientBuildOption {
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

// WithNamespace sets the kubernetes client against
// the provided namespace
func (k *Kubeclient) WithNamespace(namespace string) *Kubeclient {
	k.namespace = namespace
	return k
}

// WithKubeConfigPath sets the kubernetes client against
// the provided path
func WithKubeConfigPath(path string) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.kubeConfigPath = path
	}
}

// NewKubeclient returns a new instance of kubeclient meant for
// cstor volume replica operations
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
		return nil, err
	}
	k.clientset = c
	return k.clientset, nil
}

// List returns a list of cstor volume replica
// instances present in kubernetes cluster
func (k *Kubeclient) List(
	opts metav1.ListOptions,
) (*apis.CStorVolumeReplicaList, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.list(cli, k.namespace, opts)
}

// Get returns cstorvolumereplica object for given name
func (k *Kubeclient) Get(
	name string,
	opts metav1.GetOptions,
) (*apis.CStorVolumeReplica, error) {
	if len(name) == 0 {
		return nil,
			errors.New("failed to get cstorvolume: name can't be empty")
	}
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.get(cli, name, k.namespace, opts)
}

// Delete delete the cstorvolume replica resource
func (k *Kubeclient) Delete(name string) error {
	cli, err := k.getClientOrCached()
	if err != nil {
		return err
	}
	return k.del(cli, name, k.namespace, &metav1.DeleteOptions{})
}

// Patch patches the cstorvolume replica resource
func (k *Kubeclient) Patch(
	name, namespace string,
	pt types.PatchType,
	data []byte,
) (*apis.CStorVolumeReplica, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.patch(cli, name, k.namespace, pt, data)
}

// Create creates cstorvolumereplica resource for given object
func (k *Kubeclient) Create(
	volr *apis.CStorVolumeReplica,
) (*apis.CStorVolumeReplica, error) {
	if volr == nil {
		return nil,
			errors.New("failed to create cvr: nil cvr object")
	}
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.create(cli, k.namespace, volr)
}

// Update updates cstorvolumereplica resource for given object
func (k *Kubeclient) Update(
	volr *apis.CStorVolumeReplica,
) (*apis.CStorVolumeReplica, error) {
	if volr == nil {
		return nil,
			errors.New("failed to create cvr: nil cvr object")
	}
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.update(cli, k.namespace, volr)
}
