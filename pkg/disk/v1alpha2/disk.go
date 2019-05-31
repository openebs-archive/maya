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

// DiskState is label for disk states
type DiskState string

const (
	// DiskStateActive is active state of the disk.
	DiskStateActive DiskState = "Active"
)

// DefaultDiskCount is a map containing the default disk count of various raid types.
var DefaultDiskCount = map[string]int{
	string(apisv1alpha1.PoolTypeMirroredCPV): int(apisv1alpha1.MirroredDiskCountCPV),
	string(apisv1alpha1.PoolTypeStripedCPV):  int(apisv1alpha1.StripedDiskCountCPV),
	string(apisv1alpha1.PoolTypeRaidzCPV):    int(apisv1alpha1.RaidzDiskCountCPV),
	string(apisv1alpha1.PoolTypeRaidz2CPV):   int(apisv1alpha1.Raidz2DiskCountCPV),
}

// Disk encapsulates Disk api object.
type Disk struct {
	// actual disk object
	Object *ndmapisv1alpha1.Disk
}

// DiskList holds the list of Disk api
type DiskList struct {
	// list of disk
	ObjectList *ndmapisv1alpha1.DiskList
}

// Builder is the builder object for Disk.
type Builder struct {
	Disk *Disk
}

// ListBuilder is the builder object for DiskList.
type ListBuilder struct {
	DiskList *DiskList
}

// Predicate defines an abstraction to determine conditional checks against the provided disk instance.
type Predicate func(*Disk) bool

type predicateList []Predicate

// all returns true if all the predicates succeed against the provided disk instance.
func (l predicateList) all(c *Disk) bool {
	for _, pred := range l {
		if !pred(c) {
			return false
		}
	}
	return true
}

// HasAnnotation returns true if provided annotation key and value are present in the provided disk instance.
func HasAnnotation(key, value string) Predicate {
	return func(c *Disk) bool {
		val, ok := c.Object.GetAnnotations()[key]
		if ok {
			return val == value
		}
		return false
	}
}

// IsSparse returns true if the disk is of sparse type.
func IsSparse() Predicate {
	return func(d *Disk) bool {
		return d.Object.GetLabels()[string(apisv1alpha1.NdmDiskTypeCPK)] == string(apisv1alpha1.TypeSparseCPV)
	}
}

// IsActive returns true if the disk is active.
func IsActive() Predicate {
	return func(d *Disk) bool {
		return d.Object.Status.State == string(DiskStateActive)
	}
}

// IsUsable returns true if this disk can be used for pool provisioning.
// The argument usedDisks s a map containing key as disk cr name and value as integer.
// If the value of map is greater than 0 , then this corresponding disk is not usable.
func IsUsable(usedDisks map[string]int) Predicate {
	return func(d *Disk) bool {
		return usedDisks[d.Object.Name] == 0
	}
}

// IsUsableNode returns true if disk of this node can be used for pool provisioning.
// The argument usedNodes s a map containing key as node name and value as bool.
// If the value of map is greater than false, then this corresponding node is not usable.
func IsUsableNode(usedNodes map[string]bool) Predicate {
	return func(d *Disk) bool {
		return !usedNodes[d.GetNodeName()]
	}
}

// IsBelongToNode returns true if the disk belongs to the provided node.
func IsBelongToNode(nodeName string) Predicate {
	return func(d *Disk) bool {
		return d.GetNodeName() == nodeName
	}
}

// IsType returns true if the disk is of type same as passed argument
func IsType(diskType string) Predicate {
	return func(d *Disk) bool {
		return d.Object.GetLabels()[string(apisv1alpha1.NdmDiskTypeCPK)] == diskType
	}
}

// IsValidPoolTopology returns true if the topology is valid.
func IsValidPoolTopology(poolType string, diskCount int) bool {
	return DefaultDiskCount[poolType]%diskCount == 0
}

// IsDisk returns true if the disk is of disk type.
func IsDisk() Predicate {
	return func(d *Disk) bool {
		return d.Object.GetLabels()[string(apisv1alpha1.NdmDiskTypeCPK)] == string(apisv1alpha1.TypeDiskCPV)
	}
}

