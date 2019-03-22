package v1alpha1

import (
	"strings"

	v1 "k8s.io/api/core/v1"
)

// pvc holds the api's pvc objects
type pvc struct {
	object *v1.PersistentVolumeClaim
}

// pvcList holds the list of pvc instances
type pvcList struct {
	items []*pvc
}

// listBuilder enables building an instance of
// pvclist
type listBuilder struct {
	list    *pvcList
	filters predicateList
}

// WithAPIList builds the list of pvc
// instances based on the provided
// pvc list api instance
func (b *listBuilder) WithAPIList(pvcs *v1.PersistentVolumeClaimList) *listBuilder {
	if pvcs == nil {
		return b
	}
	b.WithAPIObject(pvcs.Items...)
	return b
}

// WithObjects builds the list of pvc
// instances based on the provided
// pvc list instance
func (b *listBuilder) WithObject(pvcs ...*pvc) *listBuilder {
	b.list.items = append(b.list.items, pvcs...)
	return b
}

// WithAPIList builds the list of pvc
// instances based on the provided
// pvc's api instances
func (b *listBuilder) WithAPIObject(pvcs ...v1.PersistentVolumeClaim) *listBuilder {
	if len(pvcs) == 0 {
		return b
	}
	for _, p := range pvcs {
		b.list.items = append(b.list.items, &pvc{&p})
	}
	return b
}

// List returns the list of pvc
// instances that was built by this
// builder
func (b *listBuilder) List() *pvcList {
	if b.filters == nil || len(b.filters) == 0 {
		return b.list
	}
	filtered := ListBuilder().List()
	for _, pvc := range b.list.items {
		if b.filters.all(pvc) {
			filtered.items = append(filtered.items, pvc)
		}
	}
	return filtered
}

// Len returns the number of items present
// in the PVCList
func (p *pvcList) Len() int {
	return len(p.items)
}

// ToAPIList converts PVCList to API PVCList
func (p *pvcList) ToAPIList() *v1.PersistentVolumeClaimList {
	plist := &v1.PersistentVolumeClaimList{}
	for _, pvc := range p.items {
		plist.Items = append(plist.Items, *pvc.object)
	}
	return plist
}

// ListBuilder returns a instance of listBuilder
func ListBuilder() *listBuilder {
	return &listBuilder{list: &pvcList{items: []*pvc{}}}
}

// predicate defines an abstraction
// to determine conditional checks
// against the provided pvc instance
type predicate func(*pvc) bool

// IsBound returns true if the pvc is bounded
func (p *pvc) IsBound() bool {
	return p.object.Status.Phase == "Bound"
}

// IsBound is a predicate to filter out pvcs
// which is bounded
func IsBound() predicate {
	return func(p *pvc) bool {
		return p.IsBound()
	}
}

// IsNil returns true if the PVC instance
// is nil
func (p *pvc) IsNil() bool {
	return p.object == nil
}

// IsNil is predicate to filter out nil PVC
// instances
func IsNil() predicate {
	return func(p *pvc) bool {
		return p.IsNil()
	}
}

// ContainsName is filter function to filter pvc's
// based on the name
func ContainsName(name string) predicate {
	return func(p *pvc) bool {
		return strings.Contains(p.object.GetName(), name)
	}
}

// predicateList holds a list of predicate
type predicateList []predicate

// all returns true if all the predicates
// succeed against the provided pvc
// instance
func (l predicateList) all(p *pvc) bool {
	for _, pred := range l {
		if !pred(p) {
			return false
		}
	}
	return true
}

// WithFilter adds filters on which the pvc's has to be filtered
func (b *listBuilder) WithFilter(pred ...predicate) *listBuilder {
	b.filters = append(b.filters, pred...)
	return b
}
