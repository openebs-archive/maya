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
	"reflect"
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

func TestListDiff(t *testing.T) {
	tests := map[string]struct {
		listA        []string
		listB        []string
		expectedLen  int
		expectedList []string
	}{
		"list diff operation - positive test case - element is present": {
			listA:        []string{"hi", "hello", "crazzy"},
			listB:        []string{"hello"},
			expectedLen:  2,
			expectedList: []string{"hi", "crazzy"},
		},
		"list diff operation - positive test case - element is not present": {
			listA:        []string{},
			listB:        []string{"there", "you", "go"},
			expectedLen:  0,
			expectedList: []string{},
		},
		"contains string - boundary test case - similar elements but not same": {
			listA:        []string{"hi there", "ok now"},
			listB:        []string{},
			expectedLen:  2,
			expectedList: []string{"hi there", "ok now"},
		},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			resultArr := ListDiff(mock.listA, mock.listB)
			if mock.expectedLen != len(resultArr) {
				t.Fatalf("failed to test %q: expected element count '%d': actual element count '%d'", name, mock.expectedLen, len(resultArr))
			}
			if !reflect.DeepEqual(resultArr, mock.expectedList) {
				t.Fatalf("failed to test %q: expected elements '%v': actual elements  '%v'", name, mock.expectedList, resultArr)
			}
		})
	}
}

func TestListIntersection(t *testing.T) {
	tests := map[string]struct {
		listA        []string
		listB        []string
		expectedLen  int
		expectedList []string
	}{
		"positive test case - element is present": {
			listA:        []string{"hi", "hello", "crazzy"},
			listB:        []string{"hello"},
			expectedLen:  1,
			expectedList: []string{"hello"},
		},
		"positive test case - ListA is empty": {
			listA:        []string{},
			listB:        []string{"there", "you", "go"},
			expectedLen:  0,
			expectedList: []string{},
		},
		"ListB is empty": {
			listA:        []string{"hi there", "ok now"},
			listB:        []string{},
			expectedLen:  0,
			expectedList: []string{},
		},
		"List is missmatch - boundary test case - similar elements but not same": {
			listA:        []string{"hi there", "ok now"},
			listB:        []string{"h ithere"},
			expectedLen:  0,
			expectedList: []string{},
		},
		"List is match in different order": {
			listA:        []string{"hi there", "ok now"},
			listB:        []string{"ok now", "hi there"},
			expectedLen:  2,
			expectedList: []string{"hi there", "ok now"},
		},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			resultArr := ListIntersection(mock.listA, mock.listB)
			if mock.expectedLen != len(resultArr) {
				t.Fatalf("failed to test %q: expected element count '%d': actual element count '%d'", name, mock.expectedLen, len(resultArr))
			}
			if !reflect.DeepEqual(resultArr, mock.expectedList) {
				t.Fatalf("failed to test %q: expected elements '%v': actual elements  '%v'", name, mock.expectedList, resultArr)
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

func TestMergeMap(t *testing.T) {
	tests := map[string]struct {
		Map1        map[string]interface{}
		Map2        map[string]interface{}
		expectedMap map[string]interface{}
	}{
		"merge map of 2 objects": {
			Map1: map[string]interface{}{
				"k1": "v1",
			},
			Map2: map[string]interface{}{
				"k2": "v2",
			},
			expectedMap: map[string]interface{}{
				"k1": "v1",
				"k2": "v2",
			},
		},
		"merge map of 2 similar key objects": {
			Map1: map[string]interface{}{
				"k1": "v1",
			},
			Map2: map[string]interface{}{
				"k1": "v2",
			},
			expectedMap: map[string]interface{}{
				"k1": "v2",
			},
		},
		"merge map of different k/v objects": {
			Map1: map[string]interface{}{
				"k1": "v1",
				"k2": "v1",
				"k4": "v4",
			},
			Map2: map[string]interface{}{
				"k1": "v2",
				"k3": "v3",
			},
			expectedMap: map[string]interface{}{
				"k1": "v2",
				"k2": "v1",
				"k4": "v4",
				"k3": "v3",
			},
		},
		"merge map of 1 nil objects": {
			Map1: map[string]interface{}{
				"k1": "v1",
				"k2": "v1",
				"k4": "v4",
			},
			Map2: nil,
			expectedMap: map[string]interface{}{
				"k1": "v1",
				"k2": "v1",
				"k4": "v4",
			},
		},
		"merge map of nil values objects": {
			Map1: map[string]interface{}{
				"k1": "v1",
				"k2": "v1",
				"k4": "v4",
			},
			Map2: map[string]interface{}{
				"k1": "",
				"k3": "",
			},

			expectedMap: map[string]interface{}{
				"k1": "",
				"k2": "v1",
				"k4": "v4",
				"k3": "",
			},
		},
	}

	for name, maps := range tests {
		t.Run(name, func(t *testing.T) {
			MergedMap := MergeMaps(maps.Map1, maps.Map2)
			if !reflect.DeepEqual(MergedMap, maps.expectedMap) {
				t.Fatalf(" failed to test MergeMaps: expected Map '%v': actual Map '%v'", maps.expectedMap, MergedMap)
			}
		})

	}
}

func TestRemoveString(t *testing.T) {
	slice1 := []string{"val1", "val2", "val3"}
	slice2 := []string{"val1", "val2", "val3", "val1"}
	slice3 := []string{"val1", "val2", "val1", "val3"}
	slice4 := []string{"val2", "val1", "val1", "val3"}
	tests := map[string]struct {
		actual      []string
		removeValue string
		expected    []string
	}{
		"value is at start":                 {slice1, "val1", []string{"val2", "val3"}},
		"value is at end":                   {slice1, "val3", []string{"val1", "val2"}},
		"value is in between":               {slice1, "val2", []string{"val1", "val3"}},
		"value is twice at start & end":     {slice2, "val1", []string{"val2", "val3"}},
		"value is twice at start & between": {slice3, "val1", []string{"val2", "val3"}},
		"value is twice in between":         {slice4, "val1", []string{"val2", "val3"}},
		"nil array and non empty value":     {nil, "val1", nil},
		"empty string to be removed":        {slice1, "", slice1},
		"nil array and empty string":        {nil, "", nil},
	}
	for name, test := range tests {
		// pinning the values
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			newSlice := RemoveString(test.actual, test.removeValue)
			if !reflect.DeepEqual(newSlice, test.expected) {
				t.Fatalf(" failed to test RemoveString: expected slice '%v': actual slice '%v'", test.expected, newSlice)
			}
		})
	}
}

func TestIsCurrentLessThanNewVersion(t *testing.T) {
	type args struct {
		old string
		new string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "old is less than new",
			args: args{
				old: "1.12.0",
				new: "2.8.0",
			},
			want: true,
		},
		{
			name: "old is greater than new",
			args: args{
				old: "2.10.0-RC2",
				new: "2.8.0",
			},
			want: false,
		},
		{
			name: "old is same as new",
			args: args{
				old: "2.8.0",
				new: "2.8.0",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsCurrentLessThanNewVersion(tt.args.old, tt.args.new); got != tt.want {
				t.Errorf("IsCurrentLessThanNewVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
