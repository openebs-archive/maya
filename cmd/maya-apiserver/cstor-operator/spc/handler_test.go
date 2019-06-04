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
	"time"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	openebsFakeClientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned/fake"
	informers "github.com/openebs/maya/pkg/client/generated/informers/externalversions"
	ndmFakeClientset "github.com/openebs/maya/pkg/client/generated/openebs.io/ndm/v1alpha1/clientset/internalclientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"testing"
)

func TestValidatePoolType(t *testing.T) {
	tests := map[string]struct {
		spc           *apis.StoragePoolClaim
		expectedError bool
	}{
		"Empty pool type": {
			spc: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-pool-claim-1",
				},
				Spec: apis.StoragePoolClaimSpec{
					PoolSpec: apis.CStorPoolAttr{
						PoolType: "",
					},
				},
			},
			expectedError: true,
		},
		"Wrong pool type": {
			spc: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-pool-claim-1",
				},
				Spec: apis.StoragePoolClaimSpec{
					PoolSpec: apis.CStorPoolAttr{
						PoolType: "test",
					},
				},
			},
			expectedError: true,
		},
		"Striped pool type": {
			spc: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-pool-claim-1",
				},
				Spec: apis.StoragePoolClaimSpec{
					PoolSpec: apis.CStorPoolAttr{
						PoolType: string(apis.PoolTypeStripedCPV),
					},
				},
			},
			expectedError: false,
		},
		"Mirrored pool type": {
			spc: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-pool-claim-1",
				},
				Spec: apis.StoragePoolClaimSpec{
					PoolSpec: apis.CStorPoolAttr{
						PoolType: string(apis.PoolTypeMirroredCPV),
					},
				},
			},
			expectedError: false,
		},
		"Raidz pool type": {
			spc: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-pool-claim-1",
				},
				Spec: apis.StoragePoolClaimSpec{
					PoolSpec: apis.CStorPoolAttr{
						PoolType: string(apis.PoolTypeRaidzCPV),
					},
				},
			},
			expectedError: false,
		},
		"Raidz2 pool type": {
			spc: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-pool-claim-1",
				},
				Spec: apis.StoragePoolClaimSpec{
					PoolSpec: apis.CStorPoolAttr{
						PoolType: string(apis.PoolTypeRaidz2CPV),
					},
				},
			},
			expectedError: false,
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			err := validatePoolType(test.spc)
			var gotError bool
			if err != nil {
				gotError = true
			}
			if gotError != test.expectedError {
				t.Errorf("Test case failed as expected error %v but got error %v", test.expectedError, gotError)
			}
		})
	}
}

func TestValidateBlockDeviceType(t *testing.T) {
	tests := map[string]struct {
		spc           *apis.StoragePoolClaim
		expectedError bool
	}{
		"Sparse pool type": {
			spc: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-pool-claim-1",
				},
				Spec: apis.StoragePoolClaimSpec{
					Type: string(apis.TypeSparseCPV),
				},
			},
			expectedError: true,
		},
		"Disk pool type": {
			spc: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-pool-claim-1",
				},
				Spec: apis.StoragePoolClaimSpec{
					Type: string(apis.TypeDiskCPV),
				},
			},
			expectedError: true,
		},
		"block device pool type": {
			spc: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-pool-claim-2",
				},
				Spec: apis.StoragePoolClaimSpec{
					Type: string(apis.TypeBlockDeviceCPV),
				},
			},
			expectedError: false,
		},
		"Empty pool type": {
			spc: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-pool-claim-1",
				},
				Spec: apis.StoragePoolClaimSpec{
					Type: "",
				},
			},
			expectedError: true,
		},
		"Invalid pool type": {
			spc: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-pool-claim-1",
				},
				Spec: apis.StoragePoolClaimSpec{
					Type: "gpd",
				},
			},
			expectedError: true,
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			err := validateDiskType(test.spc)
			var gotError bool
			if err != nil {
				gotError = true
			}
			if gotError != test.expectedError {
				t.Errorf("Test case failed as expected error %v but got error %v", test.expectedError, gotError)
			}
		})
	}
}

