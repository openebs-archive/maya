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

package v1alpha1

import (
	"k8s.io/client-go/kubernetes/fake"
	"testing"
	//openebsFakeClientset "github.com/openebs/maya/pkg/client/clientset/versioned/fake"
	"strconv"

	"github.com/golang/glog"
	ndmapis "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	openebsFakeClientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned/fake"
	ndmFakeClientset "github.com/openebs/maya/pkg/client/generated/openebs.io/ndm/v1alpha1/clientset/internalclientset/fake"
	cstorpool "github.com/openebs/maya/pkg/cstorpool/v1alpha1"
	disk "github.com/openebs/maya/pkg/disk/v1alpha1"
	sp "github.com/openebs/maya/pkg/sp/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var diskK8sClient *disk.KubernetesClient

func FakeDiskCreator(dc *disk.KubernetesClient) {
	// Create some fake disk objects over nodes.
	// For example, create 14 disk (out of 14 disks, 2 disks are sparse disks)for each of 5 nodes.
	// That meant 14*5 i.e. 70 disk objects should be created

	// diskObjectList will hold the list of disk objects
	var diskObjectList [70]*ndmapis.Disk

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
		diskObjectList[diskListIndex] = &ndmapis.Disk{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name: "disk" + diskIdentifier,
				Labels: map[string]string{
					"kubernetes.io/hostname": "gke-ashu-cstor-default-pool-a4065fd6-vxsh" + strconv.Itoa(nodeIdentifer),
					"ndm.io/disk-type":       diskLabel,
				},
			},
			Status: ndmapis.DiskStatus{
				State: DiskStateActive,
			},
		}
		_, err := dc.Create(diskObjectList[diskListIndex])
		if err != nil {
			glog.Error(err)
		}
	}

}
func fakeDiskClient() {
	diskK8sClient = &disk.KubernetesClient{
		fake.NewSimpleClientset(),
		ndmFakeClientset.NewSimpleClientset(),
	}
}
func fakeAlgorithmConfig(spc *v1alpha1.StoragePoolClaim) *Config {
	var diskClient disk.DiskInterface
	fakeDiskClient()
	FakeDiskCreator(diskK8sClient)
	if ProvisioningType(spc) == ProvisioningTypeManual {
		diskClient = &disk.SpcObjectClient{
			diskK8sClient,
			spc,
		}
	} else {
		diskClient = diskK8sClient
	}

	cspK8sClient := &cstorpool.KubernetesClient{
		fake.NewSimpleClientset(),
		openebsFakeClientset.NewSimpleClientset(),
	}
	spK8sClient := &sp.KubernetesClient{
		fake.NewSimpleClientset(),
		openebsFakeClientset.NewSimpleClientset(),
	}
	ac := &Config{
		Spc:        spc,
		DiskClient: diskClient,
		CspClient:  cspK8sClient,
		SpClient:   spK8sClient,
	}

	return ac
}
func TestNodeDiskAlloter(t *testing.T) {
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
		//Test Case #5
		"manualSPC5": {&v1alpha1.StoragePoolClaim{
			Spec: v1alpha1.StoragePoolClaimSpec{
				Type: "sparse",
				PoolSpec: v1alpha1.CStorPoolAttr{
					PoolType: "striped",
				},
				Disks: v1alpha1.DiskAttr{
					DiskList: []string{"disk0", "disk1", "disk2"},
				},
			},
		},
			2,
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
			0,
		},
		// Test Case #7
		"manualSPC7": {&v1alpha1.StoragePoolClaim{
			Spec: v1alpha1.StoragePoolClaimSpec{
				Type: "sparse",
				PoolSpec: v1alpha1.CStorPoolAttr{
					PoolType: "mirrored",
				},
				Disks: v1alpha1.DiskAttr{
					DiskList: []string{"disk1", "disk7"},
				},
			},
		},
			0,
		},
		// Test Case #8
		"manualSPC8": {&v1alpha1.StoragePoolClaim{
			Spec: v1alpha1.StoragePoolClaimSpec{
				Type: "disk",
				PoolSpec: v1alpha1.CStorPoolAttr{
					PoolType: "mirrored",
				},
				Disks: v1alpha1.DiskAttr{
					DiskList: []string{"disk2", "disk3", "disk4", "disk5"},
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
					DiskList: []string{"disk2", "disk3", "disk4"},
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
					DiskList: []string{"disk2"},
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
			0,
		},
		// Test Case #19
		"manualSPC19Raidz2": {&v1alpha1.StoragePoolClaim{
			Spec: v1alpha1.StoragePoolClaimSpec{
				Type: "disk",
				PoolSpec: v1alpha1.CStorPoolAttr{
					PoolType: "raidz2",
				},
				Disks: v1alpha1.DiskAttr{
					DiskList: []string{"disk2", "disk3", "disk4", "disk5", "disk6", "disk7", "disk8", "disk9", "disk10", "disk11", "disk12", "disk13"},
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
			ac := fakeAlgorithmConfig(test.fakeCasPool)
			diskList, _ := ac.NodeDiskSelector()
			if diskList == nil {
				t.Fatalf("Got nil disk list")
			}
			if len(diskList.Disks.Items) != test.expectedDiskListLength {
				t.Errorf("Test case failed as the expected disk list length is %d but got %d", test.expectedDiskListLength, len(diskList.Disks.Items))
			}
		})
	}
}
