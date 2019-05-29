/*
Copyright 2019 The OpenEBS Authors

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
	"reflect"
	"testing"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/upgrade/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func fakePredicate() Predicate {
	return func(p *UpgradeResult) bool {
		return true
	}
}
func TestNewBuilder(t *testing.T) {
	tests := map[string]struct {
		expectUpgradeResult bool
		expectChecks        bool
	}{
		"call NewBuilder": {
			true, true,
		},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			b := NewBuilder()
			if (b.UpgradeResult != nil) != mock.expectUpgradeResult {
				t.Fatalf("test %s failed, expect upgraderesult: %t but got: %t",
					name, mock.expectUpgradeResult, b.UpgradeResult != nil)
			}
			if (b.checks != nil) != mock.expectChecks {
				t.Fatalf("test %s failed, expect checks: %t but got: %t",
					name, mock.expectChecks, b.checks != nil)
			}
		})
	}
}

func TestAddCheck(t *testing.T) {
	tests := map[string]struct {
		input                Predicate
		expectedChecksLength int
	}{
		"When a predicate is given": {
			fakePredicate(), 1,
		},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			b := NewBuilder().AddCheck(mock.input)
			if len(b.checks) != mock.expectedChecksLength {
				t.Fatalf("test %s failed, expected checks length %+v but got : %+v",
					name, mock.expectedChecksLength, len(b.checks))
			}
		})
	}
}

func TestWithAPIList(t *testing.T) {
	inputURItems := []apis.UpgradeResult{apis.UpgradeResult{
		ObjectMeta: metav1.ObjectMeta{Name: "upgradeResultList1"}}}
	outputURItems := []*UpgradeResult{&UpgradeResult{object: &apis.UpgradeResult{
		ObjectMeta: metav1.ObjectMeta{Name: "upgradeResultList1"}}}}
	tests := map[string]struct {
		inputURList    *apis.UpgradeResultList
		expectedOutput *UpgradeResultList
	}{
		"empty upgrade result list": {&apis.UpgradeResultList{},
			&UpgradeResultList{}},
		"using nil input": {nil, &UpgradeResultList{}},
		"non-empty upgrade result list": {&apis.UpgradeResultList{Items: inputURItems},
			&UpgradeResultList{items: outputURItems}},
	}

	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			b := NewListBuilder().WithAPIList(mock.inputURList)
			if len(b.list.items) != len(mock.expectedOutput.items) {
				t.Fatalf("test %s failed, expected len: %d got: %d",
					name, len(mock.expectedOutput.items), len(b.list.items))
			}
			if !reflect.DeepEqual(b.list, mock.expectedOutput) {
				t.Fatalf("test %s failed, expected : %+v got : %+v",
					name, mock.expectedOutput, b.list)
			}
		})
	}
}

func TestList(t *testing.T) {
	inputURItems := []apis.UpgradeResult{apis.UpgradeResult{
		ObjectMeta: metav1.ObjectMeta{Name: "upgradeResultList1"}}}
	outputURItems := []*UpgradeResult{&UpgradeResult{object: &apis.UpgradeResult{
		ObjectMeta: metav1.ObjectMeta{Name: "upgradeResultList1"}}}}
	tests := map[string]struct {
		inputURList    *apis.UpgradeResultList
		expectedOutput *UpgradeResultList
	}{
		"empty upgrade result list": {&apis.UpgradeResultList{},
			&UpgradeResultList{}},
		"using nil input": {nil, &UpgradeResultList{}},
		"non-empty upgrade result list": {&apis.UpgradeResultList{Items: inputURItems},
			&UpgradeResultList{items: outputURItems}},
	}

	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			b := NewListBuilder().WithAPIList(mock.inputURList).List()
			if len(b.items) != len(mock.expectedOutput.items) {
				t.Fatalf("test %s failed, expected len: %d got: %d",
					name, len(mock.expectedOutput.items), len(b.items))
			}
			if !reflect.DeepEqual(b, mock.expectedOutput) {
				t.Fatalf("test %s failed, expected : %+v got : %+v",
					name, mock.expectedOutput, b)
			}
		})
	}
}

func TestWithTypeMeta(t *testing.T) {
	tests := map[string]struct {
		typeMeta           metav1.TypeMeta
		expectedKind       string
		expectedAPIVersion string
	}{
		"only kind present": {
			metav1.TypeMeta{
				Kind: "fake-kind",
			},
			"fake-kind",
			"",
		},
		"only api version present": {
			metav1.TypeMeta{
				APIVersion: "fake-api-version",
			},
			"",
			"fake-api-version",
		},
		"both kind and api version present": {
			metav1.TypeMeta{
				Kind:       "fake-kind",
				APIVersion: "fake-api-version",
			},
			"fake-kind",
			"fake-api-version",
		},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		b := &Builder{
			UpgradeResult: &UpgradeResult{
				object: &apis.UpgradeResult{},
			},
			checks: make(map[*Predicate]string),
		}
		b.WithTypeMeta(mock.typeMeta)
		if b.object.Kind != mock.expectedKind {
			t.Fatalf("test %s failed, expected kind %s, but got : %s",
				name, mock.expectedKind, b.object.Kind)
		}
		if b.object.APIVersion != mock.expectedAPIVersion {
			t.Fatalf("test %s failed, expected apiVersion %s, but got : %s",
				name, mock.expectedAPIVersion, b.object.APIVersion)
		}
	}
}

func TestWithObjectMeta(t *testing.T) {
	tests := map[string]struct {
		objectMeta        metav1.ObjectMeta
		expectedName      string
		expectedNamespace string
	}{
		"only name present": {
			metav1.ObjectMeta{
				Name: "fake-name",
			},
			"fake-name",
			"",
		},
		"only namespace present": {
			metav1.ObjectMeta{
				Namespace: "fake-namespace",
			},
			"",
			"fake-namespace",
		},
		"both kind and api version present": {
			metav1.ObjectMeta{
				Name:      "fake-name",
				Namespace: "fake-namespace",
			},
			"fake-name",
			"fake-namespace",
		},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		b := &Builder{
			UpgradeResult: &UpgradeResult{
				object: &apis.UpgradeResult{},
			},
			checks: make(map[*Predicate]string),
		}
		b.WithObjectMeta(mock.objectMeta)
		if b.object.Name != mock.expectedName {
			t.Fatalf("test %s failed, expected name %s, but got : %s",
				name, mock.expectedName, b.object.Name)
		}
		if b.object.Namespace != mock.expectedNamespace {
			t.Fatalf("test %s failed, expected namespace %s, but got : %s",
				name, mock.expectedNamespace, b.object.Namespace)
		}
	}
}

func TestWithTasks(t *testing.T) {
	tests := map[string]struct {
		tasks      []apis.UpgradeResultTask
		expecttask bool
	}{
		"one task present": {
			[]apis.UpgradeResultTask{
				apis.UpgradeResultTask{},
			},
			true,
		},
		"more than one tasks present": {
			[]apis.UpgradeResultTask{
				apis.UpgradeResultTask{},
				apis.UpgradeResultTask{},
				apis.UpgradeResultTask{},
			},
			true,
		},
		"no task present": {
			[]apis.UpgradeResultTask{},
			false,
		},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		b := &Builder{
			UpgradeResult: &UpgradeResult{
				object: &apis.UpgradeResult{},
			},
			checks: make(map[*Predicate]string),
		}
		b.WithTasks(mock.tasks...)
		if (len(b.object.Tasks) != 0) != mock.expecttask {
			t.Fatalf("test %s failed, expect task %t, but got : %t",
				name, mock.expecttask, len(b.object.Tasks) != 0)
		}
	}
}

func TestWithResultConfig(t *testing.T) {
	tests := map[string]struct {
		data       []apis.DataItem
		expectdata bool
	}{
		"one data present": {
			[]apis.DataItem{
				apis.DataItem{},
			},
			true,
		},
		"more than one data present": {
			[]apis.DataItem{
				apis.DataItem{},
				apis.DataItem{},
				apis.DataItem{},
			},
			true,
		},
		"no data present": {
			[]apis.DataItem{},
			false,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		b := &Builder{
			UpgradeResult: &UpgradeResult{
				object: &apis.UpgradeResult{},
			},
			checks: make(map[*Predicate]string),
		}
		b.WithResultConfig(apis.ResourceDetails{}, mock.data...)
		if (len(b.object.Config.Data) != 0) != mock.expectdata {
			t.Fatalf("test %s failed, expect data %t, but got : %t",
				name, mock.expectdata, len(b.object.Config.Data) != 0)
		}
	}
}