func TestValidateAutoSpcMaxPool(t *testing.T) {
	tests := map[string]struct {
		spc           *apis.StoragePoolClaim
		expectedError bool
	}{
		"Maxpool not specified on spc": {
			spc: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-pool-claim-1",
				},
				Spec: apis.StoragePoolClaimSpec{
					Type: string(apis.TypeSparseCPV),
				},
			},
			expectedError: true,
		},
		"Wrong maxpool specified on spc": {
			spc: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-pool-claim-1",
				},
				Spec: apis.StoragePoolClaimSpec{
					Type:     string(apis.TypeSparseCPV),
					MaxPools: newInt(-1),
				},
			},
			expectedError: true,
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			err := validateAutoSpcMaxPool(test.spc)
			var gotError bool
			if err != nil {
				gotError = true
			}
			if gotError != test.expectedError {
				t.Errorf("Test case failed as expected error %v but got error %v", test.expectedError, gotError)
			}
		})
	}
}

func TestValidateSpc(t *testing.T) {
	tests := map[string]struct {
		spc           *apis.StoragePoolClaim
		expectedError bool
	}{
		"Invalid SPC #1": {
			spc: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-pool-claim-1",
				},
				Spec: apis.StoragePoolClaimSpec{
					Type: string(apis.TypeSparseCPV),
				},
			},
			expectedError: true,
		},
		"Invalid SPC #2": {
			spc: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-pool-claim-1",
				},
				Spec: apis.StoragePoolClaimSpec{
					Type:     string(apis.TypeSparseCPV),
					MaxPools: newInt(-1),
				},
			},
			expectedError: true,
		},
		"Valid Auto SPC #1": {
			spc: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-pool-claim-1",
				},
				Spec: apis.StoragePoolClaimSpec{
					PoolSpec: apis.CStorPoolAttr{
						PoolType: string(apis.PoolTypeRaidz2CPV),
					},
					Type:     string(apis.TypeBlockDeviceCPV),
					MaxPools: newInt(3),
				},
			},
			expectedError: false,
		},
		"Valid Manual SPC #1": {
			spc: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-pool-claim-1",
				},
				Spec: apis.StoragePoolClaimSpec{
					PoolSpec: apis.CStorPoolAttr{
						PoolType: string(apis.PoolTypeRaidz2CPV),
					},
					BlockDevices: apis.BlockDeviceAttr{
						BlockDeviceList: []string{"blockdevice-1"},
					},
					Type: string(apis.TypeBlockDeviceCPV),
				},
			},
			expectedError: false,
		},
		"InValid Auto SPC with invalid pool type": {
			spc: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-pool-claim-1",
				},
				Spec: apis.StoragePoolClaimSpec{
					PoolSpec: apis.CStorPoolAttr{
						PoolType: string(apis.PoolTypeRaidz2CPV),
					},
					Type:     string(apis.TypeSparseCPV),
					MaxPools: newInt(3),
				},
			},
			expectedError: true,
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			err := validate(test.spc)
			var gotError bool
			if err != nil {
				gotError = true
			}
			if gotError != test.expectedError {
				t.Errorf("Test case failed as expected error %v but got error %v", test.expectedError, gotError)
			}
		})
	}
}