// GetNodeName returns the node name to which the disk is attached
func (d *Disk) GetNodeName() string {
	return d.Object.GetLabels()[string(apisv1alpha1.HostNameCPK)]
}

// Filter will filter the disk instances if all the predicates succeed against that disk.
func (l *DiskList) Filter(p ...Predicate) *DiskList {
	var plist predicateList
	plist = append(plist, p...)
	if len(plist) == 0 {
		return l
	}

	filtered := NewListBuilder().List()
	for _, diskAPI := range l.ObjectList.Items {
		diskAPI := diskAPI // pin it
		Disk := BuilderForAPIObject(&diskAPI).Disk
		if plist.all(Disk) {
			filtered.ObjectList.Items = append(filtered.ObjectList.Items, *Disk.Object)
		}
	}
	return filtered
}

// NewBuilder returns an empty instance of the Builder object.
func NewBuilder() *Builder {
	return &Builder{
		Disk: &Disk{&ndmapisv1alpha1.Disk{}},
	}
}

// BuilderForObject returns an instance of the Builder object based on disk object
func BuilderForObject(Disk *Disk) *Builder {
	return &Builder{
		Disk: Disk,
	}
}

// BuilderForAPIObject returns an instance of the Builder object based on disk api object.
func BuilderForAPIObject(disk *ndmapisv1alpha1.Disk) *Builder {
	return &Builder{
		Disk: &Disk{disk},
	}
}

// GetDeviceID returns the device link of the disk. If device link is not found it returns device path.
// For a cstor pool creation -- this link or path is used. For convenience, we call it as device ID.
// Hence, device ID can either be a  device link or device path depending on what was available in disk cr.
func (d *Disk) GetDeviceID() string {
	var deviceID string
	if len(d.Object.Spec.DevLinks) != 0 && len(d.Object.Spec.DevLinks[0].Links) != 0 {
		deviceID = d.Object.Spec.DevLinks[0].Links[0]
	} else {
		deviceID = d.Object.Spec.Path
	}
	return deviceID
}

// Build returns the Disk object built by this builder.
func (sb *Builder) Build() *Disk {
	return sb.Disk
}

// NewListBuilder returns a new instance of ListBuilder object.
func NewListBuilder() *ListBuilder {
	return &ListBuilder{DiskList: &DiskList{ObjectList: &ndmapisv1alpha1.DiskList{}}}
}

// WithList builds the list based on the provided *DiskList instances.
func (b *ListBuilder) WithList(disks *DiskList) *ListBuilder {
	if disks == nil {
		return b
	}
	b.DiskList.ObjectList.Items = append(b.DiskList.ObjectList.Items, disks.ObjectList.Items...)
	return b
}

// WithAPIList builds the list based on the provided *apisv1alpha1.CStorPoolList.
func (b *ListBuilder) WithAPIList(disks *ndmapisv1alpha1.DiskList) *ListBuilder {
	if disks == nil {
		return b
	}
	b.DiskList.ObjectList.Items = append(b.DiskList.ObjectList.Items, disks.Items...)
	return b
}

// List returns the list of disk instances that were built by this builder.
func (b *ListBuilder) List() *DiskList {
	return b.DiskList
}

// NewListBuilderForObjectList builds the list based on the provided *DiskList instances.
func NewListBuilderForObjectList(disk *DiskList) *ListBuilder {
	newLB := NewListBuilder()
	newLB.DiskList.ObjectList.Items = append(newLB.DiskList.ObjectList.Items, disk.ObjectList.Items...)
	return newLB
}

// NewListBuilderForAPIList returns a new instance of ListBuilderForApiList object based on csp api list.
func NewListBuilderForAPIList(diskAPIList *ndmapisv1alpha1.DiskList) *ListBuilder {
	newLb := NewListBuilder()
	newLb.DiskList.ObjectList.Items = append(newLb.DiskList.ObjectList.Items, diskAPIList.Items...)
	return newLb
}

// Len returns the length og DiskList.
func (l *DiskList) Len() int {
	return len(l.ObjectList.Items)
}
