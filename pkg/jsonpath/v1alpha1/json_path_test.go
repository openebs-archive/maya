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

func TestSelectionValue(t *testing.T) {
	tests := map[string]struct {
		values   []string
		expected string
	}{
		"101": {[]string{"hi", "there", "where"}, "hi"},
		"102": {[]string{"", "there", "where"}, ""},
		"103": {[]string{"hi", "", "where"}, "hi"},
		"104": {[]string{"hi", "where", ""}, "hi"},
		"105": {[]string{"hi", "there"}, "hi"},
		"106": {[]string{"there"}, "there"},
		"107": {[]string{}, ""},
		"108": {nil, ""},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			s := &selection{Values: mock.values}
			if s.Value() != mock.expected {
				t.Fatalf("Test '%s' failed: expected '%s': actual '%s'", name, mock.expected, s.Value())
			}
		})
	}
}

func TestSelectionList(t *testing.T) {
	tests := map[string]struct {
		aliasPaths    map[string]string
		expectedCount int
	}{
		"101": {nil, 0},
		"102": {map[string]string{}, 0},
		"103": {map[string]string{"name": ".name"}, 1},
		"104": {map[string]string{"name": ".spec.name", "namespace": ".spec.namespace"}, 2},
		"105": {map[string]string{"name": ".spec.name", "capacity": ".config.capacity", "namespace": ".spec.namespace"}, 3},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			sl := SelectionList(mock.aliasPaths)
			if len(sl) != mock.expectedCount {
				t.Fatalf("Test '%s' failed: expected count '%d': actual count '%d'", name, mock.expectedCount, len(sl))
			}
		})
	}
}

func TestSelectionListValues(t *testing.T) {
	tests := map[string]struct {
		alias          string
		values         []string
		isValuesArray  bool
		isValuesString bool
	}{
		"101": {"1", []string{"a"}, false, true},
		"102": {"1", []string{"a", "b"}, true, false},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			sl := selectionList{}
			sl = append(sl, &selection{Alias: mock.alias, Values: mock.values})
			m := sl.Values()
			val := m[mock.alias]
			if mock.isValuesArray {
				if _, ok := val.([]string); !ok {
					t.Fatalf("Test '%s' failed: expected []string values: actual '%#v'", name, val)
				}
			}
			if mock.isValuesString {
				if _, ok := val.(string); !ok {
					t.Fatalf("Test '%s' failed: expected string value: actual '%#v'", name, val)
				}
			}
		})
	}
}

func TestJSONPathValues(t *testing.T) {
	tests := map[string]struct {
		target   interface{}
		path     string
		isErr    bool
		isValues bool
	}{
		// noop jsonpath
		"101": {nil, "", false, true},
		"102": {nil, ".xyz", false, true},
		"103": {map[string]string{}, "", false, true},
		"104": {map[string]string{}, ".xyz", false, true},
		"105": {"", "", false, true},
		"106": {"", ".xyz", false, true},
		"107": {[]string{}, "", false, true},
		"108": {[]string{}, ".xyz", false, true},
		// valid jsonpath
		"201": {nil, "{.xyz}", false, true},
		"202": {map[string]string{}, "{.xyz}", false, true},
		"203": {"", "{.xyz}", false, true},
		"204": {[]string{}, "{.xyz}", false, true},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			v, err := JSONPath(name).WithTarget(mock.target).Values(mock.path)

			if !mock.isErr && err != nil {
				t.Fatalf("Test '%s' failed: expected no error: actual '%s'", name, err.Error())
			}

			if !mock.isValues && v != nil {
				t.Fatalf("Test '%s' failed: expected no values: actual '%+v'", name, v)
			}
		})
	}
}

func TestSinglePathJsonQuery(t *testing.T) {
	tests := map[string]struct {
		target interface{}
		path   string
		isVal  bool
		isWarn bool
	}{
		// no data; invalid jsonpath
		"101": {nil, "", false, true},
		"102": {nil, ".xyz", false, true},
		"103": {map[string]string{}, "", false, true},
		"104": {map[string]string{}, ".xyz", false, true},
		"105": {"", "", false, true},
		"106": {"", ".xyz", false, true},
		"107": {[]string{}, "", false, true},
		"108": {[]string{}, ".xyz", false, true},
		// no data; non-existent jsonpath
		"201": {nil, "{.xyz}", false, true},
		"202": {nil, "{.xyz}", false, true},
		"203": {map[string]string{}, "{.xyz}", false, true},
		"204": {map[string]string{}, "{.xyz}", false, true},
		"205": {"", "{.xyz}", false, true},
		"206": {"", "{.xyz}", false, true},
		"207": {[]string{}, "{.xyz}", false, true},
		"208": {[]string{}, "{.xyz}", false, true},
		// with data; invalid jsonpath
		"301": {[]string{"hi"}, "", false, true},
		"302": {[]string{"hi"}, "", false, true},
		"303": {[]string{"hi"}, ".xyz", false, true},
		"304": {[]string{"hi"}, ".xyz", false, true},
		"305": {map[string]string{"hi": "hello"}, "", false, true},
		"306": {map[string]string{"hi": "hello"}, "", false, true},
		"307": {map[string]string{"hi": "hello"}, ".xyz", false, true},
		"308": {map[string]string{"hi": "hello"}, ".xyz", false, true},
		"309": {[]byte(`["hi"]`), "", false, true},
		"310": {[]byte(`["hi"]`), "", false, true},
		"311": {[]byte(`["hi"]`), ".xyz", false, true},
		"312": {[]byte(`["hi"]`), ".xyz", false, true},
		// with data; non existent jsonpath
		"401": {[]string{"hi"}, "{.xyz}", false, true},
		"402": {[]string{"hi"}, "{.xyz}", false, true},
		"403": {map[string]string{"hi": "hello"}, "{.xyz}", false, true},
		"404": {map[string]string{"hi": "hello"}, "{.xyz}", false, true},
		"405": {[]byte(`["hi"]`), "{.xyz}", false, true},
		"406": {[]byte(`["hi"]`), "{.xyz}", false, true},
		// with data; with jsonpath
		"501": {"hi", "{..}", true, false},
		"502": {"hello world", "{..}", true, false},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			j := JSONPath(name).WithTarget(mock.target)
			s := j.Query(Selection(name, mock.path))

			if mock.isVal && len(s.Values) == 0 {
				t.Fatalf("Test '%s' failed: expected queried value(s): actual %s", name, s)
			}

			if !mock.isWarn && j.HasWarn() {
				t.Fatalf("Test '%s' failed: expected no warns: actual %s", name, j.Msgs)
			}
		})
	}
}

