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
	"strings"
	"testing"

	"github.com/openebs/maya/pkg/util"
	"github.com/openebs/maya/pkg/version"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func fakeUnstructAlways() UnstructuredPredicate {
	return func(given *unstructured.Unstructured) bool {
		return true
	}
}

func fakeUnstructNever() UnstructuredPredicate {
	return func(given *unstructured.Unstructured) bool {
		return false
	}
}

func fakeEmptyUnstruct() *unstructured.Unstructured {
	return &unstructured.Unstructured{}
}

func fakeRunTask() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind": "RunTask",
			"metadata": map[string]interface{}{
				"name": "dummyRT",
				"annotations": map[string]interface{}{
					"hi":       "runtask",
					"version":  version.Current(),
					"suffixed": "runtask-0.1.0",
				},
			},
		},
	}
}

func fakeCASTemplate() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind": "CASTemplate",
			"metadata": map[string]interface{}{
				"name": "dummyCAST",
				"annotations": map[string]interface{}{
					"hi":       "cast",
					"version":  version.Current(),
					"suffixed": "cast-0.1.0",
				},
			},
			"spec": map[string]interface{}{
				"run": map[string]interface{}{
					"tasks": []string{"task001", "task002-0.1.3", "task003", "task004"},
				},
			},
		},
	}
}

func TestUnstructuredPredicateListAll(t *testing.T) {
	tests := map[string]struct {
		predicates UnstructuredPredicateList
		expected   bool
	}{
		"101": {UnstructuredPredicateList{fakeUnstructAlways()}, true},
		"102": {UnstructuredPredicateList{fakeUnstructNever()}, false},
		"103": {UnstructuredPredicateList{fakeUnstructAlways(), fakeUnstructNever()}, false},
		"104": {UnstructuredPredicateList{fakeUnstructAlways(), fakeUnstructAlways()}, true},
		"105": {UnstructuredPredicateList{fakeUnstructNever(), fakeUnstructNever()}, false},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			actual := mock.predicates.All(nil)
			if mock.expected != actual {
				t.Fatalf("Test '%s' failed: expected '%t': actual '%t'", name, mock.expected, actual)
			}
		})
	}
}

func TestUnstructuredPredicateListAny(t *testing.T) {
	tests := map[string]struct {
		predicates UnstructuredPredicateList
		expected   bool
	}{
		"101": {UnstructuredPredicateList{fakeUnstructAlways()}, true},
		"102": {UnstructuredPredicateList{fakeUnstructNever()}, false},
		"103": {UnstructuredPredicateList{fakeUnstructAlways(), fakeUnstructNever()}, true},
		"104": {UnstructuredPredicateList{fakeUnstructAlways(), fakeUnstructAlways()}, true},
		"105": {UnstructuredPredicateList{fakeUnstructNever(), fakeUnstructNever()}, false},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			actual := mock.predicates.Any(nil)
			if mock.expected != actual {
				t.Fatalf("Test '%s' failed: expected '%t': actual '%t'", name, mock.expected, actual)
			}
		})
	}
}

func TestIsNamespaceScoped(t *testing.T) {
	tests := map[string]struct {
		given    *unstructured.Unstructured
		expected bool
	}{
		"101": {&unstructured.Unstructured{}, true},
		"102": {nil, false},
		"103": {fakeCASTemplate(), false},
		"104": {fakeRunTask(), true},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			actual := IsNamespaceScoped(mock.given)
			if mock.expected != actual {
				t.Fatalf("Test '%s' failed: expected '%t': actual '%t'", name, mock.expected, actual)
			}
		})
	}
}

func TestIsNameUnversioned(t *testing.T) {
	tests := map[string]struct {
		given    *unstructured.Unstructured
		expected bool
	}{
		"101": {&unstructured.Unstructured{}, true},
		"102": {nil, false},
		"103": {fakeCASTemplate(), true},
		"104": {fakeRunTask(), true},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			actual := IsNameUnVersioned(mock.given)
			if mock.expected != actual {
				t.Fatalf("Test '%s' failed: expected '%t': actual '%t'", name, mock.expected, actual)
			}
		})
	}
}

