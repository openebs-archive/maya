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
// NOTE on templating:
//
// BuiltIn funcs: https://golang.org/src/text/template/funcs.go
// Custom template funcs:
// - https://github.com/Masterminds/sprig/tree/master/docs
// Templating guides:
// - https://github.com/kubernetes/helm/tree/master/docs/chart_template_guide
// - https://docs.ansible.com/ansible/latest/reference_appendices/YAMLSyntax.html
package template

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
	"text/template"

	"encoding/json"
	"github.com/ghodss/yaml"
)

func TestToYaml(t *testing.T) {
	tests := map[string]struct {
		data           interface{}
		expected       string
		isErr          bool
		expectedErrMsg string
	}{
		//
		// start of test case
		//
		"Positive Test - Test by providing a simple map": {
			data: map[string]string{
				"co":      "k8s",
				"storage": "openebs",
			},
			expected: fmt.Sprintf("co: k8s \nstorage: openebs\n"),
		},
		//
		// start of test case
		//
		"Positive Test - Test by providing a nested map": {
			data: map[string]interface{}{
				"co": "k8s",
				"storage": map[string]string{
					"gen1": "jiva",
					"gen2": "cstor",
				},
			},
			expected: `
co: k8s
storage:
  gen1: jiva
  gen2: cstor`,
		},
		//
		// start of test case
		//
		"Negative Test - Test by providing a non map object": {
			data:           []interface{}{"co", "k8s", "storage", "gen1", "jiva", "gen2", "cstor"},
			isErr:          true,
			expectedErrMsg: "error unmarshaling JSON: json: cannot unmarshal array into Go value of type map[string]interface {}",
		},
		//
		// start of test case
		//
		"Negative Test - Test by providing a nil object": {
			data: nil,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			actual := fromYaml(toYaml(mock.data))
			expected := fromYaml(mock.expected)

			if mock.isErr {
				actualErrMsg, _ := actual["Error"].(string)
				if mock.expectedErrMsg != actualErrMsg {
					t.Fatalf("toYaml test failed: expected error '%s': actual error '%s'", mock.expectedErrMsg, actualErrMsg)
				}
			}

			if !mock.isErr && !reflect.DeepEqual(expected, actual) {
				t.Fatalf("toYaml test failed: expected '%#v': actual '%#v'", expected, actual)
			}
		})
	}
}

func TestNestedKeyMap(t *testing.T) {
	tests := map[string]struct {
		delimiters  string
		given       []string
		destination map[string]interface{}
		expected    map[string]interface{}
	}{
		//
		// start of test case
		//
		`Positive Test - Verify map creation with valid pairs & delimiters`: {
			delimiters:  "=",
			given:       []string{"co/co1=k8s", "co1=k8sNew", "co2=swarm", "co3=nomad"},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"co/co1": "k8s",
				"co1":    "k8sNew",
				"co2":    "swarm",
				"co3":    "nomad",
			},
		},
		//
		// start of test case
		//
		`Positive Test - Verify map creation with valid pairs, multiple values per key & delimiters`: {
			delimiters:  "=",
			given:       []string{"co/co1=k8s", "co2=swarm", "co2=swarm2", "co3=nomad"},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"co/co1": "k8s",
				"co2":    "swarm, swarm2",
				"co3":    "nomad",
			},
		},
		//
		// start of test case
		//
		`Positive Test - Verify map creation with two-step i.e. multi level keys`: {
			delimiters:  "/ =",
			given:       []string{"co/co1=k8s", "co2=swarm", "co2=swarm2", "co3=nomad"},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"co": map[string]interface{}{
					"co1": "k8s",
				},
				"co2": "swarm, swarm2",
				"co3": "nomad",
			},
		},
		//
		// start of test case
		//
		`Positive Test - Verify map creation with three-step i.e. multi level keys`: {
			delimiters:  "/ @ =",
			given:       []string{"cloud/co@co1=k8s", "co2=swarm", "co2=swarm2", "co3=nomad"},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"cloud": map[string]interface{}{
					"co": map[string]interface{}{
						"co1": "k8s",
					},
				},
				"co2": "swarm, swarm2",
				"co3": "nomad",
			},
		},
		//
		// start of test case
		//
		`Positive Test - Verify map creation with three-step i.e. multi level duplicate keys`: {
			delimiters:  "/ / =",
			given:       []string{"cloud/co/co1=k8s", "co/co2=swarm", "co2=swarm2", "co3=nomad"},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"cloud": map[string]interface{}{
					"co": map[string]interface{}{
						"co1": "k8s",
					},
				},
				"co": map[string]interface{}{
					"co2": "swarm",
				},
				"co2": "swarm2",
				"co3": "nomad",
			},
		},
		//
		// start of test case
		//
		`Negative Test - Verify map creation with keys as well as values having whitespaces`: {
			delimiters:  "=",
			given:       []string{" co/co1 =k8s", "co2 = swarm ", " co2= swarm2", "co3=nomad "},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"co/co1": "k8s",
				"co2":    "swarm, swarm2",
				"co3":    "nomad",
			},
		},
		//
		// start of test case
		//
		`Negative Test - Verify map creation with multi level keys with some keys as blank`: {
			delimiters:  "/ / =",
			given:       []string{"cloud/ /co1=k8s", "co/=swarm", "co=swarm2", "co3=nomad"},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"cloud": map[string]interface{}{
					"co1": "k8s",
				},
				"co":  "swarm, swarm2",
				"co3": "nomad",
			},
		},
		//
		// start of test case
		//
		`Negative Test - Verify map creation with some values as blank`: {
			delimiters:  "/ / =",
			given:       []string{"cloud/ /co1=k8s", "cloud/ /co1=", "co/=", "co=swarm2", "co3=nomad", "co4="},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"cloud": map[string]interface{}{
					"co1": "k8s",
				},
				"co":  "swarm2",
				"co3": "nomad",
				"co4": "",
			},
		},
		//
		// start of test case
		//
		`Negative Test - Verify map creation with some keys & values as whitespaces`: {
			delimiters:  "/ / =",
			given:       []string{"cloud/ /co1=k8s", "cloud/ /co1=  ", "  =  ", " co3 = nomad ", "co4=  "},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"cloud": map[string]interface{}{
					"co1": "k8s",
				},
				"co3": "nomad",
				"co4": "",
			},
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			result := nestedKeyMap(mock.delimiters, mock.destination, mock.given)

			if !reflect.DeepEqual(mock.expected, result) {
				t.Fatalf("test nested key map failed:\n\nexpected: '%#v' \n\nactual: '%#v'", mock.expected, result)
			}
		})
	}
}

func TestKeyMap(t *testing.T) {
	tests := map[string]struct {
		destinationFields string
		given             []string
		destination       map[string]interface{}
		expected          map[string]interface{}
	}{
		//
		// start of test case
		//
		`Positive Test - Verify map creation with valid pairs & delimiters:
      - Verify if the result matches the expected`: {
			destinationFields: "test1",
			given:             []string{"co1=k8s,co2=swarm", "co3=nomad"},
			destination:       map[string]interface{}{},
			expected: map[string]interface{}{
				"test1": map[string]interface{}{
					"pkey": map[string]interface{}{
						"co1": "k8s",
						"co2": "swarm",
						"co3": "nomad",
					},
				},
			},
		},
		//
		// start of test case
		//
		`Positive Test - Verify map creation with valid pairs, primary keys & delimiters:
      - Verify if the result matches the expected`: {
			destinationFields: "test1",
			given:             []string{"co1=k8s,co2=swarm,pkey=one", "pkey=two,co3=nomad"},
			destination:       map[string]interface{}{},
			expected: map[string]interface{}{
				"test1": map[string]interface{}{
					"one": map[string]interface{}{
						"co1": "k8s",
						"co2": "swarm",
					},
					"two": map[string]interface{}{
						"co3": "nomad",
					},
				},
			},
		},
		//
		// start of test case
		//
		`Negative Test - Verify map creation with valid pairs & delimiters & erratic whitespaces:
      - Verify if the result matches the expected`: {
			destinationFields: "test2",
			given:             []string{"co1 = k8s,  co2= swarm  ", "   co3=nomad  "},
			destination:       map[string]interface{}{},
			expected: map[string]interface{}{
				"test2": map[string]interface{}{
					"pkey": map[string]interface{}{
						"co1": "k8s",
						"co2": "swarm",
						"co3": "nomad",
					},
				},
			},
		},
		//
		// start of test case
		//
		`Negative Test - Verify map creation with pairs having missing value(s) or key(s) or both:
      - Verify if the result matches the expected`: {
			destinationFields: "test3",
			given:             []string{"= k8s,  co2=  ", " = "},
			destination:       map[string]interface{}{},
			expected: map[string]interface{}{
				"test3": map[string]interface{}{
					"pkey": map[string]interface{}{
						"co2": "",
					},
				},
			},
		},
		//
		// start of test case
		//
		`Negative Test - Verify map creation with primary key but missing primary key's value:
      - Verify if the result matches the expected`: {
			destinationFields: "test4",
			given:             []string{"pkey=,co1= k8s,  co2=,=nomad  ", " = "},
			destination:       map[string]interface{}{},
			expected: map[string]interface{}{
				"test4": map[string]interface{}{
					"pkey": map[string]interface{}{
						"co1": "k8s",
						"co2": "",
					},
				},
			},
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			result := keyMap(mock.destinationFields, mock.destination, mock.given)

			if !reflect.DeepEqual(mock.expected, result) {
				t.Fatalf("test key map failed:\n\nexpected: '%#v' \n\nactual: '%#v'", mock.expected, result)
			}
		})
	}
}

