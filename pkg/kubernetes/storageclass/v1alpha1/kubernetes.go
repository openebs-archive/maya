package v1alpha1

import (
	client "github.com/openebs/maya/pkg/kubernetes/client/v1alpha1"
	"github.com/pkg/errors"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// getClientsetFn is a typed function that
// abstracts fetching of clientset
type getClientsetFn func() (clientset *kubernetes.Clientset, err error)

// listFn is a typed function that abstracts
// listing of storageclasses
type listFn func(cli *kubernetes.Clientset, opts metav1.ListOptions) (*storagev1.StorageClassList, error)

// getFn is a typed function that abstracts to get the storageclass
type getFn func(cli *kubernetes.Clientset, name string, opts metav1.GetOptions) (*storagev1.StorageClass, error)

// Kubeclient enables kubernetes API operations on storageclass instance
type Kubeclient struct {
	// clientset refers to storageclass clientset
	// that will be responsible to
	// make kubernetes API calls
	clientset *kubernetes.Clientset

	// functions useful during mocking
	getClientset getClientsetFn
	list         listFn
	get          getFn
}

// kubeclientBuildOption defines the abstraction
// to build a kubeclient instance
type kubeclientBuildOption func(*Kubeclient)

func (k *Kubeclient) withDefaults() {
	if k.getClientset == nil {
		k.getClientset = func() (clients *kubernetes.Clientset, err error) {
			return client.New().Clientset()
		}
	}
	if k.list == nil {
		k.list = func(cli *kubernetes.Clientset, opts metav1.ListOptions) (*storagev1.StorageClassList, error) {
			return cli.StorageV1().StorageClasses().List(opts)
		}
	}
	if k.get == nil {
		k.get = func(cli *kubernetes.Clientset, name string, opts metav1.GetOptions) (*storagev1.StorageClass, error) {
			return cli.StorageV1().StorageClasses().Get(name, opts)
		}
	}
}

// KubeClient returns a new instance of kubeclient meant for storageclass
func KubeClient(opts ...kubeclientBuildOption) *Kubeclient {
	k := &Kubeclient{}
	for _, o := range opts {
		o(k)
	}
	k.withDefaults()
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
		return nil, errors.Wrapf(err, "failed to get clientset")
	}
	k.clientset = c
	return k.clientset, nil
}

// List returns a list of storageclass instances present in kubernetes cluster
func (k *Kubeclient) List(opts metav1.ListOptions) (*storagev1.StorageClassList, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list")
	}
	return k.list(cli, opts)
}

// Get return a storageclass instance present in kubernetes cluster
func (k *Kubeclient) Get(name string, opts metav1.GetOptions) (*storagev1.StorageClass, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get")
	}
	return k.get(cli, name, opts)
}
