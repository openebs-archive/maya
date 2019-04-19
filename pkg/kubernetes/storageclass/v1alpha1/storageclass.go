package v1alpha1

import (
	storagev1 "k8s.io/api/storage/v1"
)

// StorageClass holds the api's Storageclass object
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
func (b *ListBuilder) WithAPIList(storageclasses *storagev1.StorageClassList) *ListBuilder {
	if storageclasses == nil {
		return b
	}
	b.WithAPIObject(storageclasses.Items...)
	return b
}

// WithObject builds the list of StorageClass instances based on the provided
// StorageClass list instance
func (b *ListBuilder) WithObject(StorageClasses ...*StorageClass) *ListBuilder {
	b.list.items = append(b.list.items, StorageClasses...)
	return b
}

// WithAPIObject builds the list of node instances based on StorageClass api instances
func (b *ListBuilder) WithAPIObject(StorageClasses ...storagev1.StorageClass) *ListBuilder {
	for _, sc := range StorageClasses {
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
	for _, StorageClass := range b.list.items {
		if b.filters.all(StorageClass) {
			StorageClass := StorageClass // Pin it
			filtered.items = append(filtered.items, StorageClass)
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
	for _, StorageClass := range scl.items {
		StorageClass := StorageClass // Pin it
		sclist.Items = append(sclist.Items, *StorageClass.object)
	}
	return sclist
}

// all returns true if all the predicateList
// succeed against the provided StorageClass
// instance
func (l predicateList) all(n *StorageClass) bool {
	for _, pred := range l {
		if !pred(n) {
			return false
		}
	}
	return true
}

// Len returns the number of items present in the StorageClassList
func (scl *StorageClassList) Len() int {
	return len(scl.items)
}
