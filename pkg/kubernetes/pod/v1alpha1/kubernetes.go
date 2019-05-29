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
	"bytes"
	"encoding/json"

	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	client "github.com/openebs/maya/pkg/kubernetes/client/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// getClientsetFn is a typed function that
// abstracts fetching of clientset
type getClientsetFn func() (clientset *clientset.Clientset, err error)

// getClientsetFromPathFn is a typed function that
// abstracts fetching of clientset from kubeConfigPath
type getClientsetForPathFn func(kubeConfigPath string) (clientset *clientset.Clientset, err error)

// getKubeConfigFn is a typed function that abstracts fetching
// rest config
type getKubeConfigFn func() (config *rest.Config, err error)

// listFn is a typed function that abstracts
// listing of pods
type listFn func(cli *clientset.Clientset, namespace string, opts metav1.ListOptions) (*corev1.PodList, error)

// deleteFn is a typed function that abstracts
// deleting of pod
type deleteFn func(cli *clientset.Clientset, namespace, name string, opts *metav1.DeleteOptions) error

// getFn is a typed function that abstracts
// to get pod
type getFn func(cli *clientset.Clientset, namespace, name string, opts metav1.GetOptions) (*corev1.Pod, error)

// execFn is a typed function that abstracts
// pod exec
type execFn func(cli *clientset.Clientset, name, namespace string, opts *corev1.PodExecOptions) (*ExecOutput, error)

// KubeClient enables kubernetes API operations
// on pod instance
type KubeClient struct {
	// clientset refers to pod clientset
	// that will be responsible to
	// make kubernetes API calls
	clientset *clientset.Clientset

	// namespace holds the namespace on which
	// KubeClient has to operate
	namespace string

	// kubeconfig path to get kubernetes clientset
	kubeConfigPath string

	// functions useful during mocking
	getKubeConfig       getKubeConfigFn
	getClientset        getClientsetFn
	getClientsetForPath getClientsetForPathFn
	list                listFn
	del                 deleteFn
	get                 getFn
	exec                execFn
}

// KubeClientBuildOption defines the abstraction
// to build a KubeClient instance
type KubeClientBuildOption func(*KubeClient)

// withDefaults sets the default options
// of KubeClient instance
func (k *KubeClient) withDefaults() {
	if k.getKubeConfig == nil {
		k.getKubeConfig = func() (config *rest.Config, err error) {
			return client.New().Config()
		}
	}
	if k.getClientset == nil {
		k.getClientset = func() (clients *clientset.Clientset, err error) {
			return client.New().Clientset()
		}
	}
	if k.getClientsetForPath == nil {
		k.getClientsetForPath = func(kubeConfigPath string) (clients *clientset.Clientset, err error) {
			return client.New(client.WithKubeConfigPath(kubeConfigPath)).Clientset()
		}
	}
	if k.list == nil {
		k.list = func(cli *clientset.Clientset, namespace string, opts metav1.ListOptions) (*corev1.PodList, error) {
			return cli.CoreV1().Pods(namespace).List(opts)
		}
	}
	if k.del == nil {
		k.del = func(cli *clientset.Clientset, namespace, name string, opts *metav1.DeleteOptions) error {
			return cli.CoreV1().Pods(namespace).Delete(name, opts)
		}
	}
	if k.get == nil {
		k.get = func(cli *clientset.Clientset, namespace, name string, opts metav1.GetOptions) (*corev1.Pod, error) {
			return cli.CoreV1().Pods(namespace).Get(name, opts)
		}
	}
	if k.exec == nil {
		k.exec = func(cli *clientset.Clientset, name, namespace string, opts *corev1.PodExecOptions) (*ExecOutput, error) {
			return k.Execute(cli, name, namespace, opts)
		}
	}
}

// WithClientSet sets the kubernetes client against
// the KubeClient instance
func WithClientSet(c *clientset.Clientset) KubeClientBuildOption {
	return func(k *KubeClient) {
		k.clientset = c
	}
}

// WithKubeConfigPath sets the kubeConfig path
// against client instance
func WithKubeConfigPath(path string) KubeClientBuildOption {
	return func(k *KubeClient) {
		k.kubeConfigPath = path
	}
}

