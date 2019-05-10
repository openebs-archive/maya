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
	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// Builder is the builder object for Volume
type Builder struct {
	volume *Volume
	errs   []error
}

// NewBuilder returns new instance of Builder
func NewBuilder() *Builder {
	return &Builder{volume: &Volume{object: &corev1.Volume{}}}
}

// WithName sets the Name field of Volume with provided value.
func (b *Builder) WithName(name string) *Builder {
	if len(name) == 0 {
		b.errs = append(b.errs, errors.New("failed to build Volume object: missing Volume name"))
		return b
	}
	b.volume.object.Name = name
	return b
}

// WithHostDirectory sets the VolumeSource field of Volume with provided hostpath
// as type directory.
func (b *Builder) WithHostDirectory(path string) *Builder {
	if len(path) == 0 {
		b.errs = append(b.errs, errors.New("failed to build volume object: missing volume path"))
		return b
	}
	volumeSource := corev1.VolumeSource{
		HostPath: &corev1.HostPathVolumeSource{
			Path: path,
		},
	}

	b.volume.object.VolumeSource = volumeSource
	return b
}

// Build returns the Volume API instance
func (b *Builder) Build() (*corev1.Volume, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("%+v", b.errs)
	}
	return b.volume.object, nil
}
