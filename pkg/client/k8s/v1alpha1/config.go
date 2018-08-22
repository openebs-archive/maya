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

// K8sAPIConfiguration is a typed string to represent various configurations used
// to invoke kubernetes APIs
type K8sAPIConfiguration string

const (
	// EnvironmentKeyForK8sMasterIP is the environment variable key used to
	// determine the kubernetes master IP address
	EnvironmentKeyForK8sMasterIP K8sAPIConfiguration = "OPENEBS_IO_K8S_MASTER"
	// EnvironmentKeyForKubeConfig is the environment variable key used to
	// determine the kubernetes config
	EnvironmentKeyForKubeConfig K8sAPIConfiguration = "OPENEBS_IO_KUBE_CONFIG"
)

// ConfigGetter abstracts fetching of kubernetes client config
type ConfigGetter interface {
	Get() (*rest.Config, error)
}

// ConfigGetterFunc is a functional implementation of ConfigGetter
type ConfigGetterFunc func() (*rest.Config, error)

// Get is an implementation of ConfigGetter
func (fn ConfigGetterFunc) Get() (*rest.Config, error) {
	return fn()
}

// WithEnvConfigGetter returns kubernetes rest config based on kubernetes
// environment values
func WithEnvConfigGetter() ConfigGetter {
	return ConfigGetterFunc(func() (*rest.Config, error) {
		k8sMaster := env.Get(string(EnvironmentKeyForK8sMasterIP))
		kubeConfig := env.Get(string(EnvironmentKeyForKubeConfig))

		if len(strings.TrimSpace(k8sMaster)) == 0 && len(strings.TrimSpace(kubeConfig)) == 0 {
			return nil, fmt.Errorf("missing kubernetes master as well as kubeconfig: failed to get rest config")
		}

		return clientcmd.BuildConfigFromFlags(k8sMaster, kubeConfig)
	})
}

// WithRestConfigGetter returns kubernetes rest config based on
// in cluster config implementation
func WithRestConfigGetter() ConfigGetter {
	return ConfigGetterFunc(func() (*rest.Config, error) {
		return rest.InClusterConfig()
	})
}

// NewClientConfigGetter fetches the kubernetes client config that is used to
// make kubernetes API calls
//
// NOTE:
//  This makes use of multiple strategies to get the client config instance
func NewClientConfigGetter() ConfigGetter {
	return ConfigGetterFunc(func() (config *rest.Config, err error) {
		var allErrors []error

		strategies := []ConfigGetter{
			WithEnvConfigGetter(),
			WithRestConfigGetter(),
		}

		for idx, s := range strategies {
			config, err = s.Get()
			if err == nil {
				// no error means this getter has succeeded
				return
			}
			allErrors = append(allErrors, errors.Wrapf(err, "failed to get kubernetes client config via strategy '%d'", idx))
		}

		// all strategies failed
		err = fmt.Errorf("%+v", allErrors)
		err = errors.Wrap(err, "failed to get kubernetes client config")
		return
	})
}
