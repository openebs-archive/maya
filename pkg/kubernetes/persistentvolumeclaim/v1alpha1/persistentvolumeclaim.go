package v1alpha1

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// pvc holds the api's pvc objects
type PVC struct {
	object *corev1.PersistentVolumeClaim
}

// pvcList holds the list of pvc instances
type PVCList struct {
	items []*PVC
}

// listBuilder enables building an instance of
// pvclist
type ListBuilder struct {
	list    *PVCList
	filters PredicateList
}

// WithAPIList builds the list of pvc
// instances based on the provided
// pvc list api instance
func (b *ListBuilder) WithAPIList(pvcs *corev1.PersistentVolumeClaimList) *ListBuilder {
	if pvcs == nil {
		return b
	}
	b.WithAPIObject(pvcs.Items...)
	return b
}

// WithObjects builds the list of pvc
// instances based on the provided
// pvc list instance
func (b *ListBuilder) WithObject(pvcs ...*PVC) *ListBuilder {
	b.list.items = append(b.list.items, pvcs...)
	return b
}

// WithAPIList builds the list of pvc
// instances based on the provided
// pvc's api instances
func (b *ListBuilder) WithAPIObject(pvcs ...corev1.PersistentVolumeClaim) *ListBuilder {
	if len(pvcs) == 0 {
		return b
	}
	for _, p := range pvcs {
		b.list.items = append(b.list.items, &PVC{&p})
	}
	return b
}

// List returns the list of pvc
// instances that was built by this
// builder
func (b *ListBuilder) List() *PVCList {
	if b.filters == nil || len(b.filters) == 0 {
		return b.list
	}
	filtered := NewListBuilder().List()
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

// ToAPIList converts PVCList to API PVCList
func (p *PVCList) ToAPIList() *corev1.PersistentVolumeClaimList {
	plist := &corev1.PersistentVolumeClaimList{}
	for _, pvc := range p.items {
		plist.Items = append(plist.Items, *pvc.object)
	}
	return plist
}

// ListBuilder returns a instance of listBuilder
func NewListBuilder() *ListBuilder {
	return &ListBuilder{list: &PVCList{items: []*PVC{}}}
}

// NewForAPIObject returns a new instance of Builder
func NewForAPIObject(obj *corev1.PersistentVolumeClaim) *PVC {
	return &PVC{
		object: obj,
	}
}

// predicate defines an abstraction
// to determine conditional checks
// against the provided pvc instance
type Predicate func(*PVC) bool

// IsBound returns true if the pvc is bounded
func (p *PVC) IsBound() bool {
	return p.object.Status.Phase == "Bound"
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
	return p.object == nil
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
		return strings.Contains(p.object.GetName(), name)
	}
}

// predicateList holds a list of predicate
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
