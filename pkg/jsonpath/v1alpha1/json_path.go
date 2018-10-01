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
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	_ "github.com/openebs/maya/pkg/msg/v1alpha1"
	ft "k8s.io/client-go/third_party/forked/golang/template"
	jp "k8s.io/client-go/util/jsonpath"
)

type selection struct {
	Name   string   `json:"name"`   // name of selection
	Path   string   `json:"path"`   // selection path used to build jsonpath query
	Values []string `json:"values"` // resulting values due to selection path
	*Msgs
}

// Selection generates a new selection struct.
func Selection(name, path string) *selection {
	return &selection{
		Name: name,
		Path: path,
		Msgs: &Msgs{},
	}
}

func (s *selection) SetValues(rvals [][]reflect.Value) {
	if len(rvals) == 0 {
		s.AddWarn(fmt.Sprintf("no value(s) found for %s %s", s.Name, s.Path))
		s.Values = append(s.Values, "<no value>")
		return
	}
	for _, rvs := range rvals {
		for _, rv := range rvs {
			pv, ok := ft.PrintableValue(rv)
			if !ok {
				s.AddWarn(fmt.Sprintf("can not print type %s: failed to query %s %s", rv.Type(), s.Name, s.Path))
				pv = "<not printable>"
			}
			var buffer bytes.Buffer
			fmt.Fprint(&buffer, pv)
			s.Values = append(s.Values, buffer.String())
		}
	}
	return
}

// SelectionList represents a list of selection.
type SelectionList []*selection

func (l SelectionList) String() string {
	return YamlString("selectionlist", l)
}

// ValuesByName retrieves a list of values filtered by name.
func (l SelectionList) ValuesByName(name string) (vals []string) {
	for _, s := range l {
		if s.Name == name {
			return s.Values
		}
	}
	return
}

// ValueByName retrieves a single value filtering by name.
func (l SelectionList) ValueByName(name string) (value string) {
	vals := l.ValuesByName(name)
	if len(vals) == 0 {
		return
	}
	return vals[0]
}

// ValuesByPath retrieves a list of values filtering by path.
func (l SelectionList) ValuesByPath(path string) (vals []string) {
	for _, s := range l {
		if s.Path == path {
			return s.Values
		}
	}
	return
}

// ValueByPath retrieves a single value according to the given path.
func (l SelectionList) ValueByPath(path string) (value string) {
	vals := l.ValuesByPath(path)
	if len(vals) == 0 {
		return
	}
	return vals[0]
}

type jsonpath struct {
	name    string        // name given to the json querying
	target  interface{}   // target to be queried against
	jpath   *jp.JSONPath  // instance that understands querying json doc
	selects SelectionList // selective queries to be done against json doc
	*Msgs
}

// JSONPath generates an new jsonpath struct.
func JSONPath(name string) (j *jsonpath) {
	return &jsonpath{
		name:  name,
		jpath: jp.New(name).AllowMissingKeys(true),
		Msgs:  &Msgs{},
	}
}

func (j *jsonpath) WithTarget(target interface{}) (u *jsonpath) {
	j.target = target
	return j
}

func (j *jsonpath) WithTargetAsRaw(target []byte) (u *jsonpath) {
	var t interface{}
	err := json.Unmarshal(target, &t)
	if err != nil {
		j.AddError(fmt.Errorf("failed to build target for jsonpath %s: error - %s", j.name, err.Error()))
		return j
	}
	j.target = t
	return j
}

func (j *jsonpath) Values(path string) (vals [][]reflect.Value, err error) {
	err = j.jpath.Parse(path)
	if err != nil {
		return
	}
	return j.jpath.FindResults(j.target)
}

func (j *jsonpath) Query(selects SelectionList) (l SelectionList) {
	for _, s := range selects {
		vals, err := j.Values(s.Path)
		if err != nil {
			j.AddError(fmt.Errorf("failed to query %s %s: error - %s", s.Name, s.Path, err.Error()))
		}
		s.SetValues(vals)
		l = append(l, s)
		j.Msgs.Merge(s.Msgs)
	}
	return
}
