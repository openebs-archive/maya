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

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// GetClientFunc abstracts fetching kubernetes client
type GetClientFunc func() (client.Client, error)

// GetFunc abstracts fetching any kubernetes resource
type GetFunc func(c client.Client, ctx context.Context, key client.ObjectKey, obj runtime.Object) error

// CreateFunc abstracts creating any kubernetes resource
type CreateFunc func(c client.Client, ctx context.Context, obj runtime.Object) error

// DeleteFunc abstracts deleting any kubernetes resource
type DeleteFunc func(c client.Client, ctx context.Context, obj runtime.Object) error

// Handle exposes each kubernetes client
// operation as a function
//
// NOTE:
//  This structure can be embedded inside
// other structs
//
// NOTE:
//  This enables mocking specific client
// operations
//
// NOTE:
//  It is not a good practice to use suffixes
// that reflect what the variable contains.
// e.g. here `Fn` or `Func` should not be used
// as suffixes. However, if we donot use these
// suffixes then it will steal good names from
// the structure(s) that will embed Handle.
type Handle struct {
	GetClientFn GetClientFunc // handle to fetch kubernetes client
	GetFn       GetFunc       // handle to fetch any kubernetes resource
	CreateFn    CreateFunc    // handle to create any kubernetes resource
	DeleteFn    DeleteFunc    // handle to delete any kubernetes resource
}

// withDefaults sets the defaults associated
// with the provided Handle instance
func withDefaults(h *Handle) {
	if h.GetClientFn == nil {
		h.GetClientFn = func() (client.Client, error) {
			conf, err := config.GetConfig()
			if err != nil {
				return nil, err
			}
			return client.New(conf, client.Options{})
		}
	}
	if h.GetFn == nil {
		h.GetFn = func(c client.Client, ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
			return c.Get(ctx, key, obj)
		}
	}
	if h.CreateFn == nil {
		h.CreateFn = func(c client.Client, ctx context.Context, obj runtime.Object) error {
			return c.Create(ctx, obj)
		}
	}
	if h.DeleteFn == nil {
		h.DeleteFn = func(c client.Client, ctx context.Context, obj runtime.Object) error {
			return c.Delete(ctx, obj)
		}
	}
}

// handleBuildOption defines the abstraction to
// build the client Handle instance
type handleBuildOption func(*Handle)

// New returns a new instance of kubernetes
// client
func New(opts ...handleBuildOption) (*Handle, error) {
	h := &Handle{}
	for _, o := range opts {
		o(h)
	}
	withDefaults(h)
	return h, nil
}
