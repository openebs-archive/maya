package v1alpha1

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func fakeEmptyNodeList(noOfNodes int) *corev1.NodeList {
	list := &corev1.NodeList{}
	for i := 0; i < noOfNodes; i++ {
		node := corev1.Node{}
		list.Items = append(list.Items, node)
	}
	return list
}

func fakeAPINodeList(nodeNames []string) *corev1.NodeList {
	if len(nodeNames) == 0 {
		return nil
	}
	list := &corev1.NodeList{}
	for _, name := range nodeNames {
		node := corev1.Node{}
		node.SetName(name)
		list.Items = append(list.Items, node)
	}
	return list
}

func fakeAPINodeListFromNameStatusMap(nodes map[string]corev1.NodeConditionType) []*node {
	nlist := []*node{}
	for k := range nodes {
		n := &corev1.Node{}
		fakeNodeCondition := corev1.NodeCondition{
			Reason: kubeletReady,
			Type:   nodes[k],
		}
		n.SetName(k)
		n.Status.Conditions = append(n.Status.Conditions, fakeNodeCondition)
		nlist = append(nlist, &node{n})
	}
	return nlist
}

func TestListBuilderFuncWithAPIList(t *testing.T) {
	tests := map[string]struct {
		availableNodes  []string
		expectedNodeLen int
	}{
		"Node set 1":  {[]string{}, 0},
		"Node set 2":  {[]string{"node1"}, 1},
		"Node set 3":  {[]string{"node1", "node2"}, 2},
		"Node set 4":  {[]string{"node1", "node2", "node3"}, 3},
		"Node set 5":  {[]string{"node1", "node2", "node3", "node4"}, 4},
		"Node set 6":  {[]string{"node1", "node2", "node3", "node4", "node5"}, 5},
		"Node set 7":  {[]string{"node1", "node2", "node3", "node4", "node5", "node6"}, 6},
		"Node set 8":  {[]string{"node1", "node2", "node3", "node4", "node5", "node6", "node7"}, 7},
		"Node set 9":  {[]string{"node1", "node2", "node3", "node4", "node5", "node6", "node7", "node8"}, 8},
		"Node set 10": {[]string{"node1", "node2", "node3", "node4", "node5", "node6", "node7", "node8", "node9"}, 9},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			b := ListBuilderFunc().WithAPIList(fakeAPINodeList(mock.availableNodes))
			if mock.expectedNodeLen != len(b.list.items) {
				t.Fatalf("Test %v failed: expected %v got %v", name, mock.expectedNodeLen, len(b.list.items))
			}
		})
	}
}

func TestListBuilderFuncWithEmptyNodeList(t *testing.T) {
	tests := map[string]struct {
		nodeCount, expectedNodeLen int
	}{
		"Two nodes": {5, 5},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			b := ListBuilderFunc().WithAPIList(fakeEmptyNodeList(mock.nodeCount))
			if mock.expectedNodeLen != len(b.list.items) {
				t.Fatalf("Test %v failed: expected %v got %v", name, mock.expectedNodeLen, len(b.list.items))
			}
		})
	}
}

func TestListBuilderWithAPIObjects(t *testing.T) {
	tests := map[string]struct {
		availableNodes  []string
		expectedNodeLen int
	}{
		"Node set 2":  {[]string{"node1"}, 1},
		"Node set 3":  {[]string{"node1", "node2"}, 2},
		"Node set 4":  {[]string{"node1", "node2", "node3"}, 3},
		"Node set 5":  {[]string{"node1", "node2", "node3", "node4"}, 4},
		"Node set 6":  {[]string{"node1", "node2", "node3", "node4", "node5"}, 5},
		"Node set 7":  {[]string{"node1", "node2", "node3", "node4", "node5", "node6"}, 6},
		"Node set 8":  {[]string{"node1", "node2", "node3", "node4", "node5", "node6", "node7"}, 7},
		"Node set 9":  {[]string{"node1", "node2", "node3", "node4", "node5", "node6", "node7", "node8"}, 8},
		"Node set 10": {[]string{"node1", "node2", "node3", "node4", "node5", "node6", "node7", "node8", "node9"}, 9},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			b := ListBuilderFunc().WithAPIObject(fakeAPINodeList(mock.availableNodes).Items...)
			if mock.expectedNodeLen != len(b.list.items) {
				t.Fatalf("Test %v failed: expected %v got %v", name, mock.expectedNodeLen, len(b.list.items))
			}
		})
	}
}

func TestListBuilderToAPIList(t *testing.T) {
	tests := map[string]struct {
		availableNodes  []string
		expectedNodeLen int
	}{
		"Node set 1":  {[]string{}, 0},
		"Node set 2":  {[]string{"node1"}, 1},
		"Node set 3":  {[]string{"node1", "node2"}, 2},
		"Node set 4":  {[]string{"node1", "node2", "node3"}, 3},
		"Node set 5":  {[]string{"node1", "node2", "node3", "node4"}, 4},
		"Node set 6":  {[]string{"node1", "node2", "node3", "node4", "node5"}, 5},
		"Node set 7":  {[]string{"node1", "node2", "node3", "node4", "node5", "node6"}, 6},
		"Node set 8":  {[]string{"node1", "node2", "node3", "node4", "node5", "node6", "node7"}, 7},
		"Node set 9":  {[]string{"node1", "node2", "node3", "node4", "node5", "node6", "node7", "node8"}, 8},
		"Node set 10": {[]string{"node1", "node2", "node3", "node4", "node5", "node6", "node7", "node8", "node9"}, 9},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			b := ListBuilderFunc().WithAPIList(fakeAPINodeList(mock.availableNodes)).List().ToAPIList()
			if mock.expectedNodeLen != len(b.Items) {
				t.Fatalf("Test %v failed: expected %v got %v", name, mock.expectedNodeLen, len(b.Items))
			}
		})
	}
}

func TestFilterList(t *testing.T) {
	tests := map[string]struct {
		availableNodes map[string]corev1.NodeConditionType
		filteredNodes  []string
		filters        predicateList
	}{
		"Nodes Set 1": {
			availableNodes: map[string]corev1.NodeConditionType{"Node 1": corev1.NodeReady, "Node 2": corev1.NodeOutOfDisk},
			filteredNodes:  []string{"Node 1"},
			filters:        predicateList{IsReady()},
		},
		"Nodes Set 2": {
			availableNodes: map[string]corev1.NodeConditionType{"Node 1": corev1.NodeReady, "Node 2": corev1.NodeReady},
			filteredNodes:  []string{"Node 1", "Node 2"},
			filters:        predicateList{IsReady()},
		},
		"Nodes Set 3": {
			availableNodes: map[string]corev1.NodeConditionType{"Node 1": corev1.NodeDiskPressure, "Node 2": corev1.NodeMemoryPressure},
			filteredNodes:  []string{},
			filters:        predicateList{IsReady()},
		},
	}
	for name, mock := range tests {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			list := ListBuilderFunc().WithObject(fakeAPINodeListFromNameStatusMap(mock.availableNodes)...).WithFilter(mock.filters...).List()
			if len(list.items) != len(mock.filteredNodes) {
				t.Fatalf("Test %v failed: expected %v got %v", name, len(mock.filteredNodes), len(list.items))
			}
		})
	}
}
