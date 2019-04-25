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
