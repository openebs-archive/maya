// Copyright © 2018-2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"
)

func fakePredicate() Predicate {
	return func(p *Patch) bool {
		return true
	}
}

func TestNewBuilder(t *testing.T) {
	tests := map[string]struct {
		expectPatch  bool
		expectChecks bool
	}{
		"call NewBuilder": {
			true, true,
		},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			b := NewBuilder()
			if (b.patch != nil) != mock.expectPatch {
				t.Fatalf("test %s failed, expect patch: %t but got: %t",
					name, mock.expectPatch, b.patch != nil)
			}
			if (b.checks != nil) != mock.expectChecks {
				t.Fatalf("test %s failed, expect checks: %t but got: %t",
					name, mock.expectChecks, b.checks != nil)
			}
		})
	}
}

func TestBuilderForObject(t *testing.T) {
	tests := map[string]struct {
		inputType    types.PatchType
		inputObject  []byte
		expectedType types.PatchType
		expectedObj  []byte
		expectChecks bool
	}{
		"call BuilderForObject with patch type and patch object": {
			"application/json-patch+json",
			[]byte("abc"),
			"application/json-patch+json",
			[]byte("abc"),
			true,
		},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			b := BuilderForObject(mock.inputType, mock.inputObject)
			gotType := b.patch.Type
			gotObject := b.patch.Object
			if gotType != mock.expectedType {
				t.Fatalf("test %s failed, expected type %+v but got : %+v",
					name, mock.expectedType, gotType)
			}
			if string(gotObject) != string(mock.expectedObj) {
				t.Fatalf("test %s failed, expected obj %s but got : %s",
					name, string(mock.expectedObj), string(gotObject))
			}
			checks := (b.checks != nil)
			if checks != mock.expectChecks {
				t.Fatalf("test %s failed, expected non-nil checks but got : %+v",
					name, b.checks)
			}
		})
	}
}

func TestIsValidType(t *testing.T) {
	tests := map[string]struct {
		patch          *Patch
		expectedOutput bool
	}{
		"Patch with type application/json-patch+json": {
			&Patch{Type: "application/json-patch+json"},
			true,
		},
		"Patch with type json": {
			&Patch{Type: "json"},
			false,
		},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			isValid := mock.patch.IsValidType()
			if isValid != mock.expectedOutput {
				t.Fatalf("test %s failed, expected %+v but got : %+v",
					name, mock.expectedOutput, isValid)
			}
		})
	}
}

func TestAddCheck(t *testing.T) {
	tests := map[string]struct {
		input                Predicate
		expectedChecksLength int
	}{
		"When a predicate is given": {
			fakePredicate(),
			1,
		},
		"When nil is given": {
			nil,
			0,
		},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			b := NewBuilder().AddCheck(mock.input)
			if len(b.checks) != mock.expectedChecksLength {
				t.Fatalf("test %s failed, expected checks length %+v but got : %+v",
					name, mock.expectedChecksLength, len(b.checks))
			}
		})
	}
}
