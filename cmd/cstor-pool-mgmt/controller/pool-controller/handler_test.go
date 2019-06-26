/*
Copyright 2018 The OpenEBS Authors.

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
package poolcontroller

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/golang/glog"
	"github.com/openebs/maya/cmd/cstor-pool-mgmt/controller/common"
	ndmapis "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	blockdevice "github.com/openebs/maya/pkg/blockdevice/v1alpha1"
	bdc "github.com/openebs/maya/pkg/blockdeviceclaim/v1alpha1"
	openebsFakeClientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned/fake"
	informers "github.com/openebs/maya/pkg/client/generated/informers/externalversions"
	ndmFakeClientset "github.com/openebs/maya/pkg/client/generated/openebs.io/ndm/v1alpha1/clientset/internalclientset/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
)

// TestGetPoolResource checks if pool resource created is successfully got.
func TestGetPoolResource(t *testing.T) {
	fakeKubeClient := fake.NewSimpleClientset()
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()
	fakeNDMClient := ndmFakeClientset.NewSimpleClientset()

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(fakeKubeClient, time.Second*30)
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)

	// Instantiate the cStor Pool controllers.
	poolController := NewCStorPoolController(fakeKubeClient, fakeOpenebsClient, fakeNDMClient, kubeInformerFactory,
		openebsInformerFactory)

	testPoolResource := map[string]struct {
		expectedPoolName string
		test             *apis.CStorPool
	}{
		"img1PoolResource": {
			expectedPoolName: "abc",
			test: &apis.CStorPool{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name: "pool1",
					UID:  types.UID("abc"),
				},
				Spec: apis.CStorPoolSpec{
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool1.cache",
						PoolType:         "mirrored",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
		},
		"img2PoolResource": {
			expectedPoolName: "abcd",
			test: &apis.CStorPool{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name: "pool2",
					UID:  types.UID("abcd"),
				},
				Spec: apis.CStorPoolSpec{
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool2.cache",
						PoolType:         "striped",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
		},
		"img3PoolResource": {
			expectedPoolName: "existingpool",
			test: &apis.CStorPool{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name: "pool3",
					UID:  types.UID("existingpool"),
				},
				Spec: apis.CStorPoolSpec{
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool3.cache",
						PoolType:         "striped",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{
					Phase:    "Healthy",
					Capacity: apis.CStorPoolCapacityAttr{},
				},
			},
		},
	}
	for desc, ut := range testPoolResource {
		// Create Pool resource
		_, err := poolController.clientset.OpenebsV1alpha1().CStorPools().Create(ut.test)
		if err != nil {
			t.Fatalf("Desc:%v, Unable to create resource : %v", desc, ut.test.ObjectMeta.Name)
		}
		// Get the created pool resource using name
		cStorPoolObtained, err := poolController.getPoolResource(ut.test.ObjectMeta.Name)
		if string(cStorPoolObtained.ObjectMeta.UID) != ut.expectedPoolName {
			t.Fatalf("Desc:%v, PoolName mismatch, Expected:%v, Got:%v", desc, ut.expectedPoolName,
				string(cStorPoolObtained.ObjectMeta.UID))
		}
	}
}

// TestRemoveFinalizer is to remove pool resource.
func TestRemoveFinalizer(t *testing.T) {
	fakeKubeClient := fake.NewSimpleClientset()
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()
	fakeNDMClient := ndmFakeClientset.NewSimpleClientset()

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(fakeKubeClient, time.Second*30)
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)

	// Instantiate the cStor Pool controllers.
	poolController := NewCStorPoolController(fakeKubeClient, fakeOpenebsClient, fakeNDMClient, kubeInformerFactory,
		openebsInformerFactory)

	testPoolResource := map[string]struct {
		expectedError error
		test          *apis.CStorPool
	}{
		"img2PoolResource": {
			expectedError: nil,
			test: &apis.CStorPool{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:       "pool2",
					UID:        types.UID("abcd"),
					Finalizers: []string{"cstorpool.openebs.io/finalizer"},
				},
				Spec: apis.CStorPoolSpec{
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool2.cache",
						PoolType:         "striped",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
		},
	}
	for desc, ut := range testPoolResource {
		// Create Pool resource
		_, err := poolController.clientset.OpenebsV1alpha1().CStorPools().Create(ut.test)
		if err != nil {
			t.Fatalf("Desc:%v, Unable to create resource : %v", desc, ut.test.ObjectMeta.Name)
		}
		obtainedErr := poolController.removeFinalizer(ut.test)
		if obtainedErr != ut.expectedError {
			t.Fatalf("Desc:%v, Expected:%v, Got:%v", desc, ut.expectedError,
				obtainedErr)
		}
	}
}

// TestIsRightCStorPoolMgmt is to check if right sidecar does operation with env match.
func TestIsRightCStorPoolMgmt(t *testing.T) {
	testPoolResource := map[string]struct {
		expectedOutput bool
		test           *apis.CStorPool
	}{
		"img2PoolResource": {
			expectedOutput: true,
			test: &apis.CStorPool{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:       "pool2",
					UID:        types.UID("abcd"),
					Finalizers: []string{"cstorpool.openebs.io/finalizer"},
				},
				Spec: apis.CStorPoolSpec{
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool2.cache",
						PoolType:         "striped",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
		},
	}
	for desc, ut := range testPoolResource {
		os.Setenv(string(common.OpenEBSIOCStorID), string(ut.test.UID))
		obtainedOutput := IsRightCStorPoolMgmt(ut.test)
		if obtainedOutput != ut.expectedOutput {
			t.Fatalf("Desc:%v, Expected:%v, Got:%v", desc, ut.expectedOutput,
				obtainedOutput)
		}
		os.Unsetenv(string(common.OpenEBSIOCStorID))
	}
}

// TestIsRightCStorPoolMgmtNegative is to check if right sidecar does operation with env match.
func TestIsRightCStorPoolMgmtNegative(t *testing.T) {
	testPoolResource := map[string]struct {
		expectedOutput bool
		test           *apis.CStorPool
	}{
		"img2PoolResource": {
			expectedOutput: false,
			test: &apis.CStorPool{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:       "pool2",
					UID:        types.UID("abcd"),
					Finalizers: []string{"cstorpool.openebs.io/finalizer"},
				},
				Spec: apis.CStorPoolSpec{
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool2.cache",
						PoolType:         "striped",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
		},
	}
	for desc, ut := range testPoolResource {
		os.Setenv(string(common.OpenEBSIOCStorID), string("awer"))
		obtainedOutput := IsRightCStorPoolMgmt(ut.test)
		if obtainedOutput != ut.expectedOutput {
			t.Fatalf("Desc:%v, Expected:%v, Got:%v", desc, ut.expectedOutput,
				obtainedOutput)
		}
		os.Unsetenv(string(common.OpenEBSIOCStorID))
	}
}

// TestIsDestroyEvent is to test if the event is to destroy pool.
func TestIsDestroyEvent(t *testing.T) {
	deletionTimeStamp := metav1.Now()
	testPoolResource := map[string]struct {
		expectedOutput bool
		test           *apis.CStorPool
	}{
		"img2PoolResource": {
			expectedOutput: true,
			test: &apis.CStorPool{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:              "pool2",
					UID:               types.UID("abcd"),
					Finalizers:        []string{"cstorpool.openebs.io/finalizer"},
					DeletionTimestamp: &deletionTimeStamp,
				},
				Spec: apis.CStorPoolSpec{
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool2.cache",
						PoolType:         "striped",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
		},
		"img1PoolResource": {
			expectedOutput: false,
			test: &apis.CStorPool{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:              "pool1",
					UID:               types.UID("abcd"),
					Finalizers:        []string{"cstorpool.openebs.io/finalizer"},
					DeletionTimestamp: nil,
				},
				Spec: apis.CStorPoolSpec{
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool2.cache",
						PoolType:         "striped",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
		},
	}
	for desc, ut := range testPoolResource {
		obtainedOutput := IsDestroyEvent(ut.test)
		if obtainedOutput != ut.expectedOutput {
			t.Fatalf("Desc:%v, Expected:%v, Got:%v", desc, ut.expectedOutput,
				obtainedOutput)
		}
	}
}

// TestIsOnlyStatusChange is to test if the event is only status change.
func TestIsOnlyStatusChange(t *testing.T) {
	deletionTimeStamp := metav1.Now()
	testPoolResource := map[string]struct {
		expectedOutput bool
		testOld        *apis.CStorPool
		testNew        *apis.CStorPool
	}{
		"img2PoolResource": {
			expectedOutput: true,
			testOld: &apis.CStorPool{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:              "pool2",
					UID:               types.UID("abcd"),
					Finalizers:        []string{"cstorpool.openebs.io/finalizer"},
					DeletionTimestamp: &deletionTimeStamp,
				},
				Spec: apis.CStorPoolSpec{
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool2.cache",
						PoolType:         "striped",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{Phase: "init"},
			},
			testNew: &apis.CStorPool{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:              "pool2",
					UID:               types.UID("abcd"),
					Finalizers:        []string{"cstorpool.openebs.io/finalizer"},
					DeletionTimestamp: &deletionTimeStamp,
				},
				Spec: apis.CStorPoolSpec{
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool2.cache",
						PoolType:         "striped",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{Phase: "offline"},
			},
		},
		"img1PoolResource": {
			expectedOutput: false,
			testOld: &apis.CStorPool{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:              "pool1",
					UID:               types.UID("abc"),
					Finalizers:        []string{"cstorpool.openebs.io/finalizer"},
					DeletionTimestamp: &deletionTimeStamp,
				},
				Spec: apis.CStorPoolSpec{
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool2.cache",
						PoolType:         "striped",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{Phase: "init"},
			},
			testNew: &apis.CStorPool{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:              "pool2",
					UID:               types.UID("abcd"),
					Finalizers:        []string{"cstorpool.openebs.io/finalizer"},
					DeletionTimestamp: &deletionTimeStamp,
				},
				Spec: apis.CStorPoolSpec{
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool2.cache",
						PoolType:         "striped",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{Phase: "init"},
			},
		},
	}
	for desc, ut := range testPoolResource {
		obtainedOutput := IsOnlyStatusChange(ut.testOld, ut.testNew)
		if obtainedOutput != ut.expectedOutput {
			t.Fatalf("Desc:%v, Expected:%v, Got:%v", desc, ut.expectedOutput,
				obtainedOutput)
		}
	}
}

// TestIsEmptyStatus is to check if status is empty.
func TestIsEmptyStatus(t *testing.T) {
	deletionTimeStamp := metav1.Now()
	testPoolResource := map[string]struct {
		expectedOutput bool
		test           *apis.CStorPool
	}{
		"img2PoolResource": {
			expectedOutput: true,
			test: &apis.CStorPool{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:              "pool2",
					UID:               types.UID("abcd"),
					Finalizers:        []string{"cstorpool.openebs.io/finalizer"},
					DeletionTimestamp: &deletionTimeStamp,
				},
				Spec: apis.CStorPoolSpec{
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool2.cache",
						PoolType:         "striped",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{Phase: ""},
			},
		},
		"img1PoolResource": {
			expectedOutput: false,
			test: &apis.CStorPool{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:              "pool1",
					UID:               types.UID("abcde"),
					Finalizers:        []string{"cstorpool.openebs.io/finalizer"},
					DeletionTimestamp: &deletionTimeStamp,
				},
				Spec: apis.CStorPoolSpec{
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool2.cache",
						PoolType:         "striped",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{Phase: "online"},
			},
		},
	}
	for desc, ut := range testPoolResource {
		obtainedOutput := IsEmptyStatus(ut.test)
		if obtainedOutput != ut.expectedOutput {
			t.Fatalf("Desc:%v, Expected:%v, Got:%v", desc, ut.expectedOutput,
				obtainedOutput)
		}
	}
}

// TestIsDeletionFailedBefore is to check if status is only init.
func TestIsDeletionFailedBefore(t *testing.T) {
	deletionTimeStamp := metav1.Now()
	testPoolResource := map[string]struct {
		expectedOutput bool
		test           *apis.CStorPool
	}{
		"pool-deletion-failed": {
			expectedOutput: true,
			test: &apis.CStorPool{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:              "pool2",
					UID:               types.UID("abcd"),
					Finalizers:        []string{"cstorpool.openebs.io/finalizer"},
					DeletionTimestamp: &deletionTimeStamp,
				},
				Spec: apis.CStorPoolSpec{
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool2.cache",
						PoolType:         "striped",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{Phase: "DeletionFailed"},
			},
		},
		"pool-online": {
			expectedOutput: false,
			test: &apis.CStorPool{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:              "pool1",
					UID:               types.UID("abcde"),
					Finalizers:        []string{"cstorpool.openebs.io/finalizer"},
					DeletionTimestamp: &deletionTimeStamp,
				},
				Spec: apis.CStorPoolSpec{
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool2.cache",
						PoolType:         "striped",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{Phase: "Online"},
			},
		},
	}
	for desc, ut := range testPoolResource {
		obtainedOutput := IsDeletionFailedBefore(ut.test)
		if obtainedOutput != ut.expectedOutput {
			t.Fatalf("Desc:%v, Expected:%v, Got:%v", desc, ut.expectedOutput,
				obtainedOutput)
		}
	}
}

func createFakeBlockDevices(c *CStorPoolController) {
	fakeNS := "openebs"
	// bdObjectList will hold the list of blockdevice objects
	for index := 0; index < 5; index++ {
		bdIdentifier := strconv.Itoa(index)
		makeBDObj := &ndmapis.BlockDevice{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name: "blockdevice" + bdIdentifier,
				Labels: map[string]string{
					"kubernetes.io/hostname":  "node1",
					"ndm.io/blockdevice-type": "blockdevice",
				},
			},
			Spec: ndmapis.DeviceSpec{
				Details: ndmapis.DeviceDetails{
					DeviceType: "disk",
				},
				DevLinks: []ndmapis.DeviceDevLink{
					ndmapis.DeviceDevLink{
						Kind:  "by-ID",
						Links: []string{"/dev/sda" + bdIdentifier},
					},
				},
			},
			Status: ndmapis.DeviceStatus{
				State:      "Active",
				ClaimState: ndmapis.BlockDeviceClaimed,
			},
		}
		_, err := c.ndmClientset.OpenebsV1alpha1().BlockDevices(fakeNS).Create(makeBDObj)
		if err != nil {
			glog.Error(err)
		}
	}
}

func TestGetBlockDeviceList(t *testing.T) {
	fakeKubeClient := fake.NewSimpleClientset()
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()
	fakeNDMClient := ndmFakeClientset.NewSimpleClientset()

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(fakeKubeClient, time.Second*30)
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)

	// Instantiate the cStor Pool controllers.
	poolController := NewCStorPoolController(fakeKubeClient, fakeOpenebsClient, fakeNDMClient, kubeInformerFactory,
		openebsInformerFactory)

	createFakeBlockDevices(poolController)
	tests := map[string]struct {
		expectedErr bool
		pool        *apis.CStorPool
	}{
		"Pool1": {
			expectedErr: false,
			pool: &apis.CStorPool{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:       "pool1",
					UID:        types.UID("abcd"),
					Finalizers: []string{"cstorpool.openebs.io/finalizer"},
				},
				Spec: apis.CStorPoolSpec{
					Group: []apis.BlockDeviceGroup{
						apis.BlockDeviceGroup{
							Item: []apis.CspBlockDevice{
								apis.CspBlockDevice{
									Name:        "blockdevice1",
									InUseByPool: true,
									DeviceID:    "/dev/sda1",
								},
								apis.CspBlockDevice{
									Name:        "blockdevice2",
									InUseByPool: true,
									DeviceID:    "/dev/sda2",
								},
							},
						},
						apis.BlockDeviceGroup{
							Item: []apis.CspBlockDevice{
								apis.CspBlockDevice{
									Name:        "blockdevice3",
									InUseByPool: true,
									DeviceID:    "/dev/sda3",
								},
								apis.CspBlockDevice{
									Name:        "blockdevice4",
									InUseByPool: true,
									DeviceID:    "/dev/sda4",
								},
							},
						},
					},
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool2.cache",
						PoolType:         "striped",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
		},
		"Pool2": {
			expectedErr: true,
			pool: &apis.CStorPool{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:       "pool2",
					UID:        types.UID("abcd"),
					Finalizers: []string{"cstorpool.openebs.io/finalizer"},
				},
				Spec: apis.CStorPoolSpec{
					Group: []apis.BlockDeviceGroup{
						apis.BlockDeviceGroup{
							Item: []apis.CspBlockDevice{
								apis.CspBlockDevice{
									Name:        "blockdevice7",
									InUseByPool: true,
									DeviceID:    "/dev/sda1",
								},
								apis.CspBlockDevice{
									Name:        "blockdevice8",
									InUseByPool: true,
									DeviceID:    "/dev/sda2",
								},
							},
						},
					},
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool2.cache",
						PoolType:         "striped",
						OverProvisioning: false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			fakeNs := "openebs"
			_, err := poolController.getBlockDeviceList(test.pool, fakeNs)
			if test.expectedErr && err == nil {
				t.Fatalf("Test case failed expected error not to be nil")
			}
			if !test.expectedErr && err != nil {
				t.Fatalf("Test case failed expected error to be nil but %v", err)
			}
		})
	}
}

func TestValidateBlockDeviceClaimStatus(t *testing.T) {
	tests := map[string]struct {
		expectedErr   bool
		bdList        *blockdevice.BlockDeviceList
		customBDCList *bdc.ListBuilder
	}{
		"testcase1": {
			expectedErr: false,
			bdList: &blockdevice.BlockDeviceList{
				BlockDeviceList: &ndmapis.BlockDeviceList{
					TypeMeta: metav1.TypeMeta{},
					ListMeta: metav1.ListMeta{},
					Items: []ndmapis.BlockDevice{
						{
							TypeMeta: metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{
								Name: "blockdevice1",
							},
							Spec: ndmapis.DeviceSpec{
								ClaimRef: &corev1.ObjectReference{
									Name: "bdc-1",
								},
							},
							Status: ndmapis.DeviceStatus{
								State:      "Active",
								ClaimState: ndmapis.BlockDeviceClaimed,
							},
						},
						{
							TypeMeta: metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{
								Name: "blockdevice2",
							},
							Spec: ndmapis.DeviceSpec{
								ClaimRef: &corev1.ObjectReference{
									Name: "bdc-2",
								},
							},
							Status: ndmapis.DeviceStatus{
								State:      "Active",
								ClaimState: ndmapis.BlockDeviceClaimed,
							},
						},
						{
							TypeMeta: metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{
								Name: "blockdevice3",
							},
							Spec: ndmapis.DeviceSpec{
								ClaimRef: &corev1.ObjectReference{
									Name: "bdc-3",
								},
							},
							Status: ndmapis.DeviceStatus{
								State:      "Active",
								ClaimState: ndmapis.BlockDeviceClaimed,
							},
						},
					},
				},
			},
			customBDCList: &bdc.ListBuilder{
				BlockDeviceClaimList: &bdc.BlockDeviceClaimList{
					ObjectList: &ndmapis.BlockDeviceClaimList{
						TypeMeta: metav1.TypeMeta{},
						ListMeta: metav1.ListMeta{},
						Items: []ndmapis.BlockDeviceClaim{
							{
								TypeMeta: metav1.TypeMeta{},
								ObjectMeta: metav1.ObjectMeta{
									Name: "bdc-1",
								},
								Spec: ndmapis.DeviceClaimSpec{
									HostName:        "openebs-1234",
									BlockDeviceName: "blockdevice1",
								},
							},
							{
								TypeMeta: metav1.TypeMeta{},
								ObjectMeta: metav1.ObjectMeta{
									Name: "bdc-2",
								},
								Spec: ndmapis.DeviceClaimSpec{
									HostName:        "openebs-1234",
									BlockDeviceName: "blockdevice2",
								},
							},
							{
								TypeMeta: metav1.TypeMeta{},
								ObjectMeta: metav1.ObjectMeta{
									Name: "bdc-3",
								},
								Spec: ndmapis.DeviceClaimSpec{
									HostName:        "openebs-1234",
									BlockDeviceName: "blockdevice3",
								},
							},
						},
					},
				},
			},
		},
		"testcase2": {
			expectedErr: true,
			bdList: &blockdevice.BlockDeviceList{
				BlockDeviceList: &ndmapis.BlockDeviceList{
					TypeMeta: metav1.TypeMeta{},
					ListMeta: metav1.ListMeta{},
					Items: []ndmapis.BlockDevice{
						{
							TypeMeta: metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{
								Name: "blockdevice1",
							},
							Spec: ndmapis.DeviceSpec{
								ClaimRef: &corev1.ObjectReference{},
							},
							Status: ndmapis.DeviceStatus{
								State:      "Active",
								ClaimState: ndmapis.BlockDeviceUnclaimed,
							},
						},
					},
				},
			},
			customBDCList: &bdc.ListBuilder{
				BlockDeviceClaimList: &bdc.BlockDeviceClaimList{
					ObjectList: &ndmapis.BlockDeviceClaimList{
						TypeMeta: metav1.TypeMeta{},
						ListMeta: metav1.ListMeta{},
						Items: []ndmapis.BlockDeviceClaim{
							{
								TypeMeta: metav1.TypeMeta{},
								ObjectMeta: metav1.ObjectMeta{
									Name: "bdc-1",
								},
								Spec: ndmapis.DeviceClaimSpec{
									HostName:        "openebs-1234",
									BlockDeviceName: "blockdevice1",
								},
							},
						},
					},
				},
			},
		},
		"testcase3": {
			expectedErr: true,
			bdList: &blockdevice.BlockDeviceList{
				BlockDeviceList: &ndmapis.BlockDeviceList{
					TypeMeta: metav1.TypeMeta{},
					ListMeta: metav1.ListMeta{},
					Items: []ndmapis.BlockDevice{
						{
							TypeMeta: metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{
								Name: "blockdevice1",
							},
							Spec: ndmapis.DeviceSpec{
								ClaimRef: &corev1.ObjectReference{
									Name: "bdc-2",
								},
							},
							Status: ndmapis.DeviceStatus{
								State:      "Active",
								ClaimState: ndmapis.BlockDeviceClaimed,
							},
						},
					},
				},
			},
			customBDCList: &bdc.ListBuilder{
				BlockDeviceClaimList: &bdc.BlockDeviceClaimList{
					ObjectList: &ndmapis.BlockDeviceClaimList{
						TypeMeta: metav1.TypeMeta{},
						ListMeta: metav1.ListMeta{},
						Items: []ndmapis.BlockDeviceClaim{
							{
								TypeMeta: metav1.TypeMeta{},
								ObjectMeta: metav1.ObjectMeta{
									Name: "bdc-1",
								},
								Spec: ndmapis.DeviceClaimSpec{
									HostName:        "openebs-1234",
									BlockDeviceName: "blockdevice4",
								},
							},
						},
					},
				},
			},
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			err := validateBlockDeviceClaimStatus(test.bdList, test.customBDCList)
			if test.expectedErr && err == nil {
				t.Fatalf("Test case failed expected error not to be nil")
			}
			if !test.expectedErr && err != nil {
				t.Fatalf("Test case failed expected error to be nil but %v", err)
			}
		})
	}
}
