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
	"strings"
	"testing"
)

func TestValidateWithCheckf(t *testing.T) {
	tests := map[string]struct {
		path        string
		check       Predicate
		msg         string
		expectError bool
	}{
		"verify empty path": {
			"",
			IsNonRoot(),
			"missing host path",
			true,
		},
		"verify non root with msg": {
			"/pv",
			IsNonRoot(),
			"root directory should not be used",
			true,
		},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			b := NewBuilder().
				WithPath(mock.path).
				WithCheckf(mock.check, mock.msg)
			err := b.Validate()
			if (err != nil) != mock.expectError {
				t.Fatalf("test %s failed, expected error: %t but got: %v",
					name, mock.expectError, err)
			}
			if (err != nil) && !strings.Contains(err.Error(), mock.msg) {
				t.Fatalf("test %s failed, expected error to include: %v but got: %v",
					name, mock.msg, err)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := map[string]struct {
		path        string
		checks      []Predicate
		expectError bool
	}{
		"predicate returns true": {
			"/var/openebs/pv",
			[]Predicate{IsNonRoot()},
			false,
		},
		"predicate returns false": {
			"/pv",
			[]Predicate{IsNonRoot()},
			true,
		},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			b := NewBuilder().
				WithPath(mock.path).
				WithChecks(mock.checks...)
			err := b.Validate()
			if (err != nil) != mock.expectError {
				t.Fatalf("test %s failed, expected error: %t but got: %v",
					name, mock.expectError, err)
			}
		})
	}
}

func TestValidateAndBuild(t *testing.T) {
	tests := map[string]struct {
		basePath    string
		relPath     string
		expectError bool
		expectPath  string
	}{
		"verify empty path": {
			"",
			"",
			true,
			"",
		},
		"verify empty rel path": {
			"/pv",
			"",
			true,
			"",
		},
		"verify empty base path": {
			"",
			"/pv",
			true,
			"",
		},
		"verify incomplete path": {
			"abc",
			"/pv",
			true,
			"",
		},
		"verify valid path": {
			"/abc",
			"/def",
			false,
			"/abc/def",
		},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			b := NewBuilder().
				WithPathJoin(mock.basePath, mock.relPath).
				WithCheckf(IsNonRoot(), "root directory is not allowed")
			path, err := b.ValidateAndBuild()
			if (err != nil) != mock.expectError {
				t.Fatalf("test %s failed, expected error: %t but got: %v",
					name, mock.expectError, err)
			}
			if (err == nil) && strings.Compare(path, mock.expectPath) != 0 {
				t.Fatalf("test %s failed, expected error to include: %v but got: %v",
					name, mock.expectPath, path)
			}
		})
	}
}
