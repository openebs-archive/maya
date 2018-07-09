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

package util

import (
	"reflect"
	"testing"
)

func TestMergeMapOfObjects(t *testing.T) {
	tests := map[string]struct {
		destMap      map[string]interface{}
		srcMap       map[string]interface{}
		isMerge      bool
		expectedKeys []string
	}{
		"merge map of objects - +ve test case - dest & src maps are exclusive": {
			destMap: map[string]interface{}{
				"k1": "v1",
			},
			srcMap: map[string]interface{}{
				"k2": "v2",
			},
			isMerge:      true,
			expectedKeys: []string{"k1", "k2"},
		},
		"merge map of objects - +ve test case - dest & src maps are same": {
			destMap: map[string]interface{}{
				"k1": "v1",
			},
			srcMap: map[string]interface{}{
				"k1": "v2",
			},
			isMerge:      true,
			expectedKeys: []string{"k1"},
		},
		"merge map of objects - +ve test case - some elements of dest & src maps are common": {
			destMap: map[string]interface{}{
				"k1": "v1",
			},
			srcMap: map[string]interface{}{
				"k1": "v1.1",
				"k2": "v2",
			},
			isMerge:      true,
			expectedKeys: []string{"k1", "k2"},
		},
		"merge map of objects - +ve test case - dest map is empty": {
			destMap: map[string]interface{}{},
			srcMap: map[string]interface{}{
				"k2": "v2",
			},
			isMerge:      true,
			expectedKeys: []string{"k2"},
		},
		"merge map of objects - +ve test case - dest map is nil": {
			destMap: nil,
			srcMap: map[string]interface{}{
				"k2": "v2",
			},
			isMerge: false,
		},
		"merge map of objects - +ve test case - src map is empty": {
			destMap: map[string]interface{}{
				"k1": "v1",
			},
			srcMap:       nil,
			isMerge:      true,
			expectedKeys: []string{"k1"},
		},
		"merge map of objects - +ve test case - src map is nil": {
			destMap: map[string]interface{}{
				"k1": "v1",
			},
			srcMap:       nil,
			isMerge:      true,
			expectedKeys: []string{"k1"},
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			ok := MergeMapOfObjects(mock.destMap, mock.srcMap)

			if mock.isMerge != ok {
				t.Fatalf("failed to test merge map of objects: expected merge '%t': actual merge '%t'", mock.isMerge, ok)
			}

			if mock.isMerge && !ContainKeys(mock.destMap, mock.expectedKeys) {
				t.Fatalf("failed to test merge map of objects: expected keys '%s': actual 'missing key(s)': dest after merge '%#v'", mock.expectedKeys, mock.destMap)
			}

			if mock.isMerge && len(mock.destMap) != len(mock.expectedKeys) {
				t.Fatalf("failed to test merge map of objects: expected key count '%d': actual key count '%d'", len(mock.expectedKeys), len(mock.destMap))
			}
		})
	}
}

func TestSetNestedField(t *testing.T) {
	tests := map[string]struct {
		obj    map[string]interface{}
		value  interface{}
		fields []string
	}{
		"set nested field - +ve test case - set nested key value pair to empty original": {
			// this is empty
			obj: map[string]interface{}{},
			// value to be set in above obj
			value: "hello there",
			// above value will be set at path obj[k1][k1.1]
			fields: []string{"k1", "k1.1"},
		},
		"set nested field - +ve test case - set nested key with complex value to empty original": {
			// this is empty
			obj: map[string]interface{}{},
			// value to be set in above obj
			value: map[string]interface{}{
				"k1.1.1": map[string]interface{}{
					"k1.1.1.1": "hi there",
				},
			},
			// above value will be set at path obj[k1][k1.1]
			fields: []string{"k1", "k1.1"},
		},
		"set nested field - +ve test case - set a new nested key value pair": {
			obj: map[string]interface{}{
				"k1": "v1",
				"k2": map[string]string{
					"k2.1": "v2.1",
					"k2.2": "v2.2",
				},
			},
			// value to be set in above obj
			value: "v2.3",
			// above value will be set at path obj[k2][k2.3]
			fields: []string{"k2", "k2.3"},
		},
		"set nested field - +ve test case - set a new value to an existing nested key": {
			obj: map[string]interface{}{
				"k1": "v1",
				"k2": map[string]string{
					"k2.1": "v2.1",
					"k2.2": "v2.2",
				},
			},
			// value to be set in above obj
			value: "v2.2'",
			// above value will be set at path obj[k2][k2.2]
			fields: []string{"k2", "k2.2"},
		},
		"set nested field - +ve test case - set a new value to an existing parent key": {
			obj: map[string]interface{}{
				"k1": "v1",
				"k2": map[string]string{
					"k2.1": "v2.1",
					"k2.2": "v2.2",
				},
			},
			// value to be set in above obj
			value: "hello",
			// above value will be set against the obj[k2]
			fields: []string{"k2"},
		},
		"set nested field - +ve test case - add a new parent key value pair": {
			obj: map[string]interface{}{
				"k1": "v1",
				"k2": map[string]string{
					"k2.1": "v2.1",
					"k2.2": "v2.2",
				},
			},
			// value to be set in above obj
			value: "hello",
			// above value will be set against the obj[k3]
			fields: []string{"k3"},
		},
		"set nested field - +ve test case - no changes to obj": {
			obj: map[string]interface{}{
				"k1": "v1",
				"k2": map[string]string{
					"k2.1": "v2.1",
					"k2.2": "v2.2",
				},
			},
			// value to be set in above obj
			value: map[string]interface{}{
				"k1": "v1",
				"k2": map[string]string{
					"k2.1": "v2.1",
					"k2.2": "v2.2",
				},
			},
			// there will be no changes to obj since there are no fields
			fields: []string{},
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			SetNestedField(mock.obj, mock.value, mock.fields...)

			if !reflect.DeepEqual(GetNestedField(mock.obj, mock.fields...), mock.value) {
				t.Fatalf("failed to test set nested field: expected '%#v': actual '%#v'", mock.value, GetNestedField(mock.obj, mock.fields...))
			}
		})
	}
}
