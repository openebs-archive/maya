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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

// ConfigMapGetter abstracts fetching of ConfigMap instance from kubernetes
// cluster
type ConfigMapGetter func(name string, options metav1.GetOptions) (*corev1.ConfigMap, error)

// NewConfigMapGetter returns a new instance of ConfigMapGetter that is capable
// of fetching a ConfigMap from kubernetes cluster
func NewConfigMapGetter(namespace string) ConfigMapGetter {
	return func(name string, options metav1.GetOptions) (*corev1.ConfigMap, error) {
		if len(strings.TrimSpace(name)) == 0 {
			return nil, fmt.Errorf("missing config map name: failed to get config map from namespace '%s'", namespace)
		}

		cs, err := NewClientsetGetter()()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get config map '%s' from namespace '%s'", name, namespace)
		}

		cm, err := cs.CoreV1().ConfigMaps(namespace).Get(name, options)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get config map '%s' from namespace '%s'", name, namespace)
		}

		return cm, nil
	}
}
