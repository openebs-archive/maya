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
	"testing"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/upgrade/v1alpha1"
)

func fakePredicateTrue() Predicate {
	return func(p *UpgradeConfig) bool {
		return true
	}
}

func fakePredicateFalse() Predicate {
	return func(p *UpgradeConfig) bool {
		return false
	}
}

func fakeUpgradeConfigValid() *UpgradeConfig {
	return &UpgradeConfig{
		object: &apis.UpgradeConfig{
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
		expectUpgradeConfig bool
		expectChecks        bool
	}{
		"call NewBuilder": {
			true, true,
		},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			b := NewBuilder()
			if (b.UpgradeConfig != nil) != mock.expectUpgradeConfig {
				t.Fatalf("test %s failed, expect patch: %t but got: %t",
					name, mock.expectUpgradeConfig, b.UpgradeConfig != nil)
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
			b := NewBuilder().
				WithYamlString(mock.yaml)
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
			b := NewBuilder().
				WithRawBytes(mock.raw)
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
			b := NewBuilder()
			b.AddChecks(mock.checks...)
			err := b.validate()
			if (err != nil) != mock.expectedError {
				t.Fatalf("test %s failed, expected error: %t but got: %t",
					name, mock.expectedError, err)
			}
		})
	}
}

func TestIsCASTemplateNamePresent(t *testing.T) {
	tests := map[string]struct {
		config         *UpgradeConfig
		expectedOutput bool
	}{
		"valid upgrade config": {
			fakeUpgradeConfigValid(),
			true,
		},
		"invalid upgrade config": {
			&UpgradeConfig{
				object: &apis.UpgradeConfig{
					Data:      []apis.DataItem{},
					Resources: []apis.ResourceDetails{},
				},
			},
			false,
		},
	}
	for name, mock := range tests {
		op := IsCASTemplateNamePresent()(mock.config)
		if op != mock.expectedOutput {
			t.Fatalf("test %s failed, expected error: %t but got: %t",
				name, mock.expectedOutput, op)
		}
	}
}

func TestIsResourcePresent(t *testing.T) {
	tests := map[string]struct {
		config         *UpgradeConfig
		expectedOutput bool
	}{
		"valid upgrade config": {
			fakeUpgradeConfigValid(),
			true,
		},
		"invalid upgrade config": {
			&UpgradeConfig{
				object: &apis.UpgradeConfig{
					Data:      []apis.DataItem{},
					Resources: []apis.ResourceDetails{},
				},
			},
			false,
		},
	}
	for name, mock := range tests {
		op := IsResourcePresent()(mock.config)
		if op != mock.expectedOutput {
			t.Fatalf("test %s failed, expected error: %t but got: %t",
				name, mock.expectedOutput, op)
		}
	}
}

func TestIsValidResource(t *testing.T) {
	tests := map[string]struct {
		config         *UpgradeConfig
		expectedOutput bool
	}{
		"valid upgrade config": {
			fakeUpgradeConfigValid(),
			true,
		},
		"invalid upgrade config namespace not present": {
			&UpgradeConfig{
				object: &apis.UpgradeConfig{
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
			&UpgradeConfig{
				object: &apis.UpgradeConfig{
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
			&UpgradeConfig{
				object: &apis.UpgradeConfig{
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
			&UpgradeConfig{
				object: &apis.UpgradeConfig{
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
			&UpgradeConfig{
				object: &apis.UpgradeConfig{
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
			&UpgradeConfig{
				object: &apis.UpgradeConfig{
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
			&UpgradeConfig{
				object: &apis.UpgradeConfig{
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
		config         *UpgradeConfig
		expectedOutput bool
	}{
		"valid upgrade config": {
			fakeUpgradeConfigValid(),
			true,
		},
		"invalid upgrade config": {
			&UpgradeConfig{
				object: &apis.UpgradeConfig{
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
		config        *UpgradeConfig
		checks        []Predicate
		expectedError bool
	}{
		"predicate returns true": {
			&UpgradeConfig{
				object: &apis.UpgradeConfig{},
			},
			[]Predicate{fakePredicateTrue()},
			false,
		},
		"predicate returns false": {
			&UpgradeConfig{
				object: &apis.UpgradeConfig{},
			},
			[]Predicate{fakePredicateFalse()},
			true,
		},
		"predicate returns both true and false": {
			&UpgradeConfig{
				object: &apis.UpgradeConfig{},
			},
			[]Predicate{fakePredicateFalse(), fakePredicateFalse()},
			true,
		},
	}
	for name, mock := range tests {
		b := &Builder{
			UpgradeConfig: mock.config,
			checks:        make(map[*Predicate]string),
		}
		b.AddChecks(mock.checks...)
		_, err := b.Build()
		if (err != nil) != mock.expectedError {
			t.Fatalf("test %s failed, expected error: %t but got: %t",
				name, mock.expectedError, err)
		}
	}
}
