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
	"testing"
)

func TestNewWithID(t *testing.T) {
	tests := map[string]struct {
		id string
	}{
		"t1": {"101"},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			r := New(WithID(mock.id))
			if r.ID != mock.id {
				t.Fatalf("test '%s' failed: expected id '%s' actual '%s'", name, mock.id, r.ID)
			}
		})
	}
}

func TestBuilderWithID(t *testing.T) {
	tests := map[string]struct {
		id string
	}{
		"t1": {"101"},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			r := Builder().WithID(mock.id).Build()
			if r.ID != mock.id {
				t.Fatalf("test '%s' failed: expected id '%s' actual '%s'", name, mock.id, r.ID)
			}
		})
	}
}

func TestNewWithName(t *testing.T) {
	tests := map[string]struct {
		name string
	}{
		"t1": {"mytest"},
	}
	for n, mock := range tests {
		t.Run(n, func(t *testing.T) {
			r := New(WithName(mock.name))
			if r.Name != mock.name {
				t.Fatalf("test '%s' failed: expected name '%s' actual '%s'", n, mock.name, r.Name)
			}
		})
	}
}

func TestBuilderWithName(t *testing.T) {
	tests := map[string]struct {
		name string
	}{
		"t1": {"mytest"},
	}
	for n, mock := range tests {
		t.Run(n, func(t *testing.T) {
			r := Builder().WithName(mock.name).Build()
			if r.Name != mock.name {
				t.Fatalf("test '%s' failed: expected name '%s' actual '%s'", n, mock.name, r.Name)
			}
		})
	}
}

func TestNewWithAction(t *testing.T) {
	tests := map[string]struct {
		action apis.Action
	}{
		"t1": {apis.Get},
	}
	for n, mock := range tests {
		t.Run(n, func(t *testing.T) {
			r := New(WithAction(mock.action))
			if r.Action != mock.action {
				t.Fatalf("test '%s' failed: expected action '%s' actual '%s'", n, mock.action, r.Action)
			}
		})
	}
}

func TestBuilderWithAction(t *testing.T) {
	tests := map[string]struct {
		action apis.Action
	}{
		"t1": {apis.Get},
	}
	for n, mock := range tests {
		t.Run(n, func(t *testing.T) {
			r := Builder().WithAction(mock.action).Build()
			if r.Action != mock.action {
				t.Fatalf("test '%s' failed: expected action '%s' actual '%s'", n, mock.action, r.Action)
			}
		})
	}
}

func TestNewWithAPIVersion(t *testing.T) {
	tests := map[string]struct {
		apiVersion string
	}{
		"t1": {"v1beta1"},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			r := New(WithAPIVersion(mock.apiVersion))
			if r.APIVersion != mock.apiVersion {
				t.Fatalf("test '%s' failed: expected api version '%s' actual '%s'", name, mock.apiVersion, r.APIVersion)
			}
		})
	}
}

func TestBuilderWithAPIVersion(t *testing.T) {
	tests := map[string]struct {
		apiVersion string
	}{
		"t1": {"v1beta1"},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			r := Builder().WithAPIVersion(mock.apiVersion).Build()
			if r.APIVersion != mock.apiVersion {
				t.Fatalf("test '%s' failed: expected api version '%s' actual '%s'", name, mock.apiVersion, r.APIVersion)
			}
		})
	}
}
