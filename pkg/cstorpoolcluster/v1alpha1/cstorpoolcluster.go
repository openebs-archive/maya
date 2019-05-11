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
	apisv1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
)

// CSPC encapsulates CStorPoolCluster api object.
type CSPC struct {
	// actual cspc object
	Object *apisv1alpha1.CStorPoolCluster
}

// CSPCList holds the list of CStorPoolCluster api
type CSPCList struct {
	// list of cstorpoolclusters
	ObjectList *apisv1alpha1.CStorPoolClusterList
}

// Builder is the builder object for CSPC.
type Builder struct {
	CSPC *CSPC
}

// ListBuilder is the builder object for CSPCList.
type ListBuilder struct {
	CSPCList *CSPCList
}

// Predicate defines an abstraction to determine conditional checks against the provided cspc instance.
type Predicate func(*CSPC) bool

type predicateList []Predicate

// all returns true if all the predicates succeed against the provided csp instance.
func (l predicateList) all(c *CSPC) bool {
	for _, pred := range l {
		if !pred(c) {
			return false
		}
	}
	return true
}

// HasAnnotation returns true if provided annotation key and value are present in the provided cspc instance.
func HasAnnotation(key, value string) Predicate {
	return func(c *CSPC) bool {
		val, ok := c.Object.GetAnnotations()[key]
		if ok {
			return val == value
		}
		return false
	}
}

// IsProvisioningAuto returns true if the cspc is of auto provisioning type.
func IsProvisioningAuto() Predicate {
	return func(c *CSPC) bool {
		return len(c.Object.Spec.Nodes) == 0
	}
}

// IsProvisioningManual returns true if the cspc is of manual provisioning type.
func IsProvisioningManual() Predicate {
	return func(c *CSPC) bool {
		return len(c.Object.Spec.Nodes) != 0
	}
}

// IsSparse returns true if the cspc is of sparse type.
func IsSparse() Predicate {
	return func(c *CSPC) bool {
		return c.IsSparse()
	}
}

// IsSparse returns true if type is sparse
func (c *CSPC) IsSparse() bool {
	return c.Object.Spec.Type == string(apisv1alpha1.TypeSparseCPV)
}

// IsDisk returns true if the cspc is of disk type.
func IsDisk() Predicate {
	return func(c *CSPC) bool {
		return c.IsDisk()
	}
}

// IsDisk returns true if type is disk
func (c *CSPC) IsDisk() bool {
	return c.Object.Spec.Type == string(apisv1alpha1.TypeDiskCPV)
}

// IsDiskRepeated returns predicate for validation of duplicate disk entry.
func IsDiskRepeated() Predicate {
	return func(c *CSPC) bool {
		return c.IsDiskRepeated()
	}
}

// IsDiskRepeated returns true if the cspc contains duplicate disk entries.
func (c *CSPC) IsDiskRepeated() bool {
	diskCount := make(map[string]int)
	for _, node := range c.Object.Spec.Nodes {
		for _, diskGroup := range node.DiskGroups {
			for _, disk := range diskGroup.Disks {
				diskCount[disk.Name]++
				if diskCount[disk.Name] > 1 {
					return true
				}
			}
		}
	}
	return false
}

// IsPoolTypeStriped returns true if pool type is striped.
func (c *CSPC) IsPoolTypeStriped() bool {
	return c.Object.Spec.PoolSpec.PoolType == string(apisv1alpha1.PoolTypeStripedCPV)
}

// GetNodeNames returns a list of node names present in cspc
func (c *CSPC) GetNodeNames() []string {
	var nodenames []string
	nodes := c.Object.Spec.Nodes
	for _, val := range nodes {
		nodenames = append(nodenames, val.Name)
	}
	return nodenames
}

// GetDiskGroupListForNode returns a list of disk present in cspc for the specified  node
func (c *CSPC) GetDiskGroupListForNode(nodeName string) []apisv1alpha1.CStorPoolClusterDiskGroups {
	nodes := c.Object.Spec.Nodes
	for _, val := range nodes {
		if val.Name == nodeName {
			return val.DiskGroups
		}
	}
	return []apisv1alpha1.CStorPoolClusterDiskGroups{}
}

// GetPoolTypeForNode returns poolType for the node in cspc.
func (c *CSPC) GetPoolTypeForNode(nodeName string) string {
	nodes := c.Object.Spec.Nodes
	for _, val := range nodes {
		if val.Name == nodeName {
			return string(val.PoolSpec.PoolType)
		}
	}
	return ""
}

// GetPoolType returns poolType for the cspc.
// NOTE: This poolType ( spec.poolSpec.poolType ) is used in auto provisioning.
func (c *CSPC) GetPoolType() string {
	return c.Object.Spec.PoolSpec.PoolType
}

// GetAnnotations returns annotations present in cspc.
func (c *CSPC) GetAnnotations() map[string]string {
	return c.Object.GetAnnotations()
}

// GetCASTName returns a name of cas template from the cspc.
func (c *CSPC) GetCASTName() string {
	return c.Object.GetAnnotations()[string(apisv1alpha1.CreatePoolCASTemplateKey)]
}

// Filter will filter the csp instances if all the predicates succeed against that cspc.
func (l *CSPCList) Filter(p ...Predicate) *CSPCList {
	var plist predicateList
	plist = append(plist, p...)
	if len(plist) == 0 {
		return l
	}

	filtered := NewListBuilder().List()
	for _, cspcAPI := range l.ObjectList.Items {
		cspcAPI := cspcAPI // pin it
		CSPC := BuilderForAPIObject(&cspcAPI).CSPC
		if plist.all(CSPC) {
			filtered.ObjectList.Items = append(filtered.ObjectList.Items, *CSPC.Object)
		}
	}
	return filtered
}

