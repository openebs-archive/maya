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
	"testing"
	"time"

	"github.com/openebs/maya/cmd/cstor-pool-mgmt/controller/common"
	zpool "github.com/openebs/maya/cmd/cstor-pool-mgmt/pool/v1alpha2"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha2"
	openebsFakeClientset "github.com/openebs/maya/pkg/client/generated/openebs.io/v1alpha2/clientset/internalclientset/fake"
	informers "github.com/openebs/maya/pkg/client/generated/openebs.io/v1alpha2/informer/externalversions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
)

// TestGetPoolResource checks if pool resource created is successfully got.
func TestGetPoolResource(t *testing.T) {
	fakeKubeClient := fake.NewSimpleClientset()
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(fakeKubeClient, time.Second*30)
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)

	// Instantiate the cStor Pool controllers.
	poolController := NewCStorPoolController(fakeKubeClient, fakeOpenebsClient, kubeInformerFactory,
		openebsInformerFactory)

	testPoolResource := map[string]struct {
		expectedPoolName string
		test             *apis.CStorNPool
	}{
		"img1PoolResource": {
			expectedPoolName: "abc",
			test: &apis.CStorNPool{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pool1",
					UID:       types.UID("abc"),
					Namespace: common.DefaultNameSpace,
				},
				Spec: apis.PoolSpec{
					PoolConfig: apis.PoolConfig{
						CacheFile:            "/tmp/pool1.cache",
						DefaultRaidGroupType: "mirrored",
						OverProvisioning:     false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
		},
		"img2PoolResource": {
			expectedPoolName: "abcd",
			test: &apis.CStorNPool{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pool2",
					UID:       types.UID("abcd"),
					Namespace: common.DefaultNameSpace,
				},
				Spec: apis.PoolSpec{
					PoolConfig: apis.PoolConfig{
						CacheFile:            "/tmp/pool2.cache",
						DefaultRaidGroupType: "striped",
						OverProvisioning:     false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
		},
		"img3PoolResource": {
			expectedPoolName: "existingpool",
			test: &apis.CStorNPool{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pool3",
					UID:       types.UID("existingpool"),
					Namespace: common.DefaultNameSpace,
				},
				Spec: apis.PoolSpec{
					PoolConfig: apis.PoolConfig{
						CacheFile:            "/tmp/pool3.cache",
						DefaultRaidGroupType: "striped",
						OverProvisioning:     false,
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
		_, err := poolController.clientset.OpenebsV1alpha2().CStorNPools(ut.test.Namespace).Create(ut.test)
		if err != nil {
			t.Fatalf("Desc:%v, Unable to create resource : %v", desc, ut.test.ObjectMeta.Name)
		}
		// Get the created pool resource using name
		cStorPoolObtained, err := poolController.getCSPObjFromKey(ut.test.ObjectMeta.Name)
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

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(fakeKubeClient, time.Second*30)
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)

	// Instantiate the cStor Pool controllers.
	poolController := NewCStorPoolController(fakeKubeClient, fakeOpenebsClient, kubeInformerFactory,
		openebsInformerFactory)

	testPoolResource := map[string]struct {
		expectedError error
		test          *apis.CStorNPool
	}{
		"img2PoolResource": {
			expectedError: nil,
			test: &apis.CStorNPool{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:       "pool2",
					UID:        types.UID("abcd"),
					Finalizers: []string{"cstorpool.openebs.io/finalizer"},
					Namespace:  common.DefaultNameSpace,
				},
				Spec: apis.PoolSpec{
					PoolConfig: apis.PoolConfig{
						CacheFile:            "/tmp/pool2.cache",
						DefaultRaidGroupType: "striped",
						OverProvisioning:     false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
		},
	}
	for desc, ut := range testPoolResource {
		// Create Pool resource
		_, err := poolController.clientset.OpenebsV1alpha2().CStorNPools(ut.test.Namespace).Create(ut.test)
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
		test           *apis.CStorNPool
	}{
		"img2PoolResource": {
			expectedOutput: true,
			test: &apis.CStorNPool{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:       "pool2",
					UID:        types.UID("abcd"),
					Finalizers: []string{"cstorpool.openebs.io/finalizer"},
					Namespace:  common.DefaultNameSpace,
				},
				Spec: apis.PoolSpec{
					PoolConfig: apis.PoolConfig{
						CacheFile:            "/tmp/pool2.cache",
						DefaultRaidGroupType: "striped",
						OverProvisioning:     false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
		},
	}
	for desc, ut := range testPoolResource {
		os.Setenv(string(common.OpenEBSIOCStorID), string(ut.test.UID))
		obtainedOutput := zpool.IsRightCStorPoolMgmt(ut.test)
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
		test           *apis.CStorNPool
	}{
		"img2PoolResource": {
			expectedOutput: false,
			test: &apis.CStorNPool{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:       "pool2",
					UID:        types.UID("abcd"),
					Finalizers: []string{"cstorpool.openebs.io/finalizer"},
					Namespace:  common.DefaultNameSpace,
				},
				Spec: apis.PoolSpec{
					PoolConfig: apis.PoolConfig{
						CacheFile:            "/tmp/pool2.cache",
						DefaultRaidGroupType: "striped",
						OverProvisioning:     false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
		},
	}
	for desc, ut := range testPoolResource {
		os.Setenv(string(common.OpenEBSIOCStorID), string("awer"))
		obtainedOutput := zpool.IsRightCStorPoolMgmt(ut.test)
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
		test           *apis.CStorNPool
	}{
		"img2PoolResource": {
			expectedOutput: true,
			test: &apis.CStorNPool{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:              "pool2",
					UID:               types.UID("abcd"),
					Finalizers:        []string{"cstorpool.openebs.io/finalizer"},
					DeletionTimestamp: &deletionTimeStamp,
					Namespace:         common.DefaultNameSpace,
				},
				Spec: apis.PoolSpec{
					PoolConfig: apis.PoolConfig{
						CacheFile:            "/tmp/pool2.cache",
						DefaultRaidGroupType: "striped",
						OverProvisioning:     false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
		},
		"img1PoolResource": {
			expectedOutput: false,
			test: &apis.CStorNPool{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:              "pool1",
					UID:               types.UID("abcd"),
					Finalizers:        []string{"cstorpool.openebs.io/finalizer"},
					DeletionTimestamp: nil,
					Namespace:         common.DefaultNameSpace,
				},
				Spec: apis.PoolSpec{
					PoolConfig: apis.PoolConfig{
						CacheFile:            "/tmp/pool2.cache",
						DefaultRaidGroupType: "striped",
						OverProvisioning:     false,
					},
				},
				Status: apis.CStorPoolStatus{},
			},
		},
	}
	for desc, ut := range testPoolResource {
		obtainedOutput := zpool.IsDestroyed(ut.test)
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
		testOld        *apis.CStorNPool
		testNew        *apis.CStorNPool
	}{
		"img2PoolResource": {
			expectedOutput: true,
			testOld: &apis.CStorNPool{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:              "pool2",
					UID:               types.UID("abcd"),
					Finalizers:        []string{"cstorpool.openebs.io/finalizer"},
					DeletionTimestamp: &deletionTimeStamp,
					Namespace:         common.DefaultNameSpace,
				},
				Spec: apis.PoolSpec{
					PoolConfig: apis.PoolConfig{
						CacheFile:            "/tmp/pool2.cache",
						DefaultRaidGroupType: "striped",
						OverProvisioning:     false,
					},
				},
				Status: apis.CStorPoolStatus{Phase: "init"},
			},
			testNew: &apis.CStorNPool{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:              "pool2",
					UID:               types.UID("abcd"),
					Finalizers:        []string{"cstorpool.openebs.io/finalizer"},
					DeletionTimestamp: &deletionTimeStamp,
				},
				Spec: apis.PoolSpec{
					PoolConfig: apis.PoolConfig{
						CacheFile:            "/tmp/pool2.cache",
						DefaultRaidGroupType: "striped",
						OverProvisioning:     false,
					},
				},
				Status: apis.CStorPoolStatus{Phase: "offline"},
			},
		},
		"img1PoolResource": {
			expectedOutput: false,
			testOld: &apis.CStorNPool{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:              "pool1",
					UID:               types.UID("abc"),
					Finalizers:        []string{"cstorpool.openebs.io/finalizer"},
					DeletionTimestamp: &deletionTimeStamp,
					Namespace:         common.DefaultNameSpace,
				},
				Spec: apis.PoolSpec{
					PoolConfig: apis.PoolConfig{
						CacheFile:            "/tmp/pool2.cache",
						DefaultRaidGroupType: "striped",
						OverProvisioning:     false,
					},
				},
				Status: apis.CStorPoolStatus{Phase: "init"},
			},
			testNew: &apis.CStorNPool{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:              "pool2",
					UID:               types.UID("abcd"),
					Finalizers:        []string{"cstorpool.openebs.io/finalizer"},
					DeletionTimestamp: &deletionTimeStamp,
					Namespace:         common.DefaultNameSpace,
				},
				Spec: apis.PoolSpec{
					PoolConfig: apis.PoolConfig{
						CacheFile:            "/tmp/pool2.cache",
						DefaultRaidGroupType: "striped",
						OverProvisioning:     false,
					},
				},
				Status: apis.CStorPoolStatus{Phase: "init"},
			},
		},
	}
	for desc, ut := range testPoolResource {
		obtainedOutput := zpool.IsOnlyStatusChange(ut.testOld, ut.testNew)
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
		test           *apis.CStorNPool
	}{
		"img2PoolResource": {
			expectedOutput: true,
			test: &apis.CStorNPool{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:              "pool2",
					UID:               types.UID("abcd"),
					Finalizers:        []string{"cstorpool.openebs.io/finalizer"},
					DeletionTimestamp: &deletionTimeStamp,
					Namespace:         common.DefaultNameSpace,
				},
				Spec: apis.PoolSpec{
					PoolConfig: apis.PoolConfig{
						CacheFile:            "/tmp/pool2.cache",
						DefaultRaidGroupType: "striped",
						OverProvisioning:     false,
					},
				},
				Status: apis.CStorPoolStatus{Phase: ""},
			},
		},
		"img1PoolResource": {
			expectedOutput: false,
			test: &apis.CStorNPool{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:              "pool1",
					UID:               types.UID("abcde"),
					Finalizers:        []string{"cstorpool.openebs.io/finalizer"},
					DeletionTimestamp: &deletionTimeStamp,
					Namespace:         common.DefaultNameSpace,
				},
				Spec: apis.PoolSpec{
					PoolConfig: apis.PoolConfig{
						CacheFile:            "/tmp/pool2.cache",
						DefaultRaidGroupType: "striped",
						OverProvisioning:     false,
					},
				},
				Status: apis.CStorPoolStatus{Phase: "online"},
			},
		},
	}
	for desc, ut := range testPoolResource {
		obtainedOutput := zpool.IsEmptyStatus(ut.test)
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
		test           *apis.CStorNPool
	}{
		"pool-deletion-failed": {
			expectedOutput: true,
			test: &apis.CStorNPool{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:              "pool2",
					UID:               types.UID("abcd"),
					Finalizers:        []string{"cstorpool.openebs.io/finalizer"},
					DeletionTimestamp: &deletionTimeStamp,
					Namespace:         common.DefaultNameSpace,
				},
				Spec: apis.PoolSpec{
					PoolConfig: apis.PoolConfig{
						CacheFile:            "/tmp/pool2.cache",
						DefaultRaidGroupType: "striped",
						OverProvisioning:     false,
					},
				},
				Status: apis.CStorPoolStatus{Phase: "DeletionFailed"},
			},
		},
		"pool-online": {
			expectedOutput: false,
			test: &apis.CStorNPool{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:              "pool1",
					UID:               types.UID("abcde"),
					Finalizers:        []string{"cstorpool.openebs.io/finalizer"},
					DeletionTimestamp: &deletionTimeStamp,
					Namespace:         common.DefaultNameSpace,
				},
				Spec: apis.PoolSpec{
					PoolConfig: apis.PoolConfig{
						CacheFile:            "/tmp/pool2.cache",
						DefaultRaidGroupType: "striped",
						OverProvisioning:     false,
					},
				},
				Status: apis.CStorPoolStatus{Phase: "Online"},
			},
		},
	}
	for desc, ut := range testPoolResource {
		obtainedOutput := zpool.IsDeletionFailedBefore(ut.test)
		if obtainedOutput != ut.expectedOutput {
			t.Fatalf("Desc:%v, Expected:%v, Got:%v", desc, ut.expectedOutput,
				obtainedOutput)
		}
	}
}
