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

package v1alpha1

import (
	"github.com/golang/glog"
	apisv1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	csp "github.com/openebs/maya/pkg/cstorpool/v1alpha3"
	cspc "github.com/openebs/maya/pkg/cstorpoolcluster/v1alpha1"
	disk "github.com/openebs/maya/pkg/disk/v1alpha2"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetCasPool returns a CasPool object for cspc.
func (op *Operations) GetCasPool() (*apisv1alpha1.CasPool, error) {
	if cspc.IsProvisioningManual()(op.CspcObject) {
		return op.getCasPoolForManualProvisioning()
	}
	if cspc.IsProvisioningAuto()(op.CspcObject) {
		return op.getCasPoolForAutoProvisioning()
	}
	return nil, errors.Errorf("provisioning type not supported for cspc %s", op.CspcObject.Object.Name)
}

// IsPoolCreationPending returns true if pool needs to be created for a given cspc.
func (op *Operations) IsPoolCreationPending() bool {
	count, err := op.GetPendingPoolCount()
	if err != nil {
		glog.Errorf("Could not get pending pool count for cspc %s", op.CspcObject.Object.Name)
		return false
	}
	if count > 0 {
		return true
	}
	return false
}

// getCurrentPoolCount returns the current pool count for the given cspc.
func (op *Operations) getCurrentPoolCount() (int, error) {
	cspcName := op.CspcObject.Object.Name
	cspList, err := op.CspClient.List(metav1.ListOptions{LabelSelector: string(apisv1alpha1.CStorPoolClusterCPK) + "=" + cspcName})
	if err != nil {
		return 0, errors.Wrapf(err, "unable to get csp list for cspc %s", cspcName)
	}
	return csp.ListBuilderForAPIObject(cspList).CspList.Len(), nil
}

// GetPendingPoolCount returns the count the pending pool that should be provisioned for the given cspc.
func (op *Operations) GetPendingPoolCount() (int, error) {
	cspcName := op.CspcObject.Object.Name
	currentCount, err := op.getCurrentPoolCount()
	if err != nil {
		return 0, errors.Wrapf(err, "unable to get current pool count for cspc %s", cspcName)
	}
	if cspc.IsProvisioningAuto()(op.CspcObject) {
		return *op.CspcObject.Object.Spec.MaxPools - currentCount, nil
	}
	return len(op.CspcObject.Object.Spec.Nodes) - currentCount, nil
}

// getUsedDiskMap returns disks which cannot be used for pool provisioning.
func (op *Operations) getUsedDiskMap() (map[string]int, error) {
	cspAPIList, err := op.CspClient.List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get the list of cstor pool for cspc %s", op.CspcObject.Object.Name)
	}
	usedDiskMap := make(map[string]int)
	for _, csp := range cspAPIList.Items {
		for _, group := range csp.Spec.Group {
			for _, disk := range group.Item {
				usedDiskMap[disk.Name]++
			}
		}
	}
	return usedDiskMap, nil
}

// getUsedNode returns nodes where pool cannot be provisioned for a given cspc.
func (op *Operations) getUsedNode() (map[string]bool, error) {
	cspAPIList, err := op.CspClient.List(metav1.ListOptions{LabelSelector: string(apisv1alpha1.CStorPoolClusterCPK) + "=" + op.CspcObject.Object.Name})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get the list of cstor pool for cspc %s", op.CspcObject.Object.Name)
	}
	usedNode := make(map[string]bool)
	for _, cspObject := range cspAPIList.Items {
		// pin it
		cspObject := cspObject
		nodeName := csp.BuilderForAPIObject(&cspObject).Csp.GetNodeName()
		usedNode[nodeName] = true
	}
	return usedNode, nil
}

// getDiskDeviceIDMap returns a map of disk to its device id.
func (op *Operations) getDiskDeviceIDMap() (map[string]string, error) {
	diskDeviceIDMap := make(map[string]string)
	for _, node := range op.CspcObject.Object.Spec.Nodes {
		for _, group := range node.DiskGroups {
			for _, diskDetails := range group.Disks {
				diskObject, err := op.DiskClient.Get(diskDetails.Name, metav1.GetOptions{})
				if err != nil {
					return nil, errors.Wrapf(err, "could not get device ID for disk %s", diskDetails.Name)
				}
				deviceID := disk.BuilderForAPIObject(diskObject).Disk.GetDeviceID()
				if deviceID == "" {
					return nil, errors.Errorf("got empty device ID for disk %s", diskDetails.Name)
				}
				diskDeviceIDMap[diskDetails.Name] = deviceID
			}
		}
	}
	return diskDeviceIDMap, nil
}

// getDiskDeviceIDMapForDiskAPIList returns a map of disk to its device id.
func (op *Operations) getDiskDeviceIDMapForDiskAPIList(diskList *apisv1alpha1.DiskList) (map[string]string, error) {
	diskDeviceIDMap := make(map[string]string)
	for _, diskObject := range diskList.Items {
		//pin it
		diskObject := diskObject
		deviceID := disk.BuilderForAPIObject(&diskObject).Disk.GetDeviceID()
		if deviceID == "" {
			return nil, errors.Errorf("got empty device ID for disk %s", diskObject.Name)
		}
		diskDeviceIDMap[diskObject.Name] = deviceID
	}
	return diskDeviceIDMap, nil
}
