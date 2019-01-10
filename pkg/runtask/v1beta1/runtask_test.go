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

package v1beta1

import (
	apis "github.com/openebs/maya/pkg/apis/openebs.io/runtask/v1beta1"
	"github.com/pkg/errors"
	"reflect"
	"testing"
)

func TestRunTaskErrors(t *testing.T) {
	tests := map[string]struct {
		task      *runtask
		hasErrors bool
		errCount  int
	}{
		"t1": {&runtask{}, false, 0},
		"t2": {&runtask{errs: []error{}}, false, 0},
		"t3": {&runtask{errs: []error{errors.New("101")}}, true, 1},
	}
	for name, mock := range tests {
		if mock.hasErrors != mock.task.HasError() {
			t.Fatalf("test '%s' failed: expected errors '%t' actual '%t'", name, mock.hasErrors, mock.task.HasError())
		}
		if mock.errCount != len(mock.task.Errors()) {
			t.Fatalf("test '%s' failed: expected errors '%d' actual '%d'", name, mock.errCount, len(mock.task.Errors()))
		}
	}
}

func TestNew(t *testing.T) {
	r := New()
	if r == nil {
		t.Fatalf("test failed: expected non nil instance actual nil")
	}
}

func TestUpdate(t *testing.T) {
	r := Update(New())
	if r == nil {
		t.Fatalf("test failed: expected non nil instance actual nil")
	}
}

func TestBuilder(t *testing.T) {
	r, _ := Builder().Build()
	if r == nil {
		t.Fatalf("test failed: expected non nil instance actual nil")
	}
}

func TestNewWithConfig(t *testing.T) {
	tests := map[string]struct {
		config map[string]string
	}{
		"t1": {nil},
		"t2": {map[string]string{"hi": "there"}},
	}
	for name, mock := range tests {
		r := New(WithConfig(mock.config))
		if r == nil {
			t.Fatalf("test '%s' failed: expected non nil instance actual nil", name)
		}
		if !reflect.DeepEqual(r.Spec.Config, mock.config) {
			t.Fatalf("test '%s' failed: expected config '%+v' actual '%+v'", name, mock.config, r.Spec.Config)
		}
	}
}

func TestUpdateWithConfig(t *testing.T) {
	tests := map[string]struct {
		config map[string]string
	}{
		"t1": {nil},
		"t2": {map[string]string{"hi": "there"}},
	}
	for name, mock := range tests {
		r := Update(New(), WithConfig(mock.config))
		if r == nil {
			t.Fatalf("test '%s' failed: expected non nil instance actual nil", name)
		}
		if !reflect.DeepEqual(r.Spec.Config, mock.config) {
			t.Fatalf("test '%s' failed: expected config '%+v' actual '%+v'", name, mock.config, r.Spec.Config)
		}
	}
}

func TestBuilderWithConfig(t *testing.T) {
	tests := map[string]struct {
		config map[string]string
	}{
		"t1": {nil},
		"t2": {map[string]string{"hi": "there"}},
	}
	for name, mock := range tests {
		r, _ := Builder().WithConfig(mock.config).Build()
		if r == nil {
			t.Fatalf("test '%s' failed: expected non nil instance actual nil", name)
		}
		if !reflect.DeepEqual(r.Spec.Config, mock.config) {
			t.Fatalf("test '%s' failed: expected config '%+v' actual '%+v'", name, mock.config, r.Spec.Config)
		}
	}
}

func TestNewAddRunItem(t *testing.T) {
	tests := map[string]struct {
		item apis.RunItem
	}{
		"t1": {apis.RunItem{}},
		"t2": {apis.RunItem{ID: "001", Name: "Test"}},
	}
	for name, mock := range tests {
		r := New(AddRunItem(mock.item))
		if r == nil {
			t.Fatalf("test '%s' failed: expected non nil instance actual nil", name)
		}
		if !reflect.DeepEqual(r.Spec.Runs[0], mock.item) {
			t.Fatalf("test '%s' failed: expected runitem '%+v' actual '%+v'", name, mock.item, r.Spec.Runs[0])
		}
	}
}

func TestUpdateAddRunItem(t *testing.T) {
	tests := map[string]struct {
		item apis.RunItem
	}{
		"t1": {apis.RunItem{}},
		"t2": {apis.RunItem{ID: "001", Name: "Test"}},
	}
	for name, mock := range tests {
		r := Update(New(), AddRunItem(mock.item))
		if r == nil {
			t.Fatalf("test '%s' failed: expected non nil instance actual nil", name)
		}
		if !reflect.DeepEqual(r.Spec.Runs[0], mock.item) {
			t.Fatalf("test '%s' failed: expected runitem '%+v' actual '%+v'", name, mock.item, r.Spec.Runs[0])
		}
	}
}