func TestIsNameVersioned(t *testing.T) {
	tests := map[string]struct {
		given    *unstructured.Unstructured
		expected bool
	}{
		"101": {&unstructured.Unstructured{}, false},
		"102": {nil, false},
		"103": {fakeCASTemplate(), false},
		"104": {fakeRunTask(), false},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			actual := IsNameVersioned(mock.given)
			if mock.expected != actual {
				t.Fatalf("Test '%s' failed: expected '%t': actual '%t'", name, mock.expected, actual)
			}
		})
	}
}

func TestIsRunTask(t *testing.T) {
	tests := map[string]struct {
		given    *unstructured.Unstructured
		expected bool
	}{
		"101": {&unstructured.Unstructured{}, false},
		"102": {nil, false},
		"103": {fakeCASTemplate(), false},
		"104": {fakeRunTask(), true},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			actual := IsRunTask(mock.given)
			if mock.expected != actual {
				t.Fatalf("Test '%s' failed: expected '%t': actual '%t'", name, mock.expected, actual)
			}
		})
	}
}

func TestIsCASTemplate(t *testing.T) {
	tests := map[string]struct {
		given    *unstructured.Unstructured
		expected bool
	}{
		"101": {&unstructured.Unstructured{}, false},
		"102": {nil, false},
		"103": {fakeCASTemplate(), true},
		"104": {fakeRunTask(), false},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			actual := IsCASTemplate(mock.given)
			if mock.expected != actual {
				t.Fatalf("Test '%s' failed: expected '%t': actual '%t'", name, mock.expected, actual)
			}
		})
	}
}

func TestUpdateNamespace(t *testing.T) {
	tests := map[string]struct {
		given     *unstructured.Unstructured
		namespace string
		expected  string
	}{
		"101": {&unstructured.Unstructured{}, "default", "default"},
		"102": {nil, "openebs", ""},
		"103": {fakeCASTemplate(), "none", "none"},
		"104": {fakeRunTask(), "gone", "gone"},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			u := UpdateNamespace(mock.namespace)(mock.given)
			if mock.given == nil && u != nil {
				t.Fatalf("Test '%s' failed: expected 'nil instance': actual '%v'", name, u)
			}
			if u == nil {
				return
			}
			if mock.expected != u.GetNamespace() {
				t.Fatalf("Test '%s' failed: expected '%s': actual '%s'", name, mock.expected, u.GetNamespace())
			}
		})
	}
}

func TestSuffixNameWithVersion(t *testing.T) {
	ver := version.Current()
	tests := map[string]struct {
		given    *unstructured.Unstructured
		expected string
	}{
		"101": {&unstructured.Unstructured{}, "-" + ver},
		"102": {nil, ""},
		"103": {fakeCASTemplate(), "dummyCAST-" + ver},
		"104": {fakeRunTask(), "dummyRT-" + ver},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			u := SuffixNameWithVersion()(mock.given)
			if mock.given == nil && u != nil {
				t.Fatalf("Test '%s' failed: expected 'nil instance': actual '%v'", name, u)
			}
			if u == nil {
				return
			}
			if mock.expected != u.GetName() {
				t.Fatalf("Test '%s' failed: expected '%s': actual '%s'", name, mock.expected, u.GetName())
			}
		})
	}
}

func TestUpdateLabels(t *testing.T) {
	var (
		lbls     = map[string]string{"app": "openebs", "kind": "storage"}
		moreLbls = map[string]string{"a": "1", "b": "2", "c": "3"}
		noLbls   = map[string]string{}
	)

	tests := map[string]struct {
		given         *unstructured.Unstructured
		labels        map[string]string
		iteration     int
		expectedCount int
	}{
		"101": {&unstructured.Unstructured{}, lbls, 1, 2},
		"102": {&unstructured.Unstructured{}, noLbls, 1, 0},
		"103": {nil, lbls, 1, 0},
		"104": {fakeCASTemplate(), lbls, 1, 2},
		"105": {fakeCASTemplate(), moreLbls, 2, 3},
		"106": {fakeRunTask(), lbls, 1, 2},
		"107": {fakeRunTask(), moreLbls, 3, 3},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			u := mock.given
			for i := 1; i <= mock.iteration; i++ {
				u = UpdateLabels(mock.labels, false)(u)
			}
			if mock.given == nil && u != nil {
				t.Fatalf("Test '%s' failed: expected 'nil instance': actual '%v'", name, u)
			}
			if u == nil {
				return
			}
			if mock.expectedCount != len(u.GetLabels()) {
				t.Fatalf("Test '%s' failed: expected label count '%d': actual '%d'", name, mock.expectedCount, len(u.GetLabels()))
			}
		})
	}
}

