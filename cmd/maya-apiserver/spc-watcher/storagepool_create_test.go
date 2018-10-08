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
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

func TestNewCasPool(t *testing.T) {
	focs := &clientSet{
		oecs:          openebsFakeClientset.NewSimpleClientset(),
		kubeclientset: fake.NewSimpleClientset(),
	}
	// Make a map of string(key) to struct(value).
	// Key of map describes test case behaviour.
	// Value of map is the test object.
	tests := map[string]struct {
		// fakestoragepoolclaim holds the fake storagepoolcalim object in test cases.
		fakestoragepoolclaim *apis.StoragePoolClaim
		reSync               bool
		pendingPoolCount     int
		invalidDataInjection bool
		autoProvisioning     bool
	}{
		// TestCase#1
		"SPC for manual provisioning with valid data": {
			invalidDataInjection: false,
			autoProvisioning:     false,
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
					Disks: apis.DiskAttr{
						DiskList: []string{"dummy-disk1", "dummy-disk2", "dummy-disk3"},
					},
				},
			},
			reSync:           false,
			pendingPoolCount: 1,
		},
		"SPC for auto provisioning with valid data": {
			invalidDataInjection: false,
			autoProvisioning:     true,
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
			reSync:           false,
			pendingPoolCount: 1,
		},
		// TestCase#2
		"SPC for manual provisioning with invalid data#1": {
			invalidDataInjection: true,
			autoProvisioning:     false,
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
					Type:     "unknown", // Inject invalid data
					PoolSpec: apis.CStorPoolAttr{
						PoolType: "mirrored",
					},
					Disks: apis.DiskAttr{
						DiskList: []string{"dummy-disk1", "dummy-disk2", "dummy-disk3"},
					},
				},
			},
			reSync:           false,
			pendingPoolCount: 1,
		},
		// TestCase#3
		"SPC for manual provisioning with invalid data#2": {
			invalidDataInjection: true,
			autoProvisioning:     false,
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
					Type:     "sparse",
					PoolSpec: apis.CStorPoolAttr{ // Inject invalid data
						PoolType: "unknown",
					},
					Disks: apis.DiskAttr{
						DiskList: []string{"dummy-disk1", "dummy-disk2", "dummy-disk3"},
					},
				},
			},
			reSync:           false,
			pendingPoolCount: 1,
		},
		// TestCase#4
		"SPC for manual provisioning with invalid data#3": {
			invalidDataInjection: true,
			autoProvisioning:     false,
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
					Type:     "sparse",
					PoolSpec: apis.CStorPoolAttr{ // Inject invalid data
					},
					Disks: apis.DiskAttr{
						DiskList: []string{"dummy-disk1", "dummy-disk2", "dummy-disk3"},
					},
				},
			},
			reSync:           false,
			pendingPoolCount: 1,
		},
	}

	// Iterate over whole map to run the test cases.
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// newCasPool is the function under test.
			if test.autoProvisioning {
				// Get a fake openebs client set
				focs.FakeDiskCreator(0, false, "")
			}
			CasPool, err := focs.newCasPool(test.fakestoragepoolclaim, test.reSync, test.pendingPoolCount)
			if test.invalidDataInjection {
				if err == nil || CasPool != nil {
					t.Errorf("Test case failed as expected expected error but got nil")
				}
			} else {
				if err != nil || CasPool == nil {
					t.Errorf("Test case failed as expected nill error but error or CasPool object was nil")
				}
			}
		})
	}
}
