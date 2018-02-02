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

// This file provides various examples w.r.t go's
// text/template package. This is done to help contributors
// into understanding the standard templating features used
// by maya
//
// NOTE:
// BuiltIn funcs https://golang.org/src/text/template/funcs.go
// More funcs https://github.com/Masterminds/sprig
package template

import (
	"bytes"
	"reflect"
	"testing"
	"text/template"

	"github.com/ghodss/yaml"
)

// txtTplMock is the mock structure to test standard templating
// & extra templating functions provided via sprig
type txtTplMock struct {
	// values hold the data that will be fed into the ymlTpl
	// property of this instance
	values map[string]interface{}
	// ymlTpl is the yaml template that has conditionals
	// & placeholders which are set at runtime. These values
	// are set from the values property of this instance.
	ymlTpl string
	// ymlExpected is the resulting yaml document after
	// executing the ymlTpl & values properties of this
	// instance
	ymlExpected string
}

// AllValues contains a hierarchical set of data
//
// LEARNING:
//  Maya converts the Volume Policy into a similar struct
// that is taken as a input for transforming the yaml
// template's placeholders with these values
var AllValues = map[string]interface{}{
	"My": map[string]string{
		"Name": "OpenEBS",
	},
	"Storage": map[string]interface{}{
		"favorite": map[string]string{
			"block": "cstor",
			//"noblock": "",
			"nfs": "cstor again",
		},
	},
}

// YmlExpected is the expected template
// after the values are placed in the template's placeholders
var YmlExpected = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: OpenEBS-configmap
data:
  msg: Hello-OpenEBS
  block: cstor
  nfs: "cstor again"
  cool: true
`

// IfEqYmlTpl is a yaml template with
// placeholders
var IfEqYmlTpl = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .My.Name }}-configmap
data:
  msg: Hello-{{ .My.Name }}
  block: {{ .Storage.favorite.block }}
  nfs: {{ .Storage.favorite.nfs }}
  {{ if eq .Storage.favorite.block "cstor" }}cool: true{{ end }}
`

// TrimLeftWhitespaceYmlTpl is a yaml template with
// placeholders
var TrimLeftWhitespaceYmlTpl = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .My.Name }}-configmap
data:
  msg: Hello-{{ .My.Name }}
  block: {{ .Storage.favorite.block }}
  nfs: {{ .Storage.favorite.nfs }}
  {{- if eq .Storage.favorite.block "cstor" }}
  cool: true
  {{- end }}
`

// WithBlockYmlTpl is a yaml template with
// scoped placeholders
var WithBlockYmlTpl = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .My.Name }}-configmap
data:
  msg: Hello-{{ .My.Name }}
  {{- with .Storage.favorite }}
  block: {{ .block }}
  nfs: {{ .nfs }}
  {{- end }}
  cool: true
`

// SetDefaultsYmlTpl is a yaml template using
// default template function
var SetDefaultsYmlTpl = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .My.Name }}-configmap
data:
  msg: Hello-{{ .My.Name }}
  block: {{ .Storage.favorite.noblock | default "cstor" }}
  nfs: {{ .Storage.favorite.nfs }}
  {{- if eq .Storage.favorite.block "cstor" }}
  cool: true
  {{- end }}
`

func TestAll(t *testing.T) {

	tests := map[string]txtTplMock{
		"Test 'if eq' condition": {
			values:      AllValues,
			ymlTpl:      IfEqYmlTpl,
			ymlExpected: YmlExpected,
		},
		"Test '{{- ' whitespace control": {
			values:      AllValues,
			ymlTpl:      TrimLeftWhitespaceYmlTpl,
			ymlExpected: YmlExpected,
		},
		"Test '{{- with ' scope": {
			values:      AllValues,
			ymlTpl:      WithBlockYmlTpl,
			ymlExpected: YmlExpected,
		},
		"Test '| default '": {
			values:      AllValues,
			ymlTpl:      SetDefaultsYmlTpl,
			ymlExpected: YmlExpected,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			// power the standard templating with sprig
			tpl := template.New("example").Funcs(funcMap())
			tpl, err := tpl.Parse(mock.ymlTpl)
			if err != nil {
				t.Fatalf("Expected: 'no error' Actual: '%#v'", err)
			}

			// buf is an io.Writer implementation
			// as required by the template
			var buf bytes.Buffer

			// execute the parsed yaml against the values
			// & write the result into the buffer
			err = tpl.Execute(&buf, mock.values)
			if err != nil {
				t.Fatalf("Expected: 'no error' Actual: '%#v'", err)
			}

			// Any given YAML can be unmarshalled into a map of any objects
			var objActual map[string]interface{}
			err = yaml.Unmarshal(buf.Bytes(), &objActual)
			if err != nil {
				t.Fatalf("Expected: 'no error' Actual: '%#v'", err)
			}

			// Get back to expected yaml & unmarshall the yaml into
			// this object
			var objExpected map[string]interface{}
			err = yaml.Unmarshal([]byte(mock.ymlExpected), &objExpected)
			if err != nil {
				t.Fatalf("Expected: 'no error' Actual: '%#v'", err)
			}

			// Now Compare
			ok := reflect.DeepEqual(objExpected, objActual)
			if !ok {
				t.Fatalf("\nExpected: '%#v' \nActual: '%#v'", objExpected, objActual)
			}
		}) // end of run
	}
}
