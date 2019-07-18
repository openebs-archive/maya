/*
Copyright 2018-2019 The OpenEBS Authors

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
		defConf   string
		isenabled bool
		iserr     bool
	}{
		"with true, default config false": {
			value:     "true",
			defConf:   "false",
			isenabled: false,
			iserr:     false,
		},
		"with true": {
			value:     "true",
			defConf:   "true",
			isenabled: true,
			iserr:     false,
		},
		"with 1": {
			value:     "1",
			defConf:   "true",
			isenabled: true,
			iserr:     false,
		},
		"with false": {
			value:     "false",
			defConf:   "true",
			isenabled: false,
			iserr:     false,
		},
		"with 0": {
			value:     "0",
			defConf:   "true",
			isenabled: false,
			iserr:     false,
		},
		"with junk": {
			value:     "junk",
			defConf:   "true",
			isenabled: false,
			iserr:     false,
		},
		"with special chars": {
			value:     "abc:123-123",
			defConf:   "true",
			isenabled: false,
			iserr:     true,
		},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			os.Unsetenv(string(CreateDefaultStorageConfig))
			errDef := os.Setenv(string(CreateDefaultStorageConfig), mock.defConf)
			if errDef != nil {
				t.Fatalf("Test '%s' failed %+v", name, errDef)
			}
			os.Unsetenv(string(DefaultCstorSparsePool))
			err := os.Setenv(string(DefaultCstorSparsePool), mock.value)
			if err != nil {
				t.Fatalf("Test '%s' failed %+v", name, err)
			}
			actual := IsCstorSparsePoolEnabled()
			actualStgConfig := IsDefaultStorageConfigEnabled()
			if actual != mock.isenabled {
				t.Fatalf("Test '%s' failed: expected '%t' actual '%t' : storage-config '%t' ", name, actual, mock.isenabled, actualStgConfig)
			}
			os.Unsetenv(string(DefaultCstorSparsePool))
		})
	}
}

func TestCstorSparsePoolSpc070(t *testing.T) {
	tests := map[string]struct {
		value    string
		defConf  string
		expected int
		iserr    bool
	}{
		"with 1": {
			value:    "1",
			defConf:  "true",
			expected: 2,
			iserr:    false,
		},
		"with true": {
			value:    "true",
			defConf:  "true",
			expected: 2,
			iserr:    false,
		},
		"with 0": {
			value:    "0",
			defConf:  "true",
			expected: 0,
			iserr:    false,
		},
		"with false": {
			value:    "false",
			defConf:  "true",
			expected: 0,
			iserr:    false,
		},
		"with junk": {
			value:    "junk",
			defConf:  "true",
			expected: 0,
			iserr:    false,
		},
		"with special chars": {
			value:    "abc:123-123",
			defConf:  "true",
			expected: 0,
			iserr:    false,
		},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			os.Unsetenv(string(CreateDefaultStorageConfig))
			errDef := os.Setenv(string(CreateDefaultStorageConfig), mock.defConf)
			if errDef != nil {
				t.Fatalf("Test '%s' failed %+v", name, errDef)
			}
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
