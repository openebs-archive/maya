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
	. "github.com/openebs/maya/pkg/msg/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	"testing"
)

// check if storeCommand implements Interface interface
var _ Interface = &storeCommand{}

// check if kvStore implements BucketStorageCondition interface
var _ BucketStorageCondition = &kvStore{}

// mockNoopStore is an implementation of StoreRunResultFn
//func mockNoopStore(id string, key StoreKey, value interface{}) {}

// mockMapStore stores the provided key value pair against the provided id
// inside the provided storage map
//func mockMapStore(store map[string]interface{}) StorageInterface {
//	return KVStore(store)
//}

type mockRunAlwaysStore struct {
	*storeCommand
}

func (m *mockRunAlwaysStore) WillRun() (cond string, willrun bool) {
	return "runalways", true
}

type mockRunNeverStore struct {
	*storeCommand
}

func (m *mockRunNeverStore) WillRun() (cond string, willrun bool) {
	return "runnever", false
}

// mockAlwaysRun is an implementation of WillRunConditionFn
func mockAlwaysRun() bool { return true }

// mockNeverRun is an implementation of WillRunConditionFn
func mockNeverRun() bool { return false }

func TestStoreCommandRun(t *testing.T) {
	tests := map[string]struct {
		store        Interface
		id           string
		cmd          *RunCommand
		expectResult bool
		expectError  bool
		expectDebug  bool
		expectWarn   bool
	}{
		// always run
		"101": {&mockRunAlwaysStore{StoreCommand(KVStore(map[string]interface{}{}))}, "", Command(), false, true, true, true},
		"102": {&mockRunAlwaysStore{StoreCommand(KVStore(map[string]interface{}{}))}, "t102", nil, false, true, true, true},
		"103": {&mockRunAlwaysStore{StoreCommand(KVStore(map[string]interface{}{}))}, "t103", Command(), false, true, true, false},
		"104": {&mockRunAlwaysStore{StoreCommand(KVStore(map[string]interface{}{}))}, "t104", VolumeCategory()(Command()), false, true, true, false},
		"105": {&mockRunAlwaysStore{StoreCommand(KVStore(map[string]interface{}{}))}, "t105", JivaCategory()(Command()), false, true, true, false},
		"106": {&mockRunAlwaysStore{StoreCommand(KVStore(map[string]interface{}{}))}, "t106", CstorCategory()(Command()), false, true, true, false},
		// never run
		"201": {&mockRunNeverStore{StoreCommand(KVStore(map[string]interface{}{}))}, "", Command(), false, true, true, true},
		"202": {&mockRunNeverStore{StoreCommand(KVStore(map[string]interface{}{}))}, "t202", nil, false, true, true, true},
		"203": {&mockRunNeverStore{StoreCommand(KVStore(map[string]interface{}{}))}, "t203", Command(), false, true, true, false},
		"204": {&mockRunNeverStore{StoreCommand(KVStore(map[string]interface{}{}))}, "t204", VolumeCategory()(Command()), false, true, true, false},
		"205": {&mockRunNeverStore{StoreCommand(KVStore(map[string]interface{}{}))}, "t205", JivaCategory()(Command()), false, true, true, false},
		"206": {&mockRunNeverStore{StoreCommand(KVStore(map[string]interface{}{}))}, "t206", CstorCategory()(Command()), false, true, true, false},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			sc := mock.store
			sc.Map(mock.id, mock.cmd)
			res := sc.Run()
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

func TestStoreCommandStorage(t *testing.T) {
	mstore := map[string]interface{}{}
	kv := KVStore(mstore)
	sc := StoreCommand(kv)
	runalways := &mockRunAlwaysStore{sc}
	runnever := &mockRunNeverStore{sc}

	// init runs
	sc.Map("sc000", Command())
	sc.Run()

	tests := map[string]struct {
		store        Interface
		id           string
		cmd          *RunCommand
		expectResult bool
		expectError  bool
		expectDebug  bool
		expectWarn   bool
		expectSkip   bool
		expectInfo   bool
	}{
		// always run
		"101": {runalways, "", Command(), false, true, true, false, false, false},
		"102": {runalways, "t102", nil, false, true, true, false, false, false},
		"103": {runalways, "t103", Command(), false, true, true, true, true, true},
		"104": {runalways, "t104", VolumeCategory()(Command()), false, true, true, true, true, true},
		"105": {runalways, "t105", JivaCategory()(Command()), false, true, true, true, true, true},
		"106": {runalways, "t106", CstorCategory()(Command()), false, true, true, true, true, true},
		// never run
		"201": {runnever, "", Command(), false, true, true, false, false, false},
		"202": {runnever, "t202", nil, false, true, true, false, false, false},
		"203": {runnever, "t203", Command(), false, true, true, true, true, true},
		"204": {runnever, "t204", VolumeCategory()(Command()), false, true, true, true, true, true},
		"205": {runnever, "t205", JivaCategory()(Command()), false, true, true, true, true, true},
		"206": {runnever, "t206", CstorCategory()(Command()), false, true, true, true, true, true},
		// default storage command
		"301": {sc, "", Command(), false, true, true, false, false, false},
		"302": {sc, "t302", nil, false, true, true, false, false, false},
		"303": {sc, "t303", Command(), false, true, true, true, true, true},
		"304": {sc, "t304", VolumeCategory()(Command()), false, true, true, true, true, true},
		"305": {sc, "t305", JivaCategory()(Command()), false, true, true, true, true, true},
		"306": {sc, "t306", CstorCategory()(Command()), false, true, true, true, true, true},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			runner := mock.store
			runner.Map(mock.id, mock.cmd)
			runner.Run()
			runid := runner.ID()

			resp := util.GetNestedField(mstore, runid)
			if resp == nil {
				t.Fatalf("Test '%s' failed: expected response for run id '%s': actual no response: %s", name, runid, kv)
			}

			if mock.expectResult && util.GetNestedField(mstore, runid, "result") == nil {
				t.Fatalf("Test '%s' failed: expected result for run id '%s': actual no result: %s", name, runid, kv)
			}

			if mock.expectError && util.GetNestedField(mstore, runid, "error") == nil {
				t.Fatalf("Test '%s' failed: expected error for run id '%s': actual no error: %s", name, runid, kv)
			}

			d := util.GetNestedField(mstore, runid, "debug").(AllMsgs)

			if mock.expectDebug && d.IsEmpty() {
				t.Fatalf("Test '%s' failed: expected debug for run id '%s': actual no debug: %s", name, runid, kv)
			}
			if !mock.expectDebug && !d.IsEmpty() {
				t.Fatalf("Test '%s' failed: expected no debug for run id '%s': actual debug: %s", name, runid, kv)
			}

			if mock.expectError && !d.HasError() {
				t.Fatalf("Test '%s' failed: expected error for run id '%s': actual no error: %s", name, runid, kv)
			}
			if !mock.expectError && d.HasError() {
				t.Fatalf("Test '%s' failed: expected no error for run id '%s': actual error: %s", name, runid, kv)
			}

			if mock.expectWarn && !d.HasWarn() {
				t.Fatalf("Test '%s' failed: expected warn for run id '%s': actual no warn: %s", name, runid, kv)
			}
			if !mock.expectWarn && d.HasWarn() {
				t.Fatalf("Test '%s' failed: expected no warn for run id '%s': actual warn: %s", name, runid, kv)
			}

			if mock.expectSkip && !d.HasSkip() {
				t.Fatalf("Test '%s' failed: expected skip for run id '%s': actual no skip: %s", name, runid, kv)
			}
			if !mock.expectSkip && d.HasSkip() {
				t.Fatalf("Test '%s' failed: expected no skip for run id '%s': actual skip: %s", name, runid, kv)
			}

			if mock.expectInfo && !d.HasInfo() {
				t.Fatalf("Test '%s' failed: expected info for run id '%s': actual no info: %s", name, runid, kv)
			}
			if !mock.expectInfo && d.HasInfo() {
				t.Fatalf("Test '%s' failed: expected no info for run id '%s': actual info: %s", name, runid, kv)
			}
		})
	}
}
