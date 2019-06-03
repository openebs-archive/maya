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
	ndmapisv1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
)

//TODO: While using these packages UnitTest must be written to corresponding function

// BlockDeviceClaim encapsulates BlockDeviceClaim api object.
type BlockDeviceClaim struct {
	// actual block device claim object
	Object *ndmapisv1alpha1.BlockDeviceClaim
}

// BlockDeviceClaimList encapsulates BlockDeviceClaimList api object
type BlockDeviceClaimList struct {
	// list of blockdeviceclaims
	ObjectList *ndmapisv1alpha1.BlockDeviceClaimList
}

// Predicate defines an abstraction to determine conditional checks against the
// provided block device claim instance
type Predicate func(*BlockDeviceClaim) bool

// predicateList holds the list of Predicates
type predicateList []Predicate

// all returns true if all the predicates succeed against the provided block
// device instance.
func (l predicateList) all(c *BlockDeviceClaim) bool {
	for _, pred := range l {
		if !pred(c) {
			return false
		}
	}
	return true
}

// HasAnnotation returns true if provided annotation key and value are present
// in the provided block deive instance.
func HasAnnotation(key, value string) Predicate {
	return func(c *BlockDeviceClaim) bool {
		val, ok := c.Object.GetAnnotations()[key]
		if ok {
			return val == value
		}
		return false
	}
}

// HasLabel returns true if provided label
// key and value are present in the provided BDC(BlockDeviceClaim)
// instance
func HasLabel(key, value string) Predicate {
	return func(c *BlockDeviceClaim) bool {
		val, ok := c.Object.GetLabels()[key]
		if ok {
			return val == value
		}
		return false
	}
}

// IsStatus returns true if the status on block device claim matches with provided status.
func IsStatus(status string) Predicate {
	return func(c *BlockDeviceClaim) bool {
		val := c.Object.Status.Phase
		return string(val) == status
	}
}

// Filter will filter the BDC instances if all the predicates succeed
// against that BDC.
func (l *BlockDeviceClaimList) Filter(p ...Predicate) *BlockDeviceClaimList {
	var plist predicateList
	plist = append(plist, p...)
	if len(plist) == 0 {
		return l
	}

	filtered := NewListBuilder().List()
	for _, bdcAPI := range l.ObjectList.Items {
		bdcAPI := bdcAPI // pin it
		BlockDeviceClaim := BuilderForAPIObject(&bdcAPI).BDC
		if plist.all(BlockDeviceClaim) {
			filtered.ObjectList.Items = append(filtered.ObjectList.Items, *BlockDeviceClaim.Object)
		}
	}
	return filtered
}

// Len returns the length og BlockDeviceClaimList.
func (l *BlockDeviceClaimList) Len() int {
	return len(l.ObjectList.Items)
}
