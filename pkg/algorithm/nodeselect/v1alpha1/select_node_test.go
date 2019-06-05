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
	"testing"

	"k8s.io/client-go/kubernetes/fake"

	//openebsFakeClientset "github.com/openebs/maya/pkg/client/clientset/versioned/fake"
	"strconv"

	"github.com/golang/glog"
	ndmapis "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	blockdevice "github.com/openebs/maya/pkg/blockdevice/v1alpha1"
	openebsFakeClientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned/fake"
	ndmFakeClientset "github.com/openebs/maya/pkg/client/generated/openebs.io/ndm/v1alpha1/clientset/internalclientset/fake"
	cstorpool "github.com/openebs/maya/pkg/cstorpool/v1alpha1"
	sp "github.com/openebs/maya/pkg/sp/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var blockDeviceK8sClient *blockdevice.KubernetesClient

func FakeDiskCreator(bdc *blockdevice.KubernetesClient) {
	// Create some fake block device objects over nodes.
	// For example, create 14 disk (out of 14 disks, 2 disks are sparse disks)for each of 5 nodes.
	// That meant 14*5 i.e. 70 disk objects should be created

	// diskObjectList will hold the list of disk objects
	var diskObjectList [70]*ndmapis.BlockDevice

	sparseDiskCount := 2
	var key, diskLabel string

	// nodeIdentifer will help in naming a node and attaching multiple disks to a single node.
	nodeIdentifer := 0
	for diskListIndex := 0; diskListIndex < 70; diskListIndex++ {
		diskIdentifier := strconv.Itoa(diskListIndex)
		if diskListIndex%14 == 0 {
			nodeIdentifer++
			sparseDiskCount = 0
		}
		if sparseDiskCount != 2 {
			key = "ndm.io/disk-type"
			diskLabel = "sparse"
			sparseDiskCount++
		} else {
			key = "ndm.io/blockdevice-type"
			diskLabel = "blockdevice"
		}
		diskObjectList[diskListIndex] = &ndmapis.BlockDevice{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name: "blockdevice" + diskIdentifier,
				Labels: map[string]string{
					"kubernetes.io/hostname": "gke-ashu-cstor-default-pool-a4065fd6-vxsh" + strconv.Itoa(nodeIdentifer),
					key:                      diskLabel,
				},
			},
			Status: ndmapis.DeviceStatus{
				State:      DiskStateActive,
				ClaimState: ndmapis.BlockDeviceClaimed,
			},
		}
		_, err := bdc.Create(diskObjectList[diskListIndex])
		if err != nil {
			glog.Error(err)
		}
	}

}
func fakeDiskClient() {
	blockDeviceK8sClient = &blockdevice.KubernetesClient{
		Kubeclientset: fake.NewSimpleClientset(),
		Clientset:     ndmFakeClientset.NewSimpleClientset(),
		Namespace:     "fake-ns",
	}
}
func fakeAlgorithmConfig(spc *v1alpha1.StoragePoolClaim) *Config {
	var bdClient blockdevice.BlockDeviceInterface
	fakeDiskClient()
	FakeDiskCreator(blockDeviceK8sClient)
	if ProvisioningType(spc) == ProvisioningTypeManual {
		bdClient = &blockdevice.SpcObjectClient{
			KubernetesClient: blockDeviceK8sClient,
			Spc:              spc,
		}
	} else {
		bdClient = blockDeviceK8sClient
	}

	cspK8sClient := &cstorpool.KubernetesClient{
		Kubeclientset: fake.NewSimpleClientset(),
		Clientset:     openebsFakeClientset.NewSimpleClientset(),
	}
	spK8sClient := &sp.KubernetesClient{
		Kubeclientset: fake.NewSimpleClientset(),
		Clientset:     openebsFakeClientset.NewSimpleClientset(),
	}
	ac := &Config{
		Spc: spc,

		BlockDeviceClient: bdClient,
		CspClient:         cspK8sClient,
		SpClient:          spK8sClient,
	}

	return ac
}

