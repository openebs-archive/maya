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
	"strconv"

	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	openebsFakeClientset "github.com/openebs/maya/pkg/client/generated/clientset/internalclientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (focs *clientSet) FakeDiskCreator() {
	// Create some fake disk objects over nodes.
	// For example, create 14 disk (out of 14 disks 2 disks are sparse disks)for each of 5 nodes.
	// That meant 14*5 i.e. 70 disk objects should be created

	// diskObjectList will hold the list of disk objects
	var diskObjectList [70]*v1alpha1.Disk

	sparseDiskCount := 2
	var diskLabel string

	// nodeIdentifer will help in naming a node and attaching multiple disks to a single node.
	nodeIdentifer := 0
	for diskListIndex := 0; diskListIndex < 70; diskListIndex++ {
		diskIdentifier := strconv.Itoa(diskListIndex)
		if diskListIndex%14 == 0 {
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
			Status: v1alpha1.DiskStatus{
				State: DiskStateActive,
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
		fakeCasPool *v1alpha1.StoragePoolClaim
		// expectedDiskListLength holds the length of disk list
		expectedDiskListLength int
	}{
		// Test Case #1
		"autoSPC1": {&v1alpha1.StoragePoolClaim{
			Spec: v1alpha1.StoragePoolClaimSpec{
				Type: "disk",
				PoolSpec: v1alpha1.CStorPoolAttr{
					PoolType: "striped",
				},
			},
		},
			1,
		},
		// Test Case #2
		"autoSPC2": {&v1alpha1.StoragePoolClaim{
			Spec: v1alpha1.StoragePoolClaimSpec{
				Type: "disk",
				PoolSpec: v1alpha1.CStorPoolAttr{
					PoolType: "mirrored",
				},
			},
		},
			2,
		},
		// Test Case #3
		"autoSPC3": {&v1alpha1.StoragePoolClaim{
			Spec: v1alpha1.StoragePoolClaimSpec{
				Type: "sparse",
				PoolSpec: v1alpha1.CStorPoolAttr{
					PoolType: "striped",
				},
			},
		},
			1,
		},
		// Test Case #4
		"autoSPC4": {&v1alpha1.StoragePoolClaim{
			Spec: v1alpha1.StoragePoolClaimSpec{
				Type: "sparse",
				PoolSpec: v1alpha1.CStorPoolAttr{
					PoolType: "mirrored",
				},
			},
		},
			2,
		},
		// Test Case #5
		"manualSPC5": {&v1alpha1.StoragePoolClaim{
			Spec: v1alpha1.StoragePoolClaimSpec{
				Type: "sparse",
				PoolSpec: v1alpha1.CStorPoolAttr{
					PoolType: "striped",
				},
				Disks: v1alpha1.DiskAttr{
					DiskList: []string{"disk1", "disk2", "disk3"},
				},
			},
		},
			3,
		},
		// Test Case #6
		"manualSPC6": {&v1alpha1.StoragePoolClaim{
			Spec: v1alpha1.StoragePoolClaimSpec{
				Type: "sparse",
				PoolSpec: v1alpha1.CStorPoolAttr{
					PoolType: "mirrored",
				},
				Disks: v1alpha1.DiskAttr{
					DiskList: []string{"disk1", "disk2"},
				},
			},
		},
			2,
		},
		// Test Case #7
		"manualSPC7": {&v1alpha1.StoragePoolClaim{
			Spec: v1alpha1.StoragePoolClaimSpec{
				Type: "sparse",
				PoolSpec: v1alpha1.CStorPoolAttr{
					PoolType: "mirrored",
				},
				Disks: v1alpha1.DiskAttr{
					DiskList: []string{"disk1", "disk2"},
				},
			},
		},
			2,
		},
		// Test Case #8
		"manualSPC8": {&v1alpha1.StoragePoolClaim{
			Spec: v1alpha1.StoragePoolClaimSpec{
				Type: "disk",
				PoolSpec: v1alpha1.CStorPoolAttr{
					PoolType: "mirrored",
				},
				Disks: v1alpha1.DiskAttr{
					DiskList: []string{"disk1", "disk2", "disk3", "disk4"},
				},
			},
		},
			4,
		},
		// Test Case #8
		"manualSPC9": {&v1alpha1.StoragePoolClaim{
			Spec: v1alpha1.StoragePoolClaimSpec{
				Type: "disk",
				PoolSpec: v1alpha1.CStorPoolAttr{
					PoolType: "mirrored",
				},
				Disks: v1alpha1.DiskAttr{
					DiskList: []string{"disk1", "disk2", "disk3"},
				},
			},
		},
			2,
		},
		// Test Case #10
		"manualSPC10Raidz": {&v1alpha1.StoragePoolClaim{
			Spec: v1alpha1.StoragePoolClaimSpec{
				Type: "disk",
				PoolSpec: v1alpha1.CStorPoolAttr{
					PoolType: "raidz",
				},
				Disks: v1alpha1.DiskAttr{
					DiskList: []string{"disk1", "disk2", "disk3"},
				},
			},
		},
			3,
		},
		// Test Case #11
		"manualSPC11Raidz": {&v1alpha1.StoragePoolClaim{
			Spec: v1alpha1.StoragePoolClaimSpec{
				Type: "disk",
				PoolSpec: v1alpha1.CStorPoolAttr{
					PoolType: "raidz",
				},
				Disks: v1alpha1.DiskAttr{
					DiskList: []string{"disk5", "disk6"},
				},
			},
		},
			0,
		},
		// Test Case #12
		"manualSPC12Raidz": {&v1alpha1.StoragePoolClaim{
			Spec: v1alpha1.StoragePoolClaimSpec{
				Type: "disk",
				PoolSpec: v1alpha1.CStorPoolAttr{
					PoolType: "raidz",
				},
				Disks: v1alpha1.DiskAttr{
					DiskList: []string{"disk1", "disk2", "disk3", "disk4"},
				},
			},
		},
			3,
		},
		// Test Case #13
		"manualSPC13Raidz": {&v1alpha1.StoragePoolClaim{
			Spec: v1alpha1.StoragePoolClaimSpec{
				Type: "disk",
				PoolSpec: v1alpha1.CStorPoolAttr{
					PoolType: "raidz",
				},
				Disks: v1alpha1.DiskAttr{
					DiskList: []string{"disk1", "disk2", "disk3", "disk4", "disk5"},
				},
			},
		},
			3,
		},
		// Test Case #14
		"manualSPC14Raidz": {&v1alpha1.StoragePoolClaim{
			Spec: v1alpha1.StoragePoolClaimSpec{
				Type: "disk",
				PoolSpec: v1alpha1.CStorPoolAttr{
					PoolType: "raidz",
				},
				Disks: v1alpha1.DiskAttr{
					DiskList: []string{"disk1"},
				},
			},
		},
			0,
		},
		// Test Case #15
		"manualSPC15Raidz2": {&v1alpha1.StoragePoolClaim{
			Spec: v1alpha1.StoragePoolClaimSpec{
				Type: "disk",
				PoolSpec: v1alpha1.CStorPoolAttr{
					PoolType: "raidz2",
				},
				Disks: v1alpha1.DiskAttr{
					DiskList: []string{"disk1", "disk2", "disk3"},
				},
			},
		},
			0,
		},
		// Test Case #16
		"manualSPC16Raidz2": {&v1alpha1.StoragePoolClaim{
			Spec: v1alpha1.StoragePoolClaimSpec{
				Type: "disk",
				PoolSpec: v1alpha1.CStorPoolAttr{
					PoolType: "raidz2",
				},
				Disks: v1alpha1.DiskAttr{
					DiskList: []string{"disk1", "disk2"},
				},
			},
		},
			0,
		},
		// Test Case #17
		"manualSPC17Raidz2": {&v1alpha1.StoragePoolClaim{
			Spec: v1alpha1.StoragePoolClaimSpec{
				Type: "disk",
				PoolSpec: v1alpha1.CStorPoolAttr{
					PoolType: "raidz2",
				},
				Disks: v1alpha1.DiskAttr{
					DiskList: []string{"disk1", "disk2", "disk3", "disk4"},
				},
			},
		},
			0,
		},
		// Test Case #18
		"manualSPC18Raidz2": {&v1alpha1.StoragePoolClaim{
			Spec: v1alpha1.StoragePoolClaimSpec{
				Type: "disk",
				PoolSpec: v1alpha1.CStorPoolAttr{
					PoolType: "raidz2",
				},
				Disks: v1alpha1.DiskAttr{
					DiskList: []string{"disk1", "disk2", "disk3", "disk4", "disk5", "disk6"},
				},
			},
		},
			6,
		},
		// Test Case #19
		"manualSPC19Raidz2": {&v1alpha1.StoragePoolClaim{
			Spec: v1alpha1.StoragePoolClaimSpec{
				Type: "disk",
				PoolSpec: v1alpha1.CStorPoolAttr{
					PoolType: "raidz2",
				},
				Disks: v1alpha1.DiskAttr{
					DiskList: []string{"disk1", "disk2", "disk3", "disk4", "disk5", "disk6", "disk7", "disk8", "disk9", "disk10", "disk11", "disk12"},
				},
			},
		},
			12,
		},
		// Test Case #20
		"manualSPC20Raidz2": {&v1alpha1.StoragePoolClaim{
			Spec: v1alpha1.StoragePoolClaimSpec{
				Type: "disk",
				PoolSpec: v1alpha1.CStorPoolAttr{
					PoolType: "raidz2",
				},
				Disks: v1alpha1.DiskAttr{
					DiskList: []string{"disk1", "disk2", "disk3", "disk4", "disk5", "disk6", "disk7"},
				},
			},
		},
			6,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			diskList, _ := focs.nodeDiskAlloter(test.fakeCasPool)
			if len(diskList.disks.items) != test.expectedDiskListLength {
				t.Errorf("Test case: %v failed as the expected disk list length is %d but got %d", name, test.expectedDiskListLength, len(diskList.disks.items))
			}
		})
	}
}
