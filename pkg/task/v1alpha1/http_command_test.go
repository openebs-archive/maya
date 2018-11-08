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
	"testing"

	. "github.com/openebs/maya/pkg/client/http/v1alpha1"
)

var _ Runner = &httpCommand{}
var _ Runner = &httpGet{}
var _ Runner = &httpPost{}
var _ Runner = &httpPut{}
var _ Runner = &httpDelete{}
var _ Runner = &httpPatch{}

func TestHttpCommandInstance(t *testing.T) {
	tests := map[string]struct {
		action      RunCommandAction
		isSupported bool
	}{
		"101": {DeleteCommandAction, true},
		"102": {CreateCommandAction, false},
		"103": {GetCommandAction, true},
		"104": {ListCommandAction, false},
		"105": {PatchCommandAction, true},
		"106": {UpdateCommandAction, false},
		"107": {PostCommandAction, true},
		"108": {PutCommandAction, true},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := WithAction(Command(), mock.action)
			j := &httpCommand{RunCommand: cmd}
			r := j.instance()
			if r == nil {
				t.Fatalf("Test '%s' failed: expected not nil runner: actual nil runner", name)
			}

			if mock.isSupported {
				if _, nosupport := r.(*notSupportedActionCommand); nosupport {
					t.Fatalf("Test '%s' failed: expected supported command instance: actual '%#v'", name, r)
				}
			}

			if !mock.isSupported {
				if _, nosupport := r.(*notSupportedActionCommand); !nosupport {
					t.Fatalf("Test '%s' failed: expected not supported command instance: actual '%#v'", name, r)
				}
			}
		})
	}
}

func TestHttpCommandSetURL(t *testing.T) {
	tests := map[string]struct {
		origURL     string
		newURL      string
		expectedURL string
	}{
		"101": {"", "http://1.1.1.1:1010", "http://1.1.1.1:1010"},
		"102": {"http://", "http://1.1.1.1", "http://1.1.1.1"},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := Command()
			j := &httpCommand{RunCommand: cmd, url: mock.origURL}
			j.withURL(mock.newURL)

			if mock.expectedURL != j.url {
				t.Fatalf("Test '%s' failed: expected url '%s': actual url '%s'", name, mock.expectedURL, j.url)
			}
		})
	}
}

func TestHttpCommandInvokeURL(t *testing.T) {
	tests := map[string]struct {
		url  string
		verb HttpVerb
	}{
		"101": {"", DeleteAction},
		"102": {"http://", DeleteAction},
		"201": {"", GetAction},
		"202": {"http://", GetAction},
		"301": {"", PutAction},
		"302": {"http://", PutAction},
		"401": {"", PostAction},
		"402": {"http://", PostAction},
		"501": {"", PatchAction},
		"502": {"http://", PatchAction},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := Command()
			j := &httpCommand{RunCommand: cmd, url: mock.url}
			_, err := j.invokeURL(mock.verb)

			if err == nil {
				t.Fatalf("Test '%s' failed: expected error: actual no error", name)
			}
		})
	}
}

func TestHttpCommandInvokeAPI(t *testing.T) {
	tests := map[string]struct {
		url      string
		verb     HttpVerb
		resource string
	}{
		"101": {"", DeleteAction, ""},
		"102": {"http://", DeleteAction, "volume"},
		"103": {"http://", DeleteAction, "pool"},
		"201": {"", GetAction, ""},
		"202": {"http://", GetAction, "volume"},
		"203": {"http://", GetAction, "pool"},
		"301": {"", PutAction, ""},
		"302": {"http://", PutAction, "volume"},
		"303": {"http://", PutAction, "pool"},
		"401": {"", PostAction, ""},
		"402": {"http://", PostAction, "volume"},
		"403": {"http://", PostAction, "pool"},
		"501": {"", PatchAction, ""},
		"502": {"http://", PatchAction, "volume"},
		"503": {"http://", PatchAction, "pool"},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := Command()
			j := &httpCommand{RunCommand: cmd, url: mock.url}
			_, err := j.invokeAPI(mock.verb, mock.resource)

			if err == nil {
				t.Fatalf("Test '%s' failed: expected error: actual no error", name)
			}
		})
	}
}