func TestAddNameToLabels(t *testing.T) {
	tests := map[string]struct {
		given    *unstructured.Unstructured
		key      string
		override bool
		expected string
	}{
		"101": {&unstructured.Unstructured{}, "key", true, ""},
		"102": {&unstructured.Unstructured{}, "key", false, ""},
		"103": {nil, "key", false, ""},
		"104": {nil, "key", true, ""},
		"105": {fakeCASTemplate(), "key", false, "dummyCAST"},
		"106": {fakeCASTemplate(), "key", true, "dummyCAST"},
		"107": {fakeRunTask(), "key", false, "dummyRT"},
		"108": {fakeRunTask(), "key", true, "dummyRT"},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			u := AddNameToLabels(mock.key, mock.override)(mock.given)
			if mock.given == nil && u != nil {
				t.Fatalf("Test '%s' failed: expected 'nil instance': actual '%v'", name, u)
			}
			if u == nil {
				return
			}
			lbls := u.GetLabels()
			var name string
			if len(lbls) > 0 {
				name = lbls[mock.key]
			}
			if mock.expected != name {
				t.Fatalf("Test '%s' failed: expected label name '%s': actual '%s'", name, mock.expected, name)
			}
		})
	}
}

func TestSuffixWithVersionAtPath(t *testing.T) {
	v := version.Current()
	tests := map[string]struct {
		given    *unstructured.Unstructured
		path     string
		expected string
	}{
		"101": {&unstructured.Unstructured{}, "metadata.name", ""},
		"102": {nil, "metadata.name", ""},

		"201": {fakeRunTask(), "metadata.name", "dummyRT-" + v},
		"202": {fakeRunTask(), "metadata.annotations.version", v},
		"203": {fakeRunTask(), "metadata.annotations.suffixed", "runtask-0.1.0"},
		"204": {fakeRunTask(), "metadata.annotations.hi", "runtask-" + v},
		"205": {fakeRunTask(), "kind", "RunTask-" + v},

		"301": {fakeCASTemplate(), "metadata.name", "dummyCAST-" + v},
		"302": {fakeCASTemplate(), "metadata.annotations.version", v},
		"303": {fakeCASTemplate(), "metadata.annotations.suffixed", "cast-0.1.0"},
		"304": {fakeCASTemplate(), "metadata.annotations.hi", "cast-" + v},
		"305": {fakeCASTemplate(), "kind", "CASTemplate-" + v},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			u := SuffixWithVersionAtPath(mock.path)(mock.given)
			if mock.given == nil && u != nil {
				t.Fatalf("Test '%s' failed: expected 'nil instance': actual '%v'", name, u)
			}
			if u == nil {
				return
			}
			actual := util.GetNestedString(u.Object, strings.Split(mock.path, ".")...)
			if mock.expected != actual {
				t.Fatalf("Test '%s' failed: expected '%s': actual '%s': debug '%+v'", name, mock.expected, actual, u.Object)
			}
		})
	}
}

func TestSuffixStringSlicesWithVersionAtPath(t *testing.T) {
	v := version.Current()
	tests := map[string]struct {
		given    *unstructured.Unstructured
		path     string
		expected []string
	}{
		"101": {fakeCASTemplate(), "spec.run.tasks", []string{"task001-" + v, "task002-0.1.3", "task003-" + v, "task004-" + v}},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			u := SuffixWithVersionAtPath(mock.path)(mock.given)
			actual := util.GetNestedSlice(u.Object, strings.Split(mock.path, ".")...)
			if len(mock.expected) != len(actual) {
				t.Fatalf("Test '%s' failed: expected count '%d': actual '%d': debug '%#v'", name, len(mock.expected), len(actual), actual)
			}
			for i, e := range mock.expected {
				if e != actual[i] {
					t.Fatalf("Test '%s' failed: expected '%s': actual '%s': debug '%#v'", name, e, actual[i], actual)
				}
			}
		})
	}
}

