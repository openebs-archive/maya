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
type ConfigMapGetter interface {
	Get(name string, options metav1.GetOptions) (*corev1.ConfigMap, error)
}

// ConfigMapGetterFunc is a functional implementation of ConfigMapGetter
type ConfigMapGetterFunc func(name string, options metav1.GetOptions) (*corev1.ConfigMap, error)

// Get is an implementation of ConfigMapGetter
func (fn ConfigMapGetterFunc) Get(name string, options metav1.GetOptions) (*corev1.ConfigMap, error) {
	return fn(name, options)
}

// NewConfigMapGetter returns a new instance of ConfigMapGetter that is capable
// of fetching a ConfigMap from kubernetes cluster
func NewConfigMapGetter(namespace string) ConfigMapGetter {
	return ConfigMapGetterFunc(func(name string, options metav1.GetOptions) (*corev1.ConfigMap, error) {
		if len(strings.TrimSpace(name)) == 0 {
			return nil, fmt.Errorf("missing config map name: failed to get config map")
		}

		cs, err := NewClientsetGetter().Get()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get config map")
		}

		return cs.CoreV1().ConfigMaps(namespace).Get(name, options)
	})
}
