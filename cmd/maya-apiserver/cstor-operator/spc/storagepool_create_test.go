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
	"strconv"
	"testing"
	"time"

	"github.com/golang/glog"
	nodeselect "github.com/openebs/maya/pkg/algorithm/nodeselect/v1alpha1"
	ndmapis "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	openebsFakeClientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned/fake"

	informers "github.com/openebs/maya/pkg/client/generated/informers/externalversions"
	ndmFakeClientset "github.com/openebs/maya/pkg/client/generated/openebs.io/ndm/v1alpha1/clientset/internalclientset/fake"
	cstorpool "github.com/openebs/maya/pkg/cstorpool/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	//	"time"

	blockdevice "github.com/openebs/maya/pkg/blockdevice/v1alpha1"
	sp "github.com/openebs/maya/pkg/sp/v1alpha1"
	"k8s.io/client-go/kubernetes/fake"
)

var bdK8sClient *blockdevice.KubernetesClient
var fakeDiskCreateFlag bool

func FakeDiskCreator(dc *blockdevice.KubernetesClient) {
	// Create some fake block device objects over nodes.
	// For example, create 6 disk (out of 6 disks 2 disks are sparse disks)for each of 5 nodes.
	// That meant 6*5 i.e. 30 disk objects should be created

	// diskObjectList will hold the list of disk objects
	var diskObjectList [30]*ndmapis.BlockDevice

	sparseDiskCount := 2
	var diskLabel, key string

	// nodeIdentifer will help in naming a node and attaching multiple disks to a single node.
	nodeIdentifer := 0
	for diskListIndex := 0; diskListIndex < 30; diskListIndex++ {
		diskIdentifier := strconv.Itoa(diskListIndex)
		if diskListIndex%6 == 0 {
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
		_, err := dc.Create(diskObjectList[diskListIndex])
		if err != nil {
			glog.Error(err)
		}
	}
	fakeDiskCreateFlag = true
}

func (focs *PoolCreateConfig) FakeDiskCreator() {
	// Create some fake disk objects over nodes.
	// For example, create 14 disk (out of 14 disks 2 disks are sparse disks)for each of 5 nodes.
	// That meant 14*5 i.e. 70 disk objects should be created

	// diskObjectList will hold the list of disk objects
	var diskObjectList [70]*ndmapis.BlockDevice

	sparseDiskCount := 2
	var diskLabel, key string

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
				State: DiskStateActive,
			},
		}
		_, err := focs.ndmclientset.OpenebsV1alpha1().BlockDevices("fake-ns").Create(diskObjectList[diskListIndex])
		if err != nil {
			glog.Error(err)
		}
	}
}
func fakeDiskClient() {
	bdK8sClient = &blockdevice.KubernetesClient{
		Kubeclientset: fake.NewSimpleClientset(),
		Clientset:     ndmFakeClientset.NewSimpleClientset(),
		Namespace:     "fake-ns",
	}
}
func fakeAlgorithmConfig(spc *apis.StoragePoolClaim) *nodeselect.Config {
	var diskClient blockdevice.BlockDeviceInterface
	fakeDiskClient()
	FakeDiskCreator(bdK8sClient)
	if nodeselect.ProvisioningType(spc) == ProvisioningTypeManual {
		diskClient = &blockdevice.SpcObjectClient{
			KubernetesClient: bdK8sClient,
			Spc:              spc,
		}
	} else {
		diskClient = bdK8sClient
	}

	cspK8sClient := &cstorpool.KubernetesClient{
		Kubeclientset: fake.NewSimpleClientset(),
		Clientset:     openebsFakeClientset.NewSimpleClientset(),
	}
	spK8sClient := &sp.KubernetesClient{
		Kubeclientset: fake.NewSimpleClientset(),
		Clientset:     openebsFakeClientset.NewSimpleClientset(),
	}
	ac := &nodeselect.Config{
		Spc:               spc,
		BlockDeviceClient: diskClient,
		CspClient:         cspK8sClient,
		SpClient:          spK8sClient,
	}

	return ac
}

func TestNewCasPool(t *testing.T) {
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
					Type: "blockdevice",
					PoolSpec: apis.CStorPoolAttr{
						PoolType: "striped",
					},
					BlockDevices: apis.BlockDeviceAttr{
						BlockDeviceList: []string{"blockdevice1", "blockdevice2", "blockdevice3"},
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
					MaxPools: newInt(6),
					MinPools: 3,
					Type:     "blockdevice",
					PoolSpec: apis.CStorPoolAttr{
						PoolType: "mirrored",
					},
				},
			},
		},
	}

	// Iterate over whole map to run the test cases.
	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			// newCasPool is the function under test.
			fakeAlgoConf := fakeAlgorithmConfig(test.fakestoragepoolclaim)
			fakePoolConfig := &PoolCreateConfig{
				fakeAlgoConf,
				controller,
			}
			if !fakeDiskCreateFlag {
				fakePoolConfig.FakeDiskCreator()
			}
			CasPool, err := fakePoolConfig.getCasPool(test.fakestoragepoolclaim)
			if err != nil || CasPool == nil {
				t.Errorf("Test case failed as expected nil error but error or CasPool object was nil:%s", name)
			}
		})
	}
}
