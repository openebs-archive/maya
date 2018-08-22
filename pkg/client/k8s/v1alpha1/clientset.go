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
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
)

// ClientsetGetter abstracts fetching of kubernetes clientset
type ClientsetGetter interface {
	Get() (*kubernetes.Clientset, error)
}

// ClientsetGetterFunc is a functional implementation of ClientsetGetter
type ClientsetGetterFunc func() (*kubernetes.Clientset, error)

// Get is an implementation of ClientsetGetter
func (fn ClientsetGetterFunc) Get() (*kubernetes.Clientset, error) {
	return fn()
}

// NewClientsetGetter returns a ClientsetGetter instance that is capable of
// invoking kubernetes API calls
func NewClientsetGetter() ClientsetGetter {
	return ClientsetGetterFunc(func() (*kubernetes.Clientset, error) {
		config, err := NewClientConfigGetter().Get()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get kubernetes clientset")
		}

		return kubernetes.NewForConfig(config)
	})
}
