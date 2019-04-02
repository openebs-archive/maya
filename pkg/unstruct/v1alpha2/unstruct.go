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
	"strings"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// unstruct holds an object of unstructured
type unstruct struct {
	object *unstructured.Unstructured
}

// GetUnstructured converts unstruct object
// to API's unstructured
func (u *unstruct) GetUnstructured() *unstructured.Unstructured {
	return u.object
}

// builder enables building of an
// unstructured instance
type builder struct {
	unstruct *unstruct
	errs     []error
}

// Builder returns a new instance of
// empty builder
func Builder() *builder {
	return &builder{
		unstruct: &unstruct{
			&unstructured.Unstructured{
				Object: map[string]interface{}{},
			},
		},
	}
}

// BuilderForYaml returns a new instance of
// unstruct builder by making use of the provided
// YAML
func BuilderForYaml(doc string) *builder {
	b := Builder()
	err := yaml.Unmarshal([]byte(doc), &b.unstruct.object)
	if err != nil {
		b.errs = append(b.errs, err)
	}
	return b
}

// BuilderForObject returns a new instance of
// unstruct builder by making use of the provided object
func BuilderForObject(obj *unstructured.Unstructured) *builder {
	b := Builder()
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

// unstructList contains a list of unstructured
// items
type unstructList struct {
	items []*unstruct
}

// listBuilder enables building a list
// of an unstruct instance
type listBuilder struct {
	list unstructList
	errs []error
}

// ListBuilderForYamls returns a mew instance of
// list unstruct builder by making use of the provided YAMLs
func ListBuilderForYamls(docs string) *listBuilder {
	lb := &listBuilder{list: unstructList{}}
	yamls := strings.Split(docs, "---")
	for _, f := range yamls {
		if f == "\n" || f == "" {
			// ignore empty cases
			continue
		}
		f = strings.TrimSpace(f)
		a, err := BuilderForYaml(f).Build()
		if err != nil {
			lb.errs = append(lb.errs, err)
			continue
		}
		lb.list.items = append(lb.list.items, a)
	}
	return lb
}

// ListBuilderForObjects returns a mew instance of
// list unstruct builder by making use of the provided
// unstructured object
func ListBuilderForObjects(objs ...*unstructured.Unstructured) *listBuilder {
	lb := &listBuilder{}
	for _, obj := range objs {
		lb.list.items = append(lb.list.items, &unstruct{obj})
	}
	return lb
}

// Build returns the list of unstruct objects created by
// the builder
func (l *listBuilder) Build() ([]*unstruct, error) {
	if len(l.errs) > 0 {
		return nil, errors.Errorf("%v", l.errs)
	}
	return l.list.items, nil
}
