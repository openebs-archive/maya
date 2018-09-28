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

	"github.com/ghodss/yaml"
	. "github.com/openebs/maya/pkg/msg/v1alpha1"
	"github.com/openebs/maya/pkg/util"
)

// mockNoopStore is an implementation of ResultStoreFn
func mockNoopStore(id string, key string, value interface{}) {}

// mockMapStore stores the provided key value pair against the provided id
// inside the provided storage map
func mockMapStore(storage map[string]interface{}) ResultStoreFn {
	return func(id string, key string, value interface{}) {
		util.SetNestedField(storage, value, id, key)
	}
}

// mockAlwaysRun is an implementation of WillRunFn
func mockAlwaysRun() bool { return true }

// mockNeverRun is an implementation of WillRunFn
func mockNeverRun() bool { return false }

func mockRunCommandFromCategory(l []RunCommandCategory) (r *RunCommand) {
	r = Command()
	r.Category = append(r.Category, l...)
	return
}

// check if RunCommand implements CommandRunner interface
var _ CommandRunner = &RunCommand{}

func TestRunCommandWithData(t *testing.T) {
	tests := map[string]struct {
		data          map[string]interface{}
		expectedcount int
	}{
		"101": {map[string]interface{}{"first": "data"}, 1},
		"102": {map[string]interface{}{"first": "data1", "second": "data2"}, 2},
		"103": {map[string]interface{}{"first": "data1", "second": "data2", "third": "data3"}, 3},
		"104": {map[string]interface{}{"first": "data1", "second": "", "third": ""}, 3},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			c := Command()
			for k, v := range mock.data {
				c = WithData(c, k, v)
			}
			if len(c.Data) != mock.expectedcount {
				t.Fatalf("Test '%s' failed: expected count '%d': actual count '%d'", name, mock.expectedcount, len(c.Data))
			}
		})
	}
}

func TestNotSupportedCategoryCommand(t *testing.T) {
	tests := map[string]struct {
		categories          RunCommandCategoryList
		isSupportedCategory bool
	}{
		"test 101": {RunCommandCategoryList{JivaCommandCategory, CstorCommandCategory}, true},
		"test 102": {RunCommandCategoryList{VolumeCommandCategory, CstorCommandCategory}, false},
		"test 103": {RunCommandCategoryList{VolumeCommandCategory, PoolCommandCategory}, false},
		"test 104": {RunCommandCategoryList{JivaCommandCategory, PoolCommandCategory}, false},
		"test 105": {RunCommandCategoryList{JivaCommandCategory, VolumeCommandCategory}, true},
		"test 106": {RunCommandCategoryList{VolumeCommandCategory, JivaCommandCategory}, true},
		"test 107": {RunCommandCategoryList{VolumeCommandCategory, CstorCommandCategory}, false},
		"test 108": {RunCommandCategoryList{SnapshotCommandCategory, CstorCommandCategory}, true},
		"test 109": {RunCommandCategoryList{SnapshotCommandCategory, JivaCommandCategory}, false},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			c := Command()
			for _, cat := range mock.categories {
				c = WithCategory(c, cat)
			}
			result := c.Run()

			if !c.IsRun() {
				// test scenario is ignored when runtask command is not run
				return
			}

			if !mock.isSupportedCategory && result.Error() != NotSupportedCategoryError {
				t.Fatalf("Test '%s' failed: expected 'NotSupportedCategoryError': actual '%s': result '%s'", name, result.Error(), result)
			}

			if mock.isSupportedCategory && result.Error() == NotSupportedCategoryError {
				t.Fatalf("Test '%s' failed: expected 'supported category': actual 'NotSupportedCategoryError': result '%s'", name, result)
			}
		})
	}
}

