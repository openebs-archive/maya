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

// TemplateKeyValue represents a key and corresponding value to be used as
// values for templating
type TemplateKeyValue struct {
	Key   string
	Value interface{}
}

type TemplateKeyValueList struct {
	Items []TemplateKeyValue
}

func NewTemplateKeyValueList() TemplateKeyValueList {
	return TemplateKeyValueList{}
}

// AddNamespace adds namespace as a template value
func (l TemplateKeyValueList) AddNamespace(value string) TemplateKeyValueList {
	l.Items = append(l.Items, TemplateKeyValue{Key: "namespace", Value: value})
	return l
}

// AddServiceAccount adds serviceaccount as a template value
func (l TemplateKeyValueList) AddServiceAccount(value string) TemplateKeyValueList {
	l.Items = append(l.Items, TemplateKeyValue{Key: "serviceaccount", Value: value})
	return l
}

// Values creates template values to be applied over install related artifacts
func (l TemplateKeyValueList) Values() (final map[string]interface{}) {
	final = map[string]interface{}{}
	if len(l.Items) == 0 {
		final["installer"] = nil
		return
	}
	nested := map[string]interface{}{}
	for _, kv := range l.Items {
		nested[kv.Key] = kv.Value
	}
	final["installer"] = nested
	return
}
