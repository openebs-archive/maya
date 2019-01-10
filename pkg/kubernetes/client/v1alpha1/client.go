/*
Copyright 2018 The OpenEBS Authors

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

package v1alpha1

import (
	env "github.com/openebs/maya/pkg/env/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"strings"
)

const (
	// K8sMasterIPEnvironmentKey is the environment variable key used to
	// determine the kubernetes master IP address
	K8sMasterIPEnvironmentKey env.ENVKey = "OPENEBS_IO_K8S_MASTER"
	// KubeConfigEnvironmentKey is the environment variable key used to
	// determine the kubernetes config
	KubeConfigEnvironmentKey env.ENVKey = "OPENEBS_IO_KUBE_CONFIG"
)

// getInClusterConfigFunc is a typed function to get incluster config
//
// NOTE:
//  functional makes it simple to mock
type getInClusterConfigFunc func() (*rest.Config, error)

// buildConfigFromFlagsFunc is a typed function to get desired
// kubernetes config
//
// NOTE:
//  functional makes it simple to mock
type buildConfigFromFlagsFunc func(string, string) (*rest.Config, error)

// getKubeMasterIPFunc is a typed function to get kubernetes master
// IP address
//
// NOTE:
//  functional makes it simple to mock
type getKubeMasterIPFunc func(env.ENVKey) string

// getKubeConfigPathFunc is a typed function to get kubernetes config
// path
//
// NOTE:
//  functional makes it simple to mock
type getKubeConfigPathFunc func(env.ENVKey) string

// getKubernetesClientsetFunc is a typed function to get kubernetes
// clientset
//
// NOTE:
//  functional makes it simple to mock
type getKubernetesClientsetFunc func(*rest.Config) (*kubernetes.Clientset, error)

// GetConfigFunc is a typed function to get kubernetes config
//
// NOTE:
//  functional makes it simple to mock
type GetConfigFunc func(*Client) (*rest.Config, error)

// GetConfig returns kubernetes config instance
//
// NOTE:
//  This is an implementation of GetConfigFunc
func GetConfig(c *Client) (*rest.Config, error) {
	if c == nil {
		return nil, errors.New("failed to get kubernetes config: nil client was provided")
	}
	return c.Config()
}

// Client provides common kuberenetes client operations
type Client struct {
	IsInCluster            bool                       // flag to let client point to its own cluster
	getInClusterConfig     getInClusterConfigFunc     // typed func to get in cluster config
	buildConfigFromFlags   buildConfigFromFlagsFunc   // typed func to get desired kubernetes config
	getKubernetesClientset getKubernetesClientsetFunc // typed func to get kubernetes clienset
	getKubeMasterIP        getKubeMasterIPFunc        // typed func to get kubernetes master IP
	getKubeConfigPath      getKubeConfigPathFunc      // typed func to get kubernetes config path
}

// OptionFunc is a typed function that abstracts any kind of operation
// against the provided client instance
//
// This is the basic building block to create functional operations
// against the client instance
type OptionFunc func(*Client)

// New returns a new instance of client
func New(opts ...OptionFunc) *Client {
	k := &Client{}
	for _, o := range opts {
		o(k)
	}
	if k.getInClusterConfig == nil {
		k.getInClusterConfig = rest.InClusterConfig
	}
	if k.buildConfigFromFlags == nil {
		k.buildConfigFromFlags = clientcmd.BuildConfigFromFlags
	}
	if k.getKubernetesClientset == nil {
		k.getKubernetesClientset = kubernetes.NewForConfig
	}
	if k.getKubeMasterIP == nil {
		k.getKubeMasterIP = env.Get
	}
	if k.getKubeConfigPath == nil {
		k.getKubeConfigPath = env.Get
	}
	return k
}

// InCluster enables IsInCluster flag
func InCluster() OptionFunc {
	return func(c *Client) {
		c.IsInCluster = true
	}
}

// config returns the kubernetes config instance based on available criteria
func (c *Client) Config() (config *rest.Config, err error) {
	if c.IsInCluster {
		return c.getInClusterConfig()
	}
	if strings.TrimSpace(c.getKubeMasterIP(K8sMasterIPEnvironmentKey)) != "" ||
		strings.TrimSpace(c.getKubeConfigPath(KubeConfigEnvironmentKey)) != "" {
		return c.getConfigFromENV()
	}
	return c.getInClusterConfig()
}

func (c *Client) getConfigFromENV() (config *rest.Config, err error) {
	k8sMaster := c.getKubeMasterIP(K8sMasterIPEnvironmentKey)
	kubeConfig := c.getKubeConfigPath(KubeConfigEnvironmentKey)
	if strings.TrimSpace(k8sMaster) == "" &&
		strings.TrimSpace(kubeConfig) == "" {
		return nil, errors.Errorf("failed to get kubernetes config: missing environment variables: atleast one should be set: %s or %s", K8sMasterIPEnvironmentKey, KubeConfigEnvironmentKey)
	}
	return c.buildConfigFromFlags(k8sMaster, kubeConfig)
}

// Clientset returns a new instance of kubernetes clientset
func (c *Client) Clientset() (*kubernetes.Clientset, error) {
	config, err := c.Config()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get kubernetes clientset")
	}
	return c.getKubernetesClientset(config)
}
