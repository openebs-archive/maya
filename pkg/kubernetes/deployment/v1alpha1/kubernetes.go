package v1alpha1

import (
	kclient "github.com/openebs/maya/pkg/client/k8s/v1alpha1"
	api_apps_v1 "k8s.io/api/apps/v1"
	api_extn_v1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/client-go/kubernetes"
)

// getClientsetFn is a typed function that abstracts fetching of internal clientset
type getClientsetFn func() (clientset *kubernetes.Clientset, err error)

// getExtnV1Beta1Fn is a typed function that abstracts get of deployment instances of ExtnV1Beta1 api group
type getExtnV1Beta1Fn func(cli *kubernetes.Clientset, name, namespace string) (*api_extn_v1beta1.Deployment, error)

// getAppsV1Fn is a typed function that abstracts get of deployment instances of appsv1 api group
type getAppsV1Fn func(cli *kubernetes.Clientset, name, namespace string) (*api_apps_v1.Deployment, error)

// rollOutStatusExtnV1Beta1Fn is a typed function that abstracts rollout status
//  of deployment instances of ExtnV1Beta1 api group
type rollOutStatusExtnV1Beta1Fn func(cli *kubernetes.Clientset, name, namespace string) ([]byte, error)

// rollOutStatusAppsV1Fn is a typed function that abstracts rollout status
//  of deployment instances of AppsV1 api group
type rollOutStatusAppsV1Fn func(cli *kubernetes.Clientset, name, namespace string) ([]byte, error)

// kubeclient enables kubernetes API operations on deployment instance
type kubeclient struct {
	// clientset refers to kubernetes clientset. It is responsible to
	// make kubernetes API calls for crud op
	clientset *kubernetes.Clientset

	// functions useful during mocking
	getClientset getClientsetFn

	getExtnV1Beta1 getExtnV1Beta1Fn

	getAppsV1 getAppsV1Fn

	rollOutStatusExtnV1Beta1 rollOutStatusExtnV1Beta1Fn

	rollOutStatusAppsV1 rollOutStatusAppsV1Fn
}

// rolloutOutput struct contaons message and boolean value to show rolloutstatus
type rolloutOutput struct {
	IsRolledout bool   `json:"IsRolledout"`
	Message     string `json:"Message"`
}

// kubeclientBuildOption defines the abstraction to build a kubeclient instance
type kubeclientBuildOption func(*kubeclient)

// withDefaults sets the default options of kubeclient instance
func (k *kubeclient) withDefaults() {
	if k.getClientset == nil {
		k.getClientset = func() (clients *kubernetes.Clientset, err error) {
			config, err := kclient.Config().Get()
			if err != nil {
				return nil, err
			}
			return kubernetes.NewForConfig(config)
		}
	}

	if k.getExtnV1Beta1 == nil {
		k.getExtnV1Beta1 = func(cli *kubernetes.Clientset, name,
			namespace string) (d *api_extn_v1beta1.Deployment, err error) {
			d = &api_extn_v1beta1.Deployment{}
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

	if k.getAppsV1 == nil {
		k.getAppsV1 = func(cli *kubernetes.Clientset, name,
			namespace string) (d *api_apps_v1.Deployment, err error) {
			d = &api_apps_v1.Deployment{}
			err = cli.AppsV1().
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

	if k.rollOutStatusExtnV1Beta1 == nil {
		k.rollOutStatusExtnV1Beta1 = func(cli *kubernetes.Clientset, name,
			namespace string) (op []byte, err error) {
			d, err := k.getExtnV1Beta1(cli, name, namespace)
			if err != nil {
				return nil, err
			}
			return DeployExtnV1Beta1(WithExtnV1Beta1Deployment(d)).
				AddCheck(IsSyncSpecV1B1()).
				AddCheck(IsProgressDeadlineExceededV1B1()).
				AddCheck(IsTerminationInProgressV1B1()).
				AddCheck(IsUpdationInProgressV1B1()).
				AddCheck(IsOlderReplicaActiveV1B1()).
				RollOutStatus()
		}
	}

	if k.rollOutStatusAppsV1 == nil {
		k.rollOutStatusAppsV1 = func(cli *kubernetes.Clientset, name,
			namespace string) (op []byte, err error) {
			d, err := k.getAppsV1(cli, name, namespace)
			if err != nil {
				return nil, err
			}
			return DeployAppsv1(WithAppsv1Deployment(d)).
				AddCheck(IsSyncSpecV1()).
				AddCheck(IsProgressDeadlineExceededV1()).
				AddCheck(IsTerminationInProgressV1()).
				AddCheck(IsUpdationInProgressV1()).
				AddCheck(IsOlderReplicaActiveV1()).
				RollOutStatus()
		}
	}

}

// WithKubeClient sets the kubernetes client against the kubeclient instance
func WithKubeClient(c *kubernetes.Clientset) kubeclientBuildOption {
	return func(k *kubeclient) {
		k.clientset = c
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

// GetExtnV1Beta1 returns deployment(ExtnV1Beta1) object for given name and namespaces
func (k *kubeclient) GetExtnV1Beta1(name, namespace string) (*api_extn_v1beta1.Deployment, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.getExtnV1Beta1(cli, name, namespace)
}

// GetAppsV1 returns deployment(GetAppsV1) object for given name and namespaces
func (k *kubeclient) GetAppsV1(name, namespace string) (*api_apps_v1.Deployment, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.getAppsV1(cli, name, namespace)
}

// RollOutStatusExtnV1Beta1 returns deployment(ExtnV1Beta1) rollout status for given name and namespaces
func (k *kubeclient) RollOutStatusExtnV1Beta1(name, namespace string) (op []byte, err error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.rollOutStatusExtnV1Beta1(cli, name, namespace)
}

// RollOutStatusAppsV1 returns deployment(AppsV1) rollout status for given name and namespaces
func (k *kubeclient) RollOutStatusAppsV1(name, namespace string) (op []byte, err error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.rollOutStatusAppsV1(cli, name, namespace)
}