func TestRunCommandCategoryContains(t *testing.T) {
	tests := map[string]struct {
		given    []RunCommandCategory
		contains RunCommandCategory
		expected bool
	}{
		"101": {[]RunCommandCategory{JivaCommandCategory}, JivaCommandCategory, true},
		"102": {[]RunCommandCategory{CstorCommandCategory, JivaCommandCategory}, JivaCommandCategory, true},
		"103": {[]RunCommandCategory{CstorCommandCategory, JivaCommandCategory}, VolumeCommandCategory, false},
		"104": {[]RunCommandCategory{JivaCommandCategory, VolumeCommandCategory}, VolumeCommandCategory, true},
		"105": {[]RunCommandCategory{CstorCommandCategory, VolumeCommandCategory}, VolumeCommandCategory, true},
		"106": {[]RunCommandCategory{}, VolumeCommandCategory, false},
		"107": {[]RunCommandCategory{}, SnapshotCommandCategory, false},
		"108": {[]RunCommandCategory{JivaCommandCategory, SnapshotCommandCategory}, SnapshotCommandCategory, true},
		"109": {[]RunCommandCategory{CstorCommandCategory, SnapshotCommandCategory}, SnapshotCommandCategory, true},
		"110": {[]RunCommandCategory{CstorCommandCategory, SnapshotCommandCategory}, CstorCommandCategory, true},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			c := mockRunCommandFromCategory(mock.given)
			actual := c.Category.Contains(mock.contains)
			if mock.expected != actual {
				t.Fatalf("Test '%s' failed: expected '%t' actual '%t'", name, mock.expected, actual)
			}
		})
	}
}

func TestRunCommandIsJivaVolume(t *testing.T) {
	tests := map[string]struct {
		given    []RunCommandCategory
		expected bool
	}{
		"101": {[]RunCommandCategory{JivaCommandCategory}, false},
		"102": {[]RunCommandCategory{CstorCommandCategory, JivaCommandCategory}, false},
		"103": {[]RunCommandCategory{CstorCommandCategory, VolumeCommandCategory}, false},
		"104": {[]RunCommandCategory{}, false},
		"105": {[]RunCommandCategory{JivaCommandCategory, VolumeCommandCategory}, true},
		"106": {[]RunCommandCategory{JivaCommandCategory, VolumeCommandCategory, CstorCommandCategory}, true},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			c := mockRunCommandFromCategory(mock.given)
			actual := c.Category.IsJivaVolume()
			if mock.expected != actual {
				t.Fatalf("Test '%s' failed: expected '%t' actual '%t'", name, mock.expected, actual)
			}
		})
	}
}

func TestRunCommandIsCstorVolume(t *testing.T) {
	tests := map[string]struct {
		given    []RunCommandCategory
		expected bool
	}{
		"101": {[]RunCommandCategory{JivaCommandCategory}, false},
		"102": {[]RunCommandCategory{CstorCommandCategory, JivaCommandCategory}, false},
		"103": {[]RunCommandCategory{CstorCommandCategory, VolumeCommandCategory}, true},
		"104": {[]RunCommandCategory{}, false},
		"105": {[]RunCommandCategory{JivaCommandCategory, VolumeCommandCategory}, false},
		"106": {[]RunCommandCategory{JivaCommandCategory, VolumeCommandCategory, CstorCommandCategory}, true},
		"107": {[]RunCommandCategory{SnapshotCommandCategory, CstorCommandCategory}, false},
		"108": {[]RunCommandCategory{JivaCommandCategory, SnapshotCommandCategory, CstorCommandCategory}, false},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			c := mockRunCommandFromCategory(mock.given)
			actual := c.Category.IsCstorVolume()
			if mock.expected != actual {
				t.Fatalf("Test '%s' failed: expected '%t' actual '%t'", name, mock.expected, actual)
			}
		})
	}
}

