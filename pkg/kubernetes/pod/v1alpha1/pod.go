package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
)

// pod holds the api's pod objects
type pod struct {
	object *v1.Pod
}

// podList holds the list of pod instances
type podList struct {
	items []*pod
}

// ListBuilder enables building an instance of
// podlist
type ListBuilder struct {
	list    *podList
	filters predicateList
}

// WithAPIList builds the list of pod
// instances based on the provided
// pod list api instance
func (b *ListBuilder) WithAPIList(pods *v1.PodList) *ListBuilder {
	if pods == nil {
		return b
	}
	b.WithAPIObject(pods.Items...)
	return b
}

// WithObjects builds the list of pod
// instances based on the provided
// pod list instance
func (b *ListBuilder) WithObject(pods ...*pod) *ListBuilder {
	b.list.items = append(b.list.items, pods...)
	return b
}

// WithAPIList builds the list of pod
// instances based on the provided
// pod api instances
func (b *ListBuilder) WithAPIObject(pods ...v1.Pod) *ListBuilder {
	for _, p := range pods {
		p := p //pin it
		b.list.items = append(b.list.items, &pod{&p})
	}
	return b
}

// List returns the list of pod
// instances that was built by this
// builder
func (b *ListBuilder) List() *podList {
	if b.filters == nil && len(b.filters) == 0 {
		return b.list
	}
	filtered := &podList{}
	for _, pod := range b.list.items {
		if b.filters.all(pod) {
			filtered.items = append(filtered.items, pod)
		}
	}
	return filtered
}

// Len returns the number of items present in the podList
func (p *podList) Len() int {
	return len(p.items)
}

// ToAPIList converts podList to API podList
func (p *podList) ToAPIList() *v1.PodList {
	plist := &v1.PodList{}
	for _, pod := range p.items {
		plist.Items = append(plist.Items, *pod.object)
	}
	return plist
}

// ListBuilder returns a instance of ListBuilder
func NewListBuilder() *ListBuilder {
	return &ListBuilder{list: &podList{items: []*pod{}}}
}

func ListBuilderForAPIList(pods *v1.PodList) *ListBuilder {
	b := &ListBuilder{list: &podList{}}
	if pods == nil {
		return b
	}
	for _, p := range pods.Items {
		p := p
		b.list.items = append(b.list.items, &pod{object: &p})
	}
	return b
}

// predicate defines an abstraction
// to determine conditional checks
// against the provided pod instance
type predicate func(*pod) bool

// IsRunning retuns true if the pod is in running
// state
func (p *pod) IsRunning() bool {
	return p.object.Status.Phase == "Running"
}

// IsRunning is a predicate to filter out pods
// which in running state
func IsRunning() predicate {
	return func(p *pod) bool {
		return p.IsRunning()
	}
}

// IsNil returns true if the pod instance
// is nil
func (p *pod) IsNil() bool {
	return p.object == nil
}

// IsNil is predicate to filter out nil pod
// instances
func IsNil() predicate {
	return func(p *pod) bool {
		return p.IsNil()
	}
}

// predicateList holds a list of predicate
type predicateList []predicate

// all returns true if all the predicates
// succeed against the provided pod
// instance
func (l predicateList) all(p *pod) bool {
	for _, pred := range l {
		if !pred(p) {
			return false
		}
	}
	return true
}

// WithFilter add filters on which the pod
// has to be filtered
func (b *ListBuilder) WithFilter(pred ...predicate) *ListBuilder {
	b.filters = append(b.filters, pred...)
	return b
}
