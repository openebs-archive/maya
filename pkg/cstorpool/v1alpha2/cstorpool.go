package v1alpha2

import (
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
)

type csp struct {
	// actual cstor pool object
	object *apis.CStorPool
}

// predicate defines an abstraction
// to determine conditional checks
// against the provided csp instance
type predicate func(*csp) bool

type predicateList []predicate

// all returns true if all the predicates
// succeed against the provided csp
// instance
func (l predicateList) all(c *csp) bool {
	for _, pred := range l {
		if !pred(c) {
			return false
		}
	}
	return true
}

// IsNotUID returns true if provided csp
// instance's UID does not match with any
// of the provided UIDs
func IsNotUID(uids ...string) predicate {
	return func(c *csp) bool {
		for _, uid := range uids {
			if uid == string(c.object.GetUID()) {
				return false
			}
		}
		return true
	}
}

type cspList struct {
	// list of cstor pools
	items []*csp
}

// FilterUIDs will filter the csp instances
// if all the predicates succeed against that
// csp. The filtered csp instances' UIDs will
// be returned
func (l *cspList) FilterUIDs(p ...predicate) []string {
	var (
		filtered []string
		plist    predicateList
	)
	plist = append(plist, p...)
	for _, csp := range l.items {
		if plist.all(csp) {
			filtered = append(filtered, string(csp.object.GetUID()))
		}
	}
	return filtered
}

// listBuilder enables building a
// list of csp instances
type listBuilder struct {
	list *cspList
}

// ListBuilder returns a new instance of
// listBuilder object
func ListBuilder() *listBuilder {
	return &listBuilder{list: &cspList{}}
}

// WithUIDs builds a list of cstor pools
// based on the provided pool UIDs
func (b *listBuilder) WithUIDs(poolUIDs ...string) *listBuilder {
	for _, uid := range poolUIDs {
		obj := &apis.CStorPool{}
		obj.SetUID(types.UID(uid))
		item := &csp{object: obj}
		b.list.items = append(b.list.items, item)
	}
	return b
}

// List returns the list of csp
// instances that were built by
// this builder
func (b *listBuilder) List() *cspList {
	return b.list
}
