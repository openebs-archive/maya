package v1alpha1

import (
	storagev1 "k8s.io/api/storage/v1"
)

// StorageClass is a wrapper over API based
// storage class instance
type StorageClass struct {
	object *storagev1.StorageClass
}

// StorageClassList holds the list of StorageClass instances
type StorageClassList struct {
	items []*StorageClass
}

// Predicate defines an abstraction
// to determine conditional checks
// against the provided StorageClass instance
type Predicate func(*StorageClass) bool

// predicateList holds the list of predicates
type predicateList []Predicate

// ListBuilder enables building an instance of StorageClassList
type ListBuilder struct {
	list    *StorageClassList
	filters predicateList
}

// WithAPIList builds the list of StorageClass
// instances based on the provided
// StorageClass list
func (b *ListBuilder) WithAPIList(scl *storagev1.StorageClassList) *ListBuilder {
	if scl == nil {
		return b
	}
	b.WithAPIObject(scl.Items...)
	return b
}

// WithObject builds the list of StorageClass instances based on the provided
// StorageClass list instance
func (b *ListBuilder) WithObject(scs ...*StorageClass) *ListBuilder {
	b.list.items = append(b.list.items, scs...)
	return b
}

// WithAPIObject builds the list of node instances based on StorageClass api instances
func (b *ListBuilder) WithAPIObject(scs ...storagev1.StorageClass) *ListBuilder {
	for _, sc := range scs {
		sc := sc // Pin it
		b.list.items = append(b.list.items, &StorageClass{&sc})
	}
	return b
}

// WithFilter add filters on which the StorageClass has to be filtered
func (b *ListBuilder) WithFilter(pred ...Predicate) *ListBuilder {
	b.filters = append(b.filters, pred...)
	return b
}

// List returns the list of StorageClass instances that was built by this builder
func (b *ListBuilder) List() *StorageClassList {
	if b.filters == nil && len(b.filters) == 0 {
		return b.list
	}
	filtered := &StorageClassList{}
	for _, sc := range b.list.items {
		if b.filters.all(sc) {
			sc := sc // Pin it
			filtered.items = append(filtered.items, sc)
		}
	}
	return filtered
}

// NewListBuilder returns a instance of ListBuilder
func NewListBuilder() *ListBuilder {
	return &ListBuilder{list: &StorageClassList{items: []*StorageClass{}}}
}

// ToAPIList converts StorageClassList to API StorageClassList
func (scl *StorageClassList) ToAPIList() *storagev1.StorageClassList {
	sclist := &storagev1.StorageClassList{}
	for _, sc := range scl.items {
		sc := sc // Pin it
		sclist.Items = append(sclist.Items, *sc.object)
	}
	return sclist
}

// all returns true if all the predicateList
// succeed against the provided StorageClass
// instance
func (l predicateList) all(sc *StorageClass) bool {
	for _, pred := range l {
		if !pred(sc) {
			return false
		}
	}
	return true
}

// Len returns the number of items present in the StorageClassList
func (scl *StorageClassList) Len() int {
	return len(scl.items)
}