// NewBuilder returns an empty instance of the Builder object.
func NewBuilder() *Builder {
	return &Builder{
		CSPC: &CSPC{&apisv1alpha1.CStorPoolCluster{}},
	}
}

// BuilderForObject returns an instance of the Builder object based on cspc object
func BuilderForObject(CSPC *CSPC) *Builder {
	return &Builder{
		CSPC: CSPC,
	}
}

// BuilderForAPIObject returns an instance of the Builder object based on cspc api object.
func BuilderForAPIObject(cspc *apisv1alpha1.CStorPoolCluster) *Builder {
	return &Builder{
		CSPC: &CSPC{cspc},
	}
}

// WithName sets the Name field of cspc with provided argument value.
func (sb *Builder) WithName(name string) *Builder {
	sb.CSPC.Object.Name = name
	sb.CSPC.Object.Spec.Name = name
	return sb
}

// WithDiskType sets the Type field of cspc with provided argument value.
func (sb *Builder) WithDiskType(diskType string) *Builder {
	sb.CSPC.Object.Spec.Type = diskType
	return sb
}

// WithPoolType sets the poolType field of cspc with provided argument value.
func (sb *Builder) WithPoolType(poolType string) *Builder {
	sb.CSPC.Object.Spec.PoolSpec.PoolType = poolType
	return sb
}

// WithOverProvisioning sets the OverProvisioning field of cspc with provided argument value.
func (sb *Builder) WithOverProvisioning(val bool) *Builder {
	sb.CSPC.Object.Spec.PoolSpec.OverProvisioning = val
	return sb
}

// WithMaxPool sets the maxpool field of cspc with provided argument value.
func (sb *Builder) WithMaxPool(val int) *Builder {
	maxPool := newInt(val)
	sb.CSPC.Object.Spec.MaxPools = maxPool
	return sb
}

//
func (sb *Builder) WithNode(nodeName string) *Builder {
	newNodespec := &apisv1alpha1.CStorPoolClusterNodeSpec{
		Name: nodeName,
	}
	if len(sb.CSPC.Object.Spec.Nodes) == 0 {
		sb.CSPC.Object.Spec.Nodes = []apisv1alpha1.CStorPoolClusterNodeSpec{*newNodespec}
	} else {
		sb.CSPC.Object.Spec.Nodes = append(sb.CSPC.Object.Spec.Nodes, *newNodespec)
	}

	return sb
}

func (sb *Builder) WithNodeStripedType(nodeName string) *Builder {
	for k, node := range sb.CSPC.Object.Spec.Nodes {
		if node.Name == nodeName {
			sb.CSPC.Object.Spec.Nodes[k].PoolSpec.PoolType = apisv1alpha1.PoolStriped
		}
	}
	return sb
}

func (sb *Builder) WithNodeMirroredType(nodeName string) *Builder {
	for k, node := range sb.CSPC.Object.Spec.Nodes {
		if node.Name == nodeName {
			sb.CSPC.Object.Spec.Nodes[k].PoolSpec.PoolType = apisv1alpha1.PoolMirrored
		}
	}
	return sb
}

func (sb *Builder) WithNodeGroup(nodeName string, diskGroup apisv1alpha1.CStorPoolClusterDiskGroups) *Builder {
	for k, node := range sb.CSPC.Object.Spec.Nodes {
		if node.Name == nodeName {
			sb.CSPC.Object.Spec.Nodes[k].DiskGroups = append(sb.CSPC.Object.Spec.Nodes[k].DiskGroups, diskGroup)
		}
	}
	return sb
}

// newInt returns a pointer to the int value.
func newInt(val int) *int {
	newVal := val
	return &newVal
}

// Build returns the CSPC object built by this builder.
func (sb *Builder) Build() *CSPC {
	return sb.CSPC
}

// NewListBuilder returns a new instance of ListBuilder object.
func NewListBuilder() *ListBuilder {
	return &ListBuilder{CSPCList: &CSPCList{ObjectList: &apisv1alpha1.CStorPoolClusterList{}}}
}

// NewListBuilderForObjectList builds the list based on the provided *CSPCList instances.
func NewListBuilderForObjectList(pools *CSPCList) *ListBuilder {
	newLB := NewListBuilder()
	newLB.CSPCList.ObjectList.Items = append(newLB.CSPCList.ObjectList.Items, pools.ObjectList.Items...)
	return newLB
}

// NewListBuilderForAPIList builds the list based on the provided *apisv1alpha1.CStorPoolList.
func NewListBuilderForAPIList(pools *apisv1alpha1.CStorPoolClusterList) *ListBuilder {
	newLB := NewListBuilder()
	for _, pool := range pools.Items {
		pool := pool //pin it
		newLB.CSPCList.ObjectList.Items = append(newLB.CSPCList.ObjectList.Items, pool)
	}
	return newLB
}

// List returns the list of csp instances that were built by this builder.
func (b *ListBuilder) List() *CSPCList {
	return b.CSPCList
}

// Len returns the length og CSPCList.
func (l *CSPCList) Len() int {
	return len(l.ObjectList.Items)
}

// IsEmpty returns false if the CSPCList is empty.
func (l *CSPCList) IsEmpty() bool {
	return len(l.ObjectList.Items) == 0
}
