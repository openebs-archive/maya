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

package kubernetes

import (
	"bytes"
	"github.com/ghodss/yaml"
	"reflect"
	"testing"
	"text/template"
)

func TestParse(t *testing.T) {
	tests := map[string]struct {
		version string
		isError bool
	}{
		// valid kubernetes versions
		"valid 1":  {"v1.11.0", false},
		"valid 2":  {"v0.0.1", false},
		"valid 3":  {"v1.0.0", false},
		"valid 4":  {"v0.1.0", false},
		"valid 5":  {"v0.1.0-alpha", false},
		"valid 6":  {"v0.1.0-alpha.123", false},
		"valid 7":  {"v0.1.0-beta", false},
		"valid 8":  {"v0.1.0-beta.11", false},
		"valid 9":  {"v0.1.0-gke", false},
		"valid 10": {"v0.1.0-gke.123", false},
		"valid 11": {"v0.1.0-eks", false},
		"valid 12": {"v0.1.0-eks.11", false},
		"valid 13": {"v0.1.0+abcfde23", false},
		"valid 14": {"v0.1.0-dev", false},
		"valid 15": {"v0.1.0-master", false},
		"valid 16": {"v0.1.0-anything", false},
		"valid 17": {"v0.1.0-anything.junk", false},
		// invalid kubernetes versions
		"invalid 1": {"1.11.0", true},
		"invalid 2": {"v0.1", true},
		"invalid 3": {"v1", true},
		"invalid 4": {"0.1", true},
		"invalid 5": {"1.0.0-alpha", true},
		"invalid 6": {"v1.0-alpha", true},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			v := parse(mock.version)
			if !mock.isError && v.HasError() {
				t.Fatalf("test '%s' failed: expected 'no error' actual 'error':\n'%s'", name, v.Errors())
			}
		})
	}
}

func TestAsLabelValue(t *testing.T) {
	tests := map[string]struct {
		version  string
		expected string
	}{
		// valid kubernetes versions
		"valid 1":  {"v1.11.0", "v1.11.0"},
		"valid 2":  {"v0.0.1", "v0.0.1"},
		"valid 3":  {"v1.0.0", "v1.0.0"},
		"valid 4":  {"v0.1.0", "v0.1.0"},
		"valid 5":  {"v0.1.0-alpha", "v0.1.0-alpha"},
		"valid 6":  {"v0.1.0-alpha.123", "v0.1.0-alpha.123"},
		"valid 7":  {"v0.1.0-beta", "v0.1.0-beta"},
		"valid 8":  {"v0.1.0-beta.11", "v0.1.0-beta.11"},
		"valid 9":  {"v0.1.0-gke", "v0.1.0-gke"},
		"valid 10": {"v0.1.0-gke.123", "v0.1.0-gke.123"},
		"valid 11": {"v0.1.0-eks", "v0.1.0-eks"},
		"valid 12": {"v0.1.0-eks.11", "v0.1.0-eks.11"},
		"valid 13": {"v0.1.0-dev", "v0.1.0-dev"},
		"valid 14": {"v0.1.0-master", "v0.1.0-master"},
		"valid 15": {"v0.1.0-anything", "v0.1.0-anything"},
		"valid 16": {"v0.1.0-anything.junk", "v0.1.0-anything.junk"},
		// invalid kubernetes versions
		"invalid 1": {"1.11.0", "1.11.0"},
		"invalid 2": {"v0.1", "v0.1"},
		"invalid 3": {"v1", "v1"},
		"invalid 4": {"0.1", "0.1"},
		"invalid 5": {"1.0.0-alpha", "1.0.0-alpha"},
		"invalid 6": {"v1.0-alpha", "v1.0-alpha"},
		// invalid label value
		"invalid label 1": {"v0.1.0+abcfde23", "v0.1.0"},
		"invalid label 2": {"v11+abcfde23", invalidVersionValue},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			v := AsLabelValue(mock.version)
			if mock.expected != v {
				t.Fatalf("test '%s' failed: expected '%s' actual '%s'", name, mock.expected, v)
			}
		})
	}
}

