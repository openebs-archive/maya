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

// TODO Integration Tests
// 1/ Catalog in namespace A, while its resources are in namespace B & C
// - expect - resources get created
// - expect - resources get deleted

package v1alpha1

import (
	"testing"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/catalog/v1alpha1"
	kclient "github.com/openebs/maya/pkg/kubernetes/client/v1alpha2"
	provider "github.com/openebs/maya/pkg/provider/v1alpha1"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	fakeGetClientOk  = kclient.FakeGetClientOk()
	fakeGetClientErr = kclient.FakeGetClientErr()
	fakeGetOk        = kclient.FakeGetOk()
	fakeGetErr       = kclient.FakeGetErr()
	fakeCreateOk     = kclient.FakeCreateOk()
	fakeCreateErr    = kclient.FakeCreateErr()
	fakeDeleteOk     = kclient.FakeDeleteOk()
	fakeDeleteErr    = kclient.FakeDeleteErr()
)

func TestKubeClient(t *testing.T) {
	tests := map[string]struct {
		opts []kubeclientBuildOption
	}{
		"t1": {[]kubeclientBuildOption{WithKubeClient(fakeclient.NewFakeClient())}},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			k, _ := KubeClient(mock.opts...)
			if k == nil {
				t.Fatalf("test '%s' failed: expected not nil kubeclient actual nil", name)
			}
		})
	}
}

func TestGet(t *testing.T) {
	tests := map[string]struct {
		name       string
		getClient  kclient.GetClientFunc
		get        kclient.GetFunc
		getOptions []provider.GetOptionFn
		isErr      bool
	}{
		"t10": {"mytask", fakeGetClientOk, fakeGetOk, nil, false},
		"t11": {"", fakeGetClientOk, fakeGetOk, nil, true},
		"t12": {"mytask", fakeGetClientErr, nil, nil, true},
		"t13": {"mytask", fakeGetClientOk, fakeGetErr, nil, true},
		"t20": {"mytask", fakeGetClientOk, fakeGetOk, []provider.GetOptionFn{provider.WithGetNamespace("d")}, false},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			k := &kubeclient{
				Handle: &kclient.Handle{
					GetClientFn: mock.getClient,
					GetFn:       mock.get,
				},
			}
			_, err := k.Get(mock.name, mock.getOptions...)
			if mock.isErr && err == nil {
				t.Fatalf("test '%s' failed: expected error actual no error", name)
			}
			if !mock.isErr && err != nil {
				t.Fatalf("test '%s' failed: expected no error actual '%s'", name, err)
			}
		})
	}
}

func TestCreateAllResourcesOrNone(t *testing.T) {
	tests := map[string]struct {
		resources []apis.CatalogResource
		getClient kclient.GetClientFunc
		create    kclient.CreateFunc
		isErr     bool
	}{
		"t10": {[]apis.CatalogResource{apis.CatalogResource{}}, fakeGetClientOk, fakeCreateOk, false},
		"t11": {[]apis.CatalogResource{apis.CatalogResource{Template: "kind: Ping"}}, fakeGetClientErr, nil, true},
		"t12": {[]apis.CatalogResource{apis.CatalogResource{Template: "kind: Hi"}}, fakeGetClientOk, fakeCreateErr, true},
		"t13": {[]apis.CatalogResource{apis.CatalogResource{Template: "kind: Namaste"}}, fakeGetClientOk, fakeCreateOk, false},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			k := &kubeclient{
				Handle: &kclient.Handle{
					GetClientFn: mock.getClient,
					CreateFn:    mock.create,
				},
			}
			err := k.CreateAllResourcesOrNone(mock.resources...)
			if mock.isErr && err == nil {
				t.Fatalf("test '%s' failed: expected error actual no error", name)
			}
			if !mock.isErr && err != nil {
				t.Fatalf("test '%s' failed: expected no error actual '%s'", name, err)
			}
		})
	}
}

func TestCreateResource(t *testing.T) {
	tests := map[string]struct {
		resource  apis.CatalogResource
		getClient kclient.GetClientFunc
		create    kclient.CreateFunc
		isErr     bool
	}{
		"t10": {apis.CatalogResource{}, fakeGetClientOk, fakeCreateOk, false},
		"t11": {apis.CatalogResource{Template: "kind: Ping"}, fakeGetClientErr, nil, true},
		"t12": {apis.CatalogResource{Template: "kind: Hi"}, fakeGetClientOk, fakeCreateErr, true},
		"t13": {apis.CatalogResource{Template: "kind: Namaste"}, fakeGetClientOk, fakeCreateOk, false},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			k := &kubeclient{
				Handle: &kclient.Handle{
					GetClientFn: mock.getClient,
					CreateFn:    mock.create,
				},
			}
			err := k.CreateResource(mock.resource)
			if mock.isErr && err == nil {
				t.Fatalf("test '%s' failed: expected error actual no error", name)
			}
			if !mock.isErr && err != nil {
				t.Fatalf("test '%s' failed: expected no error actual '%s'", name, err)
			}
		})
	}
}

func TestDeleteAllResources(t *testing.T) {
	tests := map[string]struct {
		resources []apis.CatalogResource
		getClient kclient.GetClientFunc
		delete    kclient.DeleteFunc
		isErr     bool
	}{
		"t10": {[]apis.CatalogResource{apis.CatalogResource{}}, fakeGetClientOk, fakeDeleteOk, false},
		"t11": {[]apis.CatalogResource{apis.CatalogResource{Template: "kind: Pong"}}, fakeGetClientErr, nil, true},
		"t12": {[]apis.CatalogResource{apis.CatalogResource{Template: "kind: Bye"}}, fakeGetClientOk, fakeDeleteErr, true},
		"t13": {[]apis.CatalogResource{apis.CatalogResource{Template: "kind: Namaste"}}, fakeGetClientOk, fakeDeleteOk, false},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			k := &kubeclient{
				Handle: &kclient.Handle{
					GetClientFn: mock.getClient,
					DeleteFn:    mock.delete,
				},
			}
			err := k.DeleteAllResources(mock.resources...)
			if mock.isErr && err == nil {
				t.Fatalf("test '%s' failed: expected error actual no error", name)
			}
			if !mock.isErr && err != nil {
				t.Fatalf("test '%s' failed: expected no error actual '%s'", name, err)
			}
		})
	}
}

func TestDeleteResource(t *testing.T) {
	tests := map[string]struct {
		resource  apis.CatalogResource
		getClient kclient.GetClientFunc
		delete    kclient.DeleteFunc
		isErr     bool
	}{
		"t10": {apis.CatalogResource{}, fakeGetClientOk, fakeDeleteOk, false},
		"t11": {apis.CatalogResource{Template: "kind: Pong"}, fakeGetClientErr, nil, true},
		"t12": {apis.CatalogResource{Template: "kind: Bye"}, fakeGetClientOk, fakeDeleteErr, true},
		"t13": {apis.CatalogResource{Template: "kind: Namaste"}, fakeGetClientOk, fakeDeleteOk, false},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			k := &kubeclient{
				Handle: &kclient.Handle{
					GetClientFn: mock.getClient,
					DeleteFn:    mock.delete,
				},
			}
			err := k.DeleteResource(mock.resource)
			if mock.isErr && err == nil {
				t.Fatalf("test '%s' failed: expected error actual no error", name)
			}
			if !mock.isErr && err != nil {
				t.Fatalf("test '%s' failed: expected no error actual '%s'", name, err)
			}
		})
	}
}