func TestHttpCommandInvoke(t *testing.T) {
	tests := map[string]struct {
		url      string
		verb     HttpVerb
		resource string
	}{
		"101": {"", DeleteAction, ""},
		"102": {"http://", DeleteAction, "volume"},
		"103": {"http://", DeleteAction, "pool"},
		"201": {"", GetAction, ""},
		"202": {"http://", GetAction, "volume"},
		"203": {"http://", GetAction, "pool"},
		"301": {"", PutAction, ""},
		"302": {"http://", PutAction, "volume"},
		"303": {"http://", PutAction, "pool"},
		"401": {"", PostAction, ""},
		"402": {"http://", PostAction, "volume"},
		"403": {"http://", PostAction, "pool"},
		"501": {"", PutAction, ""},
		"502": {"http://", PutAction, "volume"},
		"503": {"http://", PutAction, "pool"},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := WithData(Command(), "name", mock.resource)
			j := &httpCommand{RunCommand: cmd, url: mock.url}
			result := j.invoke(mock.verb)

			if result.Error() == nil {
				t.Fatalf("Test '%s' failed: expected error: actual no error", name)
			}

			if result.Result() != nil {
				t.Fatalf("Test '%s' failed: expected nil result: actual '%#v'", name, result.Result())
			}
		})
	}
}

func TestHttpCommandRun(t *testing.T) {
	tests := map[string]struct {
		action   RunCommandAction
		url      string
		resource string
	}{
		"101": {DeleteCommandAction, "", ""},
		"102": {DeleteCommandAction, "http://", "volume"},
		"103": {DeleteCommandAction, "http://", "pool"},
		"201": {GetCommandAction, "", ""},
		"202": {GetCommandAction, "http://", "volume"},
		"203": {GetCommandAction, "http://", "pool"},
		"301": {PutCommandAction, "", ""},
		"302": {PutCommandAction, "http://", "volume"},
		"303": {PutCommandAction, "http://", "pool"},
		"401": {PostCommandAction, "", ""},
		"402": {PostCommandAction, "http://", "volume"},
		"403": {PostCommandAction, "http://", "pool"},
		"501": {PatchCommandAction, "", ""},
		"502": {PatchCommandAction, "http://", "volume"},
		"503": {PatchCommandAction, "http://", "pool"},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := WithData(Command(), "name", mock.resource)
			cmd = WithAction(cmd, mock.action)
			j := &httpCommand{RunCommand: cmd, url: mock.url}
			result := j.Run()

			if result.Error() == nil {
				t.Fatalf("Test '%s' failed: expected error: actual no error", name)
			}

			if result.Result() != nil {
				t.Fatalf("Test '%s' failed: expected nil result: actual '%#v'", name, result.Result())
			}
		})
	}
}

func TestHttpDeleteRun(t *testing.T) {
	tests := map[string]struct {
		url      string
		resource string
	}{
		"101": {"", ""},
		"102": {"http://", "volume"},
		"103": {"http://", "pool"},
		"201": {"", ""},
		"202": {"http://", "volume"},
		"203": {"http://", "pool"},
		"301": {"", ""},
		"302": {"http://", "volume"},
		"303": {"http://", "pool"},
		"401": {"", ""},
		"402": {"http://", "volume"},
		"403": {"http://", "pool"},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := WithData(Command(), "name", mock.resource)
			h := &httpCommand{RunCommand: cmd, url: mock.url}
			d := &httpDelete{h}
			result := d.Run()

			if result.Error() == nil {
				t.Fatalf("Test '%s' failed: expected error: actual no error", name)
			}

			if result.Result() != nil {
				t.Fatalf("Test '%s' failed: expected nil result: actual '%#v'", name, result.Result())
			}
		})
	}
}

func TestHttpGetRun(t *testing.T) {
	tests := map[string]struct {
		url      string
		resource string
	}{
		"101": {"", ""},
		"102": {"http://", "volume"},
		"103": {"http://", "pool"},
		"201": {"", ""},
		"202": {"http://", "volume"},
		"203": {"http://", "pool"},
		"301": {"", ""},
		"302": {"http://", "volume"},
		"303": {"http://", "pool"},
		"401": {"", ""},
		"402": {"http://", "volume"},
		"403": {"http://", "pool"},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := WithData(Command(), "name", mock.resource)
			h := &httpCommand{RunCommand: cmd, url: mock.url}
			g := &httpGet{h}
			result := g.Run()

			if result.Error() == nil {
				t.Fatalf("Test '%s' failed: expected error: actual no error", name)
			}

			if result.Result() != nil {
				t.Fatalf("Test '%s' failed: expected nil result: actual '%#v'", name, result.Result())
			}
		})
	}
}

