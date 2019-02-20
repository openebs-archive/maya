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
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type unmarshalType int

const (
	// UnmarshalYaml will lead to yaml based
	// unmarshaling of document
	UnmarshalYaml unmarshalType = iota + 1
)

// buildOption defines the abstraction to
// build an unstruct instance
type buildOption func(*unstruct)

type unstruct struct {
	object *unstructured.Unstructured
}

func (u *unstruct) apply(opts ...buildOption) *unstructured.Unstructured {
	for _, o := range opts {
		o(u)
	}
	return u.object
}

// Unmarshal returns the unstructured instance of
// the provided document. In addition, options if
// any are applied against this instance.
//
// NOTE:
//  Supports yaml format document only
func Unmarshal(doc string, opts ...buildOption) (*unstructured.Unstructured, error) {
	obj, err := unmarshalYaml(doc)
	if err != nil {
		return nil, err
	}
	u := &unstruct{object: obj}
	return u.apply(opts...), nil
}

// Object accepts unstructured instance as an input
// and returns the updated version of the same after
// applying the provided options.
func Object(in *unstructured.Unstructured, opts ...buildOption) *unstructured.Unstructured {
	u := &unstruct{object: in}
	return u.apply(opts...)
}

func unmarshalYaml(doc string) (*unstructured.Unstructured, error) {
	if doc == "" {
		return nil, errors.New("failed to create unstructured instance: empty doc")
	}
	m := map[string]interface{}{}
	err := yaml.Unmarshal([]byte(doc), &m)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create unstructured instance from doc:\n'%s'", doc)
	}
	return &unstructured.Unstructured{
		Object: m,
	}, nil
}
