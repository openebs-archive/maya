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

package v1alpha3

import (
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// StoragePoolKind holds the value of StoragePoolClaim
	StoragePoolKind = "StoragePoolClaim"
	// StoragePoolKindCSPC holds the value of CStorPoolCluster
	StoragePoolKindCSPC = "CStorPoolCluster"
	// APIVersion holds the value of OpenEBS version
	APIVersion = "openebs.io/v1alpha1"
)

// Builder is the builder object for CStorPool
type Builder struct {
	CSP  *CStorPool
	errs []error
}

// NewBuilder returns an empty instance of the Builder object
func NewBuilder() *Builder {
	return &Builder{
		CSP:  &CStorPool{&apis.NewTestCStorPool{}},
		errs: []error{},
	}
}

// BuilderForObject returns an instance of the Builder object based on
// CStorPool object
func BuilderForObject(CStorPool *CStorPool) *Builder {
	return &Builder{
		CSP:  CStorPool,
		errs: []error{},
	}
}

// BuilderForAPIObject returns an instance of the Builder object based on
// CStorPool api object.
func BuilderForAPIObject(csp *apis.NewTestCStorPool) *Builder {
	return &Builder{
		CSP:  &CStorPool{csp},
		errs: []error{},
	}
}

// Build returns the CStorPoolClaim instance
func (b *Builder) Build() (*CStorPool, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("%+v", b.errs)
	}
	return b.CSP, nil
}

// WithName sets the Name field of CSP with provided value.
func (b *Builder) WithName(name string) *Builder {
	if len(name) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build CSP object: missing CSP name"),
		)
		return b
	}
	b.CSP.Object.Name = name
	return b
}

// WithNamespace sets the Namespace field of CSP provided arguments
func (b *Builder) WithNamespace(namespace string) *Builder {
	if len(namespace) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build CSP object: missing CSP namespace"),
		)
		return b
	}
	b.CSP.Object.Namespace = namespace
	return b
}

// WithAnnotationsNew sets the Annotations field of CSP with provided arguments
func (b *Builder) WithAnnotationsNew(annotations map[string]string) *Builder {
	if len(annotations) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build CSP object: missing annotations"),
		)
		return b
	}
	b.CSP.Object.Annotations = make(map[string]string)
	for key, value := range annotations {
		b.CSP.Object.Annotations[key] = value
	}
	return b
}

// WithAnnotations appends or overwrites existing Annotations
// values of CSP with provided arguments
func (b *Builder) WithAnnotations(annotations map[string]string) *Builder {
	if len(annotations) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build CSP object: missing annotations"),
		)
		return b
	}
	if b.CSP.Object.Annotations == nil {
		return b.WithAnnotationsNew(annotations)
	}
	for key, value := range annotations {
		b.CSP.Object.Annotations[key] = value
	}
	return b
}

// WithLabelsNew sets the Labels field of CSP with provided arguments
func (b *Builder) WithLabelsNew(labels map[string]string) *Builder {
	if len(labels) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build CSP object: missing labels"),
		)
		return b
	}
	b.CSP.Object.Labels = make(map[string]string)
	for key, value := range labels {
		b.CSP.Object.Labels[key] = value
	}
	return b
}

// WithLabels appends or overwrites existing Labels
// values of CSP with provided arguments
func (b *Builder) WithLabels(labels map[string]string) *Builder {
	if len(labels) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build CSP object: missing labels"),
		)
		return b
	}
	if b.CSP.Object.Labels == nil {
		return b.WithLabelsNew(labels)
	}
	for key, value := range labels {
		b.CSP.Object.Labels[key] = value
	}
	return b
}

// WithNodeSelectorByReference sets the node selector field of CSP with provided argument.
func (b *Builder) WithNodeSelectorByReference(nodeSelector map[string]string) *Builder {
	if len(nodeSelector) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build CSP object: missing nodeSelector"),
		)
		return b
	}
	b.CSP.Object.Spec.NodeSelector = nodeSelector
	return b
}

// WithNodeName sets the HostName field of CSP with the provided argument.
func (b *Builder) WithNodeName(nodeName string) *Builder {
	if len(nodeName) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build CSP object: missing node name"),
		)
		return b
	}
	b.CSP.Object.Spec.HostName = nodeName
	return b
}

// WithPoolConfig sets the pool config field of the CSP with the provided config.
func (b *Builder) WithPoolConfig(poolConfig *apis.PoolConfig) *Builder {
	if poolConfig == nil {
		b.errs = append(
			b.errs,
			errors.New("failed to build CSP object: missing poolConfig"),
		)
		return b
	}
	b.CSP.Object.Spec.PoolConfig = *poolConfig
	return b
}

// WithRaidGroups sets the raid group field of the CSP with the provided raid groups.
func (b *Builder) WithRaidGroups(raidGroup []apis.RaidGroup) *Builder {
	if len(raidGroup) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build CSP object: missing raidGroups"),
		)
		return b
	}

	b.CSP.Object.Spec.RaidGroup = raidGroup
	return b
}

// WithOwnerReference sets the OwnerReference field in CSP with required
//fields
func (b *Builder) WithOwnerReference(spc *apis.StoragePoolClaim) *Builder {
	if spc == nil {
		b.errs = append(
			b.errs,
			errors.New("failed to build CSP object: spc object is nil"),
		)
		return b
	}
	trueVal := true
	reference := metav1.OwnerReference{
		APIVersion:         APIVersion,
		Kind:               StoragePoolKind,
		UID:                spc.ObjectMeta.UID,
		Name:               spc.ObjectMeta.Name,
		BlockOwnerDeletion: &trueVal,
		Controller:         &trueVal,
	}
	b.CSP.Object.OwnerReferences = append(b.CSP.Object.OwnerReferences, reference)
	return b
}

// WithCSPCOwnerReference sets the OwnerReference field in CSP with required
//fields
func (b *Builder) WithCSPCOwnerReference(cspc *apis.CStorPoolCluster) *Builder {
	if cspc == nil {
		b.errs = append(
			b.errs,
			errors.New("failed to build CSP object: cspc object is nil"),
		)
		return b
	}
	trueVal := true
	reference := metav1.OwnerReference{
		APIVersion:         APIVersion,
		Kind:               StoragePoolKindCSPC,
		UID:                cspc.ObjectMeta.UID,
		Name:               cspc.ObjectMeta.Name,
		BlockOwnerDeletion: &trueVal,
		Controller:         &trueVal,
	}
	b.CSP.Object.OwnerReferences = append(b.CSP.Object.OwnerReferences, reference)
	return b
}
