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
	//openebsFakeClientset "github.com/openebs/maya/pkg/client/clientset/versioned/fake"
	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	openebsFakeClientset "github.com/openebs/maya/pkg/client/generated/clientset/internalclientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
)

func (focs *clientSet) FakeDiskCreator() {
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

}
func TestNodeDiskAlloter(t *testing.T) {

	// Get a fake openebs client set
	focs := &clientSet{
		oecs: openebsFakeClientset.NewSimpleClientset(),
	}
	focs.FakeDiskCreator()
	tests := map[string]struct {
		// fakeCasPool holds the fake fakeCasPool object in test cases.
		fakeCasPool *v1alpha1.CasPool
		// expectedDiskListLength holds the length of disk list
		expectedDiskListLength int
		// err is a bool , true signifies presence of error and vice-versa
		err bool
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
		},
		// Test Case #2
		"CasPool2": {&v1alpha1.CasPool{
			PoolType: "striped",
			MaxPools: 3,
			Type:     "disk",
		},
			3,
			false,
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
		},
		// Test Case #7
		"CasPool8 of sparse type": {&v1alpha1.CasPool{
			PoolType: "mirrored",
			MaxPools: 6,
			MinPools: 2,
			Type:     "sparse",
		},
			10,
			false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			diskList, err := focs.nodeDiskAlloter(test.fakeCasPool)
			gotErr := false
			if err != nil {
				gotErr = true
			}
			if gotErr != test.err {
				t.Fatalf("Test case failed as the expected error %v but got %v", test.err, gotErr)
			}
			if len(diskList) != test.expectedDiskListLength {
				t.Errorf("Test case failed as the expected disk list length is %d but got %d", test.expectedDiskListLength, len(diskList))
			}
		})
	}
}
