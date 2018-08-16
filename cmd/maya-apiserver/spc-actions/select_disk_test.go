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

package storagepoolactions

import (
	"testing"
	//openebsFakeClientset "github.com/openebs/maya/pkg/client/clientset/versioned/fake"
	openebsFakeClientset "github.com/openebs/maya/pkg/client/generated/clientset/internalclientset/fake"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
	"github.com/golang/glog"
)

func TestNodeDiskAlloter(t *testing.T) {

	// Get a fake openebs client set
	focs := &clientSet{
		oecs: openebsFakeClientset.NewSimpleClientset(),
	}

	// Create some fake disk objects over nodes.
	// For example, create 4 disk for each of 5 nodes.
	// That meant 4*5 i.e. 20 disk objects should be created

	// diskObjectList will hold the list of disk objects
	var diskObjectList [20]*v1alpha1.Disk

	// nodeIdentifer will help in naming a node and attaching multiple disks to a single node.
	nodeIdentifer := 0
	for diskListIndex := 0; diskListIndex < 20; diskListIndex++ {
		diskIdentifier := strconv.Itoa(diskListIndex)
		if diskListIndex%4 == 0 {
			nodeIdentifer ++
		}
		diskObjectList[diskListIndex] = &v1alpha1.Disk{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name: "disk" + diskIdentifier,
				Labels: map[string]string{
					"kubernetes.io/hostname": "gke-ashu-cstor-default-pool-a4065fd6-vxsh" + strconv.Itoa(nodeIdentifer),
				},
			},
		}
		_, err := focs.oecs.OpenebsV1alpha1().Disks().Create(diskObjectList[diskListIndex])
		if err != nil {
			glog.Error(err)
		}
	}
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
		},
			3,
			false,
		},
		// Test Case #2
		"CasPool2": {&v1alpha1.CasPool{
			PoolType: "striped",
			MaxPools: 3,
		},
			3,
			false,
		},
		// Test Case #3
		"CasPool3": {&v1alpha1.CasPool{
			PoolType: "mirrored",
			MaxPools: 3,
			MinPools: 3,
		},
			6,
			false,
		},
		// Test Case #4
		"CasPool4": {&v1alpha1.CasPool{
			PoolType: "mirrored",
			MaxPools: 6,
			MinPools: 6,
		},
			0,
			true,
		},
		// Test Case #5
		"CasPool5": {&v1alpha1.CasPool{
			PoolType: "striped",
			MaxPools: 6,
			MinPools: 6,
		},
			0,
			true,
		},
		// Test Case #6
		"CasPool6": {&v1alpha1.CasPool{
			PoolType: "striped",
			MaxPools: 6,
			MinPools: 2,
		},
			5,
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