func TestSplitKeyMap(t *testing.T) {
	tests := map[string]struct {
		splitters   string
		destFields  string
		given       []string
		destination map[string]interface{}
		expected    map[string]interface{}
	}{
		//
		// start of test case
		//
		`Positive Test - Verify map creation with k=v pairs separated by ::
      - NOTE: If there is only one k=v pair then :: is not provided
      - Verify if the result matches the expected`: {
			splitters:   ":: =",
			destFields:  "test1",
			given:       []string{"co1=k8s::co2=swarm", "co3=nomad"},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"test1": map[string]interface{}{
					"pkey": map[string]interface{}{
						"co1": "k8s",
						"co2": "swarm",
						"co3": "nomad",
					},
				},
			},
		},
		//
		// start of test case
		//
		`Positive Test - Verify map creation with k=v pairs separated by --
      - NOTE: If there is only one k=v pair then -- is not provided
      - NOTE: primary key value is provided for some keyvalue pairs
      - Verify if the result matches the expected`: {
			splitters:   "-- =",
			destFields:  "test2",
			given:       []string{"co1=k8s--co2=swarm", "pkey=myCO--co3=nomad"},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"test2": map[string]interface{}{
					"pkey": map[string]interface{}{
						"co1": "k8s",
						"co2": "swarm",
					},
					"myCO": map[string]interface{}{
						"co3": "nomad",
					},
				},
			},
		},
		//
		// start of test case
		//
		`Positive Test - Verify map creation with k::v pairs separated by --
      - NOTE: If there is only one k::v pair then -- is not provided
      - NOTE: primary key value is provided for some keyvalue pairs
      - Verify if the result matches the expected`: {
			splitters:   "-- ::",
			destFields:  "test3",
			given:       []string{"co1::k8s--co2::swarm", "pkey::myCO--co3::nomad"},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"test3": map[string]interface{}{
					"pkey": map[string]interface{}{
						"co1": "k8s",
						"co2": "swarm",
					},
					"myCO": map[string]interface{}{
						"co3": "nomad",
					},
				},
			},
		},
		//
		// start of test case
		//
		`Positive Test - Verify map creation with k=v pairs without any splitters
      - NOTE: primary key value is provided only for some keyvalue pairs
      - Verify if the result matches the expected`: {
			splitters:   "",
			destFields:  "test4",
			given:       []string{"co1=k8s,co2=swarm", "pkey=myCO,co3=nomad", "co4=custom"},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"test4": map[string]interface{}{
					"pkey": map[string]interface{}{
						"co1": "k8s",
						"co2": "swarm",
						"co4": "custom",
					},
					"myCO": map[string]interface{}{
						"co3": "nomad",
					},
				},
			},
		},
		//
		// start of test case
		//
		`Positive Test - Verify map creation with k=v pairs without any splitters and primary key
      - Verify if the result matches the expected`: {
			splitters:   "",
			destFields:  "test5",
			given:       []string{"co1=k8s,co2=swarm", "co3=nomad", "co4=custom"},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"test5": map[string]interface{}{
					"pkey": map[string]interface{}{
						"co1": "k8s",
						"co2": "swarm",
						"co3": "nomad",
						"co4": "custom",
					},
				},
			},
		},
		//
		// start of test case
		//
		`Positive Test - Verify map creation with k=v pairs with common keys & without splitters & primary key
      - Verify if the result matches the expected`: {
			splitters:   "",
			destFields:  "test6",
			given:       []string{"co1=k8s,co2=swarm,co3=newK8s,co4=newSwarm", "co3=nomad", "co3=mySched,co4=custom"},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"test6": map[string]interface{}{
					"pkey": map[string]interface{}{
						"co1": "k8s",
						"co2": "swarm",
						"co3": "newK8s, nomad, mySched",
						"co4": "newSwarm, custom",
					},
				},
			},
		},
		//
		// start of test case
		//
		`Positive Test - Verify map creation with k=v pairs with common keys & without splitters & different primary keys
      - Verify if the result matches the expected`: {
			splitters:   "",
			destFields:  "test7",
			given:       []string{"pkey=alpha,co1=k8s,co2=swarm", "pkey=beta,co3=nomad", "pkey=delta,co3=mySched,co4=custom"},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"test7": map[string]interface{}{
					"alpha": map[string]interface{}{
						"co1": "k8s",
						"co2": "swarm",
					},
					"beta": map[string]interface{}{
						"co3": "nomad",
					},
					"delta": map[string]interface{}{
						"co3": "mySched",
						"co4": "custom",
					},
				},
			},
		},
		//
		// start of test case
		//
		`Positive Test - Verify map creation with k=v pairs without destination field
      - Verify if the result matches the expected`: {
			splitters:   "",
			destFields:  "",
			given:       []string{"pkey=alpha,co1=k8s,co2=swarm", "pkey=beta,co3=nomad", "pkey=delta,co3=mySched,co4=custom"},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"alpha": map[string]interface{}{
					"co1": "k8s",
					"co2": "swarm",
				},
				"beta": map[string]interface{}{
					"co3": "nomad",
				},
				"delta": map[string]interface{}{
					"co3": "mySched",
					"co4": "custom",
				},
			},
		},
		//
		// start of test case
		//
		`Positive Test - Verify map creation with k=v pairs without destination field & some primary keys
      - Verify if the result matches the expected`: {
			splitters:   "",
			destFields:  "",
			given:       []string{"pkey=alpha,co1=k8s,co2=swarm", "co3=nomad", "co3=mySched,co4=custom"},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"alpha": map[string]interface{}{
					"co1": "k8s",
					"co2": "swarm",
				},
				"pkey": map[string]interface{}{
					"co3": "nomad, mySched",
					"co4": "custom",
				},
			},
		},
		//
		// start of test case
		//
		`Positive Test - Verify map creation with k=v pairs without destination field & primary key
      - Verify if the result matches the expected`: {
			splitters:   "",
			destFields:  "",
			given:       []string{"co1=k8s,co2=swarm", "co3=nomad", "co3=mySched,co4=custom"},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"pkey": map[string]interface{}{
					"co1": "k8s",
					"co2": "swarm",
					"co3": "nomad, mySched",
					"co4": "custom",
				},
			},
		},
		//
		// start of test case
		//
		`Negative Test - Verify map creation with 'comma' as pair delimiter or 'comma with dangling spaces as pair delimiter'
      - Verify if the result matches the expected`: {
			splitters:   "",
			destFields:  "",
			given:       []string{"co1=k8s,co2=swarm ,  co3=newK8s", "co3=nomad", "co3=mySched, co4=custom"},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"pkey": map[string]interface{}{
					"co1": "k8s",
					"co2": "swarm",
					"co3": "newK8s, nomad, mySched",
					"co4": "custom",
				},
			},
		},
		//
		// start of test case
		//
		`Negative Test - Verify map creation with blank value(s) & 'comma with dangling space delimiter'
      - Verify if the result matches the expected`: {
			splitters:   "",
			destFields:  "",
			given:       []string{"co1=,co2=swarm ,  co3= ", "co3=nomad", "co3=mySched, co4=custom"},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"pkey": map[string]interface{}{
					"co1": "",
					"co2": "swarm",
					"co3": "nomad, mySched",
					"co4": "custom",
				},
			},
		},
		//
		// start of test case
		//
		`Negative Test - Verify map creation with blank key(s), blank value(s) & 'comma with dangling space delimiter'
      - Verify if the result matches the expected`: {
			splitters:   "",
			destFields:  "",
			given:       []string{"co1=,=swarm ,  =newK8s ", "co3=nomad", "co3=mySched, co4=custom"},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"pkey": map[string]interface{}{
					"co1": "",
					"co3": "nomad, mySched",
					"co4": "custom",
				},
			},
		},
		//
		// start of test case
		//
		`Negative Test - Verify map creation with blank string & strings with whitespaces
      - Verify if the result matches the expected`: {
			splitters:   "",
			destFields:  "",
			given:       []string{"", "  ", "  co3=mySched, co4=custom  "},
			destination: map[string]interface{}{},
			expected: map[string]interface{}{
				"pkey": map[string]interface{}{
					"co3": "mySched",
					"co4": "custom",
				},
			},
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			result := splitKeyMap(mock.splitters, mock.destFields, mock.destination, mock.given)

			if !reflect.DeepEqual(mock.expected, result) {
				t.Fatalf("test split key map failed:\n\nexpected: '%#v' \n\nactual: '%#v'", mock.expected, result)
			}
		})
	}
}

