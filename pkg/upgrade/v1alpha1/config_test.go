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
	"strings"
	"testing"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/upgrade/v1alpha1"
)

func fakePredicateTrue() Predicate {
	return func(p *Config) bool {
		return true
	}
}

func fakePredicateFalse() Predicate {
	return func(p *Config) bool {
		return false
	}
}

func fakeConfigValid() *Config {
	return &Config{
		Object: &apis.UpgradeConfig{
			CASTemplate: "fake-castemplate",
			Data: []apis.DataItem{
				apis.DataItem{
					Name:  "config-key1",
					Value: "config-value1",
				},
				apis.DataItem{
					Name:  "config-key2",
					Value: "config-value2",
				},
			},
			Resources: []apis.ResourceDetails{
				apis.ResourceDetails{
					Name:      "pool-a",
					Kind:      "CStor-pool",
					Namespace: "openebs",
				},
				apis.ResourceDetails{
					Name:      "pool-b",
					Kind:      "CStor-pool",
					Namespace: "openebs",
				},
			},
		},
	}
}

func TestNewBuilder(t *testing.T) {
	tests := map[string]struct {
		expectConfig bool
		expectChecks bool
	}{
		"call NewBuilder": {
			true, true,
		},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			b := NewConfigBuilder()
			if (b.Config != nil) != mock.expectConfig {
				t.Fatalf("test %s failed, expect patch: %t but got: %t",
					name, mock.expectConfig, b.Config != nil)
			}
			if (b.checks != nil) != mock.expectChecks {
				t.Fatalf("test %s failed, expect checks: %t but got: %t",
					name, mock.expectChecks, b.checks != nil)
			}
		})
	}
}

func TestWithYamlString(t *testing.T) {
	tests := map[string]struct {
		yaml          string
		expectedError bool
	}{
		"valid yaml string": {
			`
casTemplate: cstor-pool-081-082
resources:
- name: pool-a
  kind: cstor-pool
  nameSpace: openebs`,
			false,
		},
		"invalid yaml string": {
			`
	casTemplate: cstor-pool-081-082
resources:
	- name: pool-a
  kind: cstor-pool
  nameSpace: openebs`,
			true,
		},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			b := ConfigBuilderForYaml(mock.yaml)
			if (len(b.errors) != 0) != mock.expectedError {
				t.Fatalf("test %s failed, expect error: %v but got: %v",
					name, mock.expectedError, len(b.errors) != 0)
			}
		})
	}
}

func TestWithRawBytes(t *testing.T) {
	tests := map[string]struct {
		raw           []byte
		expectedError bool
	}{
		"valid yaml string": {
			[]byte(`
casTemplate: cstor-pool-081-082
resources:
- name: pool-a
  kind: cstor-pool
  nameSpace: openebs`),
			false,
		},
		"invalid yaml string": {
			[]byte(`
	casTemplate: cstor-pool-081-082
resources:
	- name: pool-a
  kind: cstor-pool
  nameSpace: openebs`),
			true,
		},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			b := ConfigBuilderForRaw(mock.raw)
			if (len(b.errors) != 0) != mock.expectedError {
				t.Fatalf("test %s failed, expect error: %v but got: %v",
					name, mock.expectedError, len(b.errors) != 0)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := map[string]struct {
		checks        []Predicate
		expectedError bool
	}{
		"predicate returns true": {
			[]Predicate{fakePredicateTrue()},
			false,
		},
		"predicate returns false": {
			[]Predicate{fakePredicateFalse()},
			true,
		},
		"contains mix predicate returns and false": {
			[]Predicate{fakePredicateFalse(), fakePredicateTrue()},
			true,
		},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			b := NewConfigBuilder()
			b.AddChecks(mock.checks...)
			err := b.validate()
			if (err != nil) != mock.expectedError {
				t.Fatalf("test %s failed, expected error: %t but got: %t",
					name, mock.expectedError, err)
			}
		})
	}
}

