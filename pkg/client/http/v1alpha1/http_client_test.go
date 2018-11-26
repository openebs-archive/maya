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
		verb    HttpVerb
		baseurl string
		name    string
		iserr   bool
	}{
		"101": {GetAction, "", "", true},
		"102": {GetAction, "http", "abc", true},
		"103": {GetAction, "http://", "abc", true},
		"104": {GetAction, "http:/0.0.0.0", "abc", true},
		"105": {GetAction, "http:/0.0.0.0:8080", "abc", true},
		"106": {DeleteAction, "http:/0.0.0.0:2123", "abc", true},
		"107": {PostAction, "http:/0.0.0.0:2123", "abc", true},
		"108": {PutAction, "http:/0.0.0.0:2123", "abc", true},
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
		verb  HttpVerb
		url   string
		iserr bool
	}{
		"101": {GetAction, "", true},
		"102": {GetAction, "http", true},
		"103": {GetAction, "http://", true},
		"104": {GetAction, "http:/0.0.0.0", true},
		"105": {GetAction, "http:/0.0.0.0:8080", true},
		"106": {DeleteAction, "http:/0.0.0.0:2123", true},
		"107": {PostAction, "http:/0.0.0.0:2123", true},
		"108": {PutAction, "http:/0.0.0.0:2123", true},
		// with version
		"201": {GetAction, "http:/0.0.0.0:8080/v2", true},
		"202": {DeleteAction, "http:/0.0.0.0:2123/v2", true},
		"203": {PostAction, "http:/0.0.0.0:2123/v2", true},
		"204": {PutAction, "http:/0.0.0.0:2123/v2", true},
		// with version; with resource name
		"301": {GetAction, "http:/0.0.0.0:8080/v2/vol", true},
		"302": {DeleteAction, "http:/0.0.0.0:2123/v2/vol", true},
		"303": {PostAction, "http:/0.0.0.0:2123/v2/vol", true},
		"304": {PutAction, "http:/0.0.0.0:2123/v2/vol", true},
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