// MockJsonList is a struct used during unit testing that can be marshaled from
// go object to []byte & vice-versa
//
// NOTE:
//  The properties of this struct should be CapitalCased for marshal & un-marshal
// to work properly
type MockJsonList struct {
	ApiVersion string     `json:"apiVersion"`
	Items      []MockJson `json:"items"`
}

// MockJson is a struct used during unit testing that can be marshaled from go
// object to []byte & vice-versa
//
// NOTE:
//  The properties of this struct should be CapitalCased for marshal & un-marshal
// to work properly
type MockJson struct {
	Labels    map[string]string `json:"labels"`
	Owner     string            `json:"owner"`
	Kind      string            `json:"kind"`
	Namespace string            `json:"namespace"`
	Name      string            `json:"name"`
}

func mockJsonMarshal(mock *MockJson) (op []byte) {
	op, err := json.Marshal(mock)
	if err != nil {
		fmt.Printf(err.Error())
	}
	return
}

func mockJsonListMarshal(mock *MockJsonList) (op []byte) {
	op, err := json.Marshal(mock)
	if err != nil {
		fmt.Printf(err.Error())
	}
	return
}

func TestTemplatingWithMutatingTemplateValues(t *testing.T) {
	tests := map[string]struct {
		templateInYaml         string
		templateValues         map[string]interface{}
		expectedYaml           string
		expectedTemplateValues map[string]interface{}
	}{
		//
		// start of test scenario
		//
		`Positive Test - Does not throw VerifyError error for non-empty string:
		  - Provide a non empty string to verifyErr template function
		  - Verify it should not throw any error`: {
			templateInYaml: `
{{- "I am not empty" | empty | verifyErr "empty string provided" | noop -}}
`,
			templateValues: map[string]interface{}{
				"Values": map[string]interface{}{},
			},
			expectedYaml:           ``,
			expectedTemplateValues: map[string]interface{}{},
		},
		//
		// start of test scenario
		//
		`Positive Test - Does not throw VerifyError error for non-empty list:
		  - Provide a non empty list to verifyErr template function
		  - Verify it should not throw any error`: {
			templateInYaml: `
{{- list "I am not empty" | empty | verifyErr "empty list provided" | noop -}}
`,
			templateValues: map[string]interface{}{
				"Values": map[string]interface{}{},
			},
			expectedYaml:           ``,
			expectedTemplateValues: map[string]interface{}{},
		},
		//
		// start of test scenario
		//
		`Positive Test - Does not throw VerifyError error for non-empty dict:
		  - Provide a non empty dict to verifyErr template function
		  - Verify it should not throw any error`: {
			templateInYaml: `
{{- dict "k1" "v1" "k2" "v2" | empty | verifyErr "empty dict provided" | noop -}}
`,
			templateValues: map[string]interface{}{
				"Values": map[string]interface{}{},
			},
			expectedYaml:           ``,
			expectedTemplateValues: map[string]interface{}{},
		},
		//
		// start of test scenario
		//
		`Negative Test - Throw VerifyError for empty list:
		  - Provide a empty list to verifyErr template function
		  - Verify it should throw an error`: {
			templateInYaml: `
{{- list | empty | verifyErr "empty list provided" | saveIf "err" .Values | noop -}}`,
			templateValues: map[string]interface{}{
				"Values": map[string]interface{}{},
			},
			expectedYaml: ``,
			expectedTemplateValues: map[string]interface{}{
				"err": &VerifyError{"empty list provided"},
			},
		},
		//
		// start of test scenario
		//
		`Negative Test - Throw VerifyError for empty dict:
		  - Provide a empty dict to verifyErr template function
		  - Verify it should throw an error`: {
			templateInYaml: `
{{- dict | empty | verifyErr "empty dictionary provided" | saveIf "err" .Values | noop -}}`,
			templateValues: map[string]interface{}{
				"Values": map[string]interface{}{},
			},
			expectedYaml: ``,
			expectedTemplateValues: map[string]interface{}{
				"err": &VerifyError{"empty dictionary provided"},
			},
		},
		//
		// start of test scenario
		//
		`Negative Test - Throw VerifyError for empty string:
		  - Provide a empty string to verifyErr template function
		  - Verify it should throw an error`: {
			templateInYaml: `
{{- "" | empty | verifyErr "empty string provided" | saveIf "err" .Values | noop -}}`,
			templateValues: map[string]interface{}{
				"Values": map[string]interface{}{},
			},
			expectedYaml: ``,
			expectedTemplateValues: map[string]interface{}{
				"err": &VerifyError{"empty string provided"},
			},
		},
		//
		// start of test scenario
		//
		`Positive Test: Does not throw VerifyError for non-empty kind objects: 
		  - Get a list of kinds via jsonpath as a string output, 
		  - Then trim this output which has two items i.e. output is non-empty,
		  - Then split this output via " " into an array
		  - Then throw verify error if this output array length is not 2
		  - Verify this templating does not throw VerifyError error`: {
			templateInYaml: `
{{- jsonpath .JsonDoc "{.items[*].kind}" | trim | saveAs "TaskResult.tskId.kind" .Values | noop -}}
{{- .Values.TaskResult.tskId.kind | splitList " " | isLen 2 | not | verifyErr "kind count is not two" | saveIf "err" .Values | noop -}}
value: {{ .Values.TaskResult.tskId.kind }}`,
			templateValues: map[string]interface{}{
				"JsonDoc": mockJsonListMarshal(&MockJsonList{
					ApiVersion: "v1beta1",
					Items: []MockJson{
						MockJson{
							Kind: "Pod",
						},
						MockJson{
							Kind: "Deployment",
						},
					},
				}),
				"Values": map[string]interface{}{},
			},
			expectedYaml: `
value: Pod Deployment`,
			expectedTemplateValues: map[string]interface{}{
				"err": nil,
				"TaskResult": map[string]interface{}{
					"tskId": map[string]interface{}{
						"kind": "Pod Deployment",
					},
				},
			},
		},
		//
		// start of test scenario
		//
		`Negative Test: Throw VerifyError due to empty kind objects: 
		  - Get a list of kinds via jsonpath as a string output, 
		  - Then trim this output which is empty in this case,
		  - Then split this output via " " into an array
		  - Then throw VerifyError if this output array length is not 2
		  - Verify this templating throws VerifyError during template execution`: {
			templateInYaml: `
{{- jsonpath .JsonDoc "{.items[*].kind}" | trim | saveAs "TaskResult.tskId.kind" .Values | noop -}}
{{- .Values.TaskResult.tskId.kind | splitList " " | isLen 2 | not | verifyErr "kind count is not two" | saveIf "err" .Values | noop -}}`,
			templateValues: map[string]interface{}{
				"JsonDoc": mockJsonListMarshal(&MockJsonList{
					ApiVersion: "v1beta1",
					Items: []MockJson{
						MockJson{
							Kind: "",
						},
						MockJson{
							Kind: "",
						},
					},
				}),
				"Values": map[string]interface{}{},
			},
			expectedYaml: ``,
			expectedTemplateValues: map[string]interface{}{
				"err": &VerifyError{"kind count is not two"},
				"TaskResult": map[string]interface{}{
					"tskId": map[string]interface{}{
						"kind": "",
					},
				},
			},
		},
		//
		// start of test scenario
		//
		`Negative Test: Throws VerifyError as count of kind is not 2: 
		  - Get a list of kinds via jsonpath as a string output, 
		  - Then trim this output which has one item i.e. output is non-empty,
		  - Then split this output via " " into an array
		  - Then throw VerifyError if this output array length is not 2
		  - Verify VerifyError is thrown during template execution`: {
			templateInYaml: `
{{- jsonpath .JsonDoc "{.items[*].kind}" | trim | saveAs "TaskResult.tskId.kind" .Values | noop -}}
{{- .Values.TaskResult.tskId.kind | splitList " " | isLen 2 | not | verifyErr "kind count is not two" | saveIf "err" .Values | noop -}}`,
			templateValues: map[string]interface{}{
				"JsonDoc": mockJsonListMarshal(&MockJsonList{
					ApiVersion: "v1beta1",
					Items: []MockJson{
						MockJson{
							Kind: "",
						},
						MockJson{
							Kind: "Deployment",
						},
					},
				}),
				"Values": map[string]interface{}{},
			},
			expectedYaml: ``,
			expectedTemplateValues: map[string]interface{}{
				"err": &VerifyError{"kind count is not two"},
				"TaskResult": map[string]interface{}{
					"tskId": map[string]interface{}{
						"kind": "Deployment",
					},
				},
			},
		},
		//
		// start of test scenario
		//
		`Negative Test - Throw notFoundErr for empty list:
		  - Provide a empty list to notFoundErr template function
		  - Verify it should throw an error`: {
			templateInYaml: `
{{- list | notFoundErr "empty list provided" | saveIf "err" .Values | noop -}}`,
			templateValues: map[string]interface{}{
				"Values": map[string]interface{}{},
			},
			expectedYaml: ``,
			expectedTemplateValues: map[string]interface{}{
				"err": &NotFoundError{"empty list provided"},
			},
		},
		//
		// start of test scenario
		//
		`Negative Test - Throw notFoundErr for empty dict:
		  - Provide a empty dict to notFoundErr template function
		  - Verify it should throw an error`: {
			templateInYaml: `
{{- dict | notFoundErr "empty dictionary provided" | saveIf "err" .Values | noop -}}`,
			templateValues: map[string]interface{}{
				"Values": map[string]interface{}{},
			},
			expectedYaml: ``,
			expectedTemplateValues: map[string]interface{}{
				"err": &NotFoundError{"empty dictionary provided"},
			},
		},
		//
		// start of test scenario
		//
		`Negative Test - Throw notFoundErr for empty string:
		  - Provide a empty string to notFoundErr template function
		  - Verify it should throw an error`: {
			templateInYaml: `
{{- "" | notFoundErr "empty string provided" | saveIf "err" .Values | noop -}}`,
			templateValues: map[string]interface{}{
				"Values": map[string]interface{}{},
			},
			expectedYaml: ``,
			expectedTemplateValues: map[string]interface{}{
				"err": &NotFoundError{"empty string provided"},
			},
		},
		//
		// start of test scenario
		//
		`Positive Test - Does not throw notFoundErr error for non-empty string:
		  - Provide a non empty string to notFoundErr template function
		  - Verify it should not throw any error`: {
			templateInYaml: `
{{- "I am not empty" | notFoundErr "empty string provided" | saveIf "err" .Values | noop -}}
`,
			templateValues: map[string]interface{}{
				"Values": map[string]interface{}{},
			},
			expectedYaml: ``,
			expectedTemplateValues: map[string]interface{}{
				"err": nil,
			},
		},
		//
		// start of test scenario
		//
		`Positive Test - Does not throw notFoundErr error for non-empty list:
		  - Provide a non empty list to notFoundErr template function
		  - Verify it should not throw any error`: {
			templateInYaml: `
{{- list "I am not empty" | notFoundErr "empty list provided" | saveIf "err" .Values | noop -}}
`,
			templateValues: map[string]interface{}{
				"Values": map[string]interface{}{},
			},
			expectedYaml: ``,
			expectedTemplateValues: map[string]interface{}{
				"err": nil,
			},
		},
		//
		// start of test scenario
		//
		`Positive Test - Does not throw notFoundErr error for non-empty dict:
		  - Provide a non empty dict to notFoundErr template function
		  - Verify it should not throw any error`: {
			templateInYaml: `
{{- dict "k1" "v1" "k2" "v2" | notFoundErr "empty dict provided" | saveIf "err" .Values | noop -}}
`,
			templateValues: map[string]interface{}{
				"Values": map[string]interface{}{},
			},
			expectedYaml: ``,
			expectedTemplateValues: map[string]interface{}{
				"err": nil,
			},
		},
		//
		// start of test scenario
		//
		`Positive Test - Does not throw notFoundErr error for non empty kind
		  - NOTE: jsonpath, saveAs and notFoundErr are template functions
		  - Verify go templating should not throw any error`: {
			templateInYaml: `
{{- jsonpath .JsonDoc "{.kind}" | saveAs "TaskResult.tskId.kind" .Values | notFoundErr "kind is missing" | saveIf "err" .Values | noop -}}
`,
			templateValues: map[string]interface{}{
				"JsonDoc": mockJsonMarshal(&MockJson{
					Kind: "Pod",
				}),
				"Values": map[string]interface{}{},
			},
			expectedYaml: ``,
			expectedTemplateValues: map[string]interface{}{
				"err": nil,
				"TaskResult": map[string]interface{}{
					"tskId": map[string]interface{}{
						"kind": "Pod",
					},
				},
			},
		},
		//
		// start of test scenario
		//
		`Negative Test - Throw notFoundErr error for empty kind
		  - NOTE: jsonpath, saveAs and notFoundErr are template functions
		  - Verify go templating throw error`: {
			templateInYaml: `
{{- jsonpath .JsonDoc "{.kind}" | saveAs "TaskResult.tskId.kind" .Values | notFoundErr "kind is missing" | saveIf "err" .Values | noop -}}
`,
			templateValues: map[string]interface{}{
				"JsonDoc": mockJsonMarshal(&MockJson{
					Kind: "",
				}),
				"Values": map[string]interface{}{},
			},
			expectedYaml: ``,
			expectedTemplateValues: map[string]interface{}{
				"err": &NotFoundError{"kind is missing"},
				"TaskResult": map[string]interface{}{
					"tskId": map[string]interface{}{
						"kind": "",
					},
				},
			},
		},
		//
		// start of test scenario
		//
		`Negative Test - Verify if 'nestedKeyMap' works with nil list`: {
			templateInYaml: `
{{- $kindArr := jsonpath .JsonDoc "{range .items[*]}{@.namespace}/{@.name}@kind={@.kind};{end}" | trim | splitList ";" -}}
{{- $kindArr | nestedKeyMap "/ @ =" .Values | noop -}}`,
			templateValues: map[string]interface{}{
				"JsonDoc": mockJsonListMarshal(nil),
				"Values":  map[string]interface{}{},
			},
			expectedYaml:           ``,
			expectedTemplateValues: map[string]interface{}{},
		},
		//
		// start of test scenario
		//
		`Negative Test - Verify if 'nestedKeyMap' works with empty byte`: {
			templateInYaml: `
{{- $kindArr := jsonpath .JsonDoc "{range .items[*]}{@.namespace}/{@.name}@kind={@.kind};{end}" | trim | splitList ";" -}}
{{- $kindArr | nestedKeyMap "/ @ =" .Values | noop -}}`,
			templateValues: map[string]interface{}{
				"JsonDoc": []byte{},
				"Values":  map[string]interface{}{},
			},
			expectedYaml:           ``,
			expectedTemplateValues: map[string]interface{}{},
		},
		//
		// start of test scenario
		//
		`Positive Test - Verify if template function 'nestedKeyMap' works as expected`: {
			templateInYaml: `
{{- "default/mypod@app=jiva openebs/mypod@app=cstor" | splitList " " | nestedKeyMap "@ =" .Values | noop -}}
{{- "default/mypod@backend=true default/mypod@app=jiva2" | splitList " " | nestedKeyMap "@ =" .Values | noop -}}
{{- "litmus/mypod@backend=true" | splitList " " | nestedKeyMap "/ @ =" .Values | noop -}}`,
			templateValues: map[string]interface{}{
				"Values": map[string]interface{}{},
			},
			expectedYaml: ``,
			expectedTemplateValues: map[string]interface{}{
				"default/mypod": map[string]interface{}{
					"app":     "jiva, jiva2",
					"backend": "true",
				},
				"openebs/mypod": map[string]interface{}{
					"app": "cstor",
				},
				"litmus": map[string]interface{}{
					"mypod": map[string]interface{}{
						"backend": "true",
					},
				},
			},
		},
		//
		// start of test scenario
		//
		`Positive Test - Verify if template function 'nestedKeyMap' works as expected with jsonpath`: {
			templateInYaml: `
{{- $kindArr := jsonpath .JsonDoc "{range .items[*]}{@.namespace}/{@.name}@kind={@.kind};{end}" | trim | splitList ";" -}}
{{- $kindArr | nestedKeyMap "/ @ =" .Values | noop -}}
kindOne: {{ .Values.openebs.mypod.kind }}
kindTwo: {{ .Values.default.mydeploy.kind }}`,
			templateValues: map[string]interface{}{
				"JsonDoc": mockJsonListMarshal(&MockJsonList{
					ApiVersion: "v1beta1",
					Items: []MockJson{
						MockJson{
							Kind:      "Pod",
							Namespace: "openebs",
							Name:      "mypod",
						},
						MockJson{
							Kind:      "Deployment",
							Namespace: "default",
							Name:      "mydeploy",
						},
					},
				}),
				"Values": map[string]interface{}{},
			},
			expectedYaml: `
kindOne: Pod
kindTwo: Deployment`,
			expectedTemplateValues: map[string]interface{}{
				"openebs": map[string]interface{}{
					"mypod": map[string]interface{}{
						"kind": "Pod",
					},
				},
				"default": map[string]interface{}{
					"mydeploy": map[string]interface{}{
						"kind": "Deployment",
					},
				},
			},
		},
		//
		// start of test scenario
		//
		`Positive Test - Verify if template function 'nestedKeyMap' works with multiple values for same key`: {
			templateInYaml: `
{{- $kindArr := jsonpath .JsonDoc "{range .items[*]}{@.namespace}/{@.name}@kind={@.kind};{end}" | trim | splitList ";" -}}
{{- $kindArr | nestedKeyMap "/ @ =" .Values | noop -}}
kindOne: {{ .Values.openebs.myapp.kind }}
kindTwo: {{ .Values.default.mydeploy.kind }}`,
			templateValues: map[string]interface{}{
				"JsonDoc": mockJsonListMarshal(&MockJsonList{
					ApiVersion: "v1beta1",
					Items: []MockJson{
						MockJson{
							Kind:      "Service",
							Namespace: "openebs",
							Name:      "myapp",
						},
						MockJson{
							Kind:      "Pod",
							Namespace: "openebs",
							Name:      "myapp",
						},
						MockJson{
							Kind:      "Deployment",
							Namespace: "default",
							Name:      "mydeploy",
						},
					},
				}),
				"Values": map[string]interface{}{},
			},
			expectedYaml: `
kindOne: Service, Pod
kindTwo: Deployment`,
			expectedTemplateValues: map[string]interface{}{
				"openebs": map[string]interface{}{
					"myapp": map[string]interface{}{
						"kind": "Service, Pod",
					},
				},
				"default": map[string]interface{}{
					"mydeploy": map[string]interface{}{
						"kind": "Deployment",
					},
				},
			},
		},
		//
		// start of test scenario
		//
		`Positive Test - Verify yaml rendering of the map generated via 'nestedKeyMap`: {
			templateInYaml: `
{{- $kindArr := jsonpath .JsonDoc "{range .items[*]}{@.namespace}/{@.name}@kind={@.kind};{end}" | trim | splitList ";" -}}
{{- $kindArr | nestedKeyMap "/ @ =" .Values | noop -}}
{{- .Values | toYaml -}}`,
			templateValues: map[string]interface{}{
				"JsonDoc": mockJsonListMarshal(&MockJsonList{
					ApiVersion: "v1beta1",
					Items: []MockJson{
						MockJson{
							Kind:      "Pod",
							Namespace: "openebs",
							Name:      "mypod",
						},
						MockJson{
							Kind:      "Deployment",
							Namespace: "default",
							Name:      "mydeploy",
						},
					},
				}),
				"Values": map[string]interface{}{},
			},
			expectedYaml: `
default:
  mydeploy: 
    kind: Deployment
openebs:
  mypod:
    kind: Pod`,
			expectedTemplateValues: map[string]interface{}{
				"openebs": map[string]interface{}{
					"mypod": map[string]interface{}{
						"kind": "Pod",
					},
				},
				"default": map[string]interface{}{
					"mydeploy": map[string]interface{}{
						"kind": "Deployment",
					},
				},
			},
		},
		//
		// start of test scenario
		//
		`Positive Test - Verify if 'keyMap' works as expected`: {
			templateInYaml: `
{{- "pkey=openebs,stor1=jiva,stor2=cstor" | splitList " " | keyMap "scenario" .Values | noop -}}
{{- "co1=swarm,co2=k8s" | splitList " " | keyMap "scenario" .Values | noop -}}
{{- "pkey=openebs,stor2=mstor" | splitList " " | keyMap "scenario" .Values | noop -}}`,
			templateValues: map[string]interface{}{
				"Values": map[string]interface{}{},
			},
			expectedYaml: ``,
			expectedTemplateValues: map[string]interface{}{
				"scenario": map[string]interface{}{
					"openebs": map[string]interface{}{
						"stor1": "jiva",
						"stor2": "cstor, mstor",
					},
					"pkey": map[string]interface{}{
						"co1": "swarm",
						"co2": "k8s",
					},
				},
			},
		},
		//
		// start of test scenario
		//
		`Positive Test - Verify yaml rendering of map generated via keyMap`: {
			templateInYaml: `
{{- $all := jsonpath .JsonDoc .Values.path | trim | splitList ";" -}}
{{- $all | keyMap "scenario" .Values | noop -}}
kind: MyList
apiVersion: v1alpha1
items:
# NOTE:
#   Below range and if blocks should end with }} 
# If they end with -}} a new line is not formed and makes the yaml invalid
{{- range $pkey, $val := .Values.scenario }}
  - label: {{ $pkey }}
    name: {{ pluck "name" $val | first }}
    kind: {{ pluck "kind" $val | first }}
    owner: {{ pluck "owner" $val | first }}
    count: {{ pluck "kind" $val | first | default "" | splitList ", " | len }}
# NOTE:
#   Below end statements can end with -}} as there are no more yaml items
{{- end -}}`,
			templateValues: map[string]interface{}{
				"JsonDoc": mockJsonListMarshal(&MockJsonList{
					ApiVersion: "v1beta1",
					Items: []MockJson{
						MockJson{
							Labels: map[string]string{
								"openebs.io/pv":                 "pvc-1234-abc",
								"openebs.io/controller-service": "jiva-controller-svc",
							},
							Owner: "Admin",
							Kind:  "Service",
							Name:  "myservice",
						},
						MockJson{
							Labels: map[string]string{
								"openebs.io/pv":         "pvc-1234-abc",
								"openebs.io/controller": "jiva-controller",
							},
							Owner: "User",
							Kind:  "Pod",
							Name:  "mypod",
						},
						MockJson{
							Labels: map[string]string{
								"openebs.io/pv":         "pvc-1234-abc-def",
								"openebs.io/controller": "jiva-controller-deploy",
							},
							Owner: "User",
							Kind:  "Deployment",
							Name:  "mydeploy",
						},
					},
				}),
				"Values": map[string]interface{}{
					"path": `{range .items[*]}pkey={@.labels.openebs\.io/pv},name={@.name},kind={@.kind},owner={@.owner};{end}`,
				},
			},
			expectedYaml: `
kind: MyList
apiVersion: v1alpha1
items:
  - label: pvc-1234-abc
    name: myservice, mypod
    kind: Service, Pod
    owner: Admin, User
    count: 2
  - label: pvc-1234-abc-def
    name: mydeploy
    kind: Deployment
    owner: User
    count: 1`,
			expectedTemplateValues: map[string]interface{}{
				"path": `{range .items[*]}pkey={@.labels.openebs\.io/pv},name={@.name},kind={@.kind},owner={@.owner};{end}`,
				"scenario": map[string]interface{}{
					"pvc-1234-abc": map[string]interface{}{
						"kind":  "Service, Pod",
						"owner": "Admin, User",
						"name":  "myservice, mypod",
					},
					"pvc-1234-abc-def": map[string]interface{}{
						"owner": "User",
						"name":  "mydeploy",
						"kind":  "Deployment",
					},
				},
			},
		},
		//
		// start of test scenario
		//
		`Negative Test: To test if go templating works for empty yaml
		  - Given a valid template function is invoked in yaml template, 
		  - And this template function rendering is removed via {{- and -}}
		  - And this template does not have any other yaml schema definition,
		  - Then this template should work fine when executed via go templating`: {
			templateInYaml: `
{{- jsonpath .JsonDoc "{.apiVersion}" | saveAs "TaskResult.tskId.apiVersion" .Values | noop -}}`,
			templateValues: map[string]interface{}{
				"JsonDoc": mockJsonListMarshal(&MockJsonList{
					ApiVersion: "v1beta1",
				}),
				"Values": map[string]interface{}{},
			},
			expectedYaml: ``,
			expectedTemplateValues: map[string]interface{}{
				"TaskResult": map[string]interface{}{
					"tskId": map[string]interface{}{
						"apiVersion": "v1beta1",
					},
				},
			},
		},
		//
		// start of test scenario
		//
		`Positive Test: To verify parsing of template values from within go template
		  - Set a variable called '$var' with value from this template's values,
		  - Get the apiversion via jsonpath as a string output, 
		  - Then save this apiversion value at .Values.TaskResult.tskId.<value_of_$var>,
		  - Then frame a yaml by parsing the apiversion value available in template values i.e. .Values.TaskResult.tskId.$var`: {
			templateInYaml: `
{{- $var := .Values.placeholder -}}
{{- jsonpath .JsonDoc "{.apiVersion}" | saveAs "TaskResult.tskId.myplace" .Values | noop -}}
{{- range $k, $v := .Values.TaskResult.tskId -}}
{{- if eq $k $var -}}
apiVersion: {{ $v }}
{{- end -}}
{{- end -}}`,
			templateValues: map[string]interface{}{
				"JsonDoc": mockJsonListMarshal(&MockJsonList{
					ApiVersion: "v1beta1",
				}),
				"Values": map[string]interface{}{
					"placeholder": "myplace",
				},
			},
			expectedYaml: `
apiVersion: v1beta1`,
			expectedTemplateValues: map[string]interface{}{
				"placeholder": "myplace",
				"TaskResult": map[string]interface{}{
					"tskId": map[string]interface{}{
						"myplace": "v1beta1",
					},
				},
			},
		},
		//
		// start of test scenario
		//
		"Positive Test - test by trimming the white spaced output and then check if empty": {
			templateInYaml: `
{{- jsonpath .JsonDoc "{.items[*].kind}" | trim | saveAs "TaskResult.tskId.kind" .Values | noop -}}
{{- .Values.TaskResult.tskId.kind | empty | saveAs "TaskResult.tskId.iskindempty" .Values | noop -}}
value: {{ .Values.TaskResult.tskId.kind }}
bool: {{ .Values.TaskResult.tskId.iskindempty }}`,
			templateValues: map[string]interface{}{
				"JsonDoc": mockJsonListMarshal(&MockJsonList{
					ApiVersion: "v1beta1",
					Items: []MockJson{
						MockJson{
							Kind:      "",
							Namespace: "openebs",
							Name:      "mypod",
						},
						MockJson{
							Kind:      "",
							Namespace: "openebs",
							Name:      "yourpod",
						},
					},
				}),
				"Values": map[string]interface{}{},
			},
			expectedYaml: `
value:
bool: true`,
			expectedTemplateValues: map[string]interface{}{
				"TaskResult": map[string]interface{}{
					"tskId": map[string]interface{}{
						"kind":        "",
						"iskindempty": true,
					},
				},
			},
		},
		//
		// start of test scenario
		//
		"Positive Test - test len of array of empty strings": {
			templateInYaml: `
{{- $noop := jsonpath .JsonDoc "{.items[*].kind}" | splitList " " | isLen 0 | saveAs "TaskResult.tskId.iskindlenzero" .Values -}}
show: {{ .Values.TaskResult.tskId.iskindlenzero }}`,
			templateValues: map[string]interface{}{
				"JsonDoc": mockJsonListMarshal(&MockJsonList{
					ApiVersion: "v1beta1",
					Items: []MockJson{
						MockJson{
							Kind:      "",
							Namespace: "openebs",
							Name:      "mypod",
						},
						MockJson{
							Kind:      "",
							Namespace: "openebs",
							Name:      "yourpod",
						},
					},
				}),
				"Values": map[string]interface{}{},
			},
			expectedYaml: `show: false`,
			expectedTemplateValues: map[string]interface{}{
				"TaskResult": map[string]interface{}{
					"tskId": map[string]interface{}{
						"iskindlenzero": false,
					},
				},
			},
		},
		//
		// start of test scenario
		//
		"Positive Test - test saving of array of empty strings": {
			templateInYaml: `
{{- $noop := jsonpath .JsonDoc "{.items[*].kind}" | splitList " " | saveAs "TaskResult.tskId.kinds" .Values -}}
show: {{ .Values.TaskResult.tskId.kinds }}`,
			templateValues: map[string]interface{}{
				"JsonDoc": mockJsonListMarshal(&MockJsonList{
					ApiVersion: "v1beta1",
					Items: []MockJson{
						MockJson{
							Kind:      "",
							Namespace: "openebs",
							Name:      "mypod",
						},
						MockJson{
							Kind:      "",
							Namespace: "openebs",
							Name:      "yourpod",
						},
					},
				}),
				"Values": map[string]interface{}{},
			},
			expectedYaml: `show: [ ]`,
			expectedTemplateValues: map[string]interface{}{
				"TaskResult": map[string]interface{}{
					"tskId": map[string]interface{}{
						"kinds": []string{"", ""},
					},
				},
			},
		},
		//
		// start of test scenario
		//
		"Positive Test - list of items - jsonpath | saveAs | len": {
			templateInYaml: `
apiVersion: {{ jsonpath .JsonDoc "{.apiVersion}" | saveAs "TaskResult.tskId.apiVersion" .Values }}
kinds: {{ jsonpath .JsonDoc "{.items[*].kind}" | saveAs "TaskResult.tskId.kinds" .Values }}
count: {{ .Values.TaskResult.tskId.kinds | splitList " " | len }}
`,
			templateValues: map[string]interface{}{
				"JsonDoc": mockJsonListMarshal(&MockJsonList{
					ApiVersion: "v1beta1",
					Items: []MockJson{
						MockJson{
							Kind:      "Pod",
							Namespace: "openebs",
							Name:      "mypod",
						},
						MockJson{
							Kind:      "Pod",
							Namespace: "openebs",
							Name:      "yourpod",
						},
					},
				}),
				"Values": map[string]interface{}{},
			},
			expectedYaml: `
apiVersion: v1beta1
kinds: Pod Pod
count: 2
`,
			expectedTemplateValues: map[string]interface{}{
				"TaskResult": map[string]interface{}{
					"tskId": map[string]interface{}{
						"apiVersion": "v1beta1",
						"kinds":      "Pod Pod",
					},
				},
			},
		},
		//
		// start of test scenario
		//
		"Positive Test - kind is not missing - saveAs multiple times ": {
			templateInYaml: `
{{- jsonpath .JsonDoc "{.kind}" | saveAs "TaskResult.tskId.kind" .Values | noop -}}
{{- saveAs "TaskResult.tskId.kind" .Values "Deployment" | noop -}}
kind: {{ .Values.TaskResult.tskId.kind }}
`,
			templateValues: map[string]interface{}{
				"JsonDoc": mockJsonMarshal(&MockJson{
					Kind:      "Pod",
					Namespace: "openebs",
					Name:      "",
				}),
				"Values": map[string]interface{}{},
			},
			expectedYaml: `
kind: Deployment
`,
			expectedTemplateValues: map[string]interface{}{
				"TaskResult": map[string]interface{}{
					"tskId": map[string]interface{}{
						"kind": "Deployment",
					},
				},
			},
		},
		//
		// start of test scenario
		//
		"Positive Test - kind is not missing - saveIf multiple times ": {
			templateInYaml: `
{{- jsonpath .JsonDoc "{.kind}" | saveIf "TaskResult.tskId.kind" .Values | noop -}}
{{- saveIf "TaskResult.tskId.kind" .Values "Deployment" | noop -}}
kind: {{ .Values.TaskResult.tskId.kind }}
`,
			templateValues: map[string]interface{}{
				"JsonDoc": mockJsonMarshal(&MockJson{
					Kind:      "Pod",
					Namespace: "openebs",
					Name:      "",
				}),
				"Values": map[string]interface{}{},
			},
			expectedYaml: `
kind: Pod
`,
			expectedTemplateValues: map[string]interface{}{
				"TaskResult": map[string]interface{}{
					"tskId": map[string]interface{}{
						"kind": "Pod",
					},
				},
			},
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			// augment the standard templating with sprig template functions
			tpl := template.New("TestTemplatingWithMutatingTemplateValues").Funcs(funcMap())
			tpl, err := tpl.Parse(mock.templateInYaml)
			if err != nil {
				t.Fatalf("test failed: expected 'no template parse error': actual '%s'", err.Error())
			}

			// buf is an io.Writer implementation
			// as required by the template
			var buf bytes.Buffer

			// execute the parsed yaml against the values
			// & write the result into the buffer
			err = tpl.Execute(&buf, mock.templateValues)
			if err != nil {
				t.Fatalf("test failed: expected 'no template execution error': actual '%s'", err.Error())
			}

			// buffer that represents a YAML can be unmarshalled into a map of any objects
			var objActual map[string]interface{}
			err = yaml.Unmarshal(buf.Bytes(), &objActual)
			if err != nil {
				t.Fatalf("test failed: expected 'no error on un-marshalling templated bytes to any objects': actual '%s'", err.Error())
			}

			// unmarshall the expected yaml into a map of any objects
			var objExpected map[string]interface{}
			err = yaml.Unmarshal([]byte(mock.expectedYaml), &objExpected)
			if err != nil {
				t.Fatalf("test failed: expected 'no error on un-marshalling expected yaml bytes to any objects': actual '%s'", err.Error())
			}

			// compare expected vs. actual object
			ok := reflect.DeepEqual(objExpected, objActual)
			if !ok {
				t.Fatalf("test failed:\n\nexpected yaml: '%s' \n\nactual yaml: '%s'", mock.expectedYaml, buf.Bytes())
			}

			// compare the values as values can get modified at runtime
			if !reflect.DeepEqual(mock.expectedTemplateValues, mock.templateValues["Values"]) {
				t.Fatalf("test failed:\n\nexpected template values: '%#v' \n\nactual template values: '%#v'", mock.expectedTemplateValues, mock.templateValues["Values"])
			}
		})
	}
}

