package v1alpha1

import (
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// Builder is the builder object for PVC
type Builder struct {
	Pvc *PVC
	err error
}

// NewBuilder returns new instance of Builder
func NewBuilder() *Builder {
	return &Builder{Pvc: &PVC{object: &corev1.PersistentVolumeClaim{}}}
}

// WithName sets the Name field of PVC with provided value.
func (b *Builder) WithName(name string) *Builder {
	if b.err != nil {
		return b
	}
	if len(name) == 0 {
		b.err = errors.New("failed to build PVC object: missing PVC name")
		return b
	}
	b.Pvc.object.Name = name
	return b
}

// WithNamespace sets the Namespace field of PVC provided arguments
func (b *Builder) WithNamespace(namespace string) *Builder {
	if b.err != nil {
		return b
	}
	if len(namespace) == 0 {
		namespace = "default"
	}
	b.Pvc.object.Namespace = namespace
	return b
}

// WithAnnotations sets the Annotations field of PVC with provided arguments
func (b *Builder) WithAnnotations(annotations map[string]string) *Builder {
	if b.err != nil {
		return b
	}
	b.Pvc.object.Annotations = annotations
	return b
}

// WithLabels sets the Labels field of PVC with provided arguments
func (b *Builder) WithLabels(labels map[string]string) *Builder {
	if b.err != nil {
		return b
	}
	b.Pvc.object.Labels = labels
	return b
}

// WithStorageClass sets the StorageClass field of PVC with provided arguments
func (b *Builder) WithStorageClass(scName string) *Builder {
	if b.err != nil {
		return b
	}
	if len(scName) == 0 {
		b.err = errors.New("failed to build PVC object: missing storageclass name")
		return b
	}
	b.Pvc.object.Spec.StorageClassName = &scName
	return b
}

// WithAccessModes sets the AccessMode field in PVC with provided arguments
func (b *Builder) WithAccessModes(accessMode []corev1.PersistentVolumeAccessMode) *Builder {
	if b.err != nil {
		return b
	}
	if len(accessMode) == 0 {
		b.err = errors.New("failed to build PVC object: missing accessmodes types")
		return b
	}
	b.Pvc.object.Spec.AccessModes = accessMode
	return b
}

// WithCapacity sets the Capacity field in PVC with provided arguments
func (b *Builder) WithCapacity(capacity string) *Builder {
	if b.err != nil {
		return b
	}
	resCapacity, err := resource.ParseQuantity(capacity)
	if err != nil {
		b.err = err
		return b
	}
	resourceList := corev1.ResourceList{
		corev1.ResourceName(corev1.ResourceStorage): resCapacity,
	}
	b.Pvc.object.Spec.Resources.Requests = resourceList
	return b
}

// Build returns the PVC API instance
func (b *Builder) Build() (*corev1.PersistentVolumeClaim, error) {
	if b.err != nil {
		return nil, b.err
	}
	return b.Pvc.object, nil
}
