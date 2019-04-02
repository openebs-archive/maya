package v1alpha1

import (
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
)

type labelKey string

const (
	cstorVolumeLabel labelKey = "openebs.io/persistent-volume"
)

type cv struct {
	// actual cstor volume replica
	// object
	object apis.CStorVolume
}

type cvList struct {
	// list of cstor volumes
	items []cv
}

// listBuilder enables building
// an instance of cvList
type listBuilder struct {
	list    *cvList
	filters predicateList
}

// ListBuilder returns a new instance
// of listBuilder
func ListBuilder() *listBuilder {
	return &listBuilder{list: &cvList{}}
}

// WithAPIList builds the list of cvr
// instances based on the provided
// cvr api instances
func (b *listBuilder) WithAPIList(list *apis.CStorVolumeList) *listBuilder {
	if list == nil {
		return b
	}
	for _, c := range list.Items {
		b.list.items = append(b.list.items, cv{object: c})
	}
	return b
}

// List returns the list of cstorvolume (cv)
// instances that was built by this
// builder
func (b *listBuilder) List() *cvList {
	if b.filters == nil || len(b.filters) == 0 {
		return b.list
	}
	filtered := ListBuilder().List()
	for _, cv := range b.list.items {
		if b.filters.all(&cv) {
			filtered.items = append(filtered.items, cv)
		}
	}
	return filtered
}

// Len returns the number of items present
// in the CStorVolumeList
func (l *cvList) Len() int {
	return len(l.items)
}

// predicate defines an abstraction
// to determine conditional checks
// against the provided cvr instance
type predicate func(*cv) bool

// IsHealthy returns true if the CVR is in
// healthy state
func (p *cv) IsHealthy() bool {
	return p.object.Status.Phase == "Healthy"
}

// IsHealthy is a predicate to filter out cvrs
// which is healthy
func IsHealthy() predicate {
	return func(p *cv) bool {
		return p.IsHealthy()
	}
}

// predicateList holds a list of predicate
type predicateList []predicate

// all returns true if all the predicates
// succeed against the provided cvr
// instance
func (l predicateList) all(p *cv) bool {
	for _, pred := range l {
		if !pred(p) {
			return false
		}
	}
	return true
}

// WithFilter adds filters on which the cvr's has to be filtered
func (b *listBuilder) WithFilter(pred ...predicate) *listBuilder {
	b.filters = append(b.filters, pred...)
	return b
}
