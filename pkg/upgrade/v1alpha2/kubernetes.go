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

package v1alpha2

import (
	"strings"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	errors "github.com/pkg/errors"
	kclient "github.com/openebs/maya/pkg/kubernetes/client/v1alpha1"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/upgrade/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/generated/openebs.io/upgrade/v1alpha1/clientset/internalclientset"
	"k8s.io/apimachinery/pkg/types"
)

// getClientsetFn is a typed function that
// abstracts fetching of internal clientset
type getClientsetFn func() (clientset *clientset.Clientset, err error)

// getClientsetFromPathFn is a typed function that
// abstracts fetching of clientset from kubeConfigPath
type getClientsetForPathFn func(
	kubeConfigPath string,
) (clientset *clientset.Clientset, err error)

// listFn is a typed function that abstracts
// listing of upgrade task
type listFn func(
	cli *clientset.Clientset,
	namespace string,
	opts metav1.ListOptions,
) (*apis.UpgradeTaskList, error)

// getFn is a typed function that
// abstracts fetching of upgrade task
type getFn func(
	cli *clientset.Clientset,
	namespace, name string,
	opts metav1.GetOptions,
) (*apis.UpgradeTask, error)

// createFn is a typed function that abstracts
// creation of upgrade task
type createFn func(
	cli *clientset.Clientset,
	namespace string,
	upgradeTask *apis.UpgradeTask,
) (*apis.UpgradeTask, error)

// deleteFn is a typed function that abstracts
// deletion of upgradeTasks
type deleteFn func(
	cli *clientset.Clientset,
	namespace, name string,
	opts *metav1.DeleteOptions,
) error

// patchFn is a typed function that abstracts
// to patch upgrade task
type patchFn func(
	cli *clientset.Clientset,
	namespace, name string,
	pt types.PatchType,
	data []byte,
	subresources ...string,
) (*apis.UpgradeTask, error)

// updateFn is a typed function that abstracts to update
// upgrade task
type updateFn func(
	cli *clientset.Clientset,
	namespace string,
	upgradeTask *apis.UpgradeTask,
) (*apis.UpgradeTask, error)

// make upgrade task clientset as singleton
var (
	clientsetInstance *clientset.Clientset
	once              sync.Once
)

// Kubeclient enables kubernetes API operations
// on upgrade task instance
type Kubeclient struct {
	// clientset refers to upgrade task
	// clientset that will be responsible to
	// make kubernetes API calls
	clientset *clientset.Clientset
	// kubeconfig path to get kubernetes clientset
	kubeConfigPath string
	namespace      string
	// functions useful during mocking
	getClientset        getClientsetFn
	getClientsetForPath getClientsetForPathFn
	list                listFn
	get                 getFn
	create              createFn
	del                 deleteFn
	patch               patchFn
	update              updateFn
}

// KubeclientBuildOption defines the abstraction
// to build a kubeclient instance
type KubeclientBuildOption func(*Kubeclient)

func defaultGetClientset() (clients *clientset.Clientset, err error) {
	if clientsetInstance != nil {
		return clientsetInstance, nil
	}
	config, err := kclient.New().GetConfigForPathOrDirect()
	if err != nil {
		return nil, err
	}
	upgradeTaskCS, err := clientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	once.Do(func() {
		clientsetInstance = upgradeTaskCS
	})
	return clientsetInstance, nil
}

func defaultGetClientsetForPath(
	kubeConfigPath string,
) (clients *clientset.Clientset, err error) {
	config, err := kclient.New(kclient.WithKubeConfigPath(kubeConfigPath)).
		GetConfigForPathOrDirect()
	if err != nil {
		return nil, err
	}
	return clientset.NewForConfig(config)
}

func defaultList(
	cli *clientset.Clientset,
	namespace string,
	opts metav1.ListOptions,
) (*apis.UpgradeTaskList, error) {
	return cli.OpenebsV1alpha1().UpgradeTasks(namespace).List(opts)
}

func defaultGet(
	cli *clientset.Clientset,
	namespace, name string,
	opts metav1.GetOptions,
) (*apis.UpgradeTask, error) {
	return cli.OpenebsV1alpha1().UpgradeTasks(namespace).Get(name, opts)
}

func defaultCreate(
	cli *clientset.Clientset,
	namespace string,
	upgradeTask *apis.UpgradeTask,
) (*apis.UpgradeTask, error) {
	return cli.OpenebsV1alpha1().UpgradeTasks(namespace).Create(upgradeTask)
}

func defaultDel(
	cli *clientset.Clientset,
	namespace, name string,
	opts *metav1.DeleteOptions,
) error {
	return cli.OpenebsV1alpha1().
		UpgradeTasks(namespace).
		Delete(name, opts)
}

func defaultPatch(
	cli *clientset.Clientset,
	namespace, name string,
	pt types.PatchType,
	data []byte,
	subresources ...string,
) (*apis.UpgradeTask, error) {
	return cli.OpenebsV1alpha1().
		UpgradeTasks(namespace).
		Patch(name, pt, data, subresources...)
}

