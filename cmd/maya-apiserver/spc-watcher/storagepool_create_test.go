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
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	openebsFakeClientset "github.com/openebs/maya/pkg/client/generated/clientset/internalclientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestNewCasPool(t *testing.T) {
	focs := &clientSet{
		oecs: openebsFakeClientset.NewSimpleClientset(),
	}
	focs.FakeDiskCreator()
	// Make a map of string(key) to struct(value).
	// Key of map describes test case behaviour.
	// Value of map is the test object.
	tests := map[string]struct {
		// fakestoragepoolclaim holds the fake storagepoolcalim object in test cases.
		fakestoragepoolclaim *apis.StoragePoolClaim
		autoProvisioning     bool
	}{
		// TestCase#1
		"SPC for manual provisioning with valid data": {
			autoProvisioning: false,
			fakestoragepoolclaim: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pool1",
					Annotations: map[string]string{
						"cas.openebs.io/create-pool-template": "cstor-pool-create-default-0.7.0",
						"cas.openebs.io/delete-pool-template": "cstor-pool-delete-default-0.7.0",
					},
				},
				Spec: apis.StoragePoolClaimSpec{
					Type: "disk",
					PoolSpec: apis.CStorPoolAttr{
						PoolType: "striped",
					},
					Disks: apis.DiskAttr{
						DiskList: []string{"disk1", "disk2", "disk3"},
					},
				},
			},
		},
		"SPC for auto provisioning with valid data": {
			autoProvisioning: true,
			fakestoragepoolclaim: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pool1",
					Annotations: map[string]string{
						"cas.openebs.io/create-pool-template": "cstor-pool-create-default-0.7.0",
						"cas.openebs.io/delete-pool-template": "cstor-pool-delete-default-0.7.0",
					},
				},
				Spec: apis.StoragePoolClaimSpec{
					MaxPools: 6,
					MinPools: 3,
					Type:     "disk",
					PoolSpec: apis.CStorPoolAttr{
						PoolType: "mirrored",
					},
				},
			},
		},
	}

	// Iterate over whole map to run the test cases.
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// newCasPool is the function under test.
			CasPool, err := focs.NewCasPool(test.fakestoragepoolclaim)
			if err != nil || CasPool == nil {
				t.Errorf("Test case failed as expected nil error but error or CasPool object was nil:%s", name)
			}
		})
	}
}
