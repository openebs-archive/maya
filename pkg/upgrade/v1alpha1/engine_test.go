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

	upgrade "github.com/openebs/maya/pkg/apis/openebs.io/upgrade/v1alpha1"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/pkg/errors"
)

func TestNewEngine(t *testing.T) {
	tests := map[string]struct {
		expectRuntimeConfig bool
		expectCASTemplate   bool
		expectUnitOfUpgrade bool
	}{
		"call NewBuilder": {
			false, false, false,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			b := NewEngine()
			if (b.CASTemplate != nil) != mock.expectCASTemplate {
				t.Fatalf("test %s failed, expect CASTemplate: %t but got: %t",
					name, mock.expectCASTemplate, b.CASTemplate != nil)
			}
			if (b.UnitOfUpgrade != nil) != mock.expectUnitOfUpgrade {
				t.Fatalf("test %s failed, expect UnitOfUpgrade: %t but got: %t",
					name, mock.expectUnitOfUpgrade, b.UnitOfUpgrade != nil)
			}
			if (len(b.RuntimeConfig) != 0) != mock.expectUnitOfUpgrade {
				t.Fatalf("test %s failed, expect RuntimeConfig: %t but got: %t",
					name, mock.expectUnitOfUpgrade, len(b.RuntimeConfig) != 0)
			}
		})
	}
}

func TestWithRuntimeConfig(t *testing.T) {
	tests := map[string]struct {
		runtimeConfig       []upgrade.DataItem
		expectRuntimeConfig bool
	}{
		"runtime config present": {
			[]upgrade.DataItem{
				upgrade.DataItem{
					Name:  "key-12gsf",
					Value: "value-njedr",
				},
			},
			true,
		},
		"runtime config not present": {
			[]upgrade.DataItem{},
			false,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			b := &EngineBuilder{}
			b.WithRuntimeConfig(mock.runtimeConfig)
			if mock.expectRuntimeConfig != (len(b.RuntimeConfig) != 0) {
				t.Fatalf("test %s failed, expect runtimeConfig: %t but got: %t",
					name, mock.expectRuntimeConfig, len(b.RuntimeConfig) != 0)
			}
		})
	}
}

func TestWithUnitOfUpgrade(t *testing.T) {
	tests := map[string]struct {
		unitOfUpgrade       *upgrade.ResourceDetails
		expectUnitOfUpgrade bool
	}{
		"unitOfUpgrade present": {
			&upgrade.ResourceDetails{},
			true,
		},
		"unitOfUpgrade not present": {
			nil,
			false,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			b := &EngineBuilder{}
			b.WithUnitOfUpgrade(mock.unitOfUpgrade)
			if mock.expectUnitOfUpgrade != (b.UnitOfUpgrade != nil) {
				t.Fatalf("test %s failed, expect unitOfUpgrade: %t but got: %t",
					name, mock.expectUnitOfUpgrade, b.UnitOfUpgrade != nil)
			}
		})
	}
}

func TestWithCASTemplate(t *testing.T) {
	tests := map[string]struct {
		casTemplate       *apis.CASTemplate
		expectCASTemplate bool
	}{
		"unitOfUpgrade present": {
			&apis.CASTemplate{},
			true,
		},
		"unitOfUpgrade not present": {
			nil,
			false,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			b := &EngineBuilder{}
			b.WithCASTemplate(mock.casTemplate)
			if mock.expectCASTemplate != (b.CASTemplate != nil) {
				t.Fatalf("test %s failed, expect CASTemplate: %t but got: %t",
					name, mock.expectCASTemplate, b.CASTemplate != nil)
			}
		})
	}
}

func TestValidateEngineBuilder(t *testing.T) {
	tests := map[string]struct {
		builder     *EngineBuilder
		expectError bool
	}{
		"valid builder": {
			&EngineBuilder{
				CASTemplate:   &apis.CASTemplate{},
				UnitOfUpgrade: &upgrade.ResourceDetails{},
			},
			false,
		},
		"error present in builder": {
			&EngineBuilder{
				CASTemplate:   &apis.CASTemplate{},
				UnitOfUpgrade: &upgrade.ResourceDetails{},
				errors: []error{
					errors.New("123"),
				},
			},
			true,
		},
		"castemplate not present in builder": {
			&EngineBuilder{
				UnitOfUpgrade: &upgrade.ResourceDetails{},
			},
			true,
		},
		"unit of upgrade not present in builder": {
			&EngineBuilder{
				CASTemplate: &apis.CASTemplate{},
			},
			true,
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			e := mock.builder.validate()
			if mock.expectError != (e != nil) {
				t.Fatalf("test %s failed, expect unitOfUpgrade: %t but got: %t",
					name, mock.expectError, e != nil)
			}
		})
	}
}

func TestEngineBuilderString(t *testing.T) {
	tests := map[string]struct {
		builder             *EngineBuilder
		expectedStringParts []string
	}{
		"engine builder": {
			&EngineBuilder{
				RuntimeConfig: []apis.Config{
					apis.Config{
						Name:  "key-1",
						Value: "value-1",
					},
				},
				CASTemplate: &apis.CASTemplate{},
				UnitOfUpgrade: &upgrade.ResourceDetails{
					Name:      "pool-ddas",
					Kind:      "CStorPool",
					Namespace: "openebs",
				},
			},
			[]string{"RuntimeConfig:", "name: key-1", "value: value-1",
				"UnitOfUpgrade:", "name: pool-ddas", "kind: CStorPool", "namespace: openebs"},
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			ymlstr := mock.builder.String()
			for _, expect := range mock.expectedStringParts {
				if !strings.Contains(ymlstr, expect) {
					t.Errorf("test '%s' failed: expected '%s' in '%s'", name, expect, ymlstr)
				}
			}
		})
	}
}

func TestEngineBuilderGoString(t *testing.T) {
	tests := map[string]struct {
		builder             *EngineBuilder
		expectedStringParts []string
	}{
		"engine builder": {
			&EngineBuilder{
				RuntimeConfig: []apis.Config{
					apis.Config{
						Name:  "key-1",
						Value: "value-1",
					},
				},
				CASTemplate: &apis.CASTemplate{},
				UnitOfUpgrade: &upgrade.ResourceDetails{
					Name:      "pool-ddas",
					Kind:      "CStorPool",
					Namespace: "openebs",
				},
			},
			[]string{"RuntimeConfig:", "name: key-1", "value: value-1",
				"UnitOfUpgrade:", "name: pool-ddas", "kind: CStorPool", "namespace: openebs"},
		},
	}
	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			ymlstr := mock.builder.GoString()
			for _, expect := range mock.expectedStringParts {
				if !strings.Contains(ymlstr, expect) {
					t.Errorf("test '%s' failed: expected '%s' in '%s'", name, expect, ymlstr)
				}
			}
		})
	}
}