func TestUnstructuredMap(t *testing.T) {
	var (
		lbls     = map[string]string{"app": "openebs", "kind": "storage"}
		moreLbls = map[string]string{"a": "1", "b": "2", "c": "3"}
		noLbls   = map[string]string{}
		ver      = version.Current()
	)
	tests := map[string]struct {
		g                 *unstructured.Unstructured
		m                 UnstructuredMiddleware
		p                 UnstructuredPredicateList
		expectedName      string
		expectedNamespace string
		expectedLblCount  int
	}{
		// update labels with always predicate
		"101": {fakeEmptyUnstruct(), UpdateLabels(nil, false), UnstructuredPredicateList{fakeUnstructAlways()}, "", "", 0},
		"102": {fakeEmptyUnstruct(), UpdateLabels(noLbls, false), UnstructuredPredicateList{fakeUnstructAlways()}, "", "", 0},
		"103": {fakeEmptyUnstruct(), UpdateLabels(lbls, false), UnstructuredPredicateList{fakeUnstructAlways()}, "", "", 2},
		"104": {fakeEmptyUnstruct(), UpdateLabels(moreLbls, false), UnstructuredPredicateList{fakeUnstructAlways()}, "", "", 3},
		"105": {nil, UpdateLabels(moreLbls, false), UnstructuredPredicateList{fakeUnstructAlways()}, "", "", 0},
		"106": {fakeCASTemplate(), UpdateLabels(moreLbls, false), UnstructuredPredicateList{fakeUnstructAlways()}, "dummyCAST", "", 3},
		"107": {fakeRunTask(), UpdateLabels(moreLbls, false), UnstructuredPredicateList{fakeUnstructAlways()}, "dummyRT", "", 3},
		// update label with never predicate
		"201": {fakeEmptyUnstruct(), UpdateLabels(nil, false), UnstructuredPredicateList{fakeUnstructNever()}, "", "", 0},
		"202": {fakeEmptyUnstruct(), UpdateLabels(noLbls, false), UnstructuredPredicateList{fakeUnstructNever()}, "", "", 0},
		"203": {fakeEmptyUnstruct(), UpdateLabels(lbls, false), UnstructuredPredicateList{fakeUnstructNever()}, "", "", 0},
		"204": {fakeEmptyUnstruct(), UpdateLabels(moreLbls, false), UnstructuredPredicateList{fakeUnstructNever()}, "", "", 0},
		"205": {nil, UpdateLabels(moreLbls, false), UnstructuredPredicateList{fakeUnstructNever()}, "", "", 0},
		"206": {fakeCASTemplate(), UpdateLabels(moreLbls, false), UnstructuredPredicateList{fakeUnstructNever()}, "dummyCAST", "", 0},
		"207": {fakeRunTask(), UpdateLabels(moreLbls, false), UnstructuredPredicateList{fakeUnstructNever()}, "dummyRT", "", 0},
		// update name with suffix with always predicate
		"301": {fakeEmptyUnstruct(), SuffixNameWithVersion(), UnstructuredPredicateList{fakeUnstructAlways()}, "-" + ver, "", 0},
		"302": {nil, SuffixNameWithVersion(), UnstructuredPredicateList{fakeUnstructAlways()}, "", "", 0},
		"303": {fakeCASTemplate(), SuffixNameWithVersion(), UnstructuredPredicateList{fakeUnstructAlways()}, "dummyCAST-" + ver, "", 0},
		"304": {fakeRunTask(), SuffixNameWithVersion(), UnstructuredPredicateList{fakeUnstructAlways()}, "dummyRT-" + ver, "", 0},
		// update name with suffix with never predicate
		"401": {fakeEmptyUnstruct(), SuffixNameWithVersion(), UnstructuredPredicateList{fakeUnstructNever()}, "", "", 0},
		"402": {nil, SuffixNameWithVersion(), UnstructuredPredicateList{fakeUnstructNever()}, "", "", 0},
		"403": {fakeCASTemplate(), SuffixNameWithVersion(), UnstructuredPredicateList{fakeUnstructNever()}, "dummyCAST", "", 0},
		"404": {fakeRunTask(), SuffixNameWithVersion(), UnstructuredPredicateList{fakeUnstructNever()}, "dummyRT", "", 0},
		// update namespace with always predicate
		"501": {fakeEmptyUnstruct(), UpdateNamespace(""), UnstructuredPredicateList{fakeUnstructAlways()}, "", "", 0},
		"502": {fakeEmptyUnstruct(), UpdateNamespace("def"), UnstructuredPredicateList{fakeUnstructAlways()}, "", "def", 0},
		"503": {nil, UpdateNamespace(""), UnstructuredPredicateList{fakeUnstructAlways()}, "", "", 0},
		"504": {nil, UpdateNamespace("def"), UnstructuredPredicateList{fakeUnstructAlways()}, "", "", 0},
		"505": {fakeCASTemplate(), UpdateNamespace("def"), UnstructuredPredicateList{fakeUnstructAlways()}, "dummyCAST", "def", 0},
		"506": {fakeRunTask(), UpdateNamespace("def"), UnstructuredPredicateList{fakeUnstructAlways()}, "dummyRT", "def", 0},
		// update namespace with never predicate
		"601": {fakeEmptyUnstruct(), UpdateNamespace(""), UnstructuredPredicateList{fakeUnstructNever()}, "", "", 0},
		"602": {fakeEmptyUnstruct(), UpdateNamespace("def"), UnstructuredPredicateList{fakeUnstructNever()}, "", "", 0},
		"603": {nil, UpdateNamespace(""), UnstructuredPredicateList{fakeUnstructNever()}, "", "", 0},
		"604": {nil, UpdateNamespace("def"), UnstructuredPredicateList{fakeUnstructNever()}, "", "", 0},
		"605": {fakeCASTemplate(), UpdateNamespace("def"), UnstructuredPredicateList{fakeUnstructNever()}, "dummyCAST", "", 0},
		"606": {fakeRunTask(), UpdateNamespace("def"), UnstructuredPredicateList{fakeUnstructNever()}, "dummyRT", "", 0},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			u := UnstructuredMap(mock.m, mock.p...)(mock.g)
			if mock.g == nil && u != nil {
				t.Fatalf("Test '%s' failed: expected 'nil instance': actual '%v'", name, u)
			}
			if u == nil {
				return
			}
			if mock.expectedLblCount != len(u.GetLabels()) {
				t.Fatalf("Test '%s' failed: expected label count '%d': actual '%d'", name, mock.expectedLblCount, len(u.GetLabels()))
			}
			if mock.expectedName != u.GetName() {
				t.Fatalf("Test '%s' failed: expected name '%s': actual '%s'", name, mock.expectedName, u.GetName())
			}
			if mock.expectedNamespace != u.GetNamespace() {
				t.Fatalf("Test '%s' failed: expected namespace '%s': actual '%s'", name, mock.expectedNamespace, u.GetNamespace())
			}
		})
	}
}

