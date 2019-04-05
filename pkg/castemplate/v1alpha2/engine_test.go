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

	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/util"
)

func TestUnMarshallToConfig(t *testing.T) {
	tests := map[string]struct {
		config       string
		expectNames  []string
		expectValues []string
	}{
		"101": {
			config: `
        - name: StoragePoolClaim
          value: "cstor-pool-default-0.7.0"
        - name: ReplicaCount
          value: "3"`,
			expectNames:  []string{"StoragePoolClaim", "ReplicaCount"},
			expectValues: []string{"cstor-pool-default-0.7.0", "3"},
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			c, err := UnMarshallToConfig(mock.config)
			if err != nil {
				t.Fatalf("Test '%s' failed: expected no error: actual '%#v'", name, err)
			}
			actualNames := []string{}
			actualValues := []string{}
			for _, conf := range c {
				if !util.ContainsString(mock.expectNames, conf.Name) {
					t.Fatalf("Test '%s' failed: config name '%s' was not expected", name, conf.Name)
				}
				if !util.ContainsString(mock.expectValues, conf.Value) {
					t.Fatalf("Test '%s' failed: config value '%s' was not expected", name, conf.Value)
				}
				actualNames = append(actualNames, conf.Name)
				actualValues = append(actualValues, conf.Value)
			}
			if len(actualValues) != len(mock.expectValues) {
				t.Fatalf("Test '%s' failed: expected values count '%d' actual '%d'", name, len(mock.expectValues), len(actualValues))
			}
			if len(actualNames) != len(mock.expectNames) {
				t.Fatalf("Test '%s' failed: expected names count '%d' actual '%d'", name, len(mock.expectNames), len(actualNames))
			}
		})
	}
}

func TestMergeConfig(t *testing.T) {
	p1 := v1alpha1.Config{Name: "pod", Value: "p1"}
	p2 := v1alpha1.Config{Name: "pod", Value: "p2"}
	d1 := v1alpha1.Config{Name: "deploy", Value: "d1"}
	d2 := v1alpha1.Config{Name: "deploy", Value: "d2"}
	s1 := v1alpha1.Config{Name: "service", Value: "s1"}
	s2 := v1alpha1.Config{Name: "service", Value: "s2"}

	tests := map[string]struct {
		highPriority []v1alpha1.Config
		lowPriority  []v1alpha1.Config
		expectNames  []string
		expectValues []string
	}{
		"101": {
			highPriority: []v1alpha1.Config{p1, d1},
			lowPriority:  []v1alpha1.Config{p2, d2},
			expectNames:  []string{"pod", "deploy"},
			expectValues: []string{"p1", "d1"},
		},
		"102": {
			highPriority: []v1alpha1.Config{p2, d2},
			lowPriority:  []v1alpha1.Config{p1, d1},
			expectNames:  []string{"pod", "deploy"},
			expectValues: []string{"p2", "d2"},
		},
		"103": {
			highPriority: []v1alpha1.Config{p2, d2},
			lowPriority:  []v1alpha1.Config{p1, d1, s1},
			expectNames:  []string{"pod", "deploy", "service"},
			expectValues: []string{"p2", "d2", "s1"},
		},
		"104": {
			highPriority: []v1alpha1.Config{p2, d1, s2},
			lowPriority:  []v1alpha1.Config{p1, s1},
			expectNames:  []string{"pod", "deploy", "service"},
			expectValues: []string{"p2", "d1", "s2"},
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			f := MergeConfig(mock.highPriority, mock.lowPriority)
			for _, conf := range f {
				if !util.ContainsString(mock.expectNames, conf.Name) {
					t.Fatalf("Test '%s' failed: name '%s' was not expected", name, conf.Name)
				}
				if !util.ContainsString(mock.expectValues, conf.Value) {
					t.Fatalf("Test '%s' failed: value '%s' was not expected", name, conf.Value)
				}
			}
		})
	}
}

func TestConfigToMap(t *testing.T) {
	p1 := v1alpha1.Config{Name: "pod", Value: "p1"}
	p2 := v1alpha1.Config{Name: "pod", Value: "p2"}
	d1 := v1alpha1.Config{Name: "deploy", Value: "d1"}
	d2 := v1alpha1.Config{Name: "deploy", Value: "d2"}
	s1 := v1alpha1.Config{Name: "service", Value: "s1"}

	tests := map[string]struct {
		config       []v1alpha1.Config
		expectNames  []string
		expectValues []string
	}{
		"101": {
			config:       []v1alpha1.Config{p1, d1},
			expectNames:  []string{"pod", "deploy"},
			expectValues: []string{"p1", "d1"},
		},
		"102": {
			config:       []v1alpha1.Config{p2, d2},
			expectNames:  []string{"pod", "deploy"},
			expectValues: []string{"p2", "d2"},
		},
		"103": {
			config:       []v1alpha1.Config{p2, d1, s1},
			expectNames:  []string{"pod", "deploy", "service"},
			expectValues: []string{"p2", "d1", "s1"},
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			m, err := ConfigToMap(mock.config)
			if err != nil {
				t.Fatalf("Test '%s' failed: expected no error: actual '%#v'", name, err)
			}
			for k, v := range m {
				if !util.ContainsString(mock.expectNames, k) {
					t.Fatalf("Test '%s' failed: key '%s' was not expected", name, k)
				}
				for kk, vv := range v.(map[string]string) {
					if kk == "value" && !util.ContainsString(mock.expectValues, vv) {
						t.Fatalf("Test '%s' failed: value '%s' was not expected", name, vv)
					}
				}
			}
		})
	}
}
