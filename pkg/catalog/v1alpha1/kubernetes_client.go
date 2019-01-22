/*
Copyright 2019 The OpenEBS Authors

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

package v1beta1

import (
	apis "github.com/openebs/maya/pkg/apis/openebs.io/catalog/v1alpha1"
	"github.com/openebs/maya/pkg/client/generated/openebs.io/catalog/v1alpha1/clientset/internalclientset"
	kube "github.com/openebs/maya/pkg/kubernetes/client/v1alpha1"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"strings"
)

// getClientsetFunc is a typed function to get runtask's clientset
//
// NOTE:
//  functional makes it simple to mock
type getClientsetFunc func(*rest.Config) (*internalclientset.Clientset, error)

// getClientset returns a new instance of catalog based clientset
//
// NOTE:
//  This is an implementation of getClientsetFunc
func getClientset(c *rest.Config) (*internalclientset.Clientset, error) {
	return internalclientset.NewForConfig(c)
}

// getCatalogFunc abstracts fetching an instance of
// catalog
//
// NOTE:
//  function type makes it simple to mock
type getCatalogFunc func(k *kubeclient, cs *internalclientset.Clientset, name string) (*apis.Catalog, error)

// getCatalog returns an instance of catalog corresponding
// to the provided name
//
// NOTE:
//  This is an implementation of getCatalogFunc
func getCatalog(k *kubeclient, cs *internalclientset.Clientset, name string) (*apis.Catalog, error) {
	if k == nil {
		return nil, errors.New("failed to get runtask: nil kubeclient was provided")
	}
	if cs == nil {
		return nil, errors.New("failed to get runtask: nil clientset was provided")
	}
	return cs.OpenebsV1alpha1().Catalogs(k.namespace).Get(name, v1.GetOptions{})
}

// kubeclient enables kubernetes API operations on catalog instance
type kubeclient struct {
	*kube.Client                              // embeds kubernetes client related functions
	getConfig    kube.GetConfigFunc           // handle to get kubernetes config
	getClientset getClientsetFunc             // handle to get clientset to invoke API calls against catalog
	getCatalog   getCatalogFunc               // handle to get catalog instance
	clientset    *internalclientset.Clientset // clientset instance that can be cached for reuse
	namespace    string                       // namespace to use during API calls
}

// withDefaults sets the defaults associated with the provided
// kubeclient instance
func withDefaults(k *kubeclient) {
	if k.getConfig == nil {
		k.getConfig = kube.GetConfig
	}
	if k.getClientset == nil {
		k.getClientset = getClientset
	}
	if k.getCatalog == nil {
		k.getCatalog = getCatalog
	}
}

// KubeClientOptionFunc is a typed function that abstracts any kind
// of operation against the provided client instance
//
// This is the basic building block to create functional operations
// against the kubeclient instance
type KubeClientOptionFunc func(*kubeclient)

// KubeClient returns a new instance of kubeclient meant for
// runtask operations
func KubeClient(opts ...KubeClientOptionFunc) *kubeclient {
	k := &kubeclient{Client: kube.New()}
	for _, o := range opts {
		o(k)
	}
	withDefaults(k)
	return k
}

// kubeclientBuilder helps constructing an instance of kubeclient
type kubeclientBuilder struct {
	client *kubeclient
}

// KubeClientBuilder returns a new instance of kubeclientBuilder
func KubeClientBuilder() *kubeclientBuilder {
	return &kubeclientBuilder{
		client: &kubeclient{Client: kube.New()},
	}
}

// Build returns the resulting kubeclient instance
func (b *kubeclientBuilder) Build() *kubeclient {
	withDefaults(b.client)
	return b.client
}

// InCluster enables isInCluster flag
func InCluster() KubeClientOptionFunc {
	return func(c *kubeclient) {
		c.IsInCluster = true
	}
}

// InCluster enables isInCluster flag
func (b *kubeclientBuilder) InCluster() *kubeclientBuilder {
	InCluster()(b.client)
	return b
}

// WithNamespace sets namespace that should be used during
// kuberenets API calls against runtask resource
func WithNamespace(namespace string) KubeClientOptionFunc {
	return func(k *kubeclient) {
		k.namespace = namespace
	}
}

// WithNamespace sets namespace that should be used during
// kuberenets API calls against runtask resource
func (b *kubeclientBuilder) WithNamespace(namespace string) *kubeclientBuilder {
	WithNamespace(namespace)(b.client)
	return b
}

// getInternalClientset returns runtask based clientset instance
func (k *kubeclient) getInternalClientset() (*internalclientset.Clientset, error) {
	conf, err := k.getConfig(k.Client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get catalog clientset")
	}
	return k.getClientset(conf)
}

// getCachedOrNewInternalClientset returns runtask based clientset instance
func (k *kubeclient) getCachedOrNewInternalClientset() (*internalclientset.Clientset, error) {
	if k.clientset != nil {
		return k.clientset, nil
	}
	cs, err := k.getInternalClientset()
	if err != nil {
		return nil, err
	}
	k.clientset = cs
	return k.clientset, nil
}

// Get returns a catalog instance from kubernetes cluster
func (k *kubeclient) Get(name string) (*apis.Catalog, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("failed to get catalog: missing catalog name")
	}
	cs, err := k.getCachedOrNewInternalClientset()
	if err != nil {
		return nil, err
	}
	return k.getCatalog(k, cs, name)
}
