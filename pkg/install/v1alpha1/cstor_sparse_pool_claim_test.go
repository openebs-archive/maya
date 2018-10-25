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
// Rename this file by removing the version suffix information
package v1alpha1

import (
	"os"
	"testing"
)

func TestIsCstorSparsePool(t *testing.T) {
	tests := map[string]struct {
		value     string
		isenabled bool
		iserr     bool
	}{
		"with true": {
			value:     "true",
			isenabled: true,
			iserr:     false,
		},
		"with 1": {
			value:     "1",
			isenabled: true,
			iserr:     false,
		},
		"with false": {
			value:     "false",
			isenabled: false,
			iserr:     false,
		},
		"with 0": {
			value:     "0",
			isenabled: false,
			iserr:     false,
		},
		"with junk": {
			value:     "junk",
			isenabled: false,
			iserr:     false,
		},
		"with special chars": {
			value:     "abc:123-123",
			isenabled: false,
			iserr:     true,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			os.Unsetenv(string(DefaultCstorSparsePool))
			err := os.Setenv(string(DefaultCstorSparsePool), mock.value)
			if err != nil {
				t.Fatalf("Test '%s' failed %+v", name, err)
			}
			actual := IsCstorSparsePoolEnabled()
			if actual != mock.isenabled {
				t.Fatalf("Test '%s' failed: expected '%t' actual '%t' ", name, actual, mock.isenabled)
			}
			os.Unsetenv(string(DefaultCstorSparsePool))
		})
	}
}

func TestCstorSparsePoolSpc070(t *testing.T) {
	tests := map[string]struct {
		value    string
		expected int
		iserr    bool
	}{
		"with 1": {
			value:    "1",
			expected: 2,
			iserr:    false,
		},
		"with true": {
			value:    "true",
			expected: 2,
			iserr:    false,
		},
		"with 0": {
			value:    "0",
			expected: 0,
			iserr:    false,
		},
		"with false": {
			value:    "false",
			expected: 0,
			iserr:    false,
		},
		"with junk": {
			value:    "junk",
			expected: 0,
			iserr:    false,
		},
		"with special chars": {
			value:    "abc:123-123",
			expected: 0,
			iserr:    false,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			os.Unsetenv(string(DefaultCstorSparsePool))
			err := os.Setenv(string(DefaultCstorSparsePool), mock.value)
			if err != nil {
				t.Fatalf("Test '%s' failed %+v", name, err)
			}

			l := CstorSparsePoolArtifacts()
			actual := len(l.Items)
			if actual != mock.expected {
				t.Fatalf("Test '%s' failed: expected '%d' actual '%d'", name, mock.expected, actual)
			}

			os.Unsetenv(string(DefaultCstorSparsePool))
		})
	}
}
