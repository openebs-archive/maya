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

package template

import (
	"bytes"
	"fmt"
	. "github.com/openebs/maya/pkg/task/v1alpha1"
	"strings"
	"testing"
	"text/template"
)

// TestIsValidTemplateFunctionName verifies if suggested template function names
// are:
//
// 1. accepted by go based templating
func TestIsValidTemplateFunctionName(t *testing.T) {
	tests := map[string]struct {
		name string
	}{
		// possible execution based template functions
		"'do' as function name":     {"do"},
		"'run' as function name":    {"run"},
		"'runlog' as function name": {"runlog"},
		// possible action based template functions
		"'delete' as function name": {"delete"},
		"'get' as function name":    {"get"},
		"'create' as function name": {"create"},
		"'lst' as function name":    {"lst"},
		"'patch' as function name":  {"patch"},
		"'update' as function name": {"update"},
		// possible domain based template functions
		"'jiva' as function name":      {"jiva"},
		"'cstor' as function name":     {"cstor"},
		"'volume' as function name":    {"volume"},
		"'pool' as function name":      {"pool"},
		"'url' as function name":       {"url"},
		"'namespace' as function name": {"namespace"},
		// possible selector based template functions
		"'select' as function name": {"select"},
		"'where' as function name":  {"where"},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			f := template.FuncMap{mock.name: noop}
			tpl := template.New("is.func.name.valid").Funcs(f)
			tpl, err := tpl.Parse(fmt.Sprintf("{{- %s -}}", mock.name))
			if err != nil {
				t.Fatalf("Test '%s' failed: failed to parse: err '%+v'", name, err)
			}

			// buf is an io.Writer implementation
			// as required by the template
			var buf bytes.Buffer

			// execute the parsed yaml against the values
			// & write the result into the buffer
			var values map[string]interface{}
			err = tpl.Execute(&buf, values)
			if err != nil {
				t.Fatalf("Test '%s' failed: failed to execute: err '%+v'", name, err)
			}
		})
	}
}

// TestRunCommandTemplatingCombinations verifies the supported templating
// combinations made possible via `runtask command based templating functions`
func TestRunCommandTemplatingCombinations(t *testing.T) {
	tests := map[string]struct {
		templatefunc string
	}{
		// general jiva combinations
		"test 101": {"delete jiva volume | run"},
		"test 102": {"delete volume jiva | run"},
		"test 103": {"create jiva volume | run"},
		"test 104": {"create volume jiva | run"},
		"test 105": {"get jiva volume | run"},
		"test 106": {"get volume jiva | run"},
		"test 107": {"lst jiva volume | run"},
		"test 108": {"lst volume jiva | run"},
		// general cstor combinations
		"test 201": {"delete cstor volume | run"},
		"test 202": {"delete volume cstor | run"},
		"test 203": {"create cstor volume | run"},
		"test 204": {"create volume cstor | run"},
		"test 205": {"get cstor volume | run"},
		"test 206": {"get volume cstor | run"},
		"test 207": {"lst cstor volume | run"},
		"test 208": {"lst volume cstor | run"},
		// with url as input data combinations
		"test 301": {`delete jiva volume | url "http://10.10.10.10:9501/v1" | run`},
		"test 302": {`delete volume jiva | url "http://10.10.10.10:9501/v1" | run`},
		"test 303": {`delete cstor volume | url "http://10.10.10.10:9501/v1" | run`},
		"test 304": {`delete volume cstor | url "http://10.10.10.10:9501/v1" | run`},
		// with select path combinations
		"test 401": {`select "name" "namespace" | get jiva volume | run`},
		"test 402": {`select "all" | lst jiva volume | run`},
		"test 403": {`select "name" "namespace" | get cstor volume | run`},
		"test 404": {`select "all" | lst cstor volume | run`},
		// with runlog combinations
		"test 501": {`lst volume jiva | runlog "id.result" "id.extras" .Values`},
		"test 502": {`create volume jiva | runlog "id.result" "id.extras" .Values`},
		"test 503": {`get volume jiva | runlog "id.result" "id.extras" .Values`},
		"test 504": {`update volume jiva | runlog "id.result" "id.extras" .Values`},
		"test 505": {`patch volume jiva | runlog "id.result" "id.extras" .Values`},
		"test 506": {`delete volume jiva | runlog "id.result" "id.extras" .Values`},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			allfuncs := allCustomFuncs()
			tpl := template.New(fmt.Sprintf("%s", mock.templatefunc)).Funcs(allfuncs)
			tpl, err := tpl.Parse(fmt.Sprintf("{{- %s -}}", mock.templatefunc))
			if err != nil {
				t.Fatalf("Test '%s' failed: failed to parse: err '%+v'", name, err)
			}

			// buf is an io.Writer implementation
			// as required by the template
			var buf bytes.Buffer

			// execute the parsed yaml against the values
			// & write the result into the buffer
			values := map[string]interface{}{
				"Values": map[string]interface{}{},
			}
			err = tpl.Execute(&buf, values)
			if err != nil {
				t.Fatalf("Test '%s' failed: failed to execute: err '%+v'", name, err)
			}
		})
	}
}

