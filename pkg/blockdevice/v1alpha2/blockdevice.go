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

package v1alpha2

import (
	ndmapisv1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
	apisv1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
)

//TODO: While using these packages UnitTest must be written to corresponding function

// BlockDeviceState is label for block device states
type BlockDeviceState string

const (
	// BlockDeviceStateActive is active state of the block device
	BlockDeviceStateActive BlockDeviceState = "Active"
)

// DefaultBlockDeviceCount is a map containing the default block device count of various raid types.
var DefaultBlockDeviceCount = map[string]int{
	string(apisv1alpha1.PoolTypeMirroredCPV): int(apisv1alpha1.MirroredBlockDeviceCountCPV),
	string(apisv1alpha1.PoolTypeStripedCPV):  int(apisv1alpha1.StripedBlockDeviceCountCPV),
	string(apisv1alpha1.PoolTypeRaidzCPV):    int(apisv1alpha1.RaidzBlockDeviceCountCPV),
	string(apisv1alpha1.PoolTypeRaidz2CPV):   int(apisv1alpha1.Raidz2BlockDeviceCountCPV),
}

// BlockDevice encapsulates BlockDevice api object.
type BlockDevice struct {
	// actual block device object
	Object *ndmapisv1alpha1.BlockDevice
}

// BlockDeviceList holds the list of BlockDevice api
type BlockDeviceList struct {
	// list of blockdevices
	ObjectList *ndmapisv1alpha1.BlockDeviceList
}

// Predicate defines an abstraction to determine conditional checks against the
// provided block device instance
type Predicate func(*BlockDevice) bool

// predicateList holds the list of Predicates
type predicateList []Predicate

// all returns true if all the predicates succeed against the provided block
// device instance.
func (l predicateList) all(c *BlockDevice) bool {
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
	return func(c *BlockDevice) bool {
		val, ok := c.Object.GetAnnotations()[key]
		if ok {
			return val == value
		}
		return false
	}
}

// TODO: fix below snippet when #1227 merged
//// IsSparse returns true if the block device is of sparse type.
//func IsSparse() Predicate {
//	return func(d *BlockDevice) bool {
//		return d.Object.GetLabels()[string(apisv1alpha1.NdmBlockDeviceTypeCPK)] == string(apisv1alpha1.TypeBlockDeviceCPV)
//	}
//}

// IsType returns true if the block device is of type same as passed argument
//func IsType(bdType string) Predicate {
//	return func(d *BlockDevice) bool {
//		return d.Object.GetLabels()[string(apisv1alpha1.NdmBlockDeviceTypeCPK)] == bdType
//	}
//}

// IsActive returns true if the block device is active.
func IsActive() Predicate {
	return func(bd *BlockDevice) bool {
		return bd.Object.Status.State == string(BlockDeviceStateActive)
	}
}

// IsUsable returns true if this block device can be used for pool provisioning.
// The argument usedBlockDevice is a map containing key as block device cr name and value as integer.
// If the value of map is greater than 0 , then this corresponding block device is not usable.
func IsUsable(usedBlockDevices map[string]int) Predicate {
	return func(bd *BlockDevice) bool {
		return usedBlockDevices[bd.Object.Name] == 0
	}
}

// IsUsableNode returns true if block device of this node can be used for pool provisioning.
// The argument usedNodes is a map containing key as node name and value as bool.
// If the value of map is greater than false, then this corresponding node is not usable.
func IsUsableNode(usedNodes map[string]bool) Predicate {
	return func(bd *BlockDevice) bool {
		return !usedNodes[bd.GetNodeName()]
	}
}

// IsBelongToNode returns true if the block device belongs to the provided node.
func IsBelongToNode(nodeName string) Predicate {
	return func(bd *BlockDevice) bool {
		return bd.GetNodeName() == nodeName
	}
}

// IsValidPoolTopology returns true if the topology is valid.
func IsValidPoolTopology(poolType string, diskCount int) bool {
	return DefaultBlockDeviceCount[poolType]%diskCount == 0
}

// GetNodeName returns the node name to which the block device is attached
func (bd *BlockDevice) GetNodeName() string {
	return bd.Object.GetLabels()[string(apisv1alpha1.HostNameCPK)]
}

// Filter will filter the block device instances if all the predicates succeed
// against that block device.
func (l *BlockDeviceList) Filter(p ...Predicate) *BlockDeviceList {
	var plist predicateList
	plist = append(plist, p...)
	if len(plist) == 0 {
		return l
	}

	filtered := NewListBuilder().List()
	for _, bdAPI := range l.ObjectList.Items {
		bdAPI := bdAPI // pin it
		BlockDevice := BuilderForAPIObject(&bdAPI).BlockDevice
		if plist.all(BlockDevice) {
			filtered.ObjectList.Items = append(filtered.ObjectList.Items, *BlockDevice.Object)
		}
	}
	return filtered
}

// GetDeviceID returns the device link of the block device. If device link is not found it returns device path.
// For a cstor pool creation -- this link or path is used. For convenience, we call it as device ID.
// Hence, device ID can either be a  device link or device path depending on
// what was available in block device cr.
func (bd *BlockDevice) GetDeviceID() string {
	var deviceID string
	if len(bd.Object.Spec.DevLinks) != 0 && len(bd.Object.Spec.DevLinks[0].Links) != 0 {
		deviceID = bd.Object.Spec.DevLinks[0].Links[0]
	} else {
		deviceID = bd.Object.Spec.Path
	}
	return deviceID
}

// Len returns the length og BlockDeviceList.
func (l *BlockDeviceList) Len() int {
	return len(l.ObjectList.Items)
}