func TestUnstructuredListMapAll(t *testing.T) {
	v := version.Current()
	tests := map[string]struct {
		g               []*unstructured.Unstructured
		m               []UnstructuredMiddleware
		p               UnstructuredPredicateList
		lblKey          string
		expectedLbls    bool
		expectedName    string
		expectedLblName string
	}{
		"101": {
			[]*unstructured.Unstructured{fakeCASTemplate()},
			[]UnstructuredMiddleware{SuffixNameWithVersion(), AddNameToLabels("key", false)},
			UnstructuredPredicateList{fakeUnstructAlways()},
			"key",
			true,
			"dummyCAST-" + v,
			"dummyCAST-" + v,
		},
		"102": {
			[]*unstructured.Unstructured{fakeRunTask()},
			[]UnstructuredMiddleware{SuffixNameWithVersion(), AddNameToLabels("key", false)},
			UnstructuredPredicateList{fakeUnstructAlways()},
			"key",
			true,
			"dummyRT-" + v,
			"dummyRT-" + v,
		},
		"201": {
			[]*unstructured.Unstructured{fakeCASTemplate()},
			[]UnstructuredMiddleware{SuffixNameWithVersion(), AddNameToLabels("key", false)},
			UnstructuredPredicateList{fakeUnstructNever()},
			"key",
			false,
			"dummyCAST",
			"dummyCAST",
		},
		"202": {
			[]*unstructured.Unstructured{fakeRunTask()},
			[]UnstructuredMiddleware{SuffixNameWithVersion(), AddNameToLabels("key", false)},
			UnstructuredPredicateList{fakeUnstructNever()},
			"key",
			false,
			"dummyRT",
			"dummyRT",
		},
		"301": {
			[]*unstructured.Unstructured{fakeCASTemplate()},
			[]UnstructuredMiddleware{SuffixNameWithVersion(), AddNameToLabels("key", false)},
			UnstructuredPredicateList{IsNameVersioned},
			"key",
			false,
			"dummyCAST",
			"dummyCAST",
		},
		"302": {
			[]*unstructured.Unstructured{fakeRunTask()},
			[]UnstructuredMiddleware{SuffixNameWithVersion(), AddNameToLabels("key", false)},
			UnstructuredPredicateList{IsNameVersioned},
			"key",
			false,
			"dummyRT",
			"dummyRT",
		},
		"401": {
			[]*unstructured.Unstructured{fakeCASTemplate()},
			[]UnstructuredMiddleware{SuffixNameWithVersion(), AddNameToLabels("key", false)},
			UnstructuredPredicateList{IsNamespaceScoped},
			"key",
			false,
			"dummyCAST",
			"dummyCAST",
		},
		"402": {
			[]*unstructured.Unstructured{fakeRunTask()},
			[]UnstructuredMiddleware{SuffixNameWithVersion(), AddNameToLabels("key", false)},
			UnstructuredPredicateList{IsNamespaceScoped},
			"key",
			true,
			"dummyRT-" + v,
			"dummyRT-" + v,
		},
		"501": {
			[]*unstructured.Unstructured{fakeCASTemplate()},
			[]UnstructuredMiddleware{SuffixNameWithVersion(), AddNameToLabels("key", false)},
			UnstructuredPredicateList{IsNamespaceScoped, fakeUnstructAlways()},
			"key",
			false,
			"dummyCAST",
			"dummyCAST",
		},
		"502": {
			[]*unstructured.Unstructured{fakeRunTask()},
			[]UnstructuredMiddleware{SuffixNameWithVersion(), AddNameToLabels("key", false)},
			UnstructuredPredicateList{IsNamespaceScoped, fakeUnstructAlways()},
			"key",
			true,
			"dummyRT-" + v,
			"dummyRT-" + v,
		},
		"601": {
			[]*unstructured.Unstructured{fakeCASTemplate()},
			[]UnstructuredMiddleware{SuffixNameWithVersion(), AddNameToLabels("key", false)},
			UnstructuredPredicateList{IsNamespaceScoped, fakeUnstructNever()},
			"key",
			false,
			"dummyCAST",
			"dummyCAST",
		},
		"602": {
			[]*unstructured.Unstructured{fakeRunTask()},
			[]UnstructuredMiddleware{SuffixNameWithVersion(), AddNameToLabels("key", false)},
			UnstructuredPredicateList{IsNamespaceScoped, fakeUnstructNever()},
			"key",
			false,
			"dummyRT",
			"dummyRT",
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			l := UnstructedList{mock.g}
			u := l.MapAll(mock.m, mock.p...)
			actual := u.Items[0]
			if mock.expectedName != actual.GetName() {
				t.Fatalf("Test '%s' failed: expected name '%s': actual '%s'", name, mock.expectedName, actual.GetName())
			}
			if mock.expectedLbls && len(actual.GetLabels()) == 0 {
				t.Fatalf("Test '%s' failed: expected labels: actual none", name)
			}
			if !mock.expectedLbls && len(actual.GetLabels()) != 0 {
				t.Fatalf("Test '%s' failed: expected no labels: actual '%#v'", name, actual.GetLabels())
			}
			if mock.expectedLbls && mock.expectedLblName != actual.GetLabels()[mock.lblKey] {
				t.Fatalf("Test '%s' failed: expected label name '%s': actual '%s'", name, mock.expectedLblName, actual.GetLabels()[mock.lblKey])
			}
		})
	}
}