func defaultUpdate(
	cli *clientset.Clientset,
	namespace string,
	upgradeTask *apis.UpgradeTask,
) (*apis.UpgradeTask, error) {
	return cli.OpenebsV1alpha1().UpgradeTasks(namespace).Update(upgradeTask)
}

// WithDefaults sets the default options
// of kubeclient instance
func (k *Kubeclient) WithDefaults() {
	if k.getClientset == nil {
		k.getClientset = defaultGetClientset
	}
	if k.getClientsetForPath == nil {
		k.getClientsetForPath = defaultGetClientsetForPath
	}
	if k.list == nil {
		k.list = defaultList
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

// WithKubeConfigPath sets the kubeConfig path
// against client instance
func WithKubeConfigPath(kubeConfigPath string) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.kubeConfigPath = kubeConfigPath
	}
}

// NewKubeClient returns a new instance of kubeclient meant for
// upgrade task operations
func NewKubeClient(opts ...KubeclientBuildOption) *Kubeclient {
	k := &Kubeclient{}
	for _, o := range opts {
		o(k)
	}
	k.WithDefaults()
	return k
}

func (k *Kubeclient) getClientsetForPathOrDirect() (*clientset.Clientset, error) {
	if k.kubeConfigPath != "" {
		return k.getClientsetForPath(k.kubeConfigPath)
	}
	return k.getClientset()
}

// WithNamespace sets the kubernetes namespace against
// the provided namespace
func (k *Kubeclient) WithNamespace(namespace string) *Kubeclient {
	k.namespace = namespace
	return k
}

// getClientOrCached returns either a new instance
// of kubernetes client or its cached copy
func (k *Kubeclient) getClientsetOrCached() (*clientset.Clientset, error) {
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

// List returns a list of disk
// instances present in kubernetes cluster
func (k *Kubeclient) List(opts metav1.ListOptions) (*apis.UpgradeTaskList, error) {
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to list upgradeTask in namespace {%s}",
			k.namespace,
		)
	}
	return k.list(cli, k.namespace, opts)
}

// Get returns a upgrade task object
func (k *Kubeclient) Get(
	name string,
	opts metav1.GetOptions,
) (*apis.UpgradeTask, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errors.New(
			"failed to get upgradeTask: missing upgradeTask name",
		)
	}
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to get upgradeTask {%s} in namespace {%s}",
			name,
			k.namespace,
		)
	}
	return k.get(cli, k.namespace, name, opts)
}

// Create creates a upgradeTask in specified namespace in kubernetes cluster
func (k *Kubeclient) Create(
	upgradeTask *apis.UpgradeTask,
) (*apis.UpgradeTask, error) {
	if upgradeTask == nil {
		return nil, errors.New(
			"failed to create upgradeTask: nil upgradeTask object",
		)
	}
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to create upgradeTask {%s} in namespace {%s}",
			upgradeTask.Name,
			upgradeTask.Namespace,
		)
	}
	return k.create(cli, k.namespace, upgradeTask)
}

// Delete deletes a upgradeTask instance from the
// kubecrnetes cluster
func (k *Kubeclient) Delete(name string, deleteOpts *metav1.DeleteOptions) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("failed to delete upgradeTask: missing upgradeTask name")
	}
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return errors.Wrapf(
			err,
			"failed to delete upgradeTask {%s} in namespace {%s}",
			name,
			k.namespace,
		)
	}
	return k.del(cli, k.namespace, name, deleteOpts)
}

// Patch patches the upgrade task  if present in kubernetes cluster
func (k *Kubeclient) Patch(
	name string,
	pt types.PatchType,
	data []byte,
	subresources ...string,
) (*apis.UpgradeTask, error) {
	if len(name) == 0 {
		return nil, errors.New(
			"failed to patch upgrade task : missing upgradeTask name",
		)
	}
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to patch upgradeTask: {%s}", name)
	}
	return k.patch(cli, k.namespace, name, pt, data, subresources...)
}

// Update updates the upgrade task  if present in kubernetes cluster
func (k *Kubeclient) Update(
	upgradeTask *apis.UpgradeTask,
) (*apis.UpgradeTask, error) {
	if upgradeTask == nil {
		return nil, errors.New(
			"failed to udpate upgradeTask: nil upgradeTask object",
		)
	}
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to update upgradeTask {%s} in namespace {%s}",
			upgradeTask.Name,
			upgradeTask.Namespace,
		)
	}
	return k.update(cli, k.namespace, upgradeTask)
}

// IsValidStatus is used to validate IsValidStatus
func IsValidStatus(o apis.UpgradeDetailedStatuses) bool {
	if o.Step == "" {
		return false
	}
	if o.Phase == "" {
		return false
	}
	if o.Message == "" && o.Phase != apis.StepWaiting {
		return false
	}
	if o.Reason == "" && o.Phase == apis.StepErrored {
		return false
	}
	return true
}