func TestCompare(t *testing.T) {
	tests := map[string]struct {
		version1 string
		version2 string
		expected int
	}{
		// equals
		"valid & eq - 1": {"v1.0.0", "v1.0.0", 0},
		"valid & eq - 2": {"v1.0.0-alpha", "v1.0.0-alpha", 0},
		"valid & eq - 3": {"v1.0.0-beta", "v1.0.0-beta", 0},
		"valid & eq - 4": {"v1.0.0-dev", "v1.0.0-dev", 0},
		"valid & eq - 5": {"v1.0.0-dev", "v1.0.0", 0},
		"valid & eq - 6": {"v1.0.0", "v1.0.0-ga", 0},
		// greater than
		"valid & gt - 1": {"v1.0.1", "v1.0.0", 1},
		"valid & gt - 2": {"v2.0.1", "v1.10.0", 1},
		"valid & gt - 3": {"v0.0.10", "v0.0.5", 1},
		"valid & gt - 4": {"v0.0.5-beta", "v0.0.5-alpha", 1},
		"valid & gt - 5": {"v0.0.5", "v0.0.5-alpha", 1},
		"valid & gt - 6": {"v0.0.5-beta.2", "v0.0.5-beta.1", 1},
		"valid & gt - 7": {"v1.12.7-gke.10", "v1.12.0", 1},
		// less than
		"valid & lt - 1": {"v0.0.1", "v0.0.5", -1},
		"valid & lt - 2": {"v0.1.1", "v0.2.0", -1},
		"valid & lt - 3": {"v1.1.1", "v2.0.0", -1},
		"valid & lt - 4": {"v1.0.5-alpha", "v1.0.5-beta", -1},
		"valid & lt - 5": {"v1.0.5-beta", "v1.0.5", -1},
		"valid & lt - 6": {"v1.0.5-alpha.12", "v1.0.5-alpha.21", -1},
		// invalid
		"invalid first ver":  {"1.1.1", "v2.0.0", -1},
		"invalid second ver": {"v1.1.1", "2.0.0", 1},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			c := Compare(mock.version1, mock.version2)
			if mock.expected != c {
				t.Fatalf("test '%s' failed: expected '%d' actual '%d'", name, mock.expected, c)
			}
		})
	}
}

func TestEquals(t *testing.T) {
	tests := map[string]struct {
		version1 string
		version2 string
		expected bool
	}{
		// equals
		"valid & eq - 1": {"v1.0.0", "v1.0.0", true},
		"valid & eq - 2": {"v1.0.0-alpha", "v1.0.0-alpha", true},
		"valid & eq - 3": {"v1.0.0-beta", "v1.0.0-beta", true},
		"valid & eq - 4": {"v1.0.0-dev", "v1.0.0-dev", true},
		"valid & eq - 5": {"v1.0.0-dev", "v1.0.0", true},
		"valid & eq - 6": {"v1.0.0", "v1.0.0-ga", true},
		// greater than
		"valid & gt - 1": {"v1.0.1", "v1.0.0", false},
		"valid & gt - 2": {"v2.0.1", "v1.10.0", false},
		"valid & gt - 3": {"v0.0.10", "v0.0.5", false},
		"valid & gt - 4": {"v0.0.5-beta", "v0.0.5-alpha", false},
		"valid & gt - 5": {"v0.0.5", "v0.0.5-alpha", false},
		"valid & gt - 6": {"v0.0.5-beta.2", "v0.0.5-beta.1", false},
		// less than
		"valid & lt - 1": {"v0.0.1", "v0.0.5", false},
		"valid & lt - 2": {"v0.1.1", "v0.2.0", false},
		"valid & lt - 3": {"v1.1.1", "v2.0.0", false},
		"valid & lt - 4": {"v1.0.5-alpha", "v1.0.5-beta", false},
		"valid & lt - 5": {"v1.0.5-beta", "v1.0.5", false},
		"valid & lt - 6": {"v1.0.5-alpha.12", "v1.0.5-alpha.21", false},
		// invalid
		"invalid first ver":  {"1.1.1", "v2.0.0", false},
		"invalid second ver": {"v1.1.1", "2.0.0", false},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			c := Equals(mock.version1, mock.version2)
			if mock.expected != c {
				t.Fatalf("test '%s' failed: expected '%t' actual '%t'", name, mock.expected, c)
			}
		})
	}
}

func TestGreaterThan(t *testing.T) {
	tests := map[string]struct {
		version1 string
		version2 string
		expected bool
	}{
		// equals
		"valid & eq - 1": {"v1.0.0", "v1.0.0", false},
		"valid & eq - 2": {"v1.0.0-alpha", "v1.0.0-alpha", false},
		"valid & eq - 3": {"v1.0.0-beta", "v1.0.0-beta", false},
		"valid & eq - 4": {"v1.0.0-dev", "v1.0.0-dev", false},
		"valid & eq - 5": {"v1.0.0-dev", "v1.0.0", false},
		"valid & eq - 6": {"v1.0.0", "v1.0.0-ga", false},
		// greater than
		"valid & gt - 1": {"v1.0.1", "v1.0.0", true},
		"valid & gt - 2": {"v2.0.1", "v1.10.0", true},
		"valid & gt - 3": {"v0.0.10", "v0.0.5", true},
		"valid & gt - 4": {"v0.0.5-beta", "v0.0.5-alpha", true},
		"valid & gt - 5": {"v0.0.5", "v0.0.5-alpha", true},
		"valid & gt - 6": {"v0.0.5-beta.2", "v0.0.5-beta.1", true},
		// less than
		"valid & lt - 1": {"v0.0.1", "v0.0.5", false},
		"valid & lt - 2": {"v0.1.1", "v0.2.0", false},
		"valid & lt - 3": {"v1.1.1", "v2.0.0", false},
		"valid & lt - 4": {"v1.0.5-alpha", "v1.0.5-beta", false},
		"valid & lt - 5": {"v1.0.5-beta", "v1.0.5", false},
		"valid & lt - 6": {"v1.0.5-alpha.12", "v1.0.5-alpha.21", false},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			c := GreaterThan(mock.version1, mock.version2)
			if mock.expected != c {
				t.Fatalf("test '%s' failed: expected '%t' actual '%t'", name, mock.expected, c)
			}
		})
	}
}

