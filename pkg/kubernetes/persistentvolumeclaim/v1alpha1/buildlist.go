package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

// ListBuilder enables building an instance of
// PVClist
type ListBuilder struct {
	list    *PVCList
	filters PredicateList
}

// NewListBuilder returns an instance of ListBuilder
func NewListBuilder() *ListBuilder {
	return &ListBuilder{list: &PVCList{}}
}

// Build returns the final instance of patch
// TODO add validations and error checks
func (b *ListBuilder) Build() (*PVCList, error) {
	return b.list, nil
}

// ListBuilderForAPIObjects builds the ListBuilder object based on PVC api list
func ListBuilderForAPIObjects(pvcs *corev1.PersistentVolumeClaimList) *ListBuilder {
	b := &ListBuilder{list: &PVCList{}}
	if pvcs == nil {
		return b
	}
	for _, pvc := range pvcs.Items {
		pvc := pvc
		b.list.items = append(b.list.items, &PVC{object: &pvc})
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
// in the PVCList of a builder
func (b *ListBuilder) Len() int {
	p := &PVCList{}
	return len(p.items)
}

// APIList builds core API PVC list using listbuilder
func (b *ListBuilder) APIList() (*corev1.PersistentVolumeClaimList, error) {
	l, err := b.Build()
	if err != nil {
		return nil, err
	}
	return l.ToAPIList(), nil
}
