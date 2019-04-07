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

	apis "github.com/openebs/maya/pkg/apis/openebs.io/catalog/v1alpha1"
	kclient "github.com/openebs/maya/pkg/kubernetes/client/v1alpha2"
	provider "github.com/openebs/maya/pkg/provider/v1alpha1"
	unstruct "github.com/openebs/maya/pkg/unstruct/v1alpha1"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// kubeclient enables kubernetes API operations on catalog
// instance
type kubeclient struct {
	client.Client   // kubernetes client
	*kclient.Handle // manage kubernetes client & enable mocking
}

// kubeclientBuildOption defines the abstraction to build
// a kubeclient instance
type kubeclientBuildOption func(*kubeclient)

func (k *kubeclient) withDefaults() error {
	if k.Handle == nil {
		handle, err := kclient.New()
		if err != nil {
			return err
		}
		k.Handle = handle
	}
	if k.Client == nil {
		cli, err := k.Handle.GetClientFn()
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
// catalog operations
func KubeClient(opts ...kubeclientBuildOption) (*kubeclient, error) {
	k := &kubeclient{}
	for _, o := range opts {
		o(k)
	}
	err := k.withDefaults()
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

// Get returns a catalog instance from kubernetes cluster
func (k *kubeclient) Get(name string, opts ...provider.GetOptionFn) (*apis.Catalog, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("failed to get catalog: missing catalog name")
	}
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	getopt := provider.NewGetOptions(opts...)
	var cat apis.Catalog
	key := client.ObjectKey{Namespace: getopt.Namespace, Name: name}
	err = k.GetFn(cli, context.TODO(), key, &cat)
	if err != nil {
		return nil, err
	}
	return &cat, nil
}

// CreateAllResourcesOrNone creates all the provided
// resources at kubernetes cluster or none in case
// of any error
func (k *kubeclient) CreateAllResourcesOrNone(r ...apis.CatalogResource) []error {
	var (
		errs    []error
		created []apis.CatalogResource
	)
	for _, resource := range r {
		err := k.CreateResource(resource)
		if err != nil {
			errs = append(errs, err)
			break
		}
		created = append(created, resource)
	}
	if len(errs) > 0 {
		k.DeleteAllResources(created...)
	}
	return errs
}

// DeleteAllResources deletes all the provided
// resources at kubernetes cluster
func (k *kubeclient) DeleteAllResources(r ...apis.CatalogResource) []error {
	var errs []error
	for _, resource := range r {
		err := k.DeleteResource(resource)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// CreateResource creates a resource at kubernetes cluster
func (k *kubeclient) CreateResource(r apis.CatalogResource) error {
	cli, err := k.getClientOrCached()
	if err != nil {
		return err
	}
	u, err := unstruct.BuilderForYaml(r.Template).Build()
	if err != nil {
		return err
	}
	return k.CreateFn(cli, context.TODO(), u.GetUnstructured())
}

// DeleteResource deletes the resource from kubernetes cluster
func (k *kubeclient) DeleteResource(r apis.CatalogResource) error {
	cli, err := k.getClientOrCached()
	if err != nil {
		return err
	}
	u, err := unstruct.BuilderForYaml(r.Template).Build()
	if err != nil {
		return err
	}
	return k.DeleteFn(cli, context.TODO(), u.GetUnstructured())
}