func TestGreaterThanOrEquals(t *testing.T) {
	tests := map[string]struct {
		version1 string
		version2 string
		expected bool
	}{
		// equals
		"valid & eq - 1": {"v1.0.0", "v1.0.0", true},
		"valid & eq - 2": {"v1.0.0-alpha", "v1.0.0-alpha", true},
		"valid & eq - 3": {"v1.0.0-beta", "v1.0.0-beta", true},
		"valid & eq - 4": {"v1.0.0-dev", "v1.0.0-dev", true},
		"valid & eq - 5": {"v1.0.0-dev", "v1.0.0", true},
		"valid & eq - 6": {"v1.0.0", "v1.0.0-ga", true},
		// greater than
		"valid & gt - 1": {"v1.0.1", "v1.0.0", true},
		"valid & gt - 2": {"v2.0.1", "v1.10.0", true},
		"valid & gt - 3": {"v0.0.10", "v0.0.5", true},
		"valid & gt - 4": {"v0.0.5-beta", "v0.0.5-alpha", true},
		"valid & gt - 5": {"v0.0.5", "v0.0.5-alpha", true},
		"valid & gt - 6": {"v0.0.5-beta.2", "v0.0.5-beta.1", true},
		// less than
		"valid & lt - 1": {"v0.0.1", "v0.0.5", false},
		"valid & lt - 2": {"v0.1.1", "v0.2.0", false},
		"valid & lt - 3": {"v1.1.1", "v2.0.0", false},
		"valid & lt - 4": {"v1.0.5-alpha", "v1.0.5-beta", false},
		"valid & lt - 5": {"v1.0.5-beta", "v1.0.5", false},
		"valid & lt - 6": {"v1.0.5-alpha.12", "v1.0.5-alpha.21", false},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			c := GreaterThanOrEquals(mock.version1, mock.version2)
			if mock.expected != c {
				t.Fatalf("test '%s' failed: expected '%t' actual '%t'", name, mock.expected, c)
			}
		})
	}
}

func TestLessThan(t *testing.T) {
	tests := map[string]struct {
		version1 string
		version2 string
		expected bool
	}{
		// equals
		"valid & eq - 1": {"v1.0.0", "v1.0.0", false},
		"valid & eq - 2": {"v1.0.0-alpha", "v1.0.0-alpha", false},
		"valid & eq - 3": {"v1.0.0-beta", "v1.0.0-beta", false},
		"valid & eq - 4": {"v1.0.0-dev", "v1.0.0-dev", false},
		"valid & eq - 5": {"v1.0.0-dev", "v1.0.0", false},
		"valid & eq - 6": {"v1.0.0", "v1.0.0-ga", false},
		// greater than
		"valid & gt - 1": {"v1.0.1", "v1.0.0", false},
		"valid & gt - 2": {"v2.0.1", "v1.10.0", false},
		"valid & gt - 3": {"v0.0.10", "v0.0.5", false},
		"valid & gt - 4": {"v0.0.5-beta", "v0.0.5-alpha", false},
		"valid & gt - 5": {"v0.0.5", "v0.0.5-alpha", false},
		"valid & gt - 6": {"v0.0.5-beta.2", "v0.0.5-beta.1", false},
		// less than
		"valid & lt - 1": {"v0.0.1", "v0.0.5", true},
		"valid & lt - 2": {"v0.1.1", "v0.2.0", true},
		"valid & lt - 3": {"v1.1.1", "v2.0.0", true},
		"valid & lt - 4": {"v1.0.5-alpha", "v1.0.5-beta", true},
		"valid & lt - 5": {"v1.0.5-beta", "v1.0.5", true},
		"valid & lt - 6": {"v1.0.5-alpha.12", "v1.0.5-alpha.21", true},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			c := LessThan(mock.version1, mock.version2)
			if mock.expected != c {
				t.Fatalf("test '%s' failed: expected '%t' actual '%t'", name, mock.expected, c)
			}
		})
	}
}

