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

package v1alpha2

import (
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// unstruct holds an object of unstructured
type unstruct struct {
	object *unstructured.Unstructured
}

// ToUnstructured converts unstruct object
// to API's unstructured
func (u *unstruct) ToUnstructured() *unstructured.Unstructured {
	return u.object
}

// builder enables building of an
// unstructured instance
type builder struct {
	unstruct *unstruct
	errs     []error
}

// UnstructBuilder returns a new instance of
// empty unstruct builder
func UnstructBuilder() *builder {
	return &builder{
		unstruct: &unstruct{
			&unstructured.Unstructured{
				Object: map[string]interface{}{},
			},
		},
	}
}

// UnstructBuilderForYaml returns a new instance of
// unstruct builder by making use of the provided
// YAML
func UnstructBuilderForYaml(doc string) *builder {
	b := UnstructBuilder()
	err := yaml.Unmarshal([]byte(doc), &b.unstruct.object.Object)
	b.errs = append(b.errs, err)
	return b
}

// UnstructBuilderForObject returns a new instance of
// unstruct builder by making use of the provided object
func UnstructBuilderForObject(obj *unstructured.Unstructured) *builder {
	b := UnstructBuilder()
	b.unstruct.object = obj
	return b
}

// Build returns the unstruct object created by
// the builder
func (b *builder) Build() (*unstruct, error) {
	if len(b.errs) != 0 {
		return nil, errors.Errorf("%v", b.errs)
	}
	return b.unstruct, nil
}