func TestQueryAllWithJsonBytes(t *testing.T) {
	target := []byte(`[
		{"id": "i1", "x":4, "y":-5},
		{"id": "i2", "x":-2, "y":-5, "z":1},
		{"id": "i3", "x":  8, "y":  3 },
		{"id": "i4", "x": -6, "y": -1 },
		{"id": "i5", "x":  0, "y":  2, "z": 1 },
		{"id": "i6", "x":  1, "y":  4 }
	]`)

	tests := map[string]struct {
		aliaspaths map[string]string
		isVal      bool
		isWarn     bool
	}{
		"101": {map[string]string{"s1": "{[?(@.id)].x}"}, true, false},
		"102": {map[string]string{"s1": "{[?(@.id)].x}", "s2": "{[0]['id']}"}, true, false},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			j := JSONPath(name).WithTargetAsRaw(target)
			sl := SelectionList(mock.aliaspaths)
			ul := j.QueryAll(sl)

			if mock.isVal {
				for a, p := range mock.aliaspaths {
					if len(ul.ValueByAlias(a)) == 0 {
						t.Fatalf("Test '%s' failed: expected value for select %s %s: actual %s", name, a, p, ul)
					}
					if len(ul.ValuesByAlias(a)) == 0 {
						t.Fatalf("Test '%s' failed: expected value(s) for select %s %s: actual %s", name, a, p, ul)
					}
					if len(ul.ValueByPath(p)) == 0 {
						t.Fatalf("Test '%s' failed: expected value for select %s %s: actual %s", name, a, p, ul)
					}
					if len(ul.ValuesByPath(p)) == 0 {
						t.Fatalf("Test '%s' failed: expected value(s) for select %s %s: actual %s", name, a, p, ul)
					}
				}
			}
			if !mock.isWarn && j.HasWarn() {
				t.Fatalf("Test '%s' failed: expected no warns: actual %s", name, j.Msgs)
			}
		})
	}
}

func TestQueryAllJivaPayload(t *testing.T) {
	target := []byte(`{
    "data": [
      {
        "actions": {
          "start": "http://172.18.0.2:9501/v1/volumes/c3RvcmUx?action=start",
          "deletevolume": "http://172.18.0.2:9501/v1/volumes/c3RvcmUx?action=deletevolume"
        },
        "id": "c3RvcmUx",
        "links": {
          "self": "http://172.18.0.2:9501/v1/volumes/c3RvcmUx"
        },
        "name": "def-vol-claim-mysql-76555",
        "readOnly": "false",
        "replicaCount": 2,
        "type": "volume"
      }
    ],
    "links": {
      "self": "http://172.18.0.2:9501/v1/volumes"
    },
    "resourceType": "volume",
    "type": "collection"
  }`)

	tests := map[string]struct {
		aliaspaths map[string]string
		isVal      bool
		isWarn     bool
	}{
		"101": {map[string]string{"dlink": "{.data[?(@.name=='def-vol-claim-mysql-76555')].actions.deletevolume}"}, true, false},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			j := JSONPath(name).WithTargetAsRaw(target)
			sl := SelectionList(mock.aliaspaths)
			ul := j.QueryAll(sl)

			if mock.isVal {
				for a, p := range mock.aliaspaths {
					if len(ul.ValueByAlias(a)) == 0 {
						t.Fatalf("Test '%s' failed: expected value for select %s %s: actual %s", name, a, p, ul)
					}
					if len(ul.ValuesByAlias(a)) == 0 {
						t.Fatalf("Test '%s' failed: expected value(s) for select %s %s: actual %s", name, a, p, ul)
					}
					if len(ul.ValueByPath(p)) == 0 {
						t.Fatalf("Test '%s' failed: expected value for select %s %s: actual %s", name, a, p, ul)
					}
					if len(ul.ValuesByPath(p)) == 0 {
						t.Fatalf("Test '%s' failed: expected value(s) for select %s %s: actual %s", name, a, p, ul)
					}
				}
			}
			if !mock.isWarn && j.HasWarn() {
				t.Fatalf("Test '%s' failed: expected no warns: actual %s", name, j.Msgs)
			}
		})
	}
}
