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

package v1alpha1

import (
	"context"
	"strings"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/kubeassert/v1alpha1"
	kclient "github.com/openebs/maya/pkg/kubernetes/client/v1alpha2"
	provider "github.com/openebs/maya/pkg/provider/v1alpha1"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// kubeclient enables kubernetes API operations on
// kubeassert instance
type kubeclient struct {
	client.Client  // kubernetes client
	kclient.Handle // manage kubernetes client & enable mocking
}

// kubeclientBuildOption defines the abstraction to build
// a kubeclient instance
type kubeclientBuildOption func(*kubeclient)

func withDefaults(k *kubeclient) error {
	if k.Client == nil {
		cli, err := kclient.New()
		if err != nil {
			return err
		}
		k.Client = cli
	}
	return nil
}

// WithKubeClient sets the kubernetes client against
// the kubeclient instance
func WithKubeClient(c client.Client) kubeclientBuildOption {
	return func(k *kubeclient) {
		k.Client = c
	}
}

// KubeClient returns a new instance of kubeclient meant for
// kubeassert operations
func KubeClient(opts ...kubeclientBuildOption) (*kubeclient, error) {
	k := &kubeclient{}
	for _, o := range opts {
		o(k)
	}
	err := withDefaults(k)
	if err != nil {
		return nil, err
	}
	return k, nil
}

// compile time check to ensure kubeclient
// implements catalog Interface
var _ Interface = &kubeclient{}

// getClientOrCached returns either a new instance
// of kubernetes client or its cached copy
func (k *kubeclient) getClientOrCached() (client.Client, error) {
	if k.Client != nil {
		return k.Client, nil
	}
	cli, err := k.GetClientFn()
	if err != nil {
		return nil, err
	}
	k.Client = cli
	return k.Client, nil
}

// Get returns the kubeassert instance from
// kubernetes cluster
func (k *kubeclient) Get(name string, opts ...provider.GetOptionFn) (*apis.KubeAssert, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("failed to get kubeassert: missing kubeassert name")
	}
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	getopt := provider.NewGetOptions(opts...)
	var ka apis.KubeAssert
	key := client.ObjectKey{Namespace: getopt.Namespace, Name: name}
	err = k.GetFn(cli, context.TODO(), key, &ka)
	if err != nil {
		return nil, err
	}
	return &ka, nil
}
