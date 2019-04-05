package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/generated/clientset/internalclientset"
	kclient "github.com/openebs/maya/pkg/client/k8s/v1alpha1"
)

// getClientsetFn is a typed function that
// abstracts fetching of internal clientset
type getClientsetFn func() (clientset *clientset.Clientset, err error)

// getFn is a typed function that abstracts get of cstorvolume instances
type getFn func(cli *clientset.Clientset, name, namespace string,
	opts metav1.GetOptions) (*apis.CStorVolume, error)

// listFn is a typed function that abstracts
// listing of cstor volume instances
type listFn func(cli *clientset.Clientset, namespace string, opts metav1.ListOptions) (*apis.CStorVolumeList, error)

// delFn is a typed function that abstracts delete of cstorvolume instances
type delFn func(cli *clientset.Clientset, name, namespace string, opts *metav1.DeleteOptions) error

// kubeclient enables kubernetes API operations
// on cstor volume replica instance
type kubeclient struct {
	// clientset refers to cstor volume replica's
	// clientset that will be responsible to
	// make kubernetes API calls
	clientset *clientset.Clientset

	// namespace holds the namespace on which
	// kubeclient has to operate
	namespace string

	// functions useful during mocking
	getClientset getClientsetFn
	get          getFn
	list         listFn
	del          delFn
}

// kubeclientBuildOption defines the abstraction
// to build a kubeclient instance
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
	if k.get == nil {
		k.get = func(cli *clientset.Clientset, name, namespace string, opts metav1.GetOptions) (*apis.CStorVolume, error) {
			return cli.OpenebsV1alpha1().CStorVolumes(namespace).Get(name, opts)
		}
	}

	if k.list == nil {
		k.list = func(cli *clientset.Clientset, namespace string, opts metav1.ListOptions) (*apis.CStorVolumeList, error) {
			return cli.OpenebsV1alpha1().CStorVolumes(namespace).List(opts)
		}
	}
	if k.del == nil {
		k.del = func(cli *clientset.Clientset, name, namespace string, opts *metav1.DeleteOptions) error {
			// The object exists in the key-value store until the garbage collector
			// deletes all the dependents whose ownerReference.blockOwnerDeletion=true
			// from the key-value store.  API sever will put the "foregroundDeletion"
			// finalizer on the object, and sets its deletionTimestamp.  This policy is
			// cascading, i.e., the dependents will be deleted with Foreground.
			deletePropagation := metav1.DeletePropagationForeground
			opts.PropagationPolicy = &deletePropagation
			err := cli.OpenebsV1alpha1().CStorVolumes(namespace).Delete(name, opts)
			return err
		}
	}
}

// WithKubeClient sets the kubernetes client against
// the kubeclient instance
func WithKubeClient(c *clientset.Clientset) kubeclientBuildOption {
	return func(k *kubeclient) {
		k.clientset = c
	}
}

// WithNamespace sets the kubernetes client against
// the provided namespace
func WithNamespace(namespace string) kubeclientBuildOption {
	return func(k *kubeclient) {
		k.namespace = namespace
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

// Get returns deployment object for given name
func (k *kubeclient) Get(name, namespace string) (*apis.CStorVolume, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.get(cli, name, namespace, metav1.GetOptions{})
}

// List returns a list of cstor volume replica
// instances present in kubernetes cluster
func (k *kubeclient) List(opts metav1.ListOptions) (*apis.CStorVolumeList, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.list(cli, k.namespace, opts)
}

// Delete delete the cstorvolume resource
func (k *kubeclient) Delete(name string) error {
	cli, err := k.getClientOrCached()
	if err != nil {
		return err
	}
	return k.del(cli, name, k.namespace, &metav1.DeleteOptions{})
}
