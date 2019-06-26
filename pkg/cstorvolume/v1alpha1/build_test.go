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
			builder: &Builder{cstorvolume: &CStorVolume{
				object: &apis.CStorVolume{},
			}},
			expectErr: false,
		},
		"Test Builder without name": {
			name: "",
			builder: &Builder{cstorvolume: &CStorVolume{
				object: &apis.CStorVolume{},
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
			builder: &Builder{cstorvolume: &CStorVolume{
				object: &apis.CStorVolume{},
			}},
			expectErr: false,
		},
		"Test Builder without generatename": {
			generatename: "",
			builder: &Builder{cstorvolume: &CStorVolume{
				object: &apis.CStorVolume{},
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
			builder: &Builder{cstorvolume: &CStorVolume{
				object: &apis.CStorVolume{},
			}},
			expectErr: false,
		},
		"Test Builder without namespace": {
			namespace: "",
			builder: &Builder{cstorvolume: &CStorVolume{
				object: &apis.CStorVolume{},
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
			builder: &Builder{cstorvolume: &CStorVolume{
				object: &apis.CStorVolume{},
			}},
			expectErr: false,
		},
		"Test Builderwithout annotations": {
			annotations: map[string]string{},
			builder: &Builder{cstorvolume: &CStorVolume{
				object: &apis.CStorVolume{},
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
			builder: &Builder{cstorvolume: &CStorVolume{
				object: &apis.CStorVolume{},
			}},
			expectErr: false,
		},
		"Test Builderwithout annotations": {
			annotations: map[string]string{},
			builder: &Builder{cstorvolume: &CStorVolume{
				object: &apis.CStorVolume{},
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
			builder: &Builder{cstorvolume: &CStorVolume{
				object: &apis.CStorVolume{},
			}},
			expectErr: false,
		},
		"Test Builderwithout labels": {
			labels: map[string]string{},
			builder: &Builder{cstorvolume: &CStorVolume{
				object: &apis.CStorVolume{},
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
			builder: &Builder{cstorvolume: &CStorVolume{
				object: &apis.CStorVolume{},
			}},
			expectErr: false,
		},
		"Test Builderwithout labels": {
			labels: map[string]string{},
			builder: &Builder{cstorvolume: &CStorVolume{
				object: &apis.CStorVolume{},
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
			builder: &Builder{cstorvolume: &CStorVolume{
				object: &apis.CStorVolume{},
			}},
			expectErr: false,
		},
		"Test Builder without targetip": {
			targetip: "",
			builder: &Builder{cstorvolume: &CStorVolume{
				object: &apis.CStorVolume{},
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
			builder: &Builder{cstorvolume: &CStorVolume{
				object: &apis.CStorVolume{},
			}},
			expectErr: false,
		},
		"Test Builder without capacity": {
			capacity: "",
			builder: &Builder{cstorvolume: &CStorVolume{
				object: &apis.CStorVolume{},
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

func TestBuilderWithNodeBase(t *testing.T) {
	tests := map[string]struct {
		nodebase  string
		builder   *Builder
		expectErr bool
	}{
		"Test Builder with nodebase": {
			nodebase: "iqn.2016-09.com.openebs.cstor",
			builder: &Builder{cstorvolume: &CStorVolume{
				object: &apis.CStorVolume{},
			}},
			expectErr: false,
		},
		"Test Builder without nodebase": {
			nodebase: "",
			builder: &Builder{cstorvolume: &CStorVolume{
				object: &apis.CStorVolume{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithNodeBase(mock.nodebase)
			if mock.expectErr && len(b.errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuilderWithIqn(t *testing.T) {
	tests := map[string]struct {
		iqn       string
		builder   *Builder
		expectErr bool
	}{
		"Test Builder with iqn": {
			iqn: "iqn.2016-09.com.openebs.cstor:pv-name",
			builder: &Builder{cstorvolume: &CStorVolume{
				object: &apis.CStorVolume{},
			}},
			expectErr: false,
		},
		"Test Builder without iqn": {
			iqn: "",
			builder: &Builder{cstorvolume: &CStorVolume{
				object: &apis.CStorVolume{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithIQN(mock.iqn)
			if mock.expectErr && len(b.errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuilderWithTargetPort(t *testing.T) {
	tests := map[string]struct {
		targetport string
		builder    *Builder
		expectErr  bool
	}{
		"Test Builder with targetport": {
			targetport: "3600",
			builder: &Builder{cstorvolume: &CStorVolume{
				object: &apis.CStorVolume{},
			}},
			expectErr: false,
		},
		"Test Builder without targetport": {
			targetport: "",
			builder: &Builder{cstorvolume: &CStorVolume{
				object: &apis.CStorVolume{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithTargetPort(mock.targetport)
			if mock.expectErr && len(b.errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuilderWithTargetPortal(t *testing.T) {
	tests := map[string]struct {
		targetportal string
		builder      *Builder
		expectErr    bool
	}{
		"Test Builder with targetportal": {
			targetportal: "10.8.02.13:3600",
			builder: &Builder{cstorvolume: &CStorVolume{
				object: &apis.CStorVolume{},
			}},
			expectErr: false,
		},
		"Test Builder without targetportal": {
			targetportal: "",
			builder: &Builder{cstorvolume: &CStorVolume{
				object: &apis.CStorVolume{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithTargetPortal(mock.targetportal)
			if mock.expectErr && len(b.errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuilderWithReplicationFactor(t *testing.T) {
	tests := map[string]struct {
		replicationfactor int
		builder           *Builder
		expectErr         bool
	}{
		"Test Builder with replicationfactor": {
			replicationfactor: 3,
			builder: &Builder{cstorvolume: &CStorVolume{
				object: &apis.CStorVolume{},
			}},
			expectErr: false,
		},
		"Test Builder with invalid replicationfactor": {
			replicationfactor: -1,
			builder: &Builder{cstorvolume: &CStorVolume{
				object: &apis.CStorVolume{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithReplicationFactor(mock.replicationfactor)
			if mock.expectErr && len(b.errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuilderWithConsistencyFactor(t *testing.T) {
	tests := map[string]struct {
		consistencyfactor int
		builder           *Builder
		expectErr         bool
	}{
		"Test Builder with consistencyfactor": {
			consistencyfactor: 2,
			builder: &Builder{cstorvolume: &CStorVolume{
				object: &apis.CStorVolume{},
			}},
			expectErr: false,
		},
		"Test Builder with invalid consistencyfactor": {
			consistencyfactor: -1,
			builder: &Builder{cstorvolume: &CStorVolume{
				object: &apis.CStorVolume{},
			}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithConsistencyFactor(mock.consistencyfactor)
			if mock.expectErr && len(b.errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}