func TestRunCommandIsCstorSnapshot(t *testing.T) {
	tests := map[string]struct {
		given    []RunCommandCategory
		expected bool
	}{
		"101": {[]RunCommandCategory{CstorCommandCategory}, false},
		"102": {[]RunCommandCategory{CstorCommandCategory, JivaCommandCategory}, false},
		"103": {[]RunCommandCategory{CstorCommandCategory, VolumeCommandCategory}, false},
		"104": {[]RunCommandCategory{}, false},
		"105": {[]RunCommandCategory{CstorCommandCategory, VolumeCommandCategory}, false},
		"106": {[]RunCommandCategory{CstorCommandCategory, VolumeCommandCategory, JivaCommandCategory}, false},
		"107": {[]RunCommandCategory{CstorCommandCategory, SnapshotCommandCategory}, true},
		"108": {[]RunCommandCategory{JivaCommandCategory, SnapshotCommandCategory, CstorCommandCategory}, true},
		"109": {[]RunCommandCategory{JivaCommandCategory, SnapshotCommandCategory}, false},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			c := mockRunCommandFromCategory(mock.given)
			actual := c.Category.IsCstorSnapshot()
			if mock.expected != actual {
				t.Fatalf("Test '%s' failed: expected '%t' actual '%t'", name, mock.expected, actual)
			}
		})
	}
}

func TestRunCommandIsValid(t *testing.T) {
	tests := map[string]struct {
		given    []RunCommandCategory
		expected bool
	}{
		"101": {[]RunCommandCategory{JivaCommandCategory}, true},
		"102": {[]RunCommandCategory{CstorCommandCategory, JivaCommandCategory}, false},
		"103": {[]RunCommandCategory{CstorCommandCategory, VolumeCommandCategory}, true},
		"104": {[]RunCommandCategory{}, false},
		"105": {[]RunCommandCategory{JivaCommandCategory, VolumeCommandCategory}, true},
		"106": {[]RunCommandCategory{JivaCommandCategory, VolumeCommandCategory, CstorCommandCategory}, false},
		"107": {[]RunCommandCategory{JivaCommandCategory, SnapshotCommandCategory, CstorCommandCategory}, false},
		"108": {[]RunCommandCategory{SnapshotCommandCategory, CstorCommandCategory}, true},
		"109": {[]RunCommandCategory{SnapshotCommandCategory, JivaCommandCategory}, true},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			c := mockRunCommandFromCategory(mock.given)
			actual := c.Category.IsValid()
			if mock.expected != actual {
				t.Fatalf("Test '%s' failed: expected '%t' actual '%t'", name, mock.expected, actual)
			}
		})
	}
}

// check if defaultCommandRunner implements CommandRunner interface
var _ CommandRunner = &defaultCommandRunner{}

func TestDefaultCommandRunnerRun(t *testing.T) {
	tests := map[string]struct {
		willrun      WillRunFn
		id           string
		cmd          *RunCommand
		expectResult bool
		expectError  bool
		expectDebug  bool
		expectWarn   bool
	}{
		// always run
		"101": {mockAlwaysRun, "", Command(), false, true, true, true},
		"102": {mockAlwaysRun, "t102", nil, false, true, true, true},
		"103": {mockAlwaysRun, "t103", Command(), false, true, true, false},
		"104": {mockAlwaysRun, "t104", VolumeCategory()(Command()), false, true, true, false},
		"105": {mockAlwaysRun, "t105", JivaCategory()(Command()), false, true, true, false},
		"106": {mockAlwaysRun, "t106", CstorCategory()(Command()), false, true, true, false},
		"107": {mockAlwaysRun, "t107", SnapshotCategory()(Command()), false, true, true, false},
		// never run
		"201": {mockNeverRun, "", Command(), false, true, true, true},
		"202": {mockNeverRun, "t202", nil, false, true, true, true},
		"203": {mockNeverRun, "t203", Command(), false, true, true, false},
		"204": {mockNeverRun, "t204", VolumeCategory()(Command()), false, true, true, false},
		"205": {mockNeverRun, "t205", JivaCategory()(Command()), false, true, true, false},
		"206": {mockNeverRun, "t206", CstorCategory()(Command()), false, true, true, false},
		"207": {mockNeverRun, "t207", SnapshotCategory()(Command()), false, true, true, false},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			r := DefaultCommandRunner(mockNoopStore, mock.willrun)
			res := r.Command(mock.id, mock.cmd).Run()
			if !mock.expectResult && res.Result() != nil {
				t.Fatalf("Test '%s' failed: expected nil result: actual '%#v'", name, res.Result())
			}
			if !mock.expectError && res.Error() != nil {
				t.Fatalf("Test '%s' failed: expected nil error: actual '%+v'", name, res.Error())
			}
			if !mock.expectDebug && !res.Debug().IsEmpty() {
				t.Fatalf("Test '%s' failed: expected no debug: actual '%#v'", name, res.Debug())
			}
			if !mock.expectWarn && res.Debug().HasWarn() {
				t.Fatalf("Test '%s' failed: expected no warn: actual '%#v'", name, res.Debug())
			}
		})
	}
}