func TestNotSupportedActionCommand(t *testing.T) {
	mockval := map[string]interface{}{"Values": map[string]interface{}{}}

	tests := map[string]struct {
		template string
		tvalues  map[string]interface{}
	}{
		// NOTE: If these combinations are supported in future then remove the
		// test case(s)
		"test 101": {`{{- create jiva volume | run -}}`, mockval},
		"test 102": {`{{- lst jiva volume | run -}}`, mockval},
		"test 103": {`{{- patch jiva volume | run -}}`, mockval},
		"test 104": {`{{- get jiva volume | run -}}`, mockval},
		"test 105": {`{{- update jiva volume | run -}}`, mockval},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			allfuncs := allCustomFuncs()
			tpl := template.New(mock.template).Funcs(allfuncs)
			tpl, err := tpl.Parse(mock.template)
			if err != nil {
				t.Fatalf("Test '%s' failed: failed to parse: err '%+v'", name, err)
			}

			// buf is an io.Writer implementation
			// as required by the template
			var buf bytes.Buffer

			// execute the parsed yaml against the values
			// & write the result into the buffer
			err = tpl.Execute(&buf, mock.tvalues)
			if err != nil {
				t.Fatalf("Test '%s' failed: failed to execute: err '%+v'", name, err)
			}

			if len(buf.String()) == 0 {
				t.Fatalf("Test '%s' failed: nil result", name)
			}

			if !strings.Contains(buf.String(), NotSupportedActionError.Error()) {
				t.Fatalf("Test '%s' failed: expected error '%s' actual: '%s'", name, NotSupportedActionError.Error(), buf.String())
			}
		})
	}
}

func TestNotSupportedCategoryCommand(t *testing.T) {
	mockval := map[string]interface{}{"Values": map[string]interface{}{}}

	tests := map[string]struct {
		template string
		tvalues  map[string]interface{}
	}{
		// NOTE: If these combinations are supported in future then remove the
		// test case(s)
		"test 101": {`{{- create cstor volume | run -}}`, mockval},
		"test 102": {`{{- lst cstor volume | run -}}`, mockval},
		"test 103": {`{{- patch cstor volume | run -}}`, mockval},
		"test 104": {`{{- get cstor volume | run -}}`, mockval},
		"test 105": {`{{- update cstor volume | run -}}`, mockval},
		"test 106": {`{{- delete cstor volume | run -}}`, mockval},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			allfuncs := allCustomFuncs()
			tpl := template.New(mock.template).Funcs(allfuncs)
			tpl, err := tpl.Parse(mock.template)
			if err != nil {
				t.Fatalf("Test '%s' failed: failed to parse: err '%+v'", name, err)
			}

			// buf is an io.Writer implementation
			// as required by the template
			var buf bytes.Buffer

			// execute the parsed yaml against the values
			// & write the result into the buffer
			err = tpl.Execute(&buf, mock.tvalues)
			if err != nil {
				t.Fatalf("Test '%s' failed: failed to execute: err '%+v'", name, err)
			}

			if len(buf.String()) == 0 {
				t.Fatalf("Test '%s' failed: nil result", name)
			}

			if !strings.Contains(buf.String(), NotSupportedCategoryError.Error()) {
				t.Fatalf("Test '%s' failed: expected error '%s' actual: '%s'", name, NotSupportedCategoryError.Error(), buf.String())
			}
		})
	}
}

func TestDeleteJivaVolumeCommand(t *testing.T) {
	mockval := map[string]interface{}{"Values": map[string]interface{}{}}

	tests := map[string]struct {
		template string
		tvalues  map[string]interface{}
	}{
		"test 101": {`{{- delete jiva volume | run -}}`, mockval},
		"test 102": {`{{- delete jiva volume | url "" | run -}}`, mockval},
		"test 103": {`{{- delete jiva volume | url "http://1.1.1.1" | run -}}`, mockval},
		"test 104": {`{{- delete jiva volume | url "http://1.1.1.1:1010/v1" | run -}}`, mockval},
		"test 105": {`{{- $url := "http://1.1.1.1:1010/v1" -}}
		              {{- delete jiva volume | url $url | run -}}`, mockval},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			allfuncs := allCustomFuncs()
			tpl := template.New(mock.template).Funcs(allfuncs)
			tpl, err := tpl.Parse(mock.template)
			if err != nil {
				t.Fatalf("Test '%s' failed: failed to parse: err '%+v'", name, err)
			}

			// buf is an io.Writer implementation
			// as required by the template
			var buf bytes.Buffer

			// execute the parsed yaml against the values
			// & write the result into the buffer
			err = tpl.Execute(&buf, mock.tvalues)
			if err != nil {
				t.Fatalf("Test '%s' failed: failed to execute: err '%+v'", name, err)
			}

			op := buf.String()
			if len(op) == 0 {
				t.Fatalf("Test '%s' failed: nil result", name)
			}
		})
	}
}

