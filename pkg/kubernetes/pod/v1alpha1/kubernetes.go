package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kclient "github.com/openebs/maya/pkg/kubernetes/client/v1alpha1"
	clientset "k8s.io/client-go/kubernetes"
)

// getClientsetFn is a typed function that
// abstracts fetching of clientset
type getClientsetFn func() (clientset *clientset.Clientset, err error)

// listFn is a typed function that abstracts
// listing of pods
type listFn func(cli *clientset.Clientset, namespace string, opts metav1.ListOptions) (*v1.PodList, error)

// deleteFn is a typed function that abstracts
// deleting of pod
type deleteFn func(cli *clientset.Clientset, namespace, name string, opts *metav1.DeleteOptions) error

// getFn is a typed function that abstracts
// to get pod
type getFn func(cli *clientset.Clientset, namespace, name string, opts metav1.GetOptions) (*v1.Pod, error)

// Kubeclient enables kubernetes API operations
// on pod instance
type Kubeclient struct {
	// clientset refers to pod clientset
	// that will be responsible to
	// make kubernetes API calls
	clientset *clientset.Clientset

	// namespace holds the namespace on which
	// Kubeclient has to operate
	namespace string
	// functions useful during mocking
	getClientset getClientsetFn
	list         listFn
	del          deleteFn
	get          getFn
}

// kubeclientBuildOption defines the abstraction
// to build a Kubeclient instance
type kubeclientBuildOption func(*Kubeclient)

// withDefaults sets the default options
// of Kubeclient instance
func (k *Kubeclient) withDefaults() {
	if k.getClientset == nil {
		k.getClientset = func() (clients *clientset.Clientset, err error) {
			config, err := kclient.New().Config()
			if err != nil {
				return nil, err
			}
			return clientset.NewForConfig(config)
		}
	}
	if k.list == nil {
		k.list = func(cli *clientset.Clientset, namespace string, opts metav1.ListOptions) (*v1.PodList, error) {
			return cli.CoreV1().Pods(namespace).List(opts)
		}
	}
	if k.del == nil {
		k.del = func(cli *clientset.Clientset, namespace, name string, opts *metav1.DeleteOptions) error {
			return cli.CoreV1().Pods(namespace).Delete(name, opts)
		}
	}
	if k.get == nil {
		k.get = func(cli *clientset.Clientset, namespace, name string, opts metav1.GetOptions) (*v1.Pod, error) {
			return cli.CoreV1().Pods(namespace).Get(name, opts)
		}
	}
}

// WithNamespace sets the kubernetes client against
// the provided namespace
func WithNamespace(namespace string) kubeclientBuildOption {
	return func(k *Kubeclient) {
		k.namespace = namespace
	}
}

// WithClientSet sets the kubernetes client against
// the Kubeclient instance
func WithClientSet(c *clientset.Clientset) kubeclientBuildOption {
	return func(k *Kubeclient) {
		k.clientset = c
	}
}

// KubeClient returns a new instance of Kubeclient meant for
// cstor volume replica operations
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

// List returns a list of pod
// instances present in kubernetes cluster
func (k *Kubeclient) List(opts metav1.ListOptions) (*v1.PodList, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.list(cli, k.namespace, opts)
}

// Delete deletes a pod instance present in kubernetes cluster
func (k *Kubeclient) Delete(name string, opts *metav1.DeleteOptions) error {
	cli, err := k.getClientOrCached()
	if err != nil {
		return err
	}
	return k.del(cli, k.namespace, name, opts)
}

// Get gets a pod object present in kubernetes cluster
func (k *Kubeclient) Get(name string, opts metav1.GetOptions) (*v1.Pod, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.get(cli, k.namespace, name, opts)
}