func TestIsCASTemplateName(t *testing.T) {
	tests := map[string]struct {
		config         *Config
		expectedOutput bool
	}{
		"valid upgrade config": {
			fakeConfigValid(),
			true,
		},
		"invalid upgrade config": {
			&Config{
				Object: &apis.UpgradeConfig{
					Data:      []apis.DataItem{},
					Resources: []apis.ResourceDetails{},
				},
			},
			false,
		},
	}
	for name, mock := range tests {
		op := IsCASTemplateName()(mock.config)
		if op != mock.expectedOutput {
			t.Fatalf("test %s failed, expected error: %t but got: %t",
				name, mock.expectedOutput, op)
		}
	}
}

func TestIsResource(t *testing.T) {
	tests := map[string]struct {
		config         *Config
		expectedOutput bool
	}{
		"valid upgrade config": {
			fakeConfigValid(),
			true,
		},
		"invalid upgrade config": {
			&Config{
				Object: &apis.UpgradeConfig{
					Data:      []apis.DataItem{},
					Resources: []apis.ResourceDetails{},
				},
			},
			false,
		},
	}
	for name, mock := range tests {
		op := IsResource()(mock.config)
		if op != mock.expectedOutput {
			t.Fatalf("test %s failed, expected error: %t but got: %t",
				name, mock.expectedOutput, op)
		}
	}
}

func TestIsValidResource(t *testing.T) {
	tests := map[string]struct {
		config         *Config
		expectedOutput bool
	}{
		"valid upgrade config": {
			fakeConfigValid(),
			true,
		},
		"invalid upgrade config namespace not present": {
			&Config{
				Object: &apis.UpgradeConfig{
					Data: []apis.DataItem{},
					Resources: []apis.ResourceDetails{
						apis.ResourceDetails{
							Name:      "pool-a",
							Namespace: "",
							Kind:      "CStor-pool",
						},
					},
				},
			},
			false,
		},
		"invalid upgrade config name not present": {
			&Config{
				Object: &apis.UpgradeConfig{
					Data: []apis.DataItem{},
					Resources: []apis.ResourceDetails{
						apis.ResourceDetails{
							Name:      "",
							Namespace: "fake-ns",
							Kind:      "CStor-pool",
						},
					},
				},
			},
			false,
		},
		"invalid upgrade config kind not present": {
			&Config{
				Object: &apis.UpgradeConfig{
					Data: []apis.DataItem{},
					Resources: []apis.ResourceDetails{
						apis.ResourceDetails{
							Name:      "pool-a",
							Namespace: "fake-ns",
							Kind:      "",
						},
					},
				},
			},
			false,
		},
		"invalid upgrade config name, namespace not present": {
			&Config{
				Object: &apis.UpgradeConfig{
					Data: []apis.DataItem{},
					Resources: []apis.ResourceDetails{
						apis.ResourceDetails{
							Name:      "",
							Namespace: "",
							Kind:      "CStor-pool",
						},
					},
				},
			},
			false,
		},
		"invalid upgrade config kind, namespace not present": {
			&Config{
				Object: &apis.UpgradeConfig{
					Data: []apis.DataItem{},
					Resources: []apis.ResourceDetails{
						apis.ResourceDetails{
							Name:      "pool-a",
							Namespace: "",
							Kind:      "",
						},
					},
				},
			},
			false,
		},
		"invalid upgrade config name, kind not present": {
			&Config{
				Object: &apis.UpgradeConfig{
					Data: []apis.DataItem{},
					Resources: []apis.ResourceDetails{
						apis.ResourceDetails{
							Name:      "",
							Namespace: "test-ns",
							Kind:      "",
						},
					},
				},
			},
			false,
		},
		"invalid upgrade config name, namespace, kind not present": {
			&Config{
				Object: &apis.UpgradeConfig{
					Data: []apis.DataItem{},
					Resources: []apis.ResourceDetails{
						apis.ResourceDetails{
							Name:      "",
							Namespace: "",
							Kind:      "",
						},
					},
				},
			},
			false,
		},
	}
	for name, mock := range tests {
		op := IsValidResource()(mock.config)
		if op != mock.expectedOutput {
			t.Fatalf("test %s failed, expected error: %t but got: %t",
				name, mock.expectedOutput, op)
		}
	}

}

