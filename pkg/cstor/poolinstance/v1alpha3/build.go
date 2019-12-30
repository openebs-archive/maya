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

// Builder is the builder object for CStorPoolInstance
type Builder struct {
	CSPI *CSPI
	errs []error
}

// NewBuilder returns an empty instance of the Builder object
func NewBuilder() *Builder {
	return &Builder{
		CSPI: &CSPI{&apis.CStorPoolInstance{}},
		errs: []error{},
	}
}

// BuilderForObject returns an instance of the Builder object based on
// CStorPool object
func BuilderForObject(cspi *CSPI) *Builder {
	return &Builder{
		CSPI: cspi,
		errs: []error{},
	}
}

// BuilderForAPIObject returns an instance of the Builder object based on
// CStorPool api object.
func BuilderForAPIObject(cspi *apis.CStorPoolInstance) *Builder {
	return &Builder{
		CSPI: &CSPI{cspi},
		errs: []error{},
	}
}

// Build returns the CStorPoolClaim instance
func (b *Builder) Build() (*CSPI, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("%+v", b.errs)
	}
	return b.CSPI, nil
}

// WithName sets the Name field of CSPI with provided value.
func (b *Builder) WithName(name string) *Builder {
	if len(name) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build CSPI object: missing CSPI name"),
		)
		return b
	}
	b.CSPI.Object.Name = name
	return b
}

// WithNamespace sets the Namespace field of CSPI provided arguments
func (b *Builder) WithNamespace(namespace string) *Builder {
	if len(namespace) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build CSPI object: missing CSPI namespace"),
		)
		return b
	}
	b.CSPI.Object.Namespace = namespace
	return b
}

// WithAnnotationsNew sets the Annotations field of CSPI with provided arguments
func (b *Builder) WithAnnotationsNew(annotations map[string]string) *Builder {
	if len(annotations) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build CSPI object: missing annotations"),
		)
		return b
	}
	b.CSPI.Object.Annotations = make(map[string]string)
	for key, value := range annotations {
		b.CSPI.Object.Annotations[key] = value
	}
	return b
}

// WithAnnotations appends or overwrites existing Annotations
// values of CSPI with provided arguments
func (b *Builder) WithAnnotations(annotations map[string]string) *Builder {
	if len(annotations) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build CSPI object: missing annotations"),
		)
		return b
	}
	if b.CSPI.Object.Annotations == nil {
		return b.WithAnnotationsNew(annotations)
	}
	for key, value := range annotations {
		b.CSPI.Object.Annotations[key] = value
	}
	return b
}

// WithLabelsNew sets the Labels field of CSPI with provided arguments
func (b *Builder) WithLabelsNew(labels map[string]string) *Builder {
	if len(labels) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build CSPI object: missing labels"),
		)
		return b
	}
	b.CSPI.Object.Labels = make(map[string]string)
	for key, value := range labels {
		b.CSPI.Object.Labels[key] = value
	}
	return b
}

// WithLabels appends or overwrites existing Labels
// values of CSPI with provided arguments
func (b *Builder) WithLabels(labels map[string]string) *Builder {
	if len(labels) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build CSPI object: missing labels"),
		)
		return b
	}
	if b.CSPI.Object.Labels == nil {
		return b.WithLabelsNew(labels)
	}
	for key, value := range labels {
		b.CSPI.Object.Labels[key] = value
	}
	return b
}

// WithNodeSelectorByReference sets the node selector field of CSPI with provided argument.
func (b *Builder) WithNodeSelectorByReference(nodeSelector map[string]string) *Builder {
	if len(nodeSelector) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build CSPI object: missing nodeSelector"),
		)
		return b
	}
	b.CSPI.Object.Spec.NodeSelector = nodeSelector
	return b
}

// WithNodeName sets the HostName field of CSPI with the provided argument.
func (b *Builder) WithNodeName(nodeName string) *Builder {
	if len(nodeName) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build CSPI object: missing node name"),
		)
		return b
	}
	b.CSPI.Object.Spec.HostName = nodeName
	return b
}

// WithPoolConfig sets the pool config field of the CSPI with the provided config.
func (b *Builder) WithPoolConfig(poolConfig *apis.PoolConfig) *Builder {
	if poolConfig == nil {
		b.errs = append(
			b.errs,
			errors.New("failed to build CSPI object: missing poolConfig"),
		)
		return b
	}
	b.CSPI.Object.Spec.PoolConfig = *poolConfig
	return b
}

// WithRaidGroups sets the raid group field of the CSPI with the provided raid groups.
func (b *Builder) WithRaidGroups(raidGroup []apis.RaidGroup) *Builder {
	if len(raidGroup) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build CSPI object: missing raidGroups"),
		)
		return b
	}

	b.CSPI.Object.Spec.RaidGroups = raidGroup
	return b
}

// WithFinalizer sets the finalizer field in the BDC
func (b *Builder) WithFinalizer(finalizers ...string) *Builder {
	if len(finalizers) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build CSPI object: missing finalizer"),
		)
		return b
	}
	b.CSPI.Object.Finalizers = append(b.CSPI.Object.Finalizers, finalizers...)
	return b
}

// WithOwnerReference sets the OwnerReference field in CSPI with required
//fields
func (b *Builder) WithOwnerReference(spc *apis.StoragePoolClaim) *Builder {
	if spc == nil {
		b.errs = append(
			b.errs,
			errors.New("failed to build CSPI object: spc object is nil"),
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
	b.CSPI.Object.OwnerReferences = append(b.CSPI.Object.OwnerReferences, reference)
	return b
}

// WithCSPCOwnerReference sets the OwnerReference field in CSPI with required
//fields
func (b *Builder) WithCSPCOwnerReference(cspic *apis.CStorPoolCluster) *Builder {
	if cspic == nil {
		b.errs = append(
			b.errs,
			errors.New("failed to build CSPI object: cspic object is nil"),
		)
		return b
	}
	trueVal := true
	reference := metav1.OwnerReference{
		APIVersion:         APIVersion,
		Kind:               StoragePoolKindCSPC,
		UID:                cspic.ObjectMeta.UID,
		Name:               cspic.ObjectMeta.Name,
		BlockOwnerDeletion: &trueVal,
		Controller:         &trueVal,
	}
	b.CSPI.Object.OwnerReferences = append(b.CSPI.Object.OwnerReferences, reference)
	return b
}

// WithNewVersion sets the current and desired version field of
// CSPI with provided arguments
func (b *Builder) WithNewVersion(version string) *Builder {
	if version == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build cspi object: version can't be empty",
			),
		)
		return b
	}
	b.CSPI.Object.VersionDetails.Status.Current = version
	b.CSPI.Object.VersionDetails.Desired = version
	return b
}

// WithDependentsUpgraded sets the field to true for new CSPI
func (b *Builder) WithDependentsUpgraded() *Builder {
	b.CSPI.Object.VersionDetails.Status.DependentsUpgraded = true
	return b
}
