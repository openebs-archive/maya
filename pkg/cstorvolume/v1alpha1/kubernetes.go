package v1alpha1

import (
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/generated/clientset/internalclientset"
	client "github.com/openebs/maya/pkg/kubernetes/client/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// Kubeclient enables kubernetes API operations
// on cstor volume replica instance
type Kubeclient struct {
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

// KubeclientBuildOption defines the abstraction
// to build a kubeclient instance
type KubeclientBuildOption func(*Kubeclient)

// withDefaults sets the default options
// of kubeclient instance
func (k *Kubeclient) withDefaults() {
	if k.getClientset == nil {
		k.getClientset = func() (clients *clientset.Clientset, err error) {
			config, err := client.GetConfig(client.New())
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

// getClientOrCached returns either a new instance
// of kubernetes client or its cached copy
func (k *Kubeclient) getClientOrCached() (*clientset.Clientset, error) {
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
func (k *Kubeclient) Get(name string, opts metav1.GetOptions) (*apis.CStorVolume, error) {
	if len(name) == 0 {
		return nil, errors.New("failed to get cstorvolume: name can't be empty")
	}
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.get(cli, name, k.namespace, opts)
}

// List returns a list of cstor volume replica
// instances present in kubernetes cluster
func (k *Kubeclient) List(opts metav1.ListOptions) (*apis.CStorVolumeList, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.list(cli, k.namespace, opts)
}

// Delete delete the cstorvolume resource
func (k *Kubeclient) Delete(name string) error {
	cli, err := k.getClientOrCached()
	if err != nil {
		return err
	}
	return k.del(cli, name, k.namespace, &metav1.DeleteOptions{})
}
