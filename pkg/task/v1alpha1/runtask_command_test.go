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
	cas "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	. "github.com/openebs/maya/pkg/msg/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"testing"
)

func mockRunCommandFromCategory(l []RunCommandCategory) (r *RunCommand) {
	r = Command()
	r.Category = append(r.Category, l...)
	return
}

// check if RunCommand implements Runner interface
var _ Runner = &RunCommand{}

func TestSelectPathsString(t *testing.T) {
	tests := map[string]struct {
		paths    []string
		expected string
	}{
		"101": {[]string{".metadata.name"}, "select '.metadata.name'"},
		"102": {[]string{".metadata.name", ".metadata.namespace"}, "select '.metadata.name' '.metadata.namespace'"},
		"103": {[]string{".metadata.name as name"}, "select '.metadata.name as name'"},
		"104": {[]string{".metadata.name as name", ".spec.replicas as replicas"}, "select '.metadata.name as name' '.spec.replicas as replicas'"},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			s := SelectPaths{}
			s = append(s, mock.paths...)
			if s.String() != mock.expected {
				t.Fatalf("Test '%s' failed: expected '%s': actual '%s'", name, mock.expected, s.String())
			}
		})
	}
}

func TestSelectsAliasPaths(t *testing.T) {
	tests := map[string]struct {
		paths    []string
		expected map[string]string
	}{
		"101": {nil, nil},
		"102": {[]string{}, nil},
		"103": {[]string{".metadata.name"}, map[string]string{"s0": ".metadata.name"}},
		"104": {[]string{".metadata.name", ".metadata.namespace"}, map[string]string{"s0": ".metadata.name", "s1": ".metadata.namespace"}},
		"105": {[]string{".metadata.name as name"}, map[string]string{"name": ".metadata.name"}},
		"106": {[]string{".metadata.name as name", ".spec.replicas as replicas"}, map[string]string{"name": ".metadata.name", "replicas": ".spec.replicas"}},
		// invalid aliases
		"201": {[]string{".metadata.name name"}, map[string]string{"s0": ".metadata.name name"}},
		"202": {[]string{".metadata.name is name"}, map[string]string{"s0": ".metadata.name is name"}},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			s := SelectPaths{}
			s = append(s, mock.paths...)
			m := s.aliasPaths()

			if len(s) == 0 && mock.expected != nil {
				t.Fatalf("Test '%s' failed: expected nil: actual '%#v'", name, mock.expected)
			}

			for a, p := range m {
				if mock.expected[a] != p {
					t.Fatalf("Test '%s' failed: with alias '%s': expected '%s': actual '%s'", name, a, mock.expected[a], p)
				}
			}
		})
	}
}

func TestSelectsQueryCommandResult(t *testing.T) {
	var result interface{}
	result = &cas.CStorVolumeReplica{
		ObjectMeta: v1.ObjectMeta{
			Name: "my-cstor-rep",
		},
		Spec: cas.CStorVolumeReplicaSpec{
			TargetIP: "20.10.10.10",
			Capacity: "40Gi",
		},
		Status: cas.CStorVolumeReplicaStatus{
			Phase: "Online",
		},
	}
	r := NewRunCommandResult(result, AllMsgs{})

	tests := map[string]struct {
		paths    []string
		expected map[string]interface{}
	}{
		"101": {nil, nil},
		"102": {[]string{}, nil},
		"103": {[]string{"{.Name}"}, map[string]interface{}{"s0": "my-cstor-rep"}},
		"104": {[]string{"{.Spec.TargetIP}", "{.Spec.Capacity}"}, map[string]interface{}{"s0": "20.10.10.10", "s1": "40Gi"}},
		"105": {[]string{"{..TargetIP}", "{..Capacity}"}, map[string]interface{}{"s0": "20.10.10.10", "s1": "40Gi"}},
		"106": {[]string{"{.Status.Phase}", "{..Phase}"}, map[string]interface{}{"s0": "Online", "s1": "Online"}},
		"107": {[]string{"{.Status.Phase} as phase", "{..Phase} as ph"}, map[string]interface{}{"phase": "Online", "ph": "Online"}},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			s := SelectPaths{}
			s = append(s, mock.paths...)
			u := s.QueryCommandResult(r)

			if len(mock.paths) == 0 {
				// there were no runtime errors! good !!!
				return
			}
			result := u.Result()
			m, ok := result.(map[string]interface{})
			if !ok {
				t.Fatalf("Test '%s' failed: expected map[string]interface{}: actual '%#v'", name, u)
			}
			for alias, value := range m {
				if !reflect.DeepEqual(mock.expected[alias], value) {
					t.Fatalf("Test '%s' failed for alias '%s': expected '%#v': actual '%#v'", name, alias, mock.expected[alias], value)
				}
			}
		})
	}
}

