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
)

func TestIsCurrentVersionValid(t *testing.T) {
	tests := map[string]struct {
		isvalid bool
	}{
		"current version": {
			isvalid: true,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			c := CurrentVersion()
			v := Version(string(c))

			if v == invalidVersion && mock.isvalid {
				t.Fatalf("Test '%s' failed: version '%s' is '%s'", name, c, invalidVersion)
			}
		})
	}
}

func TestVersion(t *testing.T) {
	tests := map[string]struct {
		inputVersion    string
		expectedVersion version
	}{
		"case 1": {
			inputVersion:    "0.7.0",
			expectedVersion: "0.7.0",
		},
		"case 2": {
			inputVersion:    "0.8.0",
			expectedVersion: "0.8.0",
		},
		"case 3": {
			inputVersion:    "",
			expectedVersion: "invalid.version",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			validatedVersion := Version(tt.inputVersion)
			if validatedVersion != tt.expectedVersion {
				t.Errorf("Version error, got version: %v, expected version: %v", validatedVersion, tt.expectedVersion)
			}
		})
	}
}
