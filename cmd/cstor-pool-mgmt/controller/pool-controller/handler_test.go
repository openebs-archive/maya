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
	"context"
	"os"
	"testing"
	"time"

	"github.com/openebs/maya/cmd/cstor-pool-mgmt/controller/common"
	pool "github.com/openebs/maya/cmd/cstor-pool-mgmt/pool"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	openebsFakeClientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned/fake"
	informers "github.com/openebs/maya/pkg/client/generated/informers/externalversions"
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
		_, err := poolController.clientset.OpenebsV1alpha1().CStorPools().
			Create(context.TODO(), ut.test, metav1.CreateOptions{})
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

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(fakeKubeClient, time.Second*30)
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)

	// Instantiate the cStor Pool controllers.
	poolController := NewCStorPoolController(fakeKubeClient, fakeOpenebsClient, kubeInformerFactory,
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
		_, err := poolController.clientset.OpenebsV1alpha1().CStorPools().
			Create(context.TODO(), ut.test, metav1.CreateOptions{})
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

func TestGetDevPathIfNotSlashDev(t *testing.T) {
	testGetDevPath := map[string]string{
		"abcd":             "",
		"/dev":             "",
		"/dev/":            "",
		"/dev/disk":        "",
		"/dev/disk/":       "",
		"/dev/by-id/disk":  "",
		"/dev/by-id/disk/": "",
	}
	for ip, op := range testGetDevPath {
		obtainedOutput := pool.GetDevPathIfNotSlashDev(ip)
		if obtainedOutput != op {
			t.Fatalf("IP:%v, OP:%v, Got:%v", ip, op, obtainedOutput)
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
