package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kclient "github.com/openebs/maya/pkg/client/k8s/v1alpha1"
	clientset "k8s.io/client-go/kubernetes"
)

// getClientsetFn is a typed function that
// abstracts fetching of clientset
type getClientsetFn func() (clientset *clientset.Clientset, err error)

// listFn is a typed function that abstracts
// listing of pvcs
type listFn func(cli *clientset.Clientset, namespace string, opts metav1.ListOptions) (*v1.PersistentVolumeClaimList, error)

// deleteFn is a typed function that abstracts
// deletion of pvcs
type deleteFn func(cli *clientset.Clientset, namespace string, name string, deleteOpts *metav1.DeleteOptions) error

// deleteFn is a typed function that abstracts
// deletion of pvc's collection
type deleteCollectionFn func(cli *clientset.Clientset, namespace string, listOpts metav1.ListOptions, deleteOpts *metav1.DeleteOptions) error

// kubeclient enables kubernetes API operations
// on pvc instance
type kubeclient struct {
	// clientset refers to pvc clientset
	// that will be responsible to
	// make kubernetes API calls
	clientset *clientset.Clientset

	// namespace holds the namespace on which
	// kubeclient has to operate
	namespace string

	// functions useful during mocking
	getClientset  getClientsetFn
	list          listFn
	del           deleteFn
	delCollection deleteCollectionFn
}

// kubeclientBuildOption abstracts creating an
// instance of kubeclient
type kubeclientBuildOption func(*kubeclient)

// withDefaults sets the default options
// of kubeclient instance
func (k *kubeclient) withDefaults() {
	if k.getClientset == nil {
		k.getClientset = func() (clients *clientset.Clientset, err error) {
			config, err := kclient.Config().Get()
			if err != nil {
				return nil, err
			}
			return clientset.NewForConfig(config)
		}
	}
	if k.list == nil {
		k.list = func(cli *clientset.Clientset, namespace string, opts metav1.ListOptions) (*v1.PersistentVolumeClaimList, error) {
			return cli.CoreV1().PersistentVolumeClaims(namespace).List(opts)
		}
	}
	if k.del == nil {
		k.del = func(cli *clientset.Clientset, namespace string, name string, deleteOpts *metav1.DeleteOptions) error {
			return cli.CoreV1().PersistentVolumeClaims(namespace).Delete(name, deleteOpts)
		}
	}
	if k.delCollection == nil {
		k.delCollection = func(cli *clientset.Clientset, namespace string, listOpts metav1.ListOptions, deleteOpts *metav1.DeleteOptions) error {
			return cli.CoreV1().PersistentVolumeClaims(namespace).DeleteCollection(deleteOpts, listOpts)
		}
	}
}

// WithNamespace sets the kubernetes client against
// the provided namespace
func WithNamespace(namespace string) kubeclientBuildOption {
	return func(k *kubeclient) {
		k.namespace = namespace
	}
}

// WithClientSet sets the kubernetes client against
// the kubeclient instance
func WithClientSet(c *clientset.Clientset) kubeclientBuildOption {
	return func(k *kubeclient) {
		k.clientset = c
	}
}

// KubeClient returns a new instance of kubeclient meant for
// cstor volume replica operations
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
func (k *kubeclient) getClientOrCached() (*clientset.Clientset, error) {
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

// List returns a list of pvc
// instances present in kubernetes cluster
func (k *kubeclient) List(opts metav1.ListOptions) (*v1.PersistentVolumeClaimList, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.list(cli, k.namespace, opts)
}

// Delete deletes a pvc instance from the
// kubecrnetes cluster
func (k *kubeclient) Delete(name string, deleteOpts *metav1.DeleteOptions) error {
	cli, err := k.getClientOrCached()
	if err != nil {
		return err
	}
	return k.del(cli, k.namespace, name, deleteOpts)
}

// DeleteCollection deletes collection of pvc
// instance from the kubernetes cluster
func (k *kubeclient) DeleteCollection(listOpts metav1.ListOptions, deleteOpts *metav1.DeleteOptions) error {
	cli, err := k.getClientOrCached()
	if err != nil {
		return err
	}
	return k.delCollection(cli, k.namespace, listOpts, deleteOpts)
}
