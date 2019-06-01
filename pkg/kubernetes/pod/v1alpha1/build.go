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

// Builder is the builder object for Pod
type Builder struct {
	pod  *Pod
	errs []error
}

// NewBuilder returns new instance of Builder
func NewBuilder() *Builder {
	return &Builder{pod: &Pod{object: &corev1.Pod{}}}
}

// WithName sets the Name field of Pod with provided value.
func (b *Builder) WithName(name string) *Builder {
	if len(name) == 0 {
		b.errs = append(b.errs, errors.New("failed to build Pod object: missing Pod name"))
		return b
	}
	b.pod.object.Name = name
	return b
}

// WithRestartPolicy sets the RestartPolicy field in Pod with provided arguments
func (b *Builder) WithRestartPolicy(restartPolicy corev1.RestartPolicy) *Builder {
	b.pod.object.Spec.RestartPolicy = restartPolicy
	return b
}

// WithNodeName sets the NodeName field of Pod with provided value.
func (b *Builder) WithNodeName(nodeName string) *Builder {
	if len(nodeName) == 0 {
		b.errs = append(b.errs, errors.New("failed to build Pod object: missing Pod node name"))
		return b
	}
	b.pod.object.Spec.NodeName = nodeName
	return b
}

// WithContainers sets the Containers field in Pod with provided arguments
func (b *Builder) WithContainers(containers []corev1.Container) *Builder {
	if len(containers) == 0 {
		b.errs = append(b.errs, errors.New("failed to build Pod object: missing containers"))
		return b
	}
	b.pod.object.Spec.Containers = containers
	return b
}

// WithContainer sets the Containers field in Pod with provided arguments
func (b *Builder) WithContainer(container corev1.Container) *Builder {
	return b.WithContainers([]corev1.Container{container})
}

// WithVolumes sets the Volumes field in Pod with provided arguments
func (b *Builder) WithVolumes(volumes []corev1.Volume) *Builder {
	if len(volumes) == 0 {
		b.errs = append(b.errs, errors.New("failed to build Pod object: missing volumes"))
		return b
	}
	b.pod.object.Spec.Volumes = volumes
	return b
}

// WithVolume sets the Volumes field in Pod with provided arguments
func (b *Builder) WithVolume(volume corev1.Volume) *Builder {
	return b.WithVolumes([]corev1.Volume{volume})
}

// Build returns the Pod API instance
func (b *Builder) Build() (*corev1.Pod, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("%+v", b.errs)
	}
	return b.pod.object, nil
}
