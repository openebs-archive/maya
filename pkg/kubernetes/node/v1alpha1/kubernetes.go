package v1alpha1

import (
	"github.com/openebs/maya/pkg/kubernetes/client/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clientset "k8s.io/client-go/kubernetes"
)

// getClientsetFn is a typed function that
// abstracts fetching of clientset
type getClientsetFn func() (clientset *clientset.Clientset, err error)

// listFn is a typed function that abstracts
// listing of nodes
type listFn func(cli *clientset.Clientset, opts metav1.ListOptions) (*corev1.NodeList, error)

// kubeclient enables kubernetes API operations on node instance
type kubeclient struct {
	// clientset refers to node clientset
	// that will be responsible to
	// make kubernetes API calls
	clientset *clientset.Clientset

	// functions useful during mocking
	getClientset getClientsetFn
	list         listFn
}

// kubeclientBuildOption defines the abstraction
// to build a kubeclient instance
type kubeclientBuildOption func(*kubeclient)

func (k *kubeclient) withDefaults() {
	if k.getClientset == nil {
		k.getClientset = func() (clients *clientset.Clientset, err error) {
			return v1alpha1.New(v1alpha1.NotInCluster()).Clientset()
		}
	}
	if k.list == nil {
		k.list = func(cli *clientset.Clientset, opts metav1.ListOptions) (*corev1.NodeList, error) {
			return cli.CoreV1().Nodes().List(opts)
		}
	}
}

// KubeClient returns a new instance of kubeclient meant for node
func KubeClient() *kubeclient {
	k := &kubeclient{}
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

// List returns a list of nodes instances present in kubernetes cluster
func (k *kubeclient) List(opts metav1.ListOptions) (*corev1.NodeList, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.list(cli, opts)
}