func TestUnstructuredListMapAllIfAny(t *testing.T) {
	v := version.Current()
	tests := map[string]struct {
		g               []*unstructured.Unstructured
		m               []UnstructuredMiddleware
		p               UnstructuredPredicateList
		lblKey          string
		expectedLbls    bool
		expectedName    string
		expectedLblName string
	}{
		"101": {
			[]*unstructured.Unstructured{fakeCASTemplate()},
			[]UnstructuredMiddleware{SuffixNameWithVersion(), AddNameToLabels("key", false)},
			UnstructuredPredicateList{IsNamespaceScoped, fakeUnstructAlways()},
			"key",
			true,
			"dummyCAST-" + v,
			"dummyCAST-" + v,
		},
		"102": {
			[]*unstructured.Unstructured{fakeRunTask()},
			[]UnstructuredMiddleware{SuffixNameWithVersion(), AddNameToLabels("key", false)},
			UnstructuredPredicateList{IsNamespaceScoped, fakeUnstructAlways()},
			"key",
			true,
			"dummyRT-" + v,
			"dummyRT-" + v,
		},
		"201": {
			[]*unstructured.Unstructured{fakeCASTemplate()},
			[]UnstructuredMiddleware{SuffixNameWithVersion(), AddNameToLabels("key", false)},
			UnstructuredPredicateList{IsNamespaceScoped, fakeUnstructNever()},
			"key",
			false,
			"dummyCAST",
			"dummyCAST",
		},
		"202": {
			[]*unstructured.Unstructured{fakeRunTask()},
			[]UnstructuredMiddleware{SuffixNameWithVersion(), AddNameToLabels("key", false)},
			UnstructuredPredicateList{IsNamespaceScoped, fakeUnstructNever()},
			"key",
			true,
			"dummyRT-" + v,
			"dummyRT-" + v,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			l := UnstructedList{mock.g}
			u := l.MapAllIfAny(mock.m, mock.p...)
			actual := u.Items[0]
			if mock.expectedName != actual.GetName() {
				t.Fatalf("Test '%s' failed: expected name '%s': actual '%s'", name, mock.expectedName, actual.GetName())
			}
			if mock.expectedLbls && len(actual.GetLabels()) == 0 {
				t.Fatalf("Test '%s' failed: expected labels: actual none", name)
			}
			if !mock.expectedLbls && len(actual.GetLabels()) != 0 {
				t.Fatalf("Test '%s' failed: expected no labels: actual '%#v'", name, actual.GetLabels())
			}
			if mock.expectedLbls && mock.expectedLblName != actual.GetLabels()[mock.lblKey] {
				t.Fatalf("Test '%s' failed: expected label name '%s': actual '%s'", name, mock.expectedLblName, actual.GetLabels()[mock.lblKey])
			}
		})
	}
}

