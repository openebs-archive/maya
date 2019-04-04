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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// supported represents the supported provider(s)
type supported int

const (
	// Kubernetes represents kubernetes as a supported
	// provider
	Kubernetes supported = iota
)

// buildOptionFn defines the abstraction to
// build a Provider instance
type buildOptionFn func(*Provider)

// Provider represents a service provider capable of
// serving various features
type Provider struct {
	Type       supported     // specific type of provider
	KubeClient client.Client // client if this is a kubernetes provider; is optional
}

// IsKubernetes marks the given provider as a
// Kubernetes specific provider
func IsKubernetes() buildOptionFn {
	return func(p *Provider) {
		p.Type = Kubernetes
	}
}

// WithKubeClient sets kubernetes client
func WithKubeClient(c client.Client) buildOptionFn {
	return func(p *Provider) {
		p.KubeClient = c
	}
}

// New returns a new Provider instance
func New(opts ...buildOptionFn) *Provider {
	p := &Provider{}
	for _, o := range opts {
		o(p)
	}
	return p
}

// GetOptionFn defines the abstraction to
// build a GetOptions instance
type GetOptionFn func(*GetOptions)

// WithGetNamespace sets the namespace against
// the provided GetOptions instance
func WithGetNamespace(namespace string) GetOptionFn {
	return func(opt *GetOptions) {
		opt.Namespace = namespace
	}
}

// WithGroupVersionKind sets the GKV against
// the provided GetOptions instance
func WithGroupVersionKind(gkv schema.GroupVersionKind) GetOptionFn {
	return func(opt *GetOptions) {
		opt.GKV = gkv
	}
}

// GetOptions consists of the options required
// during a get call
type GetOptions struct {
	Namespace string // scope within which 'get' call is executed
	GKV       schema.GroupVersionKind
}

// NewGetOptions returns a new instance of GetOptions
func NewGetOptions(opts ...GetOptionFn) *GetOptions {
	g := &GetOptions{}
	for _, o := range opts {
		o(g)
	}
	return g
}
