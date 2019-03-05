/*
Copyright 2019 The OpenEBS Authors

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
	informers "github.com/openebs/maya/pkg/client/generated/informer/externalversions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"reflect"
	"testing"
	"time"
)

func TestNewPoolConfig(t *testing.T) {
	// fakeKubeClient, fakeOpenebsClient, kubeInformerFactory, and openebsInformerFactory
	// are arguments that is expected by the NewController function.
	fakeKubeClient := fake.NewSimpleClientset()
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(fakeKubeClient, time.Second*30)
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)
	// Instantiate the controller by passing the valid arguments.
	controller := NewController(fakeKubeClient, fakeOpenebsClient, kubeInformerFactory,
		openebsInformerFactory)
	tests := map[string]struct {
		// fakeCasPool holds the fake fakeCasPool object in test cases.
		fakespc *apis.StoragePoolClaim
		fakecsp *apis.CStorPool
		// expectedDiskListLength holds the length of disk list
		expectedConfig *PoolConfig
	}{
		//Test Case #1
		"fakeObject#1": {
			fakespc: &apis.StoragePoolClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pool1",
				},
			},
			fakecsp: nil,
			expectedConfig: &PoolConfig{
				spc: &apis.StoragePoolClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pool1",
					},
				},
				cspList:    &apis.CStorPoolList{},
				controller: controller,
			},
		},
	}
	for name, test := range tests {
		// Pinning the variables to avoid scope lint issue.
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			gotPoolConfig, _ := controller.NewPoolConfig(test.fakespc)
			if !reflect.DeepEqual(gotPoolConfig, test.expectedConfig) {
				t.Errorf("Test case %s failed:expected %+v but got %+v", name, test.expectedConfig, gotPoolConfig)
			}
		})
	}
}

// ToDo: UT for raidz and raidz2
func TestIsTopLevelVdevLost(t *testing.T) {
	tests := map[string]struct {
		fakecsp            *apis.CStorPool
		expectedTruthValue bool
	}{
		"TestCase#1": {
			fakecsp: &apis.CStorPool{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cstor-pool1",
				},
				Spec: apis.CStorPoolSpec{
					Group: []apis.DiskGroup{
						{
							Item: []apis.CspDisk{
								{
									Name:        "Disk1",
									DeviceID:    "/var/img1",
									InUseByPool: true,
								},
								{
									Name:        "Disk2",
									DeviceID:    "/var/img2",
									InUseByPool: true,
								},
							},
						},
					},
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool1.cache",
						PoolType:         string(apis.PoolTypeMirroredCPV),
						OverProvisioning: false,
					},
				},
			},
			expectedTruthValue: false,
		},
		"TestCase#2": {
			fakecsp: &apis.CStorPool{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cstor-pool2",
				},
				Spec: apis.CStorPoolSpec{
					Group: []apis.DiskGroup{
						{
							Item: []apis.CspDisk{
								{
									Name:        "Disk1",
									DeviceID:    "/var/img1",
									InUseByPool: false,
								},
								{
									Name:        "Disk2",
									DeviceID:    "/var/img2",
									InUseByPool: true,
								},
							},
						},
					},
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool1.cache",
						PoolType:         string(apis.PoolTypeMirroredCPV),
						OverProvisioning: false,
					},
				},
			},
			expectedTruthValue: false,
		},
		"TestCase#3": {
			fakecsp: &apis.CStorPool{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cstor-pool2",
				},
				Spec: apis.CStorPoolSpec{
					Group: []apis.DiskGroup{
						{
							Item: []apis.CspDisk{
								{
									Name:        "Disk1",
									DeviceID:    "/var/img1",
									InUseByPool: false,
								},
								{
									Name:        "Disk2",
									DeviceID:    "/var/img2",
									InUseByPool: true,
								},
							},
						},
						{
							Item: []apis.CspDisk{
								{
									Name:        "Disk3",
									DeviceID:    "/var/img1",
									InUseByPool: false,
								},
								{
									Name:        "Disk4",
									DeviceID:    "/var/img2",
									InUseByPool: true,
								},
							},
						},
					},
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool1.cache",
						PoolType:         string(apis.PoolTypeMirroredCPV),
						OverProvisioning: false,
					},
				},
			},
			expectedTruthValue: false,
		},
		"TestCase#4": {
			fakecsp: &apis.CStorPool{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cstor-pool2",
				},
				Spec: apis.CStorPoolSpec{
					Group: []apis.DiskGroup{
						{
							Item: []apis.CspDisk{
								{
									Name:        "Disk1",
									DeviceID:    "/var/img1",
									InUseByPool: false,
								},
								{
									Name:        "Disk2",
									DeviceID:    "/var/img2",
									InUseByPool: false,
								},
							},
						},
						{
							Item: []apis.CspDisk{
								{
									Name:        "Disk3",
									DeviceID:    "/var/img1",
									InUseByPool: true,
								},
								{
									Name:        "Disk4",
									DeviceID:    "/var/img2",
									InUseByPool: true,
								},
							},
						},
					},
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool1.cache",
						PoolType:         string(apis.PoolTypeMirroredCPV),
						OverProvisioning: false,
					},
				},
			},
			expectedTruthValue: true,
		},
		"TestCase#5": {
			fakecsp: &apis.CStorPool{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cstor-pool2",
				},
				Spec: apis.CStorPoolSpec{
					Group: []apis.DiskGroup{
						{
							Item: []apis.CspDisk{
								{
									Name:        "Disk1",
									DeviceID:    "/var/img1",
									InUseByPool: true,
								},
							},
						},
						{
							Item: []apis.CspDisk{
								{
									Name:        "Disk4",
									DeviceID:    "/var/img2",
									InUseByPool: true,
								},
							},
						},
					},
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool1.cache",
						PoolType:         string(apis.PoolTypeStripedCPV),
						OverProvisioning: false,
					},
				},
			},
			expectedTruthValue: false,
		},
		"TestCase#6": {
			fakecsp: &apis.CStorPool{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cstor-pool2",
				},
				Spec: apis.CStorPoolSpec{
					Group: []apis.DiskGroup{
						{
							Item: []apis.CspDisk{
								{
									Name:        "Disk1",
									DeviceID:    "/var/img1",
									InUseByPool: false,
								},
							},
						},
						{
							Item: []apis.CspDisk{
								{
									Name:        "Disk4",
									DeviceID:    "/var/img2",
									InUseByPool: true,
								},
							},
						},
					},
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool1.cache",
						PoolType:         string(apis.PoolTypeStripedCPV),
						OverProvisioning: false,
					},
				},
			},
			expectedTruthValue: true,
		},
		"TestCase#7": {
			fakecsp: &apis.CStorPool{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cstor-pool2",
				},
				Spec: apis.CStorPoolSpec{
					Group: []apis.DiskGroup{
						{
							Item: []apis.CspDisk{
								{
									Name:        "Disk1",
									DeviceID:    "/var/img1",
									InUseByPool: true,
								},
								{
									Name:        "Disk2",
									DeviceID:    "/var/img2",
									InUseByPool: false,
								},
							},
						},
						{
							Item: []apis.CspDisk{
								{
									Name:        "Disk4",
									DeviceID:    "/var/img2",
									InUseByPool: true,
								},
								{
									Name:        "Disk5",
									DeviceID:    "/var/img5",
									InUseByPool: true,
								},
							},
						},
					},
					PoolSpec: apis.CStorPoolAttr{
						CacheFile:        "/tmp/pool1.cache",
						PoolType:         string(apis.PoolTypeStripedCPV),
						OverProvisioning: false,
					},
				},
			},
			expectedTruthValue: true,
		},
	}
	for name, test := range tests {
		// Pinning the variables to avoid scope lint issue.
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			gotTruthValue := isTopVdevLost(test.fakecsp)
			if gotTruthValue != test.expectedTruthValue {
				t.Errorf("Test case %s failed:expected truth value %v but got %v", name, test.expectedTruthValue, gotTruthValue)
			}
		})
	}
}

// TODO: Add sparse and disk test cases.
func TestGetDeviceId(t *testing.T) {
	tests := map[string]struct {
		fakedisk         *apis.Disk
		expectedDeviceID string
	}{
		"TestCase#1": {
			fakedisk: &apis.Disk{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cstor-pool1",
				},
			},
			expectedDeviceID: "",
		},
	}
	for name, test := range tests {
		// Pinning the variables to avoid scope lint issue.
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			gotDeviceID := getDeviceID(test.fakedisk)
			if gotDeviceID != test.expectedDeviceID {
				t.Errorf("Test case %s failed:expected device id %s but got %s", name, test.expectedDeviceID, gotDeviceID)
			}
		})
	}
}

func TestEnqueueAddOperation(t *testing.T) {
	tests := map[string]struct {
		fakecsp                      *apis.CStorPool
		diskDeviceIDs                []string
		expectedOperationSubResource []apis.CstorOperation
	}{
		"TestCase#1": {
			fakecsp: &apis.CStorPool{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cstor-pool1",
				},
			},
			diskDeviceIDs: []string{"disk1", "disk2"},
			expectedOperationSubResource: []apis.CstorOperation{
				{
					Action:   apis.PoolExpandAction,
					NewDisks: []string{"disk1", "disk2"},
					OldDisk:  nil,
					Status:   apis.PoolOperationStatusInit,
				},
			},
		},
	}
	for name, test := range tests {
		// Pinning the variables to avoid scope lint issue.
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			gotCsp := enqueueAddOperation(test.fakecsp, test.diskDeviceIDs)
			if !reflect.DeepEqual(gotCsp.Operations, test.expectedOperationSubResource) {
				t.Errorf("Test case %s failed:expected %+v but got %+v", name, test.expectedOperationSubResource, gotCsp.Operations)
			}
		})
	}
}

func TestEnqueueDeleteOperation(t *testing.T) {
	tests := map[string]struct {
		fakecsp                      *apis.CStorPool
		expectedOperationSubResource []apis.CstorOperation
	}{
		"TestCase#1": {
			fakecsp: &apis.CStorPool{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cstor-pool1",
				},
			},
			expectedOperationSubResource: []apis.CstorOperation{
				{
					Action: apis.PoolDeleteAction,
					Status: apis.PoolOperationStatusInit,
				},
			},
		},
	}
	for name, test := range tests {
		// Pinning the variables to avoid scope lint issue.
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			gotCsp := enqueueDeleteOperation(test.fakecsp)
			if !reflect.DeepEqual(gotCsp.Operations, test.expectedOperationSubResource) {
				t.Errorf("Test case %s failed:expected %+v but got %+v", name, test.expectedOperationSubResource, gotCsp.Operations)
			}
		})
	}
}