func TestIsSameKind(t *testing.T) {
	tests := map[string]struct {
		config         *Config
		expectedOutput bool
	}{
		"valid upgrade config": {
			fakeConfigValid(),
			true,
		},
		"invalid upgrade config": {
			&Config{
				Object: &apis.UpgradeConfig{
					Data: []apis.DataItem{},
					Resources: []apis.ResourceDetails{
						apis.ResourceDetails{
							Name: "pool-a",
							Kind: "CStor-pool",
						},
						apis.ResourceDetails{
							Name: "volume-b",
							Kind: "CStor-volume",
						},
					},
				},
			},
			false,
		},
	}
	for name, mock := range tests {
		op := IsSameKind()(mock.config)
		if op != mock.expectedOutput {
			t.Fatalf("test %s failed, expected error: %t but got: %t",
				name, mock.expectedOutput, op)
		}
	}
}

func TestBuild(t *testing.T) {
	tests := map[string]struct {
		config        *Config
		checks        []Predicate
		expectedError bool
	}{
		"predicate returns true": {
			&Config{
				Object: &apis.UpgradeConfig{},
			},
			[]Predicate{fakePredicateTrue()},
			false,
		},
		"predicate returns false": {
			&Config{
				Object: &apis.UpgradeConfig{},
			},
			[]Predicate{fakePredicateFalse()},
			true,
		},
		"predicate returns both true and false": {
			&Config{
				Object: &apis.UpgradeConfig{},
			},
			[]Predicate{fakePredicateFalse(), fakePredicateFalse()},
			true,
		},
	}
	for name, mock := range tests {
		b := &ConfigBuilder{
			Config: mock.config,
			checks: make(map[*Predicate]string),
		}
		b.AddChecks(mock.checks...)
		_, err := b.Build()
		if (err != nil) != mock.expectedError {
			t.Fatalf("test %s failed, expected error: %t but got: %t",
				name, mock.expectedError, err)
		}
	}
}

func TestConfigString(t *testing.T) {
	tests := map[string]struct {
		config              *Config
		expectedStringParts []string
	}{
		"upgrade config": {
			&Config{
				Object: &apis.UpgradeConfig{
					CASTemplate: "my-template",
					Data: []apis.DataItem{
						apis.DataItem{
							Name:  "key-1",
							Value: "value-1",
						},
					},
					Resources: []apis.ResourceDetails{
						apis.ResourceDetails{
							Name:      "pool-ddas",
							Kind:      "CStorPool",
							Namespace: "openebs",
						},
					},
				},
			},
			[]string{"casTemplate: my-template", "data:", "name: key-1", "value: value-1",
				"resources:", "name: pool-ddas", "kind: CStorPool", "namespace: openebs"},
		},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			ymlstr := mock.config.String()
			for _, expect := range mock.expectedStringParts {
				if !strings.Contains(ymlstr, expect) {
					t.Errorf("test '%s' failed: expected '%s' in '%s'", name, expect, ymlstr)
				}
			}
		})
	}
}

func TestConfigGoString(t *testing.T) {
	tests := map[string]struct {
		config              *Config
		expectedStringParts []string
	}{
		"upgrade config": {
			&Config{
				Object: &apis.UpgradeConfig{
					CASTemplate: "my-template",
					Data: []apis.DataItem{
						apis.DataItem{
							Name:  "key-1",
							Value: "value-1",
						},
					},
					Resources: []apis.ResourceDetails{
						apis.ResourceDetails{
							Name:      "pool-ddas",
							Kind:      "CStorPool",
							Namespace: "openebs",
						},
					},
				},
			},
			[]string{"casTemplate: my-template", "data:", "name: key-1", "value: value-1",
				"resources:", "name: pool-ddas", "kind: CStorPool", "namespace: openebs"},
		},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			ymlstr := mock.config.GoString()
			for _, expect := range mock.expectedStringParts {
				if !strings.Contains(ymlstr, expect) {
					t.Errorf("test '%s' failed: expected '%s' in '%s'", name, expect, ymlstr)
				}
			}
		})
	}
}
