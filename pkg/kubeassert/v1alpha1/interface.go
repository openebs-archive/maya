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
	apis "github.com/openebs/maya/pkg/apis/openebs.io/kubeassert/v1alpha1"
	provider "github.com/openebs/maya/pkg/provider/v1alpha1"
	"github.com/pkg/errors"
)

// Service returns service implementors of catalog
//
// NOTE:
//  Kubernetes is currently the only service provider
// for catalog
func Service(p *provider.Provider) (Interface, error) {
	if p == nil {
		return nil, errors.New("failed to get kubeassert service: nil provider")
	}
	switch p.Type {
	default:
		return nil, errors.Errorf("failed to get kubeassert service: unsupported provider '%v'", p.Type)
	case provider.Kubernetes:
		return KubeClient(WithKubeClient(p.KubeClient))
	}
}

// Interface exposes all operations from kubeassert
// namespace
type Interface interface {
	Operations
}

// Operations abstracts operations against kubeassert
// instance
type Operations interface {
	// Get fetches the required kubeassert instance
	Get(name string, opts ...provider.GetOptionFn) (*apis.KubeAssert, error)
}