func TestCurrentPoolCount(t *testing.T) {
	fakeKubeClient := fake.NewSimpleClientset()
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()
	fakeNDMClient := ndmFakeClientset.NewSimpleClientset()
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)
	controller, err := NewControllerBuilder().
		withKubeClient(fakeKubeClient).
		withOpenEBSClient(fakeOpenebsClient).
		withNDMClient(fakeNDMClient).
		withspcSynced(openebsInformerFactory).
		withSpcLister(openebsInformerFactory).
		withRecorder(fakeKubeClient).
		withWorkqueueRateLimiting().
		withEventHandler(openebsInformerFactory).
		Build()

	if err != nil {
		t.Fatalf("failed to build controller instance: %s", err)
	}
	tests := map[string]struct {
		spc               *apis.StoragePoolClaim
		expectedPoolCount int
	}{
		"Invalid SPC #1": {
			spc: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-pool-claim-1",
				},
				Spec: apis.StoragePoolClaimSpec{
					Type: string(apis.TypeSparseCPV),
				},
			},
			expectedPoolCount: 0,
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			c, err := controller.getCurrentPoolCount(test.spc)
			if err != nil {
				t.Fatalf("Test case failed duue to error %s", err)
			}
			if c != 0 {
				t.Errorf("Test case failed as expected current pool count %d but got %d", test.expectedPoolCount, c)
			}

		})
	}
}

func TestPendingPoolCount(t *testing.T) {
	fakeKubeClient := fake.NewSimpleClientset()
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()
	fakeNDMClient := ndmFakeClientset.NewSimpleClientset()
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)
	controller, err := NewControllerBuilder().
		withKubeClient(fakeKubeClient).
		withOpenEBSClient(fakeOpenebsClient).
		withNDMClient(fakeNDMClient).
		withspcSynced(openebsInformerFactory).
		withSpcLister(openebsInformerFactory).
		withRecorder(fakeKubeClient).
		withWorkqueueRateLimiting().
		withEventHandler(openebsInformerFactory).
		Build()

	if err != nil {
		t.Fatalf("failed to build controller instance: %s", err)
	}
	tests := map[string]struct {
		spc               *apis.StoragePoolClaim
		expectedPoolCount int
	}{
		"Auto SPC": {
			spc: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-pool-claim-1",
				},
				Spec: apis.StoragePoolClaimSpec{
					Type:     string(apis.TypeSparseCPV),
					MaxPools: newInt(3),
				},
			},
			expectedPoolCount: 3,
		},
		"Manual SPC": {
			spc: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-pool-claim-1",
				},
				Spec: apis.StoragePoolClaimSpec{
					Type: string(apis.TypeSparseCPV),
					BlockDevices: apis.BlockDeviceAttr{
						BlockDeviceList: []string{"blockdevice-1"},
					},
				},
			},
			expectedPoolCount: 0,
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			pc, err := controller.getPendingPoolCount(test.spc)
			if err != nil {
				t.Fatalf("Test case failed duue to error %s", err)
			}
			if pc != test.expectedPoolCount {
				t.Errorf("Test case failed as expected current pool count %d but got %d", test.expectedPoolCount, pc)
			}

		})
	}
}

func TestIsPoolPending(t *testing.T) {
	fakeKubeClient := fake.NewSimpleClientset()
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()
	fakeNDMClient := ndmFakeClientset.NewSimpleClientset()
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)
	controller, err := NewControllerBuilder().
		withKubeClient(fakeKubeClient).
		withOpenEBSClient(fakeOpenebsClient).
		withNDMClient(fakeNDMClient).
		withspcSynced(openebsInformerFactory).
		withSpcLister(openebsInformerFactory).
		withRecorder(fakeKubeClient).
		withWorkqueueRateLimiting().
		withEventHandler(openebsInformerFactory).
		Build()

	if err != nil {
		t.Fatalf("failed to build controller instance: %s", err)
	}
	tests := map[string]struct {
		spc               *apis.StoragePoolClaim
		expectedIsPending bool
	}{
		"Auto SPC": {
			spc: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-pool-claim-1",
				},
				Spec: apis.StoragePoolClaimSpec{
					Type:     string(apis.TypeBlockDeviceCPV),
					MaxPools: newInt(3),
				},
			},
			expectedIsPending: true,
		},
		"Manual SPC": {
			spc: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-pool-claim-1",
				},
				Spec: apis.StoragePoolClaimSpec{
					Type: string(apis.TypeBlockDeviceCPV),
					BlockDevices: apis.BlockDeviceAttr{
						BlockDeviceList: []string{"blockdevice-1"},
					},
				},
			},
			expectedIsPending: false,
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			gotBool := controller.isPoolPending(test.spc)
			if gotBool != test.expectedIsPending {
				t.Errorf("Test case failed as expected %v but got %v", test.expectedIsPending, gotBool)
			}

		})
	}
}

