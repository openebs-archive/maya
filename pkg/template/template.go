/*
Copyright 2017 The OpenEBS Authors
Copyright 2016 The Kubernetes Authors

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

package template

import (
	"bytes"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/ghodss/yaml"
)

// ToYaml takes a map, marshals it to yaml, and returns a string. It will
// always return a string, even on marshal error (empty string).
//
// This is designed to be called from a template.
// NOTE: Borrowed from a similar function in helm
// https://github.com/kubernetes/helm/blob/master/pkg/chartutil/files.go
func toYaml(v map[string]interface{}) string {
	data, err := yaml.Marshal(v)
	if err != nil {
		// Swallow errors inside of a template.
		return ""
	}
	return string(data)
}

// fromYaml converts a YAML document into a map[string]interface{}.
//
// This is not a general-purpose YAML parser, and will not parse all valid
// YAML documents. Additionally, because its intended use is within templates
// it tolerates errors. It will insert the returned error message string into
// m["Error"] in the returned map.
//
// NOTE: Borrowed from helm
// https://github.com/kubernetes/helm/blob/master/pkg/chartutil/files.go
func fromYaml(str string) map[string]interface{} {
	m := map[string]interface{}{}

	if err := yaml.Unmarshal([]byte(str), &m); err != nil {
		m["Error"] = err.Error()
	}
	return m
}

// funcMap returns the set of template functions supported in this library
func funcMap() template.FuncMap {
	f := sprig.TxtFuncMap()

	// Add some extra templating functions
	extra := template.FuncMap{
		"toYaml":   toYaml,
		"fromYaml": fromYaml,
	}

	for k, v := range extra {
		f[k] = v
	}

	return f
}

// AsTemplatedBytes returns a byte slice
// based on the provided yaml & values
func AsTemplatedBytes(context string, yml string, values map[string]interface{}) ([]byte, error) {
	tpl := template.New(context + "YamlTpl")

	// Any maya yaml exposes below templating functions
	tpl.Funcs(funcMap())

	tpl, err := tpl.Parse(yml)
	if err != nil {
		return nil, err
	}

	// buf is an io.Writer implementation
	// as required by the template
	var buf bytes.Buffer

	// execute the parsed yaml against this instance
	// & write the result into the buffer
	err = tpl.Execute(&buf, values)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// AsMapOfObjects returns a map of objects based on the provided yaml & values
func AsMapOfObjects(yml string, values map[string]interface{}) (map[string]interface{}, error) {
	// templated & then unmarshall-ed version of this yaml
	b, err := AsTemplatedBytes("MapOfObjects", yml, values)
	if err != nil {
		return nil, err
	}

	// Any given YAML can be unmarshalled into a map of arbitrary objects
	var obj map[string]interface{}
	err = yaml.Unmarshal(b, &obj)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

// AsMapOfStrings returns a map of strings based on the provided yaml & values
func AsMapOfStrings(context string, yml string, values map[string]interface{}) (map[string]string, error) {
	// templated & then unmarshall-ed version of this yaml
	b, err := AsTemplatedBytes(context+"MapOfStrings", yml, values)
	if err != nil {
		return nil, err
	}

	// Any given YAML can be unmarshalled into a map of strings
	var obj map[string]string
	err = yaml.Unmarshal(b, &obj)
	if err != nil {
		return nil, err
	}

	return obj, nil
}