func TestDeleteJivaVolumeSaveAs(t *testing.T) {
	tests := map[string]struct {
		template string
	}{
		// NOTE:
		//  Name of the test case and saveas key needs to be same
		"101": {`{{- delete jiva volume | run | saveas "101" .Values -}}`},
		"102": {`{{- delete jiva volume | url "" | name "" | run | saveas "102" .Values -}}`},
		"103": {`{{- delete jiva volume | url "http://1.1.1.1" | name "ab" | run | saveas "103" .Values -}}`},
		"104": {`{{- delete jiva volume | url "http://1.1.1.1:1010" | name "abc" | run | saveas "104" .Values -}}`},
		"105": {`{{- $url := "http://1.1.1.1:1010/v1" -}}
		         {{- delete jiva volume | url $url | name "abcd" | run | saveas "105" .Values -}}`},
		"106": {`{{- $url := "http://1.1.1.1:1010/v1/volumes" -}}
		         {{- delete jiva volume | url $url | name "abcde" | run | saveas "106" .Values -}}`},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			mockval := map[string]interface{}{"Values": map[string]interface{}{}}
			tpl := template.New(mock.template).Funcs(allCustomFuncs())
			tpl, err := tpl.Parse(mock.template)
			if err != nil {
				t.Fatalf("Test '%s' failed: failed to parse: err '%+v'", name, err)
			}

			// buf is an io.Writer implementation
			// as required by the template
			var buf bytes.Buffer

			// execute the parsed yaml against the values
			// & write the result into the buffer
			err = tpl.Execute(&buf, mockval)
			if err != nil {
				t.Fatalf("Test '%s' failed: failed to execute: err '%+v'", name, err)
			}

			op := buf.String()
			if len(op) == 0 {
				t.Fatalf("Test '%s' failed: nil result", name)
			}

			tvalues := mockval["Values"]
			if tvalues == nil {
				t.Fatalf("Test '%s' failed: nil template values post template execution", name)
			}

			response := tvalues.(map[string]interface{})[name]
			if response == nil {
				t.Fatalf("Test '%s' failed: nil runtask command response post template execution", name)
			}

			cmdres := response.(map[string]interface{})
			cmdresult := cmdres["result"]
			if cmdresult != nil {
				t.Fatalf("Test '%s' failed: expected nil result: actual result %#v", name, cmdresult)
			}

			cmderr := cmdres["error"]
			if cmderr == nil {
				t.Fatalf("Test '%s' failed: expected not nil error: actual nil error", name)
			}

			cmddebug := cmdres["debug"]
			if cmddebug == nil {
				t.Fatalf("Test '%s' failed: expected not nil debug: actual nil debug", name)
			}
		})
	}
}

func TestDeleteJivaVolumeSaveAsVerifyError(t *testing.T) {
	tests := map[string]struct {
		template string
	}{
		// NOTE:
		//  Name of the test case should equal to the key used by saveas & saveif
		// functions
		"t101": {`{{- $url := "http://1.1.1.1:1010" -}}
		         {{- delete jiva volume | url $url | name "myvol" | run | saveas "t101" .Values -}}
		         {{- $err := toString .Values.t101.error -}}
		         {{- $err | empty | not | verifyErr $err | saveif "t101.verifyerr" .Values | noop -}}`},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			mockval := map[string]interface{}{"Values": map[string]interface{}{}}
			tpl := template.New(mock.template).Funcs(allCustomFuncs())
			tpl, err := tpl.Parse(mock.template)
			if err != nil {
				t.Fatalf("Test '%s' failed: failed to parse: err '%+v'", name, err)
			}

			// buf is an io.Writer implementation
			// as required by the template
			var buf bytes.Buffer

			// execute the parsed yaml against the values
			// & write the result into the buffer
			err = tpl.Execute(&buf, mockval)
			if err != nil {
				t.Fatalf("Test '%s' failed: failed to execute: err '%+v'", name, err)
			}

			op := buf.String()
			if len(op) == 0 {
				t.Fatalf("Test '%s' failed: nil result", name)
			}

			tvalues := mockval["Values"]
			if tvalues == nil {
				t.Fatalf("Test '%s' failed: nil template values post template execution", name)
			}

			response := tvalues.(map[string]interface{})[name]
			if response == nil {
				t.Fatalf("Test '%s' failed: nil runtask command response post template execution", name)
			}

			cmdres := response.(map[string]interface{})
			verifyerr := cmdres["verifyerr"]
			if verifyerr == nil {
				t.Fatalf("Test '%s' failed: expected not nil verifyerr: actual verifyerr %#v", name, verifyerr)
			}

			var verr error
			verr, ok := verifyerr.(*VerifyError)
			if !ok {
				t.Fatalf("Test '%s' failed: expected VerifyErr error: actual %#v", name, verr)
			}
		})
	}
}
