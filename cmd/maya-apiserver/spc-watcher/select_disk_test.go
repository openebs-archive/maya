/*
Copyright 2018 The OpenEBS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package spc

import (
	"testing"

	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	openebsFakeClientset "github.com/openebs/maya/pkg/client/generated/clientset/internalclientset/fake"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"strconv"
)

const (
	UnschedulableNode = "unschedulable"
	UnreachableNode   = "unreachable"
)

func (focs *clientSet) FakeDiskCreator(badNodeCount int, createNodeResource bool, badNodeKind string) bool {
	// Create some fake disk objects over nodes.
	// For example, create 6 disk (out of 6 disks 2 disks are sparse disks)for each of 5 nodes.
	// That meant 6*5 i.e. 30 disk objects should be created
	// diskObjectList will hold the list of disk objects
	var diskObjectList [30]*v1alpha1.Disk

	sparseDiskCount := 2
	var diskLabel string

	// nodeIdentifer will help in naming a node and attaching multiple disks to a single node.
	nodeIdentifer := 0
	for diskListIndex := 0; diskListIndex < 30; diskListIndex++ {
		diskIdentifier := strconv.Itoa(diskListIndex)
		if diskListIndex%6 == 0 {
			nodeIdentifer++
			if badNodeCount > 0 {
				focs.FakeNodeCreator("gke-ashu-cstor-default-pool-a4065fd6-vxsh"+strconv.Itoa(nodeIdentifer), true, badNodeKind)
				badNodeCount--
			} else {
				focs.FakeNodeCreator("gke-ashu-cstor-default-pool-a4065fd6-vxsh"+strconv.Itoa(nodeIdentifer), false, badNodeKind)
			}
			sparseDiskCount = 0
		}
		if sparseDiskCount != 2 {
			diskLabel = "sparse"
			sparseDiskCount++
		} else {
			diskLabel = "disk"
		}
		diskObjectList[diskListIndex] = &v1alpha1.Disk{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name: "disk" + diskIdentifier,
				Labels: map[string]string{
					"kubernetes.io/hostname": "gke-ashu-cstor-default-pool-a4065fd6-vxsh" + strconv.Itoa(nodeIdentifer),
					"ndm.io/disk-type":       diskLabel,
				},
			},
		}
		_, err := focs.oecs.OpenebsV1alpha1().Disks().Create(diskObjectList[diskListIndex])
		if err != nil {
			glog.Error(err)
		}
	}
	return true
}

func (focs *clientSet) FakeNodeCreator(hostName string, badNode bool, badNodeKind string) {
	var condition v1.ConditionStatus
	var unschedulable bool
	condition = v1.ConditionTrue
	unschedulable = false
	if badNode {
		if badNodeKind == UnschedulableNode {
			unschedulable = true
			condition = v1.ConditionTrue
		} else {
			condition = v1.ConditionFalse
			unschedulable = false
		}
	}
	nodeObject := &v1.Node{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: hostName,
		},
		Spec: v1.NodeSpec{
			Unschedulable: unschedulable,
		},
		Status: v1.NodeStatus{
			Conditions: []v1.NodeCondition{
				v1.NodeCondition{
					Type:   v1.NodeReady,
					Status: condition,
				},
			},
		},
	}
	_, err := focs.kubeclientset.CoreV1().Nodes().Create(nodeObject)
	if err != nil {
		glog.Error(err)
	}

}
func (focs *clientSet) FakeDiskDeleter() {
	for diskListIndex := 0; diskListIndex < 30; diskListIndex++ {
		diskIdentifier := strconv.Itoa(diskListIndex)
		diskName := "disk" + diskIdentifier
		err := focs.oecs.OpenebsV1alpha1().Disks().Delete(diskName, &metav1.DeleteOptions{})
		if err != nil {
			glog.Error("Cannot delete fake disks:", err)
		}
	}
}
func (focs *clientSet) FakeNodeDeleter() {
	for nodeListIndex := 1; nodeListIndex < 6; nodeListIndex++ {
		nodeIdentifier := strconv.Itoa(nodeListIndex)
		nodeName := "gke-ashu-cstor-default-pool-a4065fd6-vxsh" + nodeIdentifier
		err := focs.kubeclientset.CoreV1().Nodes().Delete(nodeName, &metav1.DeleteOptions{})
		if err != nil {
			glog.Error("Cannot delete fake nodes:", err)
		}
	}
}
func TestNodeDiskAlloter(t *testing.T) {

	// Get a fake openebs client set
	focs := &clientSet{
		oecs:          openebsFakeClientset.NewSimpleClientset(),
		kubeclientset: fake.NewSimpleClientset(),
	}
	//focs.FakeDiskCreator()
	tests := map[string]struct {
		// fakeCasPool holds the fake fakeCasPool object in test cases.
		fakeCasPool *v1alpha1.CasPool
		// expectedDiskListLength holds the length of disk list
		expectedDiskListLength int
		// err is a bool , true signifies presence of error and vice-versa
		err                bool
		badNodeCount       int
		createNodeResource bool
		nodeDisruptionType string
	}{
		// Test Case #1
		"CasPool1": {&v1alpha1.CasPool{
			PoolType: "striped",
			MaxPools: 3,
			MinPools: 3,
			Type:     "disk",
		},
			3,
			false,
			0,
			true,
			"",
		},
		// Test Case #2
		"CasPool2": {&v1alpha1.CasPool{
			PoolType: "striped",
			MaxPools: 3,
			Type:     "disk",
		},
			3,
			false,
			0,
			true,
			"",
		},
		// Test Case #3
		"CasPool3": {&v1alpha1.CasPool{
			PoolType: "mirrored",
			MaxPools: 3,
			MinPools: 3,
			Type:     "disk",
		},
			6,
			false,
			0,
			true,
			"",
		},
		// Test Case #4
		"CasPool4": {&v1alpha1.CasPool{
			PoolType: "mirrored",
			MaxPools: 6,
			MinPools: 6,
			Type:     "disk",
		},
			0,
			true,
			0,
			true,
			"",
		},
		// Test Case #5
		"CasPool5": {&v1alpha1.CasPool{
			PoolType: "striped",
			MaxPools: 6,
			MinPools: 6,
			Type:     "disk",
		},
			0,
			true,
			0,
			true,
			"",
		},
		// Test Case #6
		"CasPool6": {&v1alpha1.CasPool{
			PoolType: "striped",
			MaxPools: 6,
			MinPools: 2,
			Type:     "disk",
		},
			5,
			false,
			0,
			true,
			"",
		},
		// Test Case #7
		"CasPool7 of sparse type": {&v1alpha1.CasPool{
			PoolType: "striped",
			MaxPools: 6,
			MinPools: 2,
			Type:     "sparse",
		},
			5,
			false,
			0,
			true,
			"",
		},
		// Test Case #8
		"CasPool8 of sparse type": {&v1alpha1.CasPool{
			PoolType: "mirrored",
			MaxPools: 6,
			MinPools: 2,
			Type:     "sparse",
		},
			10,
			false,
			0,
			true,
			"",
		},
		//Test Case #9
		"CasPool9": {&v1alpha1.CasPool{
			PoolType: "striped",
			MaxPools: 3,
			MinPools: 3,
			Type:     "disk",
		},
			3,
			false,
			2,
			true,
			UnschedulableNode,
		},
		// Test Case #10
		"CasPool10": {&v1alpha1.CasPool{
			PoolType: "striped",
			MaxPools: 3,
			MinPools: 3,
			Type:     "disk",
		},
			0,
			true,
			3,
			true,
			UnschedulableNode,
		},
		// Test Case #11
		"CasPool11": {&v1alpha1.CasPool{
			PoolType: "striped",
			MaxPools: 3,
			MinPools: 1,
			Type:     "disk",
		},
			2,
			false,
			3,
			true,
			UnschedulableNode,
		},
		// Test Case #12
		"CasPool12": {&v1alpha1.CasPool{
			PoolType: "striped",
			MaxPools: 3,
			MinPools: 3,
			Type:     "disk",
		},
			0,
			true,
			3,
			true,
			UnreachableNode,
		},
		// Test Case #13
		"CasPool13": {&v1alpha1.CasPool{
			PoolType: "striped",
			MaxPools: 3,
			MinPools: 1,
			Type:     "disk",
		},
			2,
			false,
			3,
			true,
			UnreachableNode,
		},
	}
	for name, test := range tests {
		focs.FakeDiskCreator(test.badNodeCount, test.createNodeResource, test.nodeDisruptionType)
		t.Run(name, func(t *testing.T) {
			diskList, err := focs.nodeDiskAlloter(test.fakeCasPool)
			gotErr := false
			if err != nil {
				glog.Error(err)
				gotErr = true
			}
			if gotErr != test.err {
				t.Errorf("Test case failed as the expected error %v but got %v", test.err, gotErr)
			}
			if len(diskList) != test.expectedDiskListLength {
				t.Errorf("Test case failed as the expected disk list length is %d but got %d", test.expectedDiskListLength, len(diskList))
			}
		})
		// Delete all the created node and disk resources
		focs.FakeDiskDeleter()
		focs.FakeNodeDeleter()
	}
}