func TestIsValidPendingPoolCount(t *testing.T) {
	tests := map[string]struct {
		pendingPoolCount int
		isValid          bool
	}{
		"Invalid Pending Pool Count": {
			pendingPoolCount: -1,
			isValid:          false,
		},
		"Valid Pending Pool Count": {
			pendingPoolCount: 1,
			isValid:          true,
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			gotBool := isValidPendingPoolCount(test.pendingPoolCount)
			if gotBool != test.isValid {
				t.Errorf("Test case failed as expected %v but got %v", test.isValid, gotBool)
			}

		})
	}
}

func TestIsManualProvisioning(t *testing.T) {
	tests := map[string]struct {
		spc                *apis.StoragePoolClaim
		manualProvisioning bool
	}{
		"A manual spc Config": {
			spc: &apis.StoragePoolClaim{
				Spec: apis.StoragePoolClaimSpec{
					BlockDevices: apis.BlockDeviceAttr{
						BlockDeviceList: []string{},
					},
				},
			},
			manualProvisioning: true,
		},
		"Not a manual spc config": {
			spc: &apis.StoragePoolClaim{
				Spec: apis.StoragePoolClaimSpec{
					BlockDevices: apis.BlockDeviceAttr{},
				},
			},
			manualProvisioning: false,
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			gotBool := isManualProvisioning(test.spc)
			if gotBool != test.manualProvisioning {
				t.Errorf("Test case failed as expected %v but got %v", test.manualProvisioning, gotBool)
			}

		})
	}
}

func TestIsAutoProvisioning(t *testing.T) {

	tests := map[string]struct {
		spc              *apis.StoragePoolClaim
		autoProvisioning bool
	}{
		"A auto spc Config #1": {
			spc: &apis.StoragePoolClaim{
				Spec: apis.StoragePoolClaimSpec{
					BlockDevices: apis.BlockDeviceAttr{},
				},
			},
			autoProvisioning: true,
		},
		"A auto spc Config #2": {
			spc: &apis.StoragePoolClaim{
				Spec: apis.StoragePoolClaimSpec{},
			},
			autoProvisioning: true,
		},
		"A auto spc Config #3": {
			spc: &apis.StoragePoolClaim{
				Spec: apis.StoragePoolClaimSpec{
					BlockDevices: apis.BlockDeviceAttr{
						BlockDeviceList: nil,
					},
				},
			},
			autoProvisioning: true,
		},
		"Not a auto spc config#1": {
			spc: &apis.StoragePoolClaim{
				Spec: apis.StoragePoolClaimSpec{
					BlockDevices: apis.BlockDeviceAttr{
						BlockDeviceList: []string{},
					},
				},
			},
			autoProvisioning: false,
		},
		"Not a auto spc config#2": {
			spc: &apis.StoragePoolClaim{
				Spec: apis.StoragePoolClaimSpec{
					BlockDevices: apis.BlockDeviceAttr{
						BlockDeviceList: []string{"blockdevice1"},
					},
				},
			},
			autoProvisioning: false,
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			gotBool := isAutoProvisioning(test.spc)
			if gotBool != test.autoProvisioning {
				t.Errorf("Test case failed as expected %v but got %v", test.autoProvisioning, gotBool)
			}

		})
	}
}

func newInt(val int) *int {
	value := val
	return &value
}
