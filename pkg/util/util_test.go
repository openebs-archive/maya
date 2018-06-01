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

package util

import (
	"testing"
)

func TestContainsString(t *testing.T) {
	tests := map[string]struct {
		array     []string
		element   string
		isPresent bool
	}{
		"contains string - positive test case - element is present": {
			array:     []string{"hi", "hello"},
			element:   "hello",
			isPresent: true,
		},
		"contains string - positive test case - element is not present": {
			array:     []string{"there", "you", "go"},
			element:   "there",
			isPresent: true,
		},
		"contains string - boundary test case - similar elements but not same": {
			array:     []string{"hi there", "ok now"},
			element:   "hi there ",
			isPresent: false,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			isPresent := ContainsString(mock.array, mock.element)

			if mock.isPresent != isPresent {
				t.Fatalf("failed to test contains string: expected element '%t': actual element '%t'", mock.isPresent, isPresent)
			}
		})
	}
}

func TestContainsKey(t *testing.T) {
	tests := map[string]struct {
		mapOfObjs map[string]interface{}
		searchKey string
		hasKey    bool
	}{
		"contains key - +ve test case - map having the key": {
			mapOfObjs: map[string]interface{}{
				"k1": "v1",
			},
			searchKey: "k1",
			hasKey:    true,
		},
		"contains key - +ve test case - map without the key": {
			mapOfObjs: map[string]interface{}{
				"k1": "v1",
			},
			searchKey: "k2",
			hasKey:    false,
		},
		"contains key - +ve test case - empty map": {
			mapOfObjs: map[string]interface{}{},
			searchKey: "k1",
			hasKey:    false,
		},
		"contains key - +ve test case - nil map": {
			mapOfObjs: nil,
			searchKey: "k1",
			hasKey:    false,
		},
		"contains key - +ve test case - with empty search key": {
			mapOfObjs: map[string]interface{}{
				"k1": "v1",
			},
			searchKey: "",
			hasKey:    false,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			hasKey := ContainsKey(mock.mapOfObjs, mock.searchKey)

			if hasKey != mock.hasKey {
				t.Fatalf("failed to test contains key: expected key '%s': actual 'not found'", mock.searchKey)
			}
		})
	}
}

func TestContainKeys(t *testing.T) {
	tests := map[string]struct {
		mapOfObjs  map[string]interface{}
		searchKeys []string
		hasKeys    bool
	}{
		"contains key - +ve test case - map having the key": {
			mapOfObjs: map[string]interface{}{
				"k1": "v1",
			},
			searchKeys: []string{"k1"},
			hasKeys:    true,
		},
		"contains key - +ve test case - map without the keys": {
			mapOfObjs: map[string]interface{}{
				"k1": "v1",
			},
			searchKeys: []string{"k2"},
			hasKeys:    false,
		},
		"contains key - +ve test case - empty map": {
			mapOfObjs:  map[string]interface{}{},
			searchKeys: []string{"k1"},
			hasKeys:    false,
		},
		"contains key - +ve test case - nil map": {
			mapOfObjs:  nil,
			searchKeys: []string{"k1"},
			hasKeys:    false,
		},
		"contains key - +ve test case - with no search keys": {
			mapOfObjs: map[string]interface{}{
				"k1": "v1",
			},
			searchKeys: []string{},
			hasKeys:    false,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			hasKeys := ContainKeys(mock.mapOfObjs, mock.searchKeys)

			if hasKeys != mock.hasKeys {
				t.Fatalf("failed to test contains key: expected key '%s': actual 'not found'", mock.searchKeys)
			}
		})
	}
}