func TestDefaultCommandRunnerStore(t *testing.T) {
	tests := map[string]struct {
		willrun      WillRunFn
		id           string
		cmd          *RunCommand
		expectResult bool
		expectError  bool
		expectDebug  bool
		expectWarn   bool
	}{
		// always run
		"101": {mockAlwaysRun, "", Command(), false, true, true, false},
		"102": {mockAlwaysRun, "t102", nil, false, true, true, false},
		"103": {mockAlwaysRun, "t103", Command(), false, true, true, false},
		"104": {mockAlwaysRun, "t104", VolumeCategory()(Command()), false, true, true, false},
		"105": {mockAlwaysRun, "t105", JivaCategory()(Command()), false, true, true, false},
		"106": {mockAlwaysRun, "t106", CstorCategory()(Command()), false, true, true, false},
		"107": {mockAlwaysRun, "t107", SnapshotCategory()(Command()), false, true, true, false},
		// never run
		"201": {mockNeverRun, "", Command(), false, true, true, false},
		"202": {mockNeverRun, "t202", nil, false, true, true, false},
		"203": {mockNeverRun, "t203", Command(), false, true, true, false},
		"204": {mockNeverRun, "t204", VolumeCategory()(Command()), false, true, true, false},
		"205": {mockNeverRun, "t205", JivaCategory()(Command()), false, true, true, false},
		"206": {mockNeverRun, "t206", CstorCategory()(Command()), false, true, true, false},
		"207": {mockNeverRun, "t207", SnapshotCategory()(Command()), false, true, true, false},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			stor := map[string]interface{}{}
			runner := DefaultCommandRunner(mockMapStore(stor), mock.willrun)
			runner.Command(mock.id, mock.cmd).Run()
			runid := runner.GetID()

			story, _ := yaml.Marshal(stor)
			s := string(story)

			resp := util.GetNestedField(stor, runid)
			if resp == nil {
				t.Fatalf("Test '%s' failed: expected response at id '%s': actual \n%s", name, runid, s)
			}

			if mock.expectResult && util.GetNestedField(stor, runid, "result") == nil {
				t.Fatalf("Test '%s' failed: expected result at id '%s': actual \n%s", name, runid, s)
			}

			if mock.expectDebug {
				d := util.GetNestedField(stor, runid, "debug").(AllMsgs)
				if d.IsEmpty() {
					t.Fatalf("Test '%s' failed: expected debug at id '%s': actual \n%s", name, runid, s)
				}
			}

			if mock.expectWarn {
				d := util.GetNestedField(stor, runid, "debug").(AllMsgs)
				if !d.HasWarn() {
					t.Fatalf("Test '%s' failed: expected warn at id '%s.debug': actual \n%s", name, runid, s)
				}
			}

			if mock.expectError && util.GetNestedField(stor, runid, "error") == nil {
				t.Fatalf("Test '%s' failed: expected error at id '%s': actual \n%s", name, runid, s)
			}
		})
	}
}
