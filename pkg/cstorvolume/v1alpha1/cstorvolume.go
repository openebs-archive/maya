package v1alpha1

import (
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
)

type labelKey string

const (
	cstorVolumeLabel labelKey = "openebs.io/persistent-volume"
)

// CStorVolume a wrapper for CStorVolume object
type CStorVolume struct {
	// actual cstor volume replica
	// object
	object apis.CStorVolume
}

// CStorVolumeList is a list of cstorvolume objects
type CStorVolumeList struct {
	// list of cstor volumes
	items []CStorVolume
}

// ListBuilder enables building
// an instance of cvList
type ListBuilder struct {
	list    *CStorVolumeList
	filters PredicateList
}

// NewListBuilder returns a new instance
// of listBuilder
func NewListBuilder() *ListBuilder {
	return &ListBuilder{list: &CStorVolumeList{}}
}

// WithAPIList builds the list of cvr
// instances based on the provided
// cvr api instances
func (b *ListBuilder) WithAPIList(list *apis.CStorVolumeList) *ListBuilder {
	if list == nil {
		return b
	}
	for _, c := range list.Items {
		b.list.items = append(b.list.items, CStorVolume{object: c})
	}
	return b
}

// List returns the list of cstorvolume (cv)
// instances that was built by this
// builder
func (b *ListBuilder) List() *CStorVolumeList {
	if b.filters == nil || len(b.filters) == 0 {
		return b.list
	}
	filtered := NewListBuilder().List()
	for _, cv := range b.list.items {
		if b.filters.all(&cv) {
			filtered.items = append(filtered.items, cv)
		}
	}
	return filtered
}

// Len returns the number of items present
// in the CStorVolumeList
func (l *CStorVolumeList) Len() int {
	return len(l.items)
}

// Predicate defines an abstraction
// to determine conditional checks
// against the provided cvr instance
type Predicate func(*CStorVolume) bool

// IsHealthy returns true if the CVR is in
// healthy state
func (p *CStorVolume) IsHealthy() bool {
	return p.object.Status.Phase == "Healthy"
}

// IsHealthy is a predicate to filter out cvrs
// which is healthy
func IsHealthy() Predicate {
	return func(p *CStorVolume) bool {
		return p.IsHealthy()
	}
}

// PredicateList holds a list of predicate
type PredicateList []Predicate

// all returns true if all the predicates
// succeed against the provided cvr
// instance
func (l PredicateList) all(p *CStorVolume) bool {
	for _, pred := range l {
		if !pred(p) {
			return false
		}
	}
	return true
}

// WithFilter adds filters on which the cstorvolume has to be filtered
func (b *ListBuilder) WithFilter(pred ...Predicate) *ListBuilder {
	b.filters = append(b.filters, pred...)
	return b
}