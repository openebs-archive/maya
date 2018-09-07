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

func TestAPI(t *testing.T) {
	tests := map[string]struct {
		verb    string
		baseurl string
		name    string
		iserr   bool
	}{
		"101": {"GET", "", "", true},
		"102": {"GET", "http", "abc", true},
		"103": {"GET", "http://", "abc", true},
		"104": {"GET", "http://127.0.0.1", "abc", true},
		"105": {"GET", "http://127.0.0.1:8080", "abc", true},
		"106": {"DELETE", "http://127.0.0.1:2123", "abc", true},
		"107": {"POST", "http://127.0.0.1:2123", "abc", true},
		"108": {"PUT", "http://127.0.0.1:2123", "abc", true},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := API(mock.verb, mock.baseurl, mock.name)

			if !mock.iserr && err != nil {
				t.Fatalf("Test '%s' failed: %s", name, err)
			}
		})
	}
}

func TestURL(t *testing.T) {
	tests := map[string]struct {
		verb  string
		url   string
		iserr bool
	}{
		"101": {"GET", "", true},
		"102": {"GET", "http", true},
		"103": {"GET", "http://", true},
		"104": {"GET", "http://127.0.0.1", true},
		"105": {"GET", "http://127.0.0.1:8080", true},
		"106": {"DELETE", "http://127.0.0.1:2123", true},
		"107": {"POST", "http://127.0.0.1:2123", true},
		"108": {"PUT", "http://127.0.0.1:2123", true},
		// with version
		"201": {"GET", "http://127.0.0.1:8080/v2", true},
		"202": {"DELETE", "http://127.0.0.1:2123/v2", true},
		"203": {"POST", "http://127.0.0.1:2123/v2", true},
		"204": {"PUT", "http://127.0.0.1:2123/v2", true},
		// with version; with resource name
		"301": {"GET", "http://127.0.0.1:8080/v2/vol", true},
		"302": {"DELETE", "http://127.0.0.1:2123/v2/vol", true},
		"303": {"POST", "http://127.0.0.1:2123/v2/vol", true},
		"304": {"PUT", "http://127.0.0.1:2123/v2/vol", true},
		// with server != localhost
		"401": {"GET", "http://10.0.0.1:8080/v2/vol", true},
		"402": {"DELETE", "http://10.0.0.1:2123/v2/vol", true},
		"403": {"POST", "http://10.0.0.1:2123/v2/vol", true},
		"404": {"PUT", "http://10.0.0.1:2123/v2/vol", true},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := URL(mock.verb, mock.url)

			if !mock.iserr && err != nil {
				t.Fatalf("Test '%s' failed: %s", name, err)
			}
		})
	}
}