func TestSelectsQueryCommandResultV2(t *testing.T) {
	var result interface{}
	result = struct {
		Items []*cas.CStorVolumeReplica
	}{
		Items: []*cas.CStorVolumeReplica{
			&cas.CStorVolumeReplica{
				ObjectMeta: v1.ObjectMeta{Name: "my-cstor-rep"},
				Spec:       cas.CStorVolumeReplicaSpec{TargetIP: "20.10.10.10", Capacity: "40Gi"},
				Status:     cas.CStorVolumeReplicaStatus{Phase: "Online"},
			},
			&cas.CStorVolumeReplica{
				ObjectMeta: v1.ObjectMeta{Name: "my-cstor-rep-2"},
				Spec:       cas.CStorVolumeReplicaSpec{TargetIP: "20.1.1.1", Capacity: "20Gi"},
				Status:     cas.CStorVolumeReplicaStatus{Phase: "Offline"},
			},
		},
	}
	r := NewRunCommandResult(result, AllMsgs{})

	tests := map[string]struct {
		paths    []string
		expected map[string]interface{}
	}{
		"101": {nil, nil},
		"102": {[]string{}, nil},
		"103": {[]string{"{.Items[*].Name}"}, map[string]interface{}{"s0": []string{"my-cstor-rep", "my-cstor-rep-2"}}},
		"104": {[]string{"{.Items[*].Spec.TargetIP}"}, map[string]interface{}{"s0": []string{"20.10.10.10", "20.1.1.1"}}},
		"105": {[]string{"{.Items[*]..TargetIP}"}, map[string]interface{}{"s0": []string{"20.10.10.10", "20.1.1.1"}}},
		"106": {[]string{"{.Items[*].Status.Phase}"}, map[string]interface{}{"s0": []string{"Online", "Offline"}}},
		"107": {[]string{"{.Items[*]..Phase}"}, map[string]interface{}{"s0": []string{"Online", "Offline"}}},
		"108": {[]string{"{.Items[*].Status.Phase} as phase"}, map[string]interface{}{"phase": []string{"Online", "Offline"}}},
		"109": {[]string{"{range .Items[*]..Phase}{@}{end}"}, map[string]interface{}{"s0": []string{"Online", "Offline"}}},
		"110": {[]string{"{range .Items[*].Spec}{@.TargetIP}{@.Capacity}{end}"}, map[string]interface{}{"s0": []string{"20.10.10.10", "40Gi", "20.1.1.1", "20Gi"}}},
		"111": {[]string{"{range .Items[*].Spec}{.TargetIP}{.Capacity}{end}"}, map[string]interface{}{"s0": []string{"20.10.10.10", "40Gi", "20.1.1.1", "20Gi"}}},
		"112": {[]string{"{.Items[*].Spec['.TargetIP','.Capacity']}"}, map[string]interface{}{"s0": []string{"20.10.10.10", "20.1.1.1", "40Gi", "20Gi"}}},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			s := SelectPaths{}
			s = append(s, mock.paths...)
			u := s.QueryCommandResult(r)

			if len(mock.paths) == 0 {
				// there were no runtime errors! good !!!
				return
			}
			result := u.Result()
			m, ok := result.(map[string]interface{})
			if !ok {
				t.Fatalf("Test '%s' failed: expected map[string]interface{}: actual '%#v'", name, u)
			}
			for alias, value := range m {
				if !reflect.DeepEqual(mock.expected[alias], value) {
					t.Fatalf("Test '%s' failed for alias '%s': expected '%#v': actual '%#v'", name, alias, mock.expected[alias], value)
				}
			}
		})
	}
}

