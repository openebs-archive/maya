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
	"github.com/pkg/errors"
	"strings"

	k8s "github.com/openebs/maya/pkg/client/k8s/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConfigGetterFunc abstracts fetching an instance of install config
type ConfigGetterFunc func(name string) (config *InstallConfig, err error)

// WithConfigMapConfigGetter returns an instance of ConfigGetterFunc that is
// capable of fetching install config from a kubernetes ConfigMap
//
// NOTE:
//  The name of the install config is also the name of the ConfigMap that embeds
// this install config specifications.
func WithConfigMapConfigGetter(getter k8s.ConfigMapGetter) ConfigGetterFunc {
	return func(name string) (config *InstallConfig, err error) {
		if len(strings.TrimSpace(name)) == 0 {
			return nil, fmt.Errorf("missing config map name: failed to get install config from config map")
		}

		if getter == nil {
			return nil, fmt.Errorf("nil config map getter: failed to get install config from config map")
		}

		cm, err := getter.Get(name, metav1.GetOptions{})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get install config from config map '%s'", name)
		}

		if cm == nil {
			return nil, fmt.Errorf("nil config map instance found: failed to get install config from config map '%s'", name)
		}

		installSpecs := cm.Data["install"]
		if len(strings.TrimSpace(installSpecs)) == 0 {
			return nil, fmt.Errorf("missing install config specs: failed to get install config from config map '%s'", name)
		}

		return UnmarshallConfig(installSpecs)
	}
}
