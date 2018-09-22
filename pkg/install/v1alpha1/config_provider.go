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
	k8s "github.com/openebs/maya/pkg/client/k8s/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

// ConfigProvider abstracts providing an instance of install config
type ConfigProvider interface {
	Provide() (ic *InstallConfig, err error)
}

type configFromK8s struct {
	namespace string              // namespace of install config
	name      string              // name of install config
	provider  k8s.ConfigMapGetter // kubernetes cluster provides the install config
}

// Config returns a new instance of ConfigProvider
func Config(namespace, name string) ConfigProvider {
	return &configFromK8s{name: name, namespace: namespace, provider: k8s.ConfigMap(namespace, name)}
}

// Provide provides the install config instance
func (c *configFromK8s) Provide() (ic *InstallConfig, err error) {
	if len(strings.TrimSpace(c.name)) == 0 {
		err = errors.Errorf("missing config name: failed to provide install config from config map: namespace %s", c.namespace)
		return
	}
	if c.provider == nil {
		err = errors.Errorf("nil config provider: failed to provide install config from config map %s %s", c.namespace, c.name)
		return
	}
	cm, err := c.provider.Get(metav1.GetOptions{})
	if err != nil {
		err = errors.Wrapf(err, "failed to provide install config from config map %s %s", c.namespace, c.name)
		return
	}
	if cm == nil {
		err = errors.Errorf("nil config map instance found: failed to provide install config %s %s", c.namespace, c.name)
		return
	}
	install := cm.Data["install"]
	if len(strings.TrimSpace(install)) == 0 {
		err = errors.Errorf("missing install config specs: failed to provide install config from config map %s %s", c.namespace, c.name)
		return
	}
	return UnmarshallConfig(install)
}