func TestLessThanOrEquals(t *testing.T) {
	tests := map[string]struct {
		version1 string
		version2 string
		expected bool
	}{
		// equals
		"valid & eq - 1": {"v1.0.0", "v1.0.0", true},
		"valid & eq - 2": {"v1.0.0-alpha", "v1.0.0-alpha", true},
		"valid & eq - 3": {"v1.0.0-beta", "v1.0.0-beta", true},
		"valid & eq - 4": {"v1.0.0-dev", "v1.0.0-dev", true},
		"valid & eq - 5": {"v1.0.0-dev", "v1.0.0", true},
		"valid & eq - 6": {"v1.0.0", "v1.0.0-ga", true},
		// greater than
		"valid & gt - 1": {"v1.0.1", "v1.0.0", false},
		"valid & gt - 2": {"v2.0.1", "v1.10.0", false},
		"valid & gt - 3": {"v0.0.10", "v0.0.5", false},
		"valid & gt - 4": {"v0.0.5-beta", "v0.0.5-alpha", false},
		"valid & gt - 5": {"v0.0.5", "v0.0.5-alpha", false},
		"valid & gt - 6": {"v0.0.5-beta.2", "v0.0.5-beta.1", false},
		// less than
		"valid & lt - 1": {"v0.0.1", "v0.0.5", true},
		"valid & lt - 2": {"v0.1.1", "v0.2.0", true},
		"valid & lt - 3": {"v1.1.1", "v2.0.0", true},
		"valid & lt - 4": {"v1.0.5-alpha", "v1.0.5-beta", true},
		"valid & lt - 5": {"v1.0.5-beta", "v1.0.5", true},
		"valid & lt - 6": {"v1.0.5-alpha.12", "v1.0.5-alpha.21", true},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			c := LessThanOrEquals(mock.version1, mock.version2)
			if mock.expected != c {
				t.Fatalf("test '%s' failed: expected '%t' actual '%t'", name, mock.expected, c)
			}
		})
	}
}

func TestTemplateFuncs(t *testing.T) {
	tests := map[string]struct {
		versions map[string]string
		template string
		expected string
	}{
		"all in one": {
			versions: map[string]string{
				"version1": "v1.0.1-alpha.21",
				"version2": "v1.0.1-beta",
				"version3": "v1.1.1",
				"version4": "v1.1.1-ga",
				"version5": "v1.12.7-gke.10",
				"version6": "v1.12.0",
			},
			template: `
one: {{kubeVersionEq .version3 .version4}}
two: {{kubeVersionEq .version1 .version2}}
three: {{kubeVersionGt .version3 .version4}}
four: {{kubeVersionGte .version3 .version4}}
five: {{kubeVersionLt .version3 .version4}}
six: {{kubeVersionLte .version3 .version4}}
seven: {{kubeVersionGt .version1 .version2}}
eight: {{kubeVersionLt .version1 .version2}}
nine: {{kubeVersionLt .version2 .version4}}
ten: {{kubeVersionGte .version5 .version6}}
{{- if kubeVersionGte .version5 .version6 }}
passed: true
{{- end }}
`,
			expected: `
one: true
two: false
three: false
four: true
five: false
six: true
seven: false
eight: true
nine: true
ten: true
passed: true
`,
		},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			tpl := template.New("KubeVersionTemplating").Funcs(TemplateFunctions())
			tpl, err := tpl.Parse(mock.template)
			if err != nil {
				t.Fatalf("test '%s' failed: expected 'no template parse error': actual '%s'", name, err.Error())
			}
			// buf is an io.Writer implementation
			// as required by the template
			var buf bytes.Buffer
			// execute the parsed yaml against the values
			// & write the result into the buffer
			err = tpl.Execute(&buf, mock.versions)
			if err != nil {
				t.Fatalf("test failed: expected 'no template execution error': actual '%s'", err.Error())
			}
			// buffer that represents a YAML can be unmarshalled into a map of any objects
			var objActual map[string]interface{}
			err = yaml.Unmarshal(buf.Bytes(), &objActual)
			if err != nil {
				t.Fatalf("test failed: expected 'no error on un-marshalling templated bytes': actual '%s'", err.Error())
			}
			// unmarshall the expected yaml into a map of any objects
			var objExpected map[string]interface{}
			err = yaml.Unmarshal([]byte(mock.expected), &objExpected)
			if err != nil {
				t.Fatalf("test failed: expected 'no error on un-marshalling expected bytes': actual '%s'", err.Error())
			}
			// compare expected vs. actual object
			ok := reflect.DeepEqual(objExpected, objActual)
			if !ok {
				t.Fatalf("test failed:\n\nexpected yaml: '%s' \n\nactual yaml: '%s'", mock.expected, buf.Bytes())
			}
		})
	}
}
