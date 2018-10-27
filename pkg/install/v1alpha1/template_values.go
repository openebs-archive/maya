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

// TODO
// Check if this file is required !!! Remove if not required

//Package v1alpha1 - TODO
// Move this to pkg/template/v1alpha1 and rename appropriately if required
package v1alpha1

import (
	"github.com/openebs/maya/pkg/util"
)

// templatingPair represents a key and corresponding value to be used during
// go template execution
type templatingPair struct {
	Key   string
	Value interface{}
}

// templatingPairList represents a list of templatingPair
type templatingPairList struct {
	Items []templatingPair
}

// TemplatingPairList returns an empty list of templatingPairList
func TemplatingPairList() templatingPairList {
	return templatingPairList{}
}

// Add adds a templating pair
func (l templatingPairList) Add(v interface{}, k ...string) templatingPairList {
	if len(k) == 0 {
		return l
	}
	if len(k) == 1 {
		l.Items = append(l.Items, templatingPair{Key: k[0], Value: v})
		return l
	}
	kk := append(k[:0], k[1:]...)
	vv := map[string]interface{}{}
	util.SetNestedField(vv, v, kk...)
	l.Items = append(l.Items, templatingPair{Key: k[0], Value: vv})
	return l
}

// AddNamespace adds namespace as a templating pair
func (l templatingPairList) AddNamespace(value string) templatingPairList {
	return l.Add(value, "namespace")
}

// AddServiceAccount adds serviceaccount as a template value
func (l templatingPairList) AddServiceAccount(value string) templatingPairList {
	return l.Add(value, "serviceaccount")
}

// AsMap returns a map of templating pairs that can be used during go template
// execution
func (l templatingPairList) AsMap(rootKey ...string) (final map[string]interface{}) {
	final = map[string]interface{}{}
	if len(l.Items) == 0 && len(rootKey) == 0 {
		return
	}
	nested := map[string]interface{}{}
	for _, pair := range l.Items {
		nested[pair.Key] = pair.Value
	}
	if len(rootKey) == 0 {
		return nested
	}
	final[rootKey[0]] = nested
	return
}