func TestRunCommandEnable(t *testing.T) {
	tests := map[string]struct {
		predicate RunPredicate
		willrun   bool
	}{
		"101": {On, true},
		"102": {Off, false},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			c := Command()
			isRun := c.Enable(mock.predicate).IsRun()
			if isRun != mock.willrun {
				t.Fatalf("Test '%s' failed: expected willrun '%t': actual willrun '%t'", name, mock.willrun, isRun)
			}
		})
	}
}

func TestRunCommandPostRun(t *testing.T) {
	tests := map[string]struct {
		result   RunCommandResult
		selects  SelectPaths
		expected map[string]interface{}
	}{
		"101": {RunCommandResult{}, nil, nil},
		"102": {RunCommandResult{}, []string{"{.}"}, nil},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			c := WithSelect(Command(), mock.selects)
			u := c.postRun(mock.result)
			result := u.Result()
			if mock.expected == nil && result != nil {
				t.Fatalf("Test '%s' failed: expected nil result: actual '%#v'", name, result)
			}
			if mock.expected == nil {
				return
			}
			// this test is designed to have result as a map
			m, ok := result.(map[string]interface{})
			if !ok {
				t.Fatalf("Test '%s' failed: expected map[string]interface{}: actual '%#v'", name, result)
			}
			// test each value within the map
			for alias, value := range m {
				if !reflect.DeepEqual(mock.expected[alias], value) {
					t.Fatalf("Test '%s' failed for alias '%s': expected '%#v': actual '%#v'", name, alias, mock.expected[alias], value)
				}
			}
		})
	}
}

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
		"test 101": {RunCommandCategoryList{JivaCommandCategory, CstorCommandCategory}, false},
		"test 102": {RunCommandCategoryList{VolumeCommandCategory, CstorCommandCategory}, false},
		"test 103": {RunCommandCategoryList{VolumeCommandCategory, PoolCommandCategory}, false},
		"test 104": {RunCommandCategoryList{JivaCommandCategory, PoolCommandCategory}, false},
		"test 105": {RunCommandCategoryList{JivaCommandCategory, VolumeCommandCategory}, true},
		"test 106": {RunCommandCategoryList{VolumeCommandCategory, JivaCommandCategory}, true},
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

			if !mock.isSupportedCategory && result.Error() != ErrorNotSupportedCategory {
				t.Fatalf("Test '%s' failed: expected 'ErrorNotSupportedCategory': actual '%s': result '%s'", name, result.Error(), result)
			}

			if mock.isSupportedCategory && result.Error() == ErrorNotSupportedCategory {
				t.Fatalf("Test '%s' failed: expected 'supported category': actual 'ErrorNotSupportedCategory': result '%s'", name, result)
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

func TestRunCommandIsValid(t *testing.T) {
	tests := map[string]struct {
		given    []RunCommandCategory
		expected bool
	}{
		"101": {[]RunCommandCategory{JivaCommandCategory}, true},
		"102": {[]RunCommandCategory{CstorCommandCategory, JivaCommandCategory}, false},
		"103": {[]RunCommandCategory{CstorCommandCategory, VolumeCommandCategory}, true},
		"104": {[]RunCommandCategory{}, true},
		"105": {[]RunCommandCategory{JivaCommandCategory, VolumeCommandCategory}, true},
		"106": {[]RunCommandCategory{JivaCommandCategory, VolumeCommandCategory, CstorCommandCategory}, false},
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

func TestRunCommandIsEmpty(t *testing.T) {
	tests := map[string]struct {
		given    []RunCommandCategory
		expected bool
	}{
		"101": {[]RunCommandCategory{JivaCommandCategory}, false},
		"102": {[]RunCommandCategory{CstorCommandCategory, JivaCommandCategory}, false},
		"103": {[]RunCommandCategory{CstorCommandCategory, VolumeCommandCategory}, false},
		"104": {[]RunCommandCategory{}, true},
		"105": {[]RunCommandCategory{JivaCommandCategory, VolumeCommandCategory}, false},
		"106": {[]RunCommandCategory{JivaCommandCategory, VolumeCommandCategory, CstorCommandCategory}, false},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			c := mockRunCommandFromCategory(mock.given)
			actual := c.Category.IsEmpty()
			if mock.expected != actual {
				t.Fatalf("Test '%s' failed: expected '%t' actual '%t'", name, mock.expected, actual)
			}
		})
	}
}