func TestUnstructuredListMapAllNamesIfAny(t *testing.T) {
	v := version.Current()
	tests := map[string]struct {
		g            []*unstructured.Unstructured
		m            []UnstructuredMiddleware
		p            UnstructuredPredicateList
		lblKey       string
		expectedLbls bool
		nameContains string
	}{
		"101": {
			[]*unstructured.Unstructured{fakeCASTemplate(), fakeRunTask()},
			[]UnstructuredMiddleware{SuffixNameWithVersion(), AddNameToLabels("key", false)},
			UnstructuredPredicateList{fakeUnstructAlways()},
			"key",
			true,
			"-" + v,
		},
		"102": {
			[]*unstructured.Unstructured{fakeRunTask(), fakeCASTemplate()},
			[]UnstructuredMiddleware{SuffixNameWithVersion(), AddNameToLabels("key", false)},
			UnstructuredPredicateList{IsNamespaceScoped, fakeUnstructAlways()},
			"key",
			true,
			"-" + v,
		},
		"103": {
			[]*unstructured.Unstructured{fakeRunTask(), fakeCASTemplate()},
			[]UnstructuredMiddleware{SuffixNameWithVersion(), AddNameToLabels("key", false)},
			UnstructuredPredicateList{fakeUnstructAlways(), fakeUnstructNever()},
			"key",
			true,
			"-" + v,
		},
		"104": {
			[]*unstructured.Unstructured{fakeRunTask(), fakeCASTemplate()},
			[]UnstructuredMiddleware{SuffixNameWithVersion(), AddNameToLabels("key", false)},
			UnstructuredPredicateList{fakeUnstructNever()},
			"key",
			false,
			"dummy",
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			l := UnstructedList{mock.g}
			u := l.MapAllIfAny(mock.m, mock.p...)
			for i, actual := range u.Items {
				if !strings.Contains(actual.GetName(), mock.nameContains) {
					t.Fatalf("Test '%s-%d' failed: expected name contains '%s': actual '%s'", name, i, mock.nameContains, actual.GetName())
				}
				if mock.expectedLbls && len(actual.GetLabels()) == 0 {
					t.Fatalf("Test '%s-%d' failed: expected labels: actual none", name, i)
				}
				if !mock.expectedLbls && len(actual.GetLabels()) != 0 {
					t.Fatalf("Test '%s-%d' failed: expected no labels: actual '%#v'", name, i, actual.GetLabels())
				}
				if mock.expectedLbls && !strings.Contains(actual.GetLabels()[mock.lblKey], mock.nameContains) {
					t.Fatalf("Test '%s-%d' failed: expected label name contains '%s': actual '%s'", name, i, mock.nameContains, actual.GetLabels()[mock.lblKey])
				}
			}
		})
	}
}
