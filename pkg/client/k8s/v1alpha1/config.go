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
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	env "github.com/openebs/maya/pkg/env/v1alpha1"
)

// ConfigGetter abstracts fetching of kubernetes client config
type ConfigGetter func() (*rest.Config, error)

// WithEnvConfigGetter returns kubernetes rest config based on kubernetes
// environment values
func WithEnvConfigGetter() ConfigGetter {
	return func() (*rest.Config, error) {
		k8sMaster := env.Get(K8sMasterIPEnvironmentKey)
		kubeConfig := env.Get(KubeConfigEnvironmentKey)

		if len(strings.TrimSpace(k8sMaster)) == 0 && len(strings.TrimSpace(kubeConfig)) == 0 {
			return nil, fmt.Errorf("missing kubernetes master as well as kubeconfig: failed to get rest config")
		}

		return clientcmd.BuildConfigFromFlags(k8sMaster, kubeConfig)
	}
}

// WithRestConfigGetter returns kubernetes rest config based on
// in cluster config implementation
func WithRestConfigGetter() ConfigGetter {
	return func() (*rest.Config, error) {
		return rest.InClusterConfig()
	}
}

// newClientConfigGetter fetches the kubernetes client config that is used to
// make kubernetes API calls
//
// NOTE:
//  This makes use of multiple strategies to get the client config instance
func newClientConfigGetter(strategies map[string]ConfigGetter) ConfigGetter {
	return func() (config *rest.Config, err error) {
		var allErrors []error

		for name, strategy := range strategies {
			config, err = strategy()
			if err == nil {
				// no error means this succeeded
				return
			}
			allErrors = append(allErrors, errors.Wrapf(err, "failed to get kubernetes client config via strategy '%s'", name))
		}

		// all strategies failed
		err = fmt.Errorf("%+v", allErrors)
		err = errors.Wrap(err, "failed to get kubernetes client config")
		return
	}
}

// NewClientConfigGetter fetches the kubernetes client config that is used to
// make kubernetes API calls
func NewClientConfigGetter() ConfigGetter {
	strategies := map[string]ConfigGetter{
		"env-based":  WithEnvConfigGetter(),
		"rest-based": WithRestConfigGetter(),
	}

	return newClientConfigGetter(strategies)
}