func TestBuilderAddRunItems(t *testing.T) {
	tests := map[string]struct {
		items []apis.RunItem
		count int
	}{
		"t1": {[]apis.RunItem{apis.RunItem{}}, 1},
		"t2": {[]apis.RunItem{apis.RunItem{ID: "001", Name: "Test"}}, 1},
		"t3": {[]apis.RunItem{apis.RunItem{ID: "002", Name: "Test2"}, apis.RunItem{}}, 2},
	}
	for name, mock := range tests {
		r, _ := Builder().AddRunItems(mock.items).Build()
		if r == nil {
			t.Fatalf("test '%s' failed: expected non nil instance actual nil", name)
		}
		if len(r.Spec.Runs) != len(mock.items) {
			t.Fatalf("test '%s' failed: expected runitem count '%d' actual '%d'", name, len(mock.items), len(r.Spec.Runs))
		}
		if !reflect.DeepEqual(r.Spec.Runs[0], mock.items[0]) {
			t.Fatalf("test '%s' failed: expected runitem '%+v' actual '%+v'", name, mock.items[0], r.Spec.Runs[0])
		}
	}
}

func TestNewWithSpec(t *testing.T) {
	tests := map[string]struct {
		spec *apis.RunTask
	}{
		"t1": {
			&apis.RunTask{
				Spec: apis.RunTaskSpec{
					Runs: []apis.RunItem{
						apis.RunItem{
							ID:   "101",
							Name: "test",
						},
					},
				},
			},
		},
	}
	for name, mock := range tests {
		r := New(WithSpec(mock.spec))
		if r == nil {
			t.Fatalf("test '%s' failed: expected non nil instance actual nil", name)
		}
		if !reflect.DeepEqual(r, mock.spec) {
			t.Fatalf("test '%s' failed: expected spec '%+v' actual '%+v'", name, mock.spec, r)
		}
	}
}

func TestUpdateWithSpec(t *testing.T) {
	tests := map[string]struct {
		spec *apis.RunTask
	}{
		"t1": {
			&apis.RunTask{
				Spec: apis.RunTaskSpec{
					Runs: []apis.RunItem{
						apis.RunItem{
							ID:   "101",
							Name: "test",
						},
					},
				},
			},
		},
	}
	for name, mock := range tests {
		r := Update(New(), WithSpec(mock.spec))
		if r == nil {
			t.Fatalf("test '%s' failed: expected non nil instance actual nil", name)
		}
		if !reflect.DeepEqual(r, mock.spec) {
			t.Fatalf("test '%s' failed: expected spec '%+v' actual '%+v'", name, mock.spec, r)
		}
	}
}

func TestBuilderWithSpec(t *testing.T) {
	tests := map[string]struct {
		spec *apis.RunTask
	}{
		"t1": {
			&apis.RunTask{
				Spec: apis.RunTaskSpec{
					Runs: []apis.RunItem{
						apis.RunItem{
							ID:   "101",
							Name: "test",
						},
					},
				},
			},
		},
	}
	for name, mock := range tests {
		r, _ := Builder().WithSpec(mock.spec).Build()
		if r == nil {
			t.Fatalf("test '%s' failed: expected non nil instance actual nil", name)
		}
		if !reflect.DeepEqual(r, mock.spec) {
			t.Fatalf("test '%s' failed: expected spec '%+v' actual '%+v'", name, mock.spec, r)
		}
	}
}

func TestNewWithUnmarshal(t *testing.T) {
	tests := map[string]struct {
		yaml             string
		expectedName     string
		expectedNS       string
		expectedRunCount int
	}{
		"t1": {`
apiVersion: openebs.io/v1beta1
kind: RunTask
metadata:
  name: myrun
  namespace: openebs
  labels:
    app: mstor
spec:
  runs:
  - id: "101"
    name: myrun1
  - id: "102"
    name: myrun2
`,
			"myrun",
			"openebs",
			2},
	}
	for name, mock := range tests {
		r := New(WithUnmarshal(mock.yaml))
		if r == nil {
			t.Fatalf("test '%s' failed: expected non nil instance actual nil", name)
		}
		if r.Name != mock.expectedName {
			t.Fatalf("test '%s' failed: expected name '%s' actual '%s'", name, mock.expectedName, r.Name)
		}
		if r.Namespace != mock.expectedNS {
			t.Fatalf("test '%s' failed: expected namespace '%s' actual '%s'", name, mock.expectedNS, r.Namespace)
		}
		if len(r.Spec.Runs) != mock.expectedRunCount {
			t.Fatalf("test '%s' failed: expected namespace '%d' actual '%d'", name, mock.expectedRunCount, len(r.Spec.Runs))
		}
	}
}

