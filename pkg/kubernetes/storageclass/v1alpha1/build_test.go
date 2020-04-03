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

	storagev1 "k8s.io/api/storage/v1"
)

func TestBuilderWithName(t *testing.T) {
	tests := map[string]struct {
		name      string
		builder   *Builder
		expectErr bool
	}{
		"Test SC with name": {
			name:      "SC1",
			builder:   &Builder{sc: &StorageClass{object: &storagev1.StorageClass{}}},
			expectErr: false,
		},
		"Test SC without name": {
			name:      "",
			builder:   &Builder{sc: &StorageClass{object: &storagev1.StorageClass{}}},
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

func TestBuilderWithAnnotations(t *testing.T) {
	tests := map[string]struct {
		annotations map[string]string
		builder     *Builder
		expectErr   bool
	}{
		"Test SC with annotations": {
			annotations: map[string]string{"persistent-volume": "PV"},
			builder:     &Builder{sc: &StorageClass{object: &storagev1.StorageClass{}}},
			expectErr:   false,
		},
		"Test SC without annotations": {
			annotations: map[string]string{},
			builder:     &Builder{sc: &StorageClass{object: &storagev1.StorageClass{}}},
			expectErr:   true,
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

func TestBuilderWithProvisioner(t *testing.T) {
	tests := map[string]struct {
		name      string
		builder   *Builder
		expectErr bool
	}{
		"Test SC with name": {
			name:      "openebs.io/provisioner-iscsi",
			builder:   &Builder{sc: &StorageClass{object: &storagev1.StorageClass{}}},
			expectErr: false,
		},
		"Test SC without name": {
			name:      "",
			builder:   &Builder{sc: &StorageClass{object: &storagev1.StorageClass{}}},
			expectErr: true,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithProvisioner(mock.name)
			if mock.expectErr && len(b.errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestBuilderWithVolumeBind(t *testing.T) {
	tests := map[string]struct {
		bindmode  storagev1.VolumeBindingMode
		builder   *Builder
		expectErr bool
	}{
		"Test SC with immediate binding mode": {
			bindmode:  storagev1.VolumeBindingImmediate,
			builder:   &Builder{sc: &StorageClass{object: &storagev1.StorageClass{}}},
			expectErr: false,
		},
		"Test SC with waitforconsumer binding mode": {
			bindmode:  storagev1.VolumeBindingWaitForFirstConsumer,
			builder:   &Builder{sc: &StorageClass{object: &storagev1.StorageClass{}}},
			expectErr: false,
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := mock.builder.WithVolumeBindingMode(mock.bindmode)
			if mock.expectErr && len(b.errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && len(b.errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}
