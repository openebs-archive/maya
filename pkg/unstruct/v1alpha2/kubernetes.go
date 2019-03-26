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

package v1alpha2

import (
	"context"
	"strings"

	kclient "github.com/openebs/maya/pkg/kubernetes/client/v1alpha2"
	provider "github.com/openebs/maya/pkg/provider/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// kubeclient enables kubernetes API operations on catalog
// instance
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

// WithClient sets the kubernetes client against
// the kubeclient instance
func WithClient(c client.Client) kubeclientBuildOption {
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
	err := withDefaults(k)
	if err != nil {
		return nil, err
	}
	return k, nil
}

// compile time check to ensure kubeclient
// implements unstruct Interface
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

// Get returns an unstructured instance from kubernetes
// cluster
func (k *kubeclient) Get(name string, opts ...provider.GetOptionFn) (*unstructured.Unstructured, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("failed to get unstructured instance: missing name")
	}
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	getopt := provider.NewGetOptions(opts...)
	var un unstructured.Unstructured
	key := client.ObjectKey{Namespace: getopt.Namespace, Name: name}
	err = k.GetFn(cli, context.TODO(), key, &un)
	if err != nil {
		return nil, err
	}
	return &un, nil
}

// CreateAllOrNone creates all the provided
// unstructured instances at kubernetes cluster
// or none in case of any error
func (k *kubeclient) CreateAllOrNone(u ...*unstructured.Unstructured) []error {
	var (
		errs    []error
		created []*unstructured.Unstructured
	)
	for _, ustruct := range u {
		err := k.Create(ustruct)
		if err != nil {
			errs = append(errs, err)
			break
		}
		created = append(created, ustruct)
	}
	if len(errs) > 0 {
		k.DeleteAll(created...)
	}
	return errs
}

// DeleteAll deletes all the provided unstructured
// instances at kubernetes cluster
func (k *kubeclient) DeleteAll(u ...*unstructured.Unstructured) []error {
	var errs []error
	for _, ustruct := range u {
		err := k.Delete(ustruct)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// Create creates an unstructured instance at
// kubernetes cluster
func (k *kubeclient) Create(u *unstructured.Unstructured) error {
	cli, err := k.getClientOrCached()
	if err != nil {
		return err
	}
	return k.CreateFn(cli, context.TODO(), u)
}

// Delete deletes the unstructured instance from
// kubernetes cluster
func (k *kubeclient) Delete(u *unstructured.Unstructured) error {
	cli, err := k.getClientOrCached()
	if err != nil {
		return err
	}
	return k.DeleteFn(cli, context.TODO(), u)
}