func TestHttpPostRun(t *testing.T) {
	tests := map[string]struct {
		url      string
		resource string
	}{
		"101": {"", ""},
		"102": {"http://", "volume"},
		"103": {"http://", "pool"},
		"201": {"", ""},
		"202": {"http://", "volume"},
		"203": {"http://", "pool"},
		"301": {"", ""},
		"302": {"http://", "volume"},
		"303": {"http://", "pool"},
		"401": {"", ""},
		"402": {"http://", "volume"},
		"403": {"http://", "pool"},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := WithData(Command(), "name", mock.resource)
			h := &httpCommand{RunCommand: cmd, url: mock.url}
			p := &httpPost{h}
			result := p.Run()

			if result.Error() == nil {
				t.Fatalf("Test '%s' failed: expected error: actual no error", name)
			}

			if result.Result() != nil {
				t.Fatalf("Test '%s' failed: expected nil result: actual '%#v'", name, result.Result())
			}
		})
	}
}

func TestHttpPutRun(t *testing.T) {
	tests := map[string]struct {
		url      string
		resource string
	}{
		"101": {"", ""},
		"102": {"http://", "volume"},
		"103": {"http://", "pool"},
		"201": {"", ""},
		"202": {"http://", "volume"},
		"203": {"http://", "pool"},
		"301": {"", ""},
		"302": {"http://", "volume"},
		"303": {"http://", "pool"},
		"401": {"", ""},
		"402": {"http://", "volume"},
		"403": {"http://", "pool"},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := WithData(Command(), "name", mock.resource)
			h := &httpCommand{RunCommand: cmd, url: mock.url}
			p := &httpPut{h}
			result := p.Run()

			if result.Error() == nil {
				t.Fatalf("Test '%s' failed: expected error: actual no error", name)
			}

			if result.Result() != nil {
				t.Fatalf("Test '%s' failed: expected nil result: actual '%#v'", name, result.Result())
			}
		})
	}
}

func TestHttpPatchRun(t *testing.T) {
	tests := map[string]struct {
		url      string
		resource string
	}{
		"101": {"", ""},
		"102": {"http://", "volume"},
		"103": {"http://", "pool"},
		"201": {"", ""},
		"202": {"http://", "volume"},
		"203": {"http://", "pool"},
		"301": {"", ""},
		"302": {"http://", "volume"},
		"303": {"http://", "pool"},
		"401": {"", ""},
		"402": {"http://", "volume"},
		"403": {"http://", "pool"},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := WithData(Command(), "name", mock.resource)
			h := &httpCommand{RunCommand: cmd, url: mock.url}
			p := &httpPatch{h}
			result := p.Run()

			if result.Error() == nil {
				t.Fatalf("Test '%s' failed: expected error: actual no error", name)
			}

			if result.Result() != nil {
				t.Fatalf("Test '%s' failed: expected nil result: actual '%#v'", name, result.Result())
			}
		})
	}
}

func TestWithunmarshal(t *testing.T) {
	tests := map[string]struct {
		command        *httpCommand
		expectedOutput bool
	}{
		"When unmarshal options is invoked": {
			command: &httpCommand{
				RunCommand: &RunCommand{
					Data: RunCommandDataMap{
						"unmarshal": false,
					},
				},
			},
			expectedOutput: false,
		},
		"When unmarshal options not invoked": {
			command: &httpCommand{
				RunCommand: &RunCommand{
					Data: RunCommandDataMap{},
				},
			},
			expectedOutput: true,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			c := mock.command.withUnmarshal()

			if c.doUnmarshal != mock.expectedOutput {
				t.Fatalf("Test Name: %v Expected: %v Got: %v", name, mock.expectedOutput, c.doUnmarshal)
			}
		})
	}
}
