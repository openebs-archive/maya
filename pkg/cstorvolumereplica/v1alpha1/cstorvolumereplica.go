package v1alpha1

import (
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
)

const (
	cstorPoolUIDLabelKey string = "cstorpool.openebs.io/uid"
)

type cvr struct {
	// actual cstor volume replica
	// object
	object apis.CStorVolumeReplica
}

type cvrList struct {
	// list of cstor volume replicas
	items []cvr
}

// GetPoolUIDs returns a list of cstor pool
// UIDs corresponding to cstor volume replica
// instances
func (l *cvrList) GetPoolUIDs() []string {
	var uids []string
	for _, cvr := range l.items {
		uid := cvr.object.GetLabels()[cstorPoolUIDLabelKey]
		uids = append(uids, uid)
	}
	return uids
}

// listBuilder enables building
// an instance of cvrList
type listBuilder struct {
	list *cvrList
}

// ListBuilder returns a new instance
// of listBuilder
func ListBuilder() *listBuilder {
	return &listBuilder{list: &cvrList{}}
}

// WithListObject builds the list of cvr
// instances based on the provided
// cvr api instances
func (b *listBuilder) WithListObject(list *apis.CStorVolumeReplicaList) *listBuilder {
	if list == nil {
		return b
	}
	for _, c := range list.Items {
		b.list.items = append(b.list.items, cvr{object: c})
	}
	return b
}

// List returns the list of cvr
// instances that was built by this
// builder
func (b *listBuilder) List() *cvrList {
	return b.list
}
