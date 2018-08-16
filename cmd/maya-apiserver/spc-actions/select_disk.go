/*
Copyright 2017 The OpenEBS Authors

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
	mach_apis_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//openebs "github.com/openebs/maya/pkg/client/clientset/versioned"
	"errors"
	"fmt"
	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	openebs "github.com/openebs/maya/pkg/client/generated/clientset/internalclientset"
	"github.com/openebs/maya/pkg/client/k8s"
)

// clientset struct holds the interface of internalclientset
// i.e. openebs.
// This struct will be binded to method ListDisk and is useful in mocking
// and unit testing.
type clientSet struct {
	oecs openebs.Interface
}

// nodeDisk struct will be used as a value for a map nodeDiskMap (map defined in ListDisk function)
// The struct will be useful in forming the data structure nodeDiskMap which will be manipulated
// to efficiently select the nodes and disk for dynamic pool provisioning.
type nodeDisk struct {
	// diskCount is the count of usable disks that can be used in storagepool provisioning.
	diskCount int
	//diskList is the list of usable disks that can be used in storagepool provisioning.
	diskList []string
}

// spareAllotment holds the value for the remaining node allotments
// for pool provisioning.
var spareAllotment int16

// getDiskList is a wrapper function which will receive list of disks
// that can be used for dynamic pool provisioning at runtime.
// The function finally returns the disk list to the caller (i.e. getCasPoolDisk function).
func getDiskList(cp *v1alpha1.CasPool) ([]string, error) {

	// Get kubernetes clientset
	// namespaces is not required, hence passed empty.
	newK8sClient, err := k8s.NewK8sClient("")

	if err != nil {
		return nil, err
	}
	// Get openebs clientset using a getter method (i.e. GetOECS() ) as
	// the openebs clientset is not exported.
	newOecsClient := newK8sClient.GetOECS()

	// Create instance of clientset struct defined above which binds
	// ListDisk method and fill it with openebs clienset (i.e.newOecsClient ).
	newClientSet := clientSet{
		oecs: newOecsClient,
	}

	// if no minimum pools were specified it will default to 1.
	if cp.MinPools <= 0 {
		glog.Warning("invalid or 0 min pool specified, defaulting to 1")
		cp.MinPools = 1
	}

	// nodeDiskAlloter will try to return a list of disks so that maxpool number of storagepool
	// is provisioned.
	diskList, err := newClientSet.nodeDiskAlloter(cp)
	if err != nil {
		return nil, err
	}

	return diskList, nil
}

// nodeDiskAlloter will try to allot nodes for pool creation as specified in
// maxPool field of the storagepoolclaim.

// For exapmle, if maxPool=5 and minPool=3, it will try to search for 5 nodes that will qualify for
// pool provisioning. At least 3 node should qualify else pool will not be provisioned and pool creation
// will be aborted gracefully with proper log messages.

// If no minPool field is present,at least one node must qalify for pool provisioning.

// modeDiskAlloter can be made more intelligent as per the required pool constraints for alloting nodes.
func (k *clientSet) nodeDiskAlloter(cp *v1alpha1.CasPool) ([]string, error) {

	// assign maxPools to spareAllotment as right now maxPool is the number of allotments
	// that needs to be done.
	spareAllotment = cp.MaxPools

	// Request kube-apiserver for the list of disk (powered by NDM)
	// Currently, all the disks are returned,but the disk that is already a part of pool
	// should not be returned.
	listDisk, err := k.oecs.OpenebsV1alpha1().Disks().List(mach_apis_meta_v1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error in getting the disk list:%v", err)
	}
	if listDisk == nil {
		return nil, errors.New("no disk object found")
	}
	nodeDiskMap := nodeSelector(listDisk, cp.PoolType)
	gotAllotment := cp.MaxPools - spareAllotment
	if spareAllotment > cp.MaxPools-cp.MinPools {
		return nil, fmt.Errorf("no node qualified for pool:only %d node could be alloted but required is %d", gotAllotment, cp.MinPools)
	}
	if gotAllotment < cp.MaxPools {
		glog.Warning("partial node allotment done:spared node allotment:", spareAllotment)
	}
	selectedDisk := diskSelector(nodeDiskMap, cp.PoolType)
	return selectedDisk, nil
}

// nodeSelector function will select candidate nodes that will qualify for storagepool provisioning in accordance
// with the pool constraints.

// NOTE: Not all the selected nodes may qualify.

// Finally diskSelector function will vote for qualified nodes.

func nodeSelector(listDisk *v1alpha1.DiskList, poolType string) map[string]*nodeDisk {

	// nodeDiskMap is the data structure holding host name as key
	// and nodeDisk struct as value
	nodeDiskMap := make(map[string]*nodeDisk)
	for _, value := range listDisk.Items {
		// if no more allotment is required, stop processing
		if spareAllotment == 0 {
			glog.Info("required pool allotment done")
			break
		}
		if nodeDiskMap[value.Labels[string(v1alpha1.HostNameCPK)]] == nil {
			// Entry to this block means first time the hostname will be mapped for the first time.
			// Obviously, this entry of hostname(node) is for a usable disk and initialize diskCount to 1.
			nodeDiskMap[value.Labels[string(v1alpha1.HostNameCPK)]] = &nodeDisk{diskList: []string{value.Name}, diskCount: 1}
			// If pool type is striped the node qualifies for pool creation hence spareAllotment decremented.
			if poolType == string(v1alpha1.PoolTypeStripedCPK) {
				spareAllotment--
			}
		} else {
			// Entry to this block means the hostname was already mapped and it has more than one disk and at least two disks.
			nodeDisk := nodeDiskMap[value.Labels[string(v1alpha1.HostNameCPK)]]
			// Increment the disk count
			nodeDisk.diskCount++
			// Add the current disk to the diskList for this node.
			nodeDisk.diskList = append(nodeDisk.diskList, value.Name)
			// If pool type is mirrored the node qualifies for pool creation hence spareAllotment decremented.
			if poolType == string(v1alpha1.PoolTypeMirroredCPK) {
				if nodeDisk.diskCount == int(v1alpha1.MirroredDiskCountCPK) {
					spareAllotment--
				}
			}
		}

	}
	return nodeDiskMap
}

// diskSelector is the function that will select the required number of disks from qualified nodes
// so as to provision storagepool

func diskSelector(nodeDiskMap map[string]*nodeDisk, poolType string) []string {

	// selectedDisk will hold a list of disk that will be used to provision storage pool, after a
	// minimum number of node qualifies
	var selectedDisk []string

	// requiredDiskCount will hold the required number of disk that should be selcted from a qualified
	// node for specific pool type
	var requiredDiskCount int

	// Range over the nodeDiskMap map to get the list of disks
	for _, val := range nodeDiskMap {
		// If pool type is striped, 1 disk should be selected
		if poolType == string(v1alpha1.PoolTypeStripedCPK) {
			requiredDiskCount = int(v1alpha1.StripedDiskCountCPK)
		}
		// If pool type is striped, 2 disks should be selected
		if poolType == string(v1alpha1.PoolTypeMirroredCPK) {
			requiredDiskCount = int(v1alpha1.MirroredDiskCountCPK)
			// If the current disk count on the node is less than the required disks
			// then this is a dirty node and it will not qualify.
			if len(val.diskList) < requiredDiskCount {
				continue
			}
		}
		// Select the required disk from qualified nodes.
		for i := 0; i < requiredDiskCount; i++ {
			selectedDisk = append(selectedDisk, val.diskList[i])
		}
	}
	return selectedDisk
}
