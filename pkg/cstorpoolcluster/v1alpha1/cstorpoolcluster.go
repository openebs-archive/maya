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

package v1beta1

import (
	apisv1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
)

// SPC encapsulates CStorPoolCluster api object.
type SPC struct {
	// actual spc object
	Object *apisv1alpha1.CStorPoolCluster
}

// SPCList holds the list of CStorPoolCluster api
type SPCList struct {
	// list of cstorpoolclusters
	ObjectList *apisv1alpha1.CStorPoolClusterList
}

// Builder is the builder object for SPC.
type Builder struct {
	SPC *SPC
}

// ListBuilder is the builder object for SPCList.
type ListBuilder struct {
	SPCList *SPCList
}

// Predicate defines an abstraction to determine conditional checks against the provided spc instance.
type Predicate func(*SPC) bool

type predicateList []Predicate

// all returns true if all the predicates succeed against the provided csp instance.
func (l predicateList) all(c *SPC) bool {
	for _, pred := range l {
		if !pred(c) {
			return false
		}
	}
	return true
}

// HasAnnotation returns true if provided annotation key and value are present in the provided spc instance.
func HasAnnotation(key, value string) Predicate {
	return func(c *SPC) bool {
		val, ok := c.Object.GetAnnotations()[key]
		if ok {
			return val == value
		}
		return false
	}
}

// IsProvisioningAuto returns true if the spc is of auto provisioning type.
func IsProvisioningAuto() Predicate {
	return func(c *SPC) bool {
		return len(c.Object.Spec.Nodes) == 0
	}
}

// IsProvisioningManual returns true if the spc is of manual provisioning type.
func IsProvisioningManual() Predicate {
	return func(c *SPC) bool {
		return len(c.Object.Spec.Nodes) != 0
	}
}

// IsSparse returns true if the spc is of sparse type.
func IsSparse() Predicate {
	return func(c *SPC) bool {
		return c.Object.Spec.Type == "sparse"
	}
}

// IsDisk returns true if the spc is of disk type.
func IsDisk() Predicate {
	return func(c *SPC) bool {
		return c.Object.Spec.Type == "disk"
	}
}

// GetNodeNames returns a list of node names present in spc
func (s *SPC) GetNodeNames() []string {
	var nodenames []string
	nodes := s.Object.Spec.Nodes
	for _, val := range nodes {
		nodenames = append(nodenames, val.Name)
	}
	return nodenames
}

// GetDiskGroupListForNode returns a list of disk present in spc for the specified  node
func (s *SPC) GetDiskGroupListForNode(nodeName string) []apisv1alpha1.CStorPoolClusterDiskGroups {
	nodes := s.Object.Spec.Nodes
	for _, val := range nodes {
		if val.Name == nodeName {
			return val.DiskGroups
		}
	}
	return []apisv1alpha1.CStorPoolClusterDiskGroups{}
}

// GetPoolTypeForNode returns poolType for the node in spc.
func (s *SPC) GetPoolTypeForNode(nodeName string) string {
	nodes := s.Object.Spec.Nodes
	for _, val := range nodes {
		if val.Name == nodeName {
			return val.PoolSpec.PoolType
		}
	}
	return ""
}

// GetAnnotations returns annotations present in spc.
func (s *SPC) GetAnnotations() map[string]string {
	return s.Object.GetAnnotations()
}

// GetCASTName returns a name of cas template from the spc.
func (s *SPC) GetCASTName() string {
	return s.Object.GetAnnotations()[string(apisv1alpha1.CreatePoolCASTemplateKey)]
}

// Filter will filter the csp instances if all the predicates succeed against that spc.
func (l *SPCList) Filter(p ...Predicate) *SPCList {
	var plist predicateList
	plist = append(plist, p...)
	if len(plist) == 0 {
		return l
	}

	filtered := NewListBuilder().List()
	for _, spcAPI := range l.ObjectList.Items {
		spcAPI := spcAPI // pin it
		SPC := BuilderForAPIObject(&spcAPI).SPC
		if plist.all(SPC) {
			filtered.ObjectList.Items = append(filtered.ObjectList.Items, *SPC.Object)
		}
	}
	return filtered
}

// NewBuilder returns an empty instance of the Builder object.
func NewBuilder() *Builder {
	return &Builder{
		SPC: &SPC{&apisv1alpha1.CStorPoolCluster{}},
	}
}

// BuilderForObject returns an instance of the Builder object based on spc object
func BuilderForObject(SPC *SPC) *Builder {
	return &Builder{
		SPC: SPC,
	}
}

// BuilderForAPIObject returns an instance of the Builder object based on spc api object.
func BuilderForAPIObject(spc *apisv1alpha1.CStorPoolCluster) *Builder {
	return &Builder{
		SPC: &SPC{spc},
	}
}

// WithName sets the Name field of spc with provided argument value.
func (sb *Builder) WithName(name string) *Builder {
	sb.SPC.Object.Name = name
	sb.SPC.Object.Spec.Name = name
	return sb
}

// WithDiskType sets the Type field of spc with provided argument value.
func (sb *Builder) WithDiskType(diskType string) *Builder {
	sb.SPC.Object.Spec.Type = diskType
	return sb
}

// WithPoolType sets the poolType field of spc with provided argument value.
func (sb *Builder) WithPoolType(poolType string) *Builder {
	sb.SPC.Object.Spec.PoolSpec.PoolType = poolType
	return sb
}

// WithOverProvisioning sets the OverProvisioning field of spc with provided argument value.
func (sb *Builder) WithOverProvisioning(val bool) *Builder {
	sb.SPC.Object.Spec.PoolSpec.OverProvisioning = val
	return sb
}

// WithMaxPool sets the maxpool field of spc with provided argument value.
func (sb *Builder) WithMaxPool(val int) *Builder {
	maxPool := newInt(val)
	sb.SPC.Object.Spec.MaxPools = maxPool
	return sb
}

// newInt returns a pointer to the int value.
func newInt(val int) *int {
	newVal := val
	return &newVal
}

// Build returns the SPC object built by this builder.
func (sb *Builder) Build() *SPC {
	return sb.SPC
}

// NewListBuilder returns a new instance of ListBuilder object.
func NewListBuilder() *ListBuilder {
	return &ListBuilder{SPCList: &SPCList{ObjectList: &apisv1alpha1.CStorPoolClusterList{}}}
}

// NewListBuilderForObjectList builds the list based on the provided *SPCList instances.
func NewListBuilderForObjectList(pools *SPCList) *ListBuilder {
	newLB := NewListBuilder()
	newLB.SPCList.ObjectList.Items = append(newLB.SPCList.ObjectList.Items, pools.ObjectList.Items...)
	return newLB
}

// NewListBuilderForAPIList builds the list based on the provided *apisv1alpha1.CStorPoolList.
func NewListBuilderForAPIList(pools *apisv1alpha1.CStorPoolClusterList) *ListBuilder {
	newLB := NewListBuilder()
	for _, pool := range pools.Items {
		pool := pool //pin it
		newLB.SPCList.ObjectList.Items = append(newLB.SPCList.ObjectList.Items, pool)
	}
	return newLB
}

// List returns the list of csp instances that were built by this builder.
func (b *ListBuilder) List() *SPCList {
	return b.SPCList
}

// Len returns the length og SPCList.
func (l *SPCList) Len() int {
	return len(l.ObjectList.Items)
}

// IsEmpty returns false if the SPCList is empty.
func (l *SPCList) IsEmpty() bool {
	return len(l.ObjectList.Items) == 0
}
