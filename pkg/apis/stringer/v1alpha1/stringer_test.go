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
	"math"
	"strings"
	"testing"
)

type fakeStruct struct {
	String string
	Int    int
	Bool   bool
	Float  float64
}

func TestTypeMetaString(t *testing.T) {
	tests := map[string]struct {
		object              interface{}
		expectedStringParts []string
	}{
		"struct as input": {
			fakeStruct{
				String: "fake-string",
				Int:    123,
				Bool:   true,
				Float:  64.64,
			},
			[]string{"Int: 123", "String: fake-string", "Bool: true", "Float: 64.64"},
		},
		"struct pointer as input": {
			&fakeStruct{
				String: "fake-string",
				Int:    123,
				Bool:   true,
				Float:  64.64,
			},
			[]string{"Int: 123", "String: fake-string", "Bool: true", "Float: 64.64"},
		},
		"integer as input": {
			64,
			[]string{"64"},
		},
		"integer pointer as input": {
			func() *int {
				i := 64
				return &i
			}(),
			[]string{"64"},
		},
		"float as input": {
			64.64,
			[]string{"64.64"},
		},
		"float pointer as input": {
			func() *float64 {
				f := 64.64
				return &f
			}(),
			[]string{"64.64"},
		},
		"bool as input": {
			true,
			[]string{"true"},
		},
		"bool pointer as input": {
			func() *bool {
				b := true
				return &b
			}(),
			[]string{"true"},
		},
		"string as input": {
			"fake-string",
			[]string{"fake-string"},
		},
		"string pointer as input": {
			func() *string {
				s := "fake-string"
				return &s
			}(),
			[]string{"fake-string"},
		},
		"channel as input": {
			make(chan int),
			[]string{"{nil}"},
		},
		"unsupported float64 value as input": {
			math.Inf(1),
			[]string{"{nil}"},
		},
		"nil object": {
			nil,
			[]string{"{nil}"},
		},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			ymlstr := Yaml("context", mock.object)
			for _, expect := range mock.expectedStringParts {
				if !strings.Contains(ymlstr, expect) {
					t.Errorf("test '%s' failed: expected '%s' in '%s'", name, expect, ymlstr)
				}
			}
		})
	}
}

func TestJSONIndent(t *testing.T) {
	tests := map[string]struct {
		object              interface{}
		expectedStringParts []string
	}{
		"struct as input": {
			fakeStruct{
				String: "fake-string",
				Int:    123,
				Bool:   true,
				Float:  64.64,
			},
			[]string{`"Int": 123`, `"Bool": true`, `"Float": 64.64`, `"String": "fake-string"`},
		},
		"struct pointer as input": {
			&fakeStruct{
				String: "fake-string",
				Int:    123,
				Bool:   true,
				Float:  64.64,
			},
			[]string{`"Int": 123`, `"Bool": true`, `"Float": 64.64`, `"String": "fake-string"`},
		},
		"integer as input": {
			64,
			[]string{"64"},
		},
		"integer pointer as input": {
			func() *int {
				i := 64
				return &i
			}(),
			[]string{"64"},
		},
		"float as input": {
			64.64,
			[]string{"64.64"},
		},
		"float pointer as input": {
			func() *float64 {
				f := 64.64
				return &f
			}(),
			[]string{"64.64"},
		},
		"bool as input": {
			true,
			[]string{"true"},
		},
		"bool pointer as input": {
			func() *bool {
				b := true
				return &b
			}(),
			[]string{"true"},
		},
		"string as input": {
			"fake-string",
			[]string{"fake-string"},
		},
		"string pointer as input": {
			func() *string {
				s := "fake-string"
				return &s
			}(),
			[]string{"fake-string"},
		},
		"channel as input": {
			make(chan int),
			[]string{"{nil}"},
		},
		"unsupported float64 value as input": {
			math.Inf(1),
			[]string{"{nil}"},
		},
		"nil object": {
			nil,
			[]string{"{nil}"},
		},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			ymlstr := JSONIndent("context", mock.object)
			for _, expect := range mock.expectedStringParts {
				if !strings.Contains(ymlstr, expect) {
					t.Errorf("test '%s' failed: expected '%s' in '%s'", name, expect, ymlstr)
				}
			}
		})
	}
}
