package v1alpha1

import (
	"github.com/pkg/errors"
	storagev1 "k8s.io/api/storage/v1"
)

// Builder enables building an instance of StorageClass
type Builder struct {
	sc   *StorageClass
	errs []error
}

// NewBuilder returns new instance of Builder
func NewBuilder() *Builder {
	return &Builder{sc: &StorageClass{object: &storagev1.StorageClass{}}}
}

// WithName sets the Name field of SC with provided argument.
func (b *Builder) WithName(name string) *Builder {
	if len(name) == 0 {
		b.errs = append(b.errs, errors.New("failed to build SC object: missing storageclass name"))
		return b
	}
	b.sc.object.Name = name
	return b
}

// WithAnnotations sets the Annotations field of SC with provided value.
func (b *Builder) WithAnnotations(annotations map[string]string) *Builder {
	if len(annotations) == 0 {
		b.errs = append(b.errs, errors.New("failed to build SC object: missing annotations"))
	}
	b.sc.object.Annotations = annotations
	return b
}

// WithProvisioner sets the Provisioner field of SC with provided argument.
func (b *Builder) WithProvisioner(provisioner string) *Builder {
	if len(provisioner) == 0 {
		b.errs = append(b.errs, errors.New("failed to build storageclass: missing provisioner name"))
		return b
	}
	b.sc.object.Provisioner = provisioner
	return b
}

// Build returns the StorageClass API instance
func (b *Builder) Build() (*storagev1.StorageClass, []error) {
	if len(b.errs) > 0 {
		return nil, b.errs
	}
	return b.sc.object, nil
}
