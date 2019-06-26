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

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
)

func TestBuilderWithName(t *testing.T) {
	tests := map[string]struct {
		name      string
		builder   *Builder
		expectErr bool
	}{
		"Test Builder with name": {
			name: "PVC1",
			builder: &Builder{cvr: &CVR{
				object: &apis.CStorVolumeReplica{},
			}},
			expectErr: false,
		},
		"Test Builder without name": {
			name: "",
			builder: &Builder{cvr: &CVR{
				object: &apis.CStorVolumeReplica{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithName(mock.name)
			if mock.expectErr && len(b.errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuilderWithGenerateName(t *testing.T) {
	tests := map[string]struct {
		generatename string
		builder      *Builder
		expectErr    bool
	}{
		"Test Builder with generatename": {
			generatename: "PVC1",
			builder: &Builder{cvr: &CVR{
				object: &apis.CStorVolumeReplica{},
			}},
			expectErr: false,
		},
		"Test Builder without generatename": {
			generatename: "",
			builder: &Builder{cvr: &CVR{
				object: &apis.CStorVolumeReplica{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithGenerateName(mock.generatename)
			if mock.expectErr && len(b.errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuilderWithNamespace(t *testing.T) {
	tests := map[string]struct {
		namespace string
		builder   *Builder
		expectErr bool
	}{
		"Test Builder with namespace": {
			namespace: "PVC1",
			builder: &Builder{cvr: &CVR{
				object: &apis.CStorVolumeReplica{},
			}},
			expectErr: false,
		},
		"Test Builder without namespace": {
			namespace: "",
			builder: &Builder{cvr: &CVR{
				object: &apis.CStorVolumeReplica{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithNamespace(mock.namespace)
			if mock.expectErr && len(b.errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuildWithAnnotations(t *testing.T) {
	tests := map[string]struct {
		annotations map[string]string
		builder     *Builder
		expectErr   bool
	}{
		"Test Builderwith annotations": {
			annotations: map[string]string{"persistent-volume": "PV",
				"application": "percona"},
			builder: &Builder{cvr: &CVR{
				object: &apis.CStorVolumeReplica{},
			}},
			expectErr: false,
		},
		"Test Builderwithout annotations": {
			annotations: map[string]string{},
			builder: &Builder{cvr: &CVR{
				object: &apis.CStorVolumeReplica{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithAnnotations(mock.annotations)
			if mock.expectErr && len(b.errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuildWithAnnotationsNew(t *testing.T) {
	tests := map[string]struct {
		annotations map[string]string
		builder     *Builder
		expectErr   bool
	}{
		"Test Builderwith annotations": {
			annotations: map[string]string{"persistent-volume": "PV",
				"application": "percona"},
			builder: &Builder{cvr: &CVR{
				object: &apis.CStorVolumeReplica{},
			}},
			expectErr: false,
		},
		"Test Builderwithout annotations": {
			annotations: map[string]string{},
			builder: &Builder{cvr: &CVR{
				object: &apis.CStorVolumeReplica{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithAnnotationsNew(mock.annotations)
			if mock.expectErr && len(b.errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuildWithLabels(t *testing.T) {
	tests := map[string]struct {
		labels    map[string]string
		builder   *Builder
		expectErr bool
	}{
		"Test Builderwith labels": {
			labels: map[string]string{"persistent-volume": "PV",
				"application": "percona"},
			builder: &Builder{cvr: &CVR{
				object: &apis.CStorVolumeReplica{},
			}},
			expectErr: false,
		},
		"Test Builderwithout labels": {
			labels: map[string]string{},
			builder: &Builder{cvr: &CVR{
				object: &apis.CStorVolumeReplica{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithLabels(mock.labels)
			if mock.expectErr && len(b.errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuildWithLabelsNew(t *testing.T) {
	tests := map[string]struct {
		labels    map[string]string
		builder   *Builder
		expectErr bool
	}{
		"Test Builderwith labels": {
			labels: map[string]string{"persistent-volume": "PV",
				"application": "percona"},
			builder: &Builder{cvr: &CVR{
				object: &apis.CStorVolumeReplica{},
			}},
			expectErr: false,
		},
		"Test Builderwithout labels": {
			labels: map[string]string{},
			builder: &Builder{cvr: &CVR{
				object: &apis.CStorVolumeReplica{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithLabelsNew(mock.labels)
			if mock.expectErr && len(b.errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuilderWithTargetIP(t *testing.T) {
	tests := map[string]struct {
		targetip  string
		builder   *Builder
		expectErr bool
	}{
		"Test Builder with targetip": {
			targetip: "10.8.02.13",
			builder: &Builder{cvr: &CVR{
				object: &apis.CStorVolumeReplica{},
			}},
			expectErr: false,
		},
		"Test Builder without targetip": {
			targetip: "",
			builder: &Builder{cvr: &CVR{
				object: &apis.CStorVolumeReplica{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithTargetIP(mock.targetip)
			if mock.expectErr && len(b.errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuilderWithCapacity(t *testing.T) {
	tests := map[string]struct {
		capacity  string
		builder   *Builder
		expectErr bool
	}{
		"Test Builder with capacity": {
			capacity: "10G",
			builder: &Builder{cvr: &CVR{
				object: &apis.CStorVolumeReplica{},
			}},
			expectErr: false,
		},
		"Test Builder without capacity": {
			capacity: "",
			builder: &Builder{cvr: &CVR{
				object: &apis.CStorVolumeReplica{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithCapacity(mock.capacity)
			if mock.expectErr && len(b.errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuilderWithFinalizers(t *testing.T) {
	tests := map[string]struct {
		finalizers []string
		builder    *Builder
		expectErr  bool
	}{
		"Test Builder with finalizers": {
			finalizers: []string{
				"cstorvolumereplica.openebs.io/finalizer",
			},
			builder: &Builder{cvr: &CVR{
				object: &apis.CStorVolumeReplica{},
			}},
			expectErr: false,
		},
		"Test Builder without finalizers": {
			finalizers: []string{},
			builder: &Builder{cvr: &CVR{
				object: &apis.CStorVolumeReplica{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithFinalizers(mock.finalizers)
			if mock.expectErr && len(b.errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}
