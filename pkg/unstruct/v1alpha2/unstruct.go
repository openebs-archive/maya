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

// Unstruct holds an object of Unstructured
type Unstruct struct {
	object *unstructured.Unstructured
}

// GetUnstructured converts Unstruct object
// to API's Unstructured
func (u *Unstruct) GetUnstructured() *unstructured.Unstructured {
	return u.object
}

// Builder enables building of an
// Unstructured instance
type Builder struct {
	Unstruct *Unstruct
	errs     []error
}

// NewBuilder returns a new instance of
// empty Builder
func NewBuilder() *Builder {
	return &Builder{
		Unstruct: &Unstruct{
			&unstructured.Unstructured{
				Object: map[string]interface{}{},
			},
		},
	}
}

// BuilderForYaml returns a new instance of
// Unstruct Builder by making use of the provided
// YAML
func BuilderForYaml(doc string) *Builder {
	b := NewBuilder()
	err := yaml.Unmarshal([]byte(doc), &b.Unstruct.object)
	if err != nil {
		b.errs = append(b.errs, err)
	}
	return b
}

// BuilderForObject returns a new instance of
// Unstruct Builder by making use of the provided object
func BuilderForObject(obj *unstructured.Unstructured) *Builder {
	b := NewBuilder()
	b.Unstruct.object = obj
	return b
}

// Build returns the Unstruct object created by
// the Builder
func (b *Builder) Build() (*Unstruct, error) {
	if len(b.errs) != 0 {
		return nil, errors.Errorf("%v", b.errs)
	}
	return b.Unstruct, nil
}

// UnstructList contains a list of Unstructured
// items
type UnstructList struct {
	items []*Unstruct
}

// ListBuilder enables building a list
// of an Unstruct instance
type ListBuilder struct {
	list UnstructList
	errs []error
}

// ListBuilderForYamls returns a mew instance of
// list Unstruct Builder by making use of the provided YAMLs
func ListBuilderForYamls(docs string) *ListBuilder {
	lb := &ListBuilder{list: UnstructList{}}
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
// list Unstruct Builder by making use of the provided
// Unstructured object
func ListBuilderForObjects(objs ...*unstructured.Unstructured) *ListBuilder {
	lb := &ListBuilder{}
	for _, obj := range objs {
		lb.list.items = append(lb.list.items, &Unstruct{obj})
	}
	return lb
}

// Build returns the list of Unstruct objects created by
// the Builder
func (l *ListBuilder) Build() ([]*Unstruct, error) {
	if len(l.errs) > 0 {
		return nil, errors.Errorf("%v", l.errs)
	}
	return l.list.items, nil
}
