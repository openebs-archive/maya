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
	ndm "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
)

// BlockDeviceClaim encapsulates BlockDeviceClaim api object.
type BlockDeviceClaim struct {
	// actual block device claim object
	Object *ndm.BlockDeviceClaim
}

// BlockDeviceClaimList encapsulates BlockDeviceClaimList api object
type BlockDeviceClaimList struct {
	// list of blockdeviceclaims
	ObjectList *ndm.BlockDeviceClaimList
}

// Predicate defines an abstraction to determine conditional checks against the
// provided block device claim instance
type Predicate func(*BlockDeviceClaim) bool

// PredicateList holds the list of Predicates
type PredicateList []Predicate

// all returns true if all the predicates succeed against the provided block
// device instance.
func (l PredicateList) all(c *BlockDeviceClaim) bool {
	for _, pred := range l {
		if !pred(c) {
			return false
		}
	}
	return true
}

// HasAnnotation is predicate to filter out based on
// annotation in BDC instances
func HasAnnotation(key, value string) Predicate {
	return func(bdc *BlockDeviceClaim) bool {
		return bdc.HasAnnotation(key, value)
	}
}

// HasAnnotation return true if provided annotation
// key and value are present in the the provided BDCList
// instance
func (bdc *BlockDeviceClaim) HasAnnotation(key, value string) bool {
	val, ok := bdc.Object.GetAnnotations()[key]
	if ok {
		return val == value
	}
	return false
}

// HasLabel is predicate to filter out labeled
// BDC instances
func HasLabel(key, value string) Predicate {
	return func(bdc *BlockDeviceClaim) bool {
		return bdc.HasLabel(key, value)
	}
}

// HasLabel returns true if provided label
// key and value are present in the provided BDC(BlockDeviceClaim)
// instance
func (bdc *BlockDeviceClaim) HasLabel(key, value string) bool {
	val, ok := bdc.Object.GetLabels()[key]
	if ok {
		return val == value
	}
	return false
}

// IsStatus is predicate to filter out BDC instances based on argument provided
func IsStatus(status string) Predicate {
	return func(bdc *BlockDeviceClaim) bool {
		return bdc.IsStatus(status)
	}
}

// IsStatus returns true if the status on
// block device claim matches with provided status.
func (bdc *BlockDeviceClaim) IsStatus(status string) bool {
	return string(bdc.Object.Status.Phase) == status
}

// Len returns the length og BlockDeviceClaimList.
func (bdcl *BlockDeviceClaimList) Len() int {
	return len(bdcl.ObjectList.Items)
}

// GetBlockDeviceNamesByNode returns map of node name and corresponding block devices to that
// node from block device claim list
func (bdcl *BlockDeviceClaimList) GetBlockDeviceNamesByNode() map[string][]string {
	newNodeBDList := make(map[string][]string)
	if bdcl == nil {
		return newNodeBDList
	}
	for _, bdc := range bdcl.ObjectList.Items {
		newNodeBDList[bdc.Spec.HostName] = append(newNodeBDList[bdc.Spec.HostName], bdc.Spec.BlockDeviceName)
	}
	return newNodeBDList
}