func TestUpdateWithUnmarshal(t *testing.T) {
	tests := map[string]struct {
		yaml             string
		expectedName     string
		expectedNS       string
		expectedRunCount int
	}{
		"t1": {`
apiVersion: openebs.io/v1beta1
kind: RunTask
metadata:
  name: myrun
  namespace: openebs
  labels:
    app: mstor
spec:
  runs:
  - id: "101"
    name: myrun1
  - id: "102"
    name: myrun2
`,
			"myrun",
			"openebs",
			2},
	}
	for name, mock := range tests {
		r := Update(New(), WithUnmarshal(mock.yaml))
		if r == nil {
			t.Fatalf("test '%s' failed: expected non nil instance actual nil", name)
		}
		if r.Name != mock.expectedName {
			t.Fatalf("test '%s' failed: expected name '%s' actual '%s'", name, mock.expectedName, r.Name)
		}
		if r.Namespace != mock.expectedNS {
			t.Fatalf("test '%s' failed: expected namespace '%s' actual '%s'", name, mock.expectedNS, r.Namespace)
		}
		if len(r.Spec.Runs) != mock.expectedRunCount {
			t.Fatalf("test '%s' failed: expected namespace '%d' actual '%d'", name, mock.expectedRunCount, len(r.Spec.Runs))
		}
	}
}

func TestBuilderWithUnmarshal(t *testing.T) {
	tests := map[string]struct {
		yaml             string
		expectedName     string
		expectedNS       string
		expectedRunCount int
	}{
		"t1": {`
apiVersion: openebs.io/v1beta1
kind: RunTask
metadata:
  name: myrun
  namespace: openebs
  labels:
    app: mstor
spec:
  runs:
  - id: "101"
    name: myrun1
  - id: "102"
    name: myrun2
`,
			"myrun",
			"openebs",
			2},
	}
	for name, mock := range tests {
		r, _ := Builder().WithUnmarshal(mock.yaml).Build()
		if r == nil {
			t.Fatalf("test '%s' failed: expected non nil instance actual nil", name)
		}
		if r.Name != mock.expectedName {
			t.Fatalf("test '%s' failed: expected name '%s' actual '%s'", name, mock.expectedName, r.Name)
		}
		if r.Namespace != mock.expectedNS {
			t.Fatalf("test '%s' failed: expected namespace '%s' actual '%s'", name, mock.expectedNS, r.Namespace)
		}
		if len(r.Spec.Runs) != mock.expectedRunCount {
			t.Fatalf("test '%s' failed: expected namespace '%d' actual '%d'", name, mock.expectedRunCount, len(r.Spec.Runs))
		}
	}
}

func TestNewWithStatus(t *testing.T) {
	tests := map[string]struct {
		status apis.RunTaskStatus
	}{
		"t1": {apis.RunTaskStatus{Phase: "init"}},
	}
	for name, mock := range tests {
		r := New(WithStatus(mock.status))
		if r == nil {
			t.Fatalf("test '%s' failed: expected non nil instance actual nil", name)
		}
		if !reflect.DeepEqual(r.Status, mock.status) {
			t.Fatalf("test '%s' failed: expected status '%+v' actual '%+v'", name, mock.status, r.Status)
		}
	}
}

func TestUpdateWithStatus(t *testing.T) {
	tests := map[string]struct {
		status apis.RunTaskStatus
	}{
		"t1": {apis.RunTaskStatus{Phase: "init"}},
	}
	for name, mock := range tests {
		r := Update(New(), WithStatus(mock.status))
		if r == nil {
			t.Fatalf("test '%s' failed: expected non nil instance actual nil", name)
		}
		if !reflect.DeepEqual(r.Status, mock.status) {
			t.Fatalf("test '%s' failed: expected status '%+v' actual '%+v'", name, mock.status, r.Status)
		}
	}
}

func TestBuilderWithStatus(t *testing.T) {
	tests := map[string]struct {
		status apis.RunTaskStatus
	}{
		"t1": {apis.RunTaskStatus{Phase: "init"}},
	}
	for name, mock := range tests {
		r, _ := Builder().WithStatus(mock.status).Build()
		if r == nil {
			t.Fatalf("test '%s' failed: expected non nil instance actual nil", name)
		}
		if !reflect.DeepEqual(r.Status, mock.status) {
			t.Fatalf("test '%s' failed: expected status '%+v' actual '%+v'", name, mock.status, r.Status)
		}
	}
}