// NewKubeClient returns a new instance of KubeClient meant for
// cstor volume replica operations
func NewKubeClient(opts ...KubeClientBuildOption) *KubeClient {
	k := &KubeClient{}
	for _, o := range opts {
		o(k)
	}
	k.withDefaults()
	return k
}

// WithNamespace sets the kubernetes namespace against
// the provided namespace
func (k *KubeClient) WithNamespace(namespace string) *KubeClient {
	k.namespace = namespace
	return k
}

func (k *KubeClient) getClientsetForPathOrDirect() (*clientset.Clientset, error) {
	if k.kubeConfigPath != "" {
		return k.getClientsetForPath(k.kubeConfigPath)
	}
	return k.getClientset()
}

// getClientsetOrCached returns either a new instance
// of kubernetes client or its cached copy
func (k *KubeClient) getClientsetOrCached() (*clientset.Clientset, error) {
	if k.clientset != nil {
		return k.clientset, nil
	}

	cs, err := k.getClientsetForPathOrDirect()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get clientset")
	}
	k.clientset = cs
	return k.clientset, nil
}

// List returns a list of pod
// instances present in kubernetes cluster
func (k *KubeClient) List(opts metav1.ListOptions) (*corev1.PodList, error) {
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list pods")
	}
	return k.list(cli, k.namespace, opts)
}

// Delete deletes a pod instance present in kubernetes cluster
func (k *KubeClient) Delete(name string, opts *metav1.DeleteOptions) error {
	if len(name) == 0 {
		return errors.New("failed to delete pod: missing pod name")
	}
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return errors.Wrapf(err, "failed to delete pod {%s}: failed to get clientset", name)
	}
	return k.del(cli, k.namespace, name, opts)
}

// Get gets a pod object present in kubernetes cluster
func (k *KubeClient) Get(name string, opts metav1.GetOptions) (*corev1.Pod, error) {
	if len(name) == 0 {
		return nil, errors.New("failed to get pod: missing pod name")
	}
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get pod {%s}: failed to get clientset", name)
	}
	return k.get(cli, k.namespace, name, opts)
}

// GetRaw gets pod object for a given name and namespace present
// in kubernetes cluster and returns result in raw byte.
func (k *KubeClient) GetRaw(name string, opts metav1.GetOptions) ([]byte, error) {
	p, err := k.Get(name, opts)
	if err != nil {
		return nil, err
	}
	return json.Marshal(p)
}

// ExecOutput struct contains stdout and stderr
type ExecOutput struct {
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
}

// Exec runs a command remotely in a container of a pod
func (k *KubeClient) Exec(name string,
	opts *corev1.PodExecOptions) (*ExecOutput, error) {
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, err
	}
	return k.exec(cli, name, k.namespace, opts)
}

// ExecRaw runs a command remotely in a container of a pod
// and returns raw output
func (k *KubeClient) ExecRaw(name string,
	opts *corev1.PodExecOptions) ([]byte, error) {
	execOutput, err := k.Exec(name, opts)
	if err != nil {
		return nil, err
	}
	return json.Marshal(execOutput)
}

// Execute executes the given commands inside a container of a pod and returns the
// output or error if any.
func (k *KubeClient) Execute(cli *clientset.Clientset, name, namespace string,
	opts *corev1.PodExecOptions) (*ExecOutput, error) {
	var stdout, stderr bytes.Buffer
	req := cli.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(name).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(opts, scheme.ParameterCodec)
	config, err := k.getKubeConfig()
	if err != nil {
		return nil, errors.Wrapf(err,
			"failed to exec into pod {%s}: failed to get kube config", name)
	}
	// create exec executor which is an interface for transporting shell-style streams.
	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return nil, errors.Wrapf(err,
			"failed to exec into pod {%s}: failed to connect to the provided server", name)
	}
	// Stream initiates the transport of the standard shell streams. It will transport any
	// non-nil stream to a remote system, and return an error if a problem occurs.
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    opts.TTY,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to exec into pod {%s}: failed to stream", name)
	}
	execOutput := &ExecOutput{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
	return execOutput, nil
}
