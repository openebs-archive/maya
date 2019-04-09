package v1alpha1

import "k8s.io/api/core/v1"

// node holds the api's node object
type node struct {
	object *v1.Node
}

// node list holds the list of node instances
type nodeList struct {
	items []*node
}

// Predicate defines an abstraction
// to determine conditional checks
// against the provided node instance
type Predicate func(*node) bool

// predicateList holds the list of predicates
type predicateList []Predicate

// listBuilder enables building an instance of nodeList
type listBuilder struct {
	list    *nodeList
	filters predicateList
}

// WithAPIList builds the list of node
// instances based on the provided
// node list
func (b *listBuilder) WithAPIList(nodes *v1.NodeList) *listBuilder {
	if nodes == nil {
		return b
	}
	b.WithAPIObject(nodes.Items...)
	return b
}

func (b *listBuilder) WithObject(nodes ...*node) *listBuilder {
	b.list.items = append(b.list.items, nodes...)
	return b
}

// WithAPIObject builds the list of node instances based on node api instances
func (b *listBuilder) WithAPIObject(nodes ...v1.Node) *listBuilder {
	for _, n := range nodes {
		b.list.items = append(b.list.items, &node{&n})
	}
	return b
}

// List returns the list of node instances that was built by this builder
func (b *listBuilder) List() *nodeList {
	if b.filters == nil && len(b.filters) == 0 {
		return b.list
	}
	filtered := &nodeList{}
	for _, node := range b.list.items {
		if b.filters.all(node) {
			filtered.items = append(filtered.items, node)
		}
	}
	return filtered
}

// ListBuilder returns a instance of listBuilder
func ListBuilder() *listBuilder {
	return &listBuilder{list: &nodeList{items: []*node{}}}
}

// ToAPIList converts nodeList to API nodeList
func (n *nodeList) ToAPIList() *v1.NodeList {
	nlist := &v1.NodeList{}
	for _, node := range n.items {
		nlist.Items = append(nlist.Items, *node.object)
	}
	return nlist
}

// all returns true if all the predicateList
// succeed against the provided node
// instance
func (l predicateList) all(n *node) bool {
	for _, pred := range l {
		if !pred(n) {
			return false
		}
	}
	return true
}

// IsReady retuns true if the node is in running state
func (n *node) IsReady() bool {
	for _, nodeCond := range n.object.Status.Conditions {
		if nodeCond.Reason == "KubeletReady" && nodeCond.Type == v1.NodeReady {
			return true
		}
	}
	return false
}

// IsReady is a Predicate to filter out nodes which are in running state
func IsReady() Predicate {
	return func(n *node) bool {
		return n.IsReady()
	}
}

// WithFilter add filters on which the node has to be filtered
func (b *listBuilder) WithFilter(pred ...Predicate) *listBuilder {
	b.filters = append(b.filters, pred...)
	return b
}

// Len returns the number of items present in the nodeList
func (n *nodeList) Len() int {
	return len(n.items)
}