func TestDynamicTemplating(t *testing.T) {
	tests := map[string]struct {
		ymlTpl      string
		values      map[string]interface{}
		ymlExpected string
	}{
		//
		// start of test scenario
		//
		"test dynamic templating - +ve - pluck | first ": {
			ymlTpl: `
{{- $ns := .defaultNamespace -}}
version: v1
kind: {{ .kind }}
label: {{ pluck $ns .labels | first }}
`,
			values: map[string]interface{}{
				"defaultNamespace": "openebs",
				"kind":             "pod",
				"labels": map[string]interface{}{
					"openebs": "cas-volume",
					"local":   "host-volume",
				},
			},
			ymlExpected: `
version: v1
kind: pod
label: cas-volume
`,
		},
		//
		// start of test scenario
		//
		"test dynamic templating - +ve - pluck then first": {
			ymlTpl: `
{{- $ns := .defaultNamespace -}}
{{- $nsLbl := pluck $ns .labels -}}
version: v1
kind: {{ .kind }}
label: {{ first $nsLbl }}
`,
			values: map[string]interface{}{
				"defaultNamespace": "openebs",
				"kind":             "pod",
				"labels": map[string]interface{}{
					"openebs": "cas-volume",
					"local":   "host-volume",
				},
			},
			ymlExpected: `
version: v1
kind: pod
label: cas-volume
`,
		},
		//
		// start of test scenario
		//
		"test dynamic templating - +ve - pick then nested ranges": {
			ymlTpl: `
{{- $ns := .runNamespace -}}
{{- $results := pick .taskResult.taskid $ns -}}
version: v1
kind: {{ .kind }}
{{- range $k, $v := $results }}
{{- range $kk, $vv := $v }}
objectName: {{ if eq $kk "objectName" }}{{ $vv }}{{ end }}
{{- end }}
{{- end }}`,
			values: map[string]interface{}{
				"runNamespace": "default-ns",
				"kind":         "pod",
				"taskResult": map[string]interface{}{
					"taskid": map[string]interface{}{
						"default-ns": map[string]string{
							"objectName": "my-replica-pod",
							"hostName":   "lenovo-laptop",
						},
						"openebs-ns": map[string]string{
							"objectName": "my-controller-pod",
							"hostName":   "k8s-minion-1",
						},
					},
				},
			},
			ymlExpected: `version: v1
kind: pod
objectName: my-replica-pod`,
		},
		//
		// start of test scenario
		//
		"test dynamic templating - +ve - pick then nested ranges then split then range": {
			ymlTpl: `
{{- $ns := .runNamespace -}}
{{- $nsResults := pick .taskResult.taskid $ns -}}
version: v1
requiredDuringSchedulingIgnoredDuringExecution:
  nodeSelectorTerms:
  - matchExpressions:
    - key: kubernetes.io/hostname
      operator: In
      values:
      {{- range $k, $v := $nsResults }}
      {{- range $kk, $vv := $v }}
      {{- if eq $kk "nodeNames" }}
      {{- $nodeNames := $vv }}
      {{- if ne $nodeNames "" }}
      {{- $nodeNamesMap := $nodeNames | split " " }}
      {{- range $kkk, $vvv := $nodeNamesMap }}
      - {{ $vvv }}
      {{- end }}
      {{- end }}
      {{- end }}
      {{- end }}
      {{- end }}
`,
			values: map[string]interface{}{
				"runNamespace": "openebs-ns",
				"taskResult": map[string]interface{}{
					"taskid": map[string]interface{}{
						"default-ns": map[string]string{
							"objectName": "my-replica-pod",
							"nodeNames":  "lenovo-laptop",
						},
						"openebs-ns": map[string]string{
							"objectName": "my-controller-pod",
							"nodeNames":  "k8s-minion-1 lenovo-laptop hp-laptop",
						},
					},
				},
			},
			ymlExpected: `version: v1
requiredDuringSchedulingIgnoredDuringExecution:
  nodeSelectorTerms:
  - matchExpressions:
    - key: kubernetes.io/hostname
      operator: In
      values:
      - k8s-minion-1
      - lenovo-laptop
      - hp-laptop`,
		},
		//
		// start of test scenario
		//
		"test dynamic templating - +ve - pick then nested ranges then splitList | first": {
			ymlTpl: `{{- $ns := .runNamespace -}}
{{- $nsResults := pick .taskResult.taskid $ns -}}
{{- range $k, $v := $nsResults }}
{{- range $kk, $vv := $v }}
{{- if eq $kk "nodeNames" }}
version: v1
requiredDuringSchedulingIgnoredDuringExecution:
  nodeSelectorTerms:
  - matchExpressions:
    - key: kubernetes.io/hostname
      operator: In
      values:
      {{- if ne $vv "" }}
      - {{ splitList " " $vv | first }}
      {{- end }}
{{- end }}
{{- end }}
{{- end }}`,
			values: map[string]interface{}{
				"runNamespace": "openebs-ns",
				"taskResult": map[string]interface{}{
					"taskid": map[string]interface{}{
						"default-ns": map[string]string{
							"objectName": "my-replica-pod",
							"nodeNames":  "lenovo-laptop",
						},
						"openebs-ns": map[string]string{
							"objectName": "my-controller-pod",
							"nodeNames":  "k8s-minion-1 lenovo-laptop hp-laptop",
						},
					},
				},
			},
			ymlExpected: `version: v1
requiredDuringSchedulingIgnoredDuringExecution:
  nodeSelectorTerms:
  - matchExpressions:
    - key: kubernetes.io/hostname
      operator: In
      values:
      - k8s-minion-1`,
		},
		//
		// start of test scenario
		//
		"test dynamic templating - boundary - pick then nested ranges then splitList of single-value | first": {
			ymlTpl: `{{- $ns := .runNamespace -}}
{{- $nsResults := pick .taskResult.taskid $ns -}}
{{- range $k, $v := $nsResults }}
{{- range $kk, $vv := $v }}
{{- if eq $kk "nodeNames" }}
version: v1
requiredDuringSchedulingIgnoredDuringExecution:
  nodeSelectorTerms:
  - matchExpressions:
    - key: kubernetes.io/hostname
      operator: In
      values:
      {{- if ne $vv "" }}
      - {{ splitList " " $vv | first }}
      {{- end }}
{{- end }}
{{- end }}
{{- end }}`,
			values: map[string]interface{}{
				"runNamespace": "default-ns",
				"taskResult": map[string]interface{}{
					"taskid": map[string]interface{}{
						"default-ns": map[string]string{
							"objectName": "my-replica-pod",
							// this is single value nodeNames
							"nodeNames": "lenovo-laptop",
						},
						"openebs-ns": map[string]string{
							"objectName": "my-controller-pod",
							"nodeNames":  "k8s-minion-1 lenovo-laptop hp-laptop",
						},
					},
				},
			},
			ymlExpected: `version: v1
requiredDuringSchedulingIgnoredDuringExecution:
  nodeSelectorTerms:
  - matchExpressions:
    - key: kubernetes.io/hostname
      operator: In
      values:
      - lenovo-laptop`,
		},
		//
		// start of test scenario
		//
		"test dynamic templating - boundary - pick then nested ranges then splitList of single-value-with-dangling-space | first": {
			ymlTpl: `{{- $ns := .runNamespace -}}
{{- $nsResults := pick .taskResult.taskid $ns -}}
{{- range $k, $v := $nsResults }}
{{- range $kk, $vv := $v }}
{{- if eq $kk "nodeNames" }}
version: v1
requiredDuringSchedulingIgnoredDuringExecution:
  nodeSelectorTerms:
  - matchExpressions:
    - key: kubernetes.io/hostname
      operator: In
      values:
      {{- if ne $vv "" }}
      - {{ splitList " " $vv | first }}
      {{- end }}
{{- end }}
{{- end }}
{{- end }}`,
			values: map[string]interface{}{
				"runNamespace": "default-ns",
				"taskResult": map[string]interface{}{
					"taskid": map[string]interface{}{
						"default-ns": map[string]string{
							"objectName": "my-replica-pod",
							// this is single value with dangling space
							"nodeNames": "lenovo-laptop ",
						},
						"openebs-ns": map[string]string{
							"objectName": "my-controller-pod",
							"nodeNames":  "k8s-minion-1 lenovo-laptop hp-laptop",
						},
					},
				},
			},
			ymlExpected: `version: v1
requiredDuringSchedulingIgnoredDuringExecution:
  nodeSelectorTerms:
  - matchExpressions:
    - key: kubernetes.io/hostname
      operator: In
      values:
      - lenovo-laptop`,
		},
		//
		// start of test scenario
		//
		"test dynamic templating - +ve - pick then nested ranges then splitList | len": {
			ymlTpl: `{{- $ns := .runNamespace -}}
{{- $nsResults := pick .taskResult.taskid $ns -}}
{{- range $k, $v := $nsResults }}
{{- range $kk, $vv := $v }}
{{- if eq $kk "nodeNames" }}
version: v1
requiredDuringSchedulingIgnoredDuringExecution:
  nodeCount: {{ if ne $vv "" }}{{ splitList " " $vv | len }}{{ end }}
{{- end }}
{{- end }}
{{- end }}`,
			values: map[string]interface{}{
				"runNamespace": "openebs-ns",
				"taskResult": map[string]interface{}{
					"taskid": map[string]interface{}{
						"default-ns": map[string]string{
							"objectName": "my-replica-pod",
							"nodeNames":  "lenovo-laptop ",
						},
						"openebs-ns": map[string]string{
							"objectName": "my-controller-pod",
							"nodeNames":  "k8s-minion-1 lenovo-laptop hp-laptop",
						},
					},
				},
			},
			ymlExpected: `version: v1
requiredDuringSchedulingIgnoredDuringExecution:
  nodeCount: 3`,
		},
		//
		// start of test scenario
		//
		"test dynamic templating - +ve - pick then nested ranges then splitList | join": {
			ymlTpl: `{{- $ns := .runNamespace -}}
{{- $nsResults := pick .taskResult.taskid $ns -}}
{{- range $k, $v := $nsResults }}
{{- range $kk, $vv := $v }}
{{- if eq $kk "nodeNames" }}
version: v1
requiredDuringSchedulingIgnoredDuringExecution:
  nodes: {{ if ne $vv "" }}{{ splitList " " $vv | join "," }}{{ end }}
{{- end }}
{{- end }}
{{- end }}`,
			values: map[string]interface{}{
				"runNamespace": "openebs-ns",
				"taskResult": map[string]interface{}{
					"taskid": map[string]interface{}{
						"default-ns": map[string]string{
							"objectName": "my-replica-pod",
							"nodeNames":  "lenovo-laptop ",
						},
						"openebs-ns": map[string]string{
							"objectName": "my-controller-pod",
							"nodeNames":  "k8s-minion-1 lenovo-laptop hp-laptop",
						},
					},
				},
			},
			ymlExpected: `version: v1
requiredDuringSchedulingIgnoredDuringExecution:
  nodes: k8s-minion-1,lenovo-laptop,hp-laptop`,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			// augment the standard templating with sprig template functions
			tpl := template.New("testdynamicvartemplating").Funcs(funcMap())
			tpl, err := tpl.Parse(mock.ymlTpl)
			if err != nil {
				t.Fatalf("failed to test dynamic templating: expected 'no instantiation error': actual '%s'", err.Error())
			}

			// buf is an io.Writer implementation
			// as required by the template
			var buf bytes.Buffer

			// execute the parsed yaml against the values
			// & write the result into the buffer
			err = tpl.Execute(&buf, mock.values)
			if err != nil {
				t.Fatalf("failed to test dynamic templating: expected 'no execution error': actual '%s'", err.Error())
			}

			// buffer that represents a YAML can be unmarshalled into a map of any objects
			var objActual map[string]interface{}
			err = yaml.Unmarshal(buf.Bytes(), &objActual)
			if err != nil {
				t.Fatalf("failed to test dynamic templating: expected 'no error w.r.t unmarshalling bytes to any objects': actual '%s'", err.Error())
			}

			// unmarshall the expected yaml into a map of any objects
			var objExpected map[string]interface{}
			err = yaml.Unmarshal([]byte(mock.ymlExpected), &objExpected)
			if err != nil {
				t.Fatalf("failed to test dynamic templating: expected 'no error w.r.t unmarshalling expected yaml to any objects': actual '%s'", err.Error())
			}

			// compare expected vs. actual object
			ok := reflect.DeepEqual(objExpected, objActual)
			if !ok {
				t.Fatalf("failed to test dynamic templating:\n\nexpected: '%s' \n\nactual: '%s'", mock.ymlExpected, buf.Bytes())
			}
		})
	}
}
