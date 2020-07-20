/*
Copyright 2018 The OpenEBS Authors.

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

package version

import (
	"testing"
)

func TestIsNotVersioned(t *testing.T) {
	tests := map[string]struct {
		given    string
		expected bool
	}{
		// "version as prefix":            {"0.7.0-maya", true},
		// "version in between":  {"openebs-0.7.0-maya", true},
		"version as suffix 1": {"maya-0.7.0", false},
		"version as suffix 2": {"maya-0.7.9", false},
		"version as suffix 3": {"maya-1.0.0", false},
		"version as suffix 4": {"maya-0.0.1", false},
		"version as suffix 5": {"openebs-maya-2.2.1", false},
		"version as suffix 6": {"maya-1.11.1", false},
		// "version as suffix 7": {"abc-121-1232-maya-10.0.13", false},
		// "version as suffix 8":          {"abc-345-1232-11.20.13", false},
		"no version 1":                 {"maya", true},
		"no version 2":                 {"maya-", true},
		"in-valid version as suffix 1": {"maya-0.8.a", true},
		"in-valid version as suffix 2": {"maya-1", true},
		"in-valid version as suffix 3": {"maya-0.8", true},
		"version with invalid suffix":  {"maya0.8.0", true},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			if mock.expected != IsNotVersioned(mock.given) {
				t.Fatalf("Test '%s' failed: expected '%t': actual '%t'", name, mock.expected, IsNotVersioned(mock.given))
			}
		})
	}
}

func TestWithSuffixIf(t *testing.T) {
	v := Current()
	tests := map[string]struct {
		given    string
		expected string
	}{
		"no version":             {"maya", "maya-" + v},
		"with version 1":         {"maya-0.11.0", "maya-0.11.0"},
		"with version 2":         {"maya-10.5.0", "maya-10.5.0"},
		"with version 3":         {"maya-0.0.12", "maya-0.0.12"},
		"with invalid version 1": {"maya-0.5", "maya-0.5-" + v},
		"with invalid version 2": {"maya-10.0", "maya-10.0-" + v},
		"with invalid version 3": {"maya-0.10", "maya-0.10-" + v},
		// "with version in between 1": {"maya-0.5.0-", "maya-0.5.0--" + v},
		// "with version in between 2": {"maya-0.5.0-openebs", "maya-0.5.0-openebs-" + v},
		"with empty": {"", "-" + v},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			u := WithSuffixIf(mock.given, IsNotVersioned)
			if mock.expected != u {
				t.Fatalf("Test '%s' failed: expected '%s': actual '%s'", name, mock.expected, u)
			}
		})
	}
}

func TestWithSuffixesIf(t *testing.T) {
	v := Current()
	tests := map[string]struct {
		given    []string
		expected []string
	}{
		"101": {[]string{"maya", "openebs-maya"}, []string{"maya-" + v, "openebs-maya-" + v}},
		"102": {[]string{"maya-0.4.0", "openebs-maya"}, []string{"maya-0.4.0", "openebs-maya-" + v}},
		"103": {[]string{"maya", "openebs-maya-0.11.1"}, []string{"maya-" + v, "openebs-maya-0.11.1"}},
		"104": {[]string{"maya-0.5.0", "openebs-maya-0.11.1"}, []string{"maya-0.5.0", "openebs-maya-0.11.1"}},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			u := WithSuffixesIf(mock.given, IsNotVersioned)
			if len(u) != len(mock.given) {
				t.Fatalf("Test '%s' failed: expected count '%d': actual '%d'", name, len(mock.given), len(u))
			}
			for i, actual := range u {
				if actual != mock.expected[i] {
					t.Fatalf("Test '%s' failed: expected '%s': actual '%s'", name, mock.expected[i], actual)
				}
			}
		})
	}
}
