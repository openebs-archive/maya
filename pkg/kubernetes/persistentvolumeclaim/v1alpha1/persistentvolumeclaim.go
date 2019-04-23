// Copyright © 2018-2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// PVC is a wrapper over persistentvolumeclaim api
// object. It provides build, validations and other common
// logic to be used by various feature specific callers.
type PVC struct {
	Object *corev1.PersistentVolumeClaim
	Err    error
}

// PVCList is a wrapper over persistentvolumeclaim api
// object. It provides build, validations and other common
// logic to be used by various feature specific callers.
type PVCList struct {
	items []*PVC
}

// ListBuilder enables building an instance of
// PVClist
type ListBuilder struct {
	list    *PVCList
	filters PredicateList
}

// Build returns the final instance of patch
// TODO add validations and error checks
func (b *ListBuilder) Build() (*PVCList, error) {
	return b.list, nil
}

// ListBuilderForAPIObjects ...
func ListBuilderForAPIObjects(pvcs *corev1.PersistentVolumeClaimList) *ListBuilder {
	b := &ListBuilder{list: &PVCList{}}
	if pvcs == nil {
		return b
	}
	for _, pvc := range pvcs.Items {
		pvc := pvc
		b.list.items = append(b.list.items, &PVC{Object: &pvc})
	}
	return b
}

// ListBuilderForObjects builds the list of pvc
// instances based on the provided PVC's
func ListBuilderForObjects(pvcs *PVCList) *ListBuilder {
	b := &ListBuilder{}
	if pvcs == nil {
		return b
	}
	b.list = pvcs
	return b
}

// List returns the list of pvc
// instances that was built by this
// builder
func (b *ListBuilder) List() *PVCList {
	if b.filters == nil || len(b.filters) == 0 {
		return b.list
	}
	filtered := &PVCList{}
	for _, pvc := range b.list.items {
		if b.filters.all(pvc) {
			filtered.items = append(filtered.items, pvc)
		}
	}
	return filtered
}

// Len returns the number of items present
// in the PVCList
func (p *PVCList) Len() int {
	return len(p.items)
}

// Len returns the number of items present
// in the PVCList of a builder
func (b *ListBuilder) Len() int {
	p := &PVCList{}
	return len(p.items)
}

// ToAPIList converts PVCList to API PVCList
func (p *PVCList) ToAPIList() *corev1.PersistentVolumeClaimList {
	plist := &corev1.PersistentVolumeClaimList{}
	for _, pvc := range p.items {
		plist.Items = append(plist.Items, *pvc.Object)
	}
	return plist
}

// APIList builds core API PVC list using listbuilder
func (b *ListBuilder) APIList() (*corev1.PersistentVolumeClaimList, error) {
	l, err := b.Build()
	if err != nil {
		return nil, err
	}
	return l.ToAPIList(), nil
}

// NewListBuilder returns an instance of ListBuilder
func NewListBuilder() *ListBuilder {
	return &ListBuilder{list: &PVCList{}}
}

type pvcBuildOption func(*PVC)

// NewForAPIObject returns a new instance of PVC
func NewForAPIObject(obj *corev1.PersistentVolumeClaim, opts ...pvcBuildOption) *PVC {
	p := &PVC{Object: obj}
	for _, o := range opts {
		o(p)
	}
	return p
}

// Predicate defines an abstraction
// to determine conditional checks
// against the provided pvc instance
type Predicate func(*PVC) bool

// IsBound returns true if the pvc is bounded
func (p *PVC) IsBound() bool {
	return p.Object.Status.Phase == corev1.ClaimBound
}

// IsBound is a predicate to filter out pvcs
// which is bounded
func IsBound() Predicate {
	return func(p *PVC) bool {
		return p.IsBound()
	}
}

// IsNil returns true if the PVC instance
// is nil
func (p *PVC) IsNil() bool {
	return p.Object == nil
}

// IsNil is predicate to filter out nil PVC
// instances
func IsNil() Predicate {
	return func(p *PVC) bool {
		return p.IsNil()
	}
}

// ContainsName is filter function to filter pvc's
// based on the name
func ContainsName(name string) Predicate {
	return func(p *PVC) bool {
		return strings.Contains(p.Object.GetName(), name)
	}
}

// PredicateList holds a list of predicate
type PredicateList []Predicate

// all returns true if all the predicates
// succeed against the provided pvc
// instance
func (l PredicateList) all(p *PVC) bool {
	for _, pred := range l {
		if !pred(p) {
			return false
		}
	}
	return true
}

// WithFilter adds filters on which the pvc's has to be filtered
func (b *ListBuilder) WithFilter(pred ...Predicate) *ListBuilder {
	b.filters = append(b.filters, pred...)
	return b
}

// NewPVC returns a PVC instance
func NewPVC() *PVC {
	return &PVC{Object: &corev1.PersistentVolumeClaim{}}
}

// WithName returns a PVC instance with name
func (p *PVC) WithName(name string) *PVC {
	if p.Err != nil {
		return p
	}
	if len(name) == 0 {
		p.Err = errors.New("PVC name shouldn't be empty")
		return p
	}
	p.Object.Name = name
	return p
}

// WithNamespace returns a PVC instance with namespace
func (p *PVC) WithNamespace(namespace string) *PVC {
	if p.Err != nil {
		return p
	}
	if len(namespace) == 0 {
		namespace = "default"
	}
	p.Object.Namespace = namespace
	return p
}

// WithAnnotations returns a PVC instance with annotations
func (p *PVC) WithAnnotations(annotations map[string]string) *PVC {
	if p.Err != nil {
		return p
	}
	p.Object.Annotations = annotations
	return p
}

// WithLabels returns a PVC instance with labels
func (p *PVC) WithLabels(labels map[string]string) *PVC {
	if p.Err != nil {
		return p
	}
	p.Object.Labels = labels
	return p
}

// WithStorageClass returns a PVC instance with storageclass
func (p *PVC) WithStorageClass(scName string) *PVC {
	if p.Err != nil {
		return p
	}
	if len(scName) == 0 {
		p.Err = errors.New("PVC storageclass name shouldn't be empty")
		return p
	}
	p.Object.Spec.StorageClassName = &scName
	return p
}

// WithAccessModes returns a PVC instance with annotations
func (p *PVC) WithAccessModes(accessMode []corev1.PersistentVolumeAccessMode) *PVC {
	if p.Err != nil {
		return p
	}
	if len(accessMode) == 0 {
		p.Err = errors.New("PVC accessMode shouldn't be empty")
		return p
	}
	p.Object.Spec.AccessModes = accessMode
	return p
}

// WithCapacity returns a PVC instance with capacity
func (p *PVC) WithCapacity(capacity string) *PVC {
	if p.Err != nil {
		return p
	}
	resCapacity, err := resource.ParseQuantity(capacity)
	if err != nil {
		p.Err = err
		return p
	}
	resourceList := corev1.ResourceList{
		corev1.ResourceName(corev1.ResourceStorage): resCapacity,
	}
	p.Object.Spec.Resources.Requests = resourceList
	return p
}
