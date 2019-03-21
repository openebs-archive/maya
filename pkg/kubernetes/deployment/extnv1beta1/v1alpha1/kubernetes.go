package v1alpha1

import (
	client "github.com/openebs/maya/pkg/kubernetes/client/v1alpha1"
	extnv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/client-go/kubernetes"
)

// getClientsetFn is a typed function that abstracts fetching of internal clientset
type getClientsetFn func() (clientset *kubernetes.Clientset, err error)

// getFn is a typed function that abstracts get of deployment instances
type getFn func(cli *kubernetes.Clientset, name, namespace string) (*extnv1beta1.Deployment, error)

// rolloutStatusFn is a typed function that abstracts rollout status of deployment instances
type rolloutStatusFn func(d *extnv1beta1.Deployment) ([]byte, error)

// kubeclient enables kubernetes API operations on deployment instance
type kubeclient struct {
	// clientset refers to kubernetes clientset. It is responsible to
	// make kubernetes API calls for crud op
	clientset *kubernetes.Clientset
	namespace string

	// functions useful during mocking
	getClientset  getClientsetFn
	get           getFn
	rolloutStatus rolloutStatusFn
}

// rolloutOutput struct contaons message and boolean value to show rolloutstatus
type rolloutOutput struct {
	IsRolledout bool   `json:"isRolledout"`
	Message     string `json:"message"`
}

// kubeclientBuildOption defines the abstraction to build a kubeclient instance
type kubeclientBuildOption func(*kubeclient)

// withDefaults sets the default options of kubeclient instance
func (k *kubeclient) withDefaults() {
	if k.getClientset == nil {
		k.getClientset = func() (clients *kubernetes.Clientset, err error) {
			config, err := client.GetConfig(client.New())
			if err != nil {
				return nil, err
			}
			return kubernetes.NewForConfig(config)
		}
	}

	if k.get == nil {
		k.get = func(cli *kubernetes.Clientset, name,
			namespace string) (d *extnv1beta1.Deployment, err error) {
			d = &extnv1beta1.Deployment{}
			err = cli.ExtensionsV1beta1().
				RESTClient().
				Get().
				Namespace(namespace).
				Name(name).
				Resource("deployments").
				Do().
				Into(d)
			return
		}
	}

	if k.rolloutStatus == nil {
		k.rolloutStatus = func(d *extnv1beta1.Deployment) (op []byte, err error) {
			return New(WithAPIObject(d)).
				RolloutStatusf()
		}
	}

}

// WithClientset sets the kubernetes client against the kubeclient instance
func WithClientset(c *kubernetes.Clientset) kubeclientBuildOption {
	return func(k *kubeclient) {
		k.clientset = c
	}
}

// WithNamespace set namespace in kubeclient object
func WithNamespace(namespace string) kubeclientBuildOption {
	return func(k *kubeclient) {
		k.namespace = namespace
	}
}

// KubeClient returns a new instance of kubeclient meant for deployment.
// caller can configure it with different kubeclientBuildOption
func KubeClient(opts ...kubeclientBuildOption) *kubeclient {
	k := &kubeclient{}
	for _, o := range opts {
		o(k)
	}
	k.withDefaults()
	return k
}

// getClientOrCached returns either a new instance of kubernetes client or its cached copy
func (k *kubeclient) getClientOrCached() (*kubernetes.Clientset, error) {
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
func (k *kubeclient) Get(name string) (*extnv1beta1.Deployment, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.get(cli, name, k.namespace)
}

// RolloutStatus returns deployment's rollout status for given name
func (k *kubeclient) RolloutStatus(name string) (op []byte, err error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	d, err := k.get(cli, name, k.namespace)
	if err != nil {
		return nil, err
	}
	return k.rolloutStatus(d)
}