func TestNodeBlockDeviceAlloter(t *testing.T) {
	tests := map[string]struct {
		// fakeCasPool holds the fake fakeCasPool object in test cases.
		fakeCasPool *v1alpha1.StoragePoolClaim
		// expectedDiskListLength holds the length of disk list
		expectedDiskListLength int
	}{
		// Test Case #1
		"autoSPC1": {&v1alpha1.StoragePoolClaim{
			Spec: v1alpha1.StoragePoolClaimSpec{
				Type: "blockdevice",
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
				Type: "blockdevice",
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
				BlockDevices: v1alpha1.BlockDeviceAttr{
					BlockDeviceList: []string{"blockdevice0", "blockdevice1", "blockdevice2"},
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
				BlockDevices: v1alpha1.BlockDeviceAttr{
					BlockDeviceList: []string{"blockdevice1", "blockdevice2"},
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
				BlockDevices: v1alpha1.BlockDeviceAttr{
					BlockDeviceList: []string{"blockdevice1", "blockdevice7"},
				},
			},
		},
			0,
		},
		// Test Case #8
		"manualSPC8": {&v1alpha1.StoragePoolClaim{
			Spec: v1alpha1.StoragePoolClaimSpec{
				Type: "blockdevice",
				PoolSpec: v1alpha1.CStorPoolAttr{
					PoolType: "mirrored",
				},
				BlockDevices: v1alpha1.BlockDeviceAttr{
					BlockDeviceList: []string{"blockdevice2", "blockdevice3", "blockdevice4", "blockdevice5"},
				},
			},
		},
			4,
		},
		// Test Case #8
		"manualSPC9": {&v1alpha1.StoragePoolClaim{
			Spec: v1alpha1.StoragePoolClaimSpec{
				Type: "blockdevice",
				PoolSpec: v1alpha1.CStorPoolAttr{
					PoolType: "mirrored",
				},
				BlockDevices: v1alpha1.BlockDeviceAttr{
					BlockDeviceList: []string{"blockdevice1", "blockdevice2", "blockdevice3"},
				},
			},
		},
			2,
		},
		// Test Case #10
		"manualSPC10Raidz": {&v1alpha1.StoragePoolClaim{
			Spec: v1alpha1.StoragePoolClaimSpec{
				Type: "blockdevice",
				PoolSpec: v1alpha1.CStorPoolAttr{
					PoolType: "raidz",
				},
				BlockDevices: v1alpha1.BlockDeviceAttr{
					BlockDeviceList: []string{"blockdevice2", "blockdevice3", "blockdevice4"},
				},
			},
		},
			3,
		},
		// Test Case #11
		"manualSPC11Raidz": {&v1alpha1.StoragePoolClaim{
			Spec: v1alpha1.StoragePoolClaimSpec{
				Type: "blockdevice",
				PoolSpec: v1alpha1.CStorPoolAttr{
					PoolType: "raidz",
				},
				BlockDevices: v1alpha1.BlockDeviceAttr{
					BlockDeviceList: []string{"blockdevice5", "blockdevice6"},
				},
			},
		},
			0,
		},
		// Test Case #12
		"manualSPC12Raidz": {&v1alpha1.StoragePoolClaim{
			Spec: v1alpha1.StoragePoolClaimSpec{
				Type: "blockdevice",
				PoolSpec: v1alpha1.CStorPoolAttr{
					PoolType: "raidz",
				},
				BlockDevices: v1alpha1.BlockDeviceAttr{
					BlockDeviceList: []string{"blockdevice1", "blockdevice2", "blockdevice3", "blockdevice4"},
				},
			},
		},
			3,
		},
		// Test Case #13
		"manualSPC13Raidz": {&v1alpha1.StoragePoolClaim{
			Spec: v1alpha1.StoragePoolClaimSpec{
				Type: "blockdevice",
				PoolSpec: v1alpha1.CStorPoolAttr{
					PoolType: "raidz",
				},
				BlockDevices: v1alpha1.BlockDeviceAttr{
					BlockDeviceList: []string{"blockdevice1", "blockdevice2", "blockdevice3", "blockdevice4", "blockdevice5"},
				},
			},
		},
			3,
		},
		// Test Case #14
		"manualSPC14Raidz": {&v1alpha1.StoragePoolClaim{
			Spec: v1alpha1.StoragePoolClaimSpec{
				Type: "blockdevice",
				PoolSpec: v1alpha1.CStorPoolAttr{
					PoolType: "raidz",
				},
				BlockDevices: v1alpha1.BlockDeviceAttr{
					BlockDeviceList: []string{"blockdevice2"},
				},
			},
		},
			0,
		},
		// Test Case #15
		"manualSPC15Raidz2": {&v1alpha1.StoragePoolClaim{
			Spec: v1alpha1.StoragePoolClaimSpec{
				Type: "blockdevice",
				PoolSpec: v1alpha1.CStorPoolAttr{
					PoolType: "raidz2",
				},
				BlockDevices: v1alpha1.BlockDeviceAttr{
					BlockDeviceList: []string{"blockdevice1", "blockdevice2", "blockdevice3"},
				},
			},
		},
			0,
		},
		// Test Case #16
		"manualSPC16Raidz2": {&v1alpha1.StoragePoolClaim{
			Spec: v1alpha1.StoragePoolClaimSpec{
				Type: "blockdevice",
				PoolSpec: v1alpha1.CStorPoolAttr{
					PoolType: "raidz2",
				},
				BlockDevices: v1alpha1.BlockDeviceAttr{
					BlockDeviceList: []string{"blockdevice1", "blockdevice2"},
				},
			},
		},
			0,
		},
		// Test Case #17
		"manualSPC17Raidz2": {&v1alpha1.StoragePoolClaim{
			Spec: v1alpha1.StoragePoolClaimSpec{
				Type: "blockdevice",
				PoolSpec: v1alpha1.CStorPoolAttr{
					PoolType: "raidz2",
				},
				BlockDevices: v1alpha1.BlockDeviceAttr{
					BlockDeviceList: []string{"blockdevice1", "blockdevice2", "blockdevice3", "blockdevice4"},
				},
			},
		},
			0,
		},
		// Test Case #18
		"manualSPC18Raidz2": {&v1alpha1.StoragePoolClaim{
			Spec: v1alpha1.StoragePoolClaimSpec{
				Type: "blockdevice",
				PoolSpec: v1alpha1.CStorPoolAttr{
					PoolType: "raidz2",
				},
				BlockDevices: v1alpha1.BlockDeviceAttr{
					BlockDeviceList: []string{"blockdevice1", "blockdevice2", "blockdevice3", "blockdevice4", "blockdevice5", "blockdevice6"},
				},
			},
		},
			0,
		},
		// Test Case #19
		"manualSPC19Raidz2": {&v1alpha1.StoragePoolClaim{
			Spec: v1alpha1.StoragePoolClaimSpec{
				Type: "blockdevice",
				PoolSpec: v1alpha1.CStorPoolAttr{
					PoolType: "raidz2",
				},
				BlockDevices: v1alpha1.BlockDeviceAttr{
					BlockDeviceList: []string{"blockdevice2", "blockdevice3", "blockdevice4", "blockdevice5", "blockdevice6", "blockdevice7", "blockdevice8", "blockdevice9", "blockdevice10", "blockdevice11", "blockdevice12", "blockdevice13"},
				},
			},
		},
			12,
		},
		// Test Case #20
		"manualSPC20Raidz2": {&v1alpha1.StoragePoolClaim{
			Spec: v1alpha1.StoragePoolClaimSpec{
				Type: "blockdevice",
				PoolSpec: v1alpha1.CStorPoolAttr{
					PoolType: "raidz2",
				},
				BlockDevices: v1alpha1.BlockDeviceAttr{
					BlockDeviceList: []string{"blockdevice1", "blockdevice2", "blockdevice3", "blockdevice4", "blockdevice5", "blockdevice6", "blockdevice7"},
				},
			},
		},
			6,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			ac := fakeAlgorithmConfig(test.fakeCasPool)
			blockdeviceList, _ := ac.NodeBlockDeviceSelector()
			if blockdeviceList == nil {
				t.Fatalf("Got nil blockdevice list")
			}
			if len(blockdeviceList.BlockDevices.Items) != test.expectedDiskListLength {
				t.Errorf("Test case failed as the expected blockdevice list length is %d but got %d", test.expectedDiskListLength, len(blockdeviceList.BlockDevices.Items))
			}
		})
	}
}
