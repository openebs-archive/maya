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
	provider "github.com/openebs/maya/pkg/provider/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Service returns service implementors of unstruct
//
// NOTE:
//  Kubernetes is currently the only service provider
func Service(p *provider.Provider) (Interface, error) {
	if p == nil {
		return nil, errors.New("failed to get unstruct service: nil provider")
	}
	switch p.Type {
	default:
		return nil, errors.Errorf("failed to get unstruct service: unsupported provider '%v'", p.Type)
	case provider.Kubernetes:
		return KubeClient(WithKubeClient(p.KubeClient))
	}
}

// Interface exposes all operations from unstruct
// namespace
type Interface interface {
	Operations
}

// Operations abstracts operations against
// unstruct instance(s)
type Operations interface {
	// Get fetches a unstruct instance
	Get(name string, opts ...provider.GetOptionFn) (*unstructured.Unstructured, error)

	// Create creates an unstructured instance
	Create(u *unstructured.Unstructured) error

	// CreateAllOrNone creates the provided
	// unstructured instances
	CreateAllOrNone(u ...*unstructured.Unstructured) []error

	// Delete deletes the unstructured instance
	Delete(u *unstructured.Unstructured) error

	// DeleteAll deletes the provided
	// unstructured instances
	DeleteAll(u ...*unstructured.Unstructured) []error
}
