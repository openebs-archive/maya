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
	"fmt"
	"time"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
)

// SupportedDiskTypes is a map containing the valid disk type
var SupportedDiskTypes = map[apis.CasPoolValString]bool{
	apis.TypeSparseCPV: true,
	apis.TypeDiskCPV:   true,
}

// SPC encapsulates StoragePoolClaim api object.
type SPC struct {
	// actual spc object
	Object *apis.StoragePoolClaim
}

// SPCList holds the list of StoragePoolClaim api
type SPCList struct {
	// list of storagepoolclaims
	ObjectList *apis.StoragePoolClaimList
}

// Builder is the builder object for SPC.
type Builder struct {
	Spc *SPC
}

// ListBuilder is the builder object for SPCList.
type ListBuilder struct {
	SpcList *SPCList
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
		SPC := BuilderForAPIObject(&spcAPI).Spc
		if plist.all(SPC) {
			filtered.ObjectList.Items = append(filtered.ObjectList.Items, *SPC.Object)
		}
	}
	return filtered
}

// NewBuilder returns an empty instance of the Builder object.
func NewBuilder() *Builder {
	return &Builder{
		Spc: &SPC{&apis.StoragePoolClaim{}},
	}
}

// BuilderForObject returns an instance of the Builder object based on spc object
func BuilderForObject(SPC *SPC) *Builder {
	return &Builder{
		Spc: SPC,
	}
}

// BuilderForAPIObject returns an instance of the Builder object based on spc api object.
func BuilderForAPIObject(spc *apis.StoragePoolClaim) *Builder {
	return &Builder{
		Spc: &SPC{spc},
	}
}

// WithName sets the Name field of spc with provided argument value.
func (sb *Builder) WithName(name string) *Builder {
	sb.Spc.Object.Name = name
	sb.Spc.Object.Spec.Name = name
	return sb
}

// WithGenerateName appends a random string after the name
func (sb *Builder) WithGenerateName(name string) *Builder {
	name = name + "-" + fmt.Sprintf("%d", time.Now().UnixNano())
	return sb.WithName(name)
}

// WithDiskType sets the Type field of spc with provided argument value.
func (sb *Builder) WithDiskType(diskType string) *Builder {
	sb.Spc.Object.Spec.Type = diskType
	return sb
}

// WithPoolType sets the poolType field of spc with provided argument value.
func (sb *Builder) WithPoolType(poolType string) *Builder {
	sb.Spc.Object.Spec.PoolSpec.PoolType = poolType
	return sb
}

// WithOverProvisioning sets the OverProvisioning field of spc with provided argument value.
func (sb *Builder) WithOverProvisioning(val bool) *Builder {
	sb.Spc.Object.Spec.PoolSpec.OverProvisioning = val
	return sb
}

// WithPool sets the poolType field of spc with provided argument value.
func (sb *Builder) WithPool(poolType string) *Builder {
	sb.Spc.Object.Spec.PoolSpec.PoolType = poolType
	return sb
}

// WithMaxPool sets the maxpool field of spc with provided argument value.
func (sb *Builder) WithMaxPool(val int) *Builder {
	maxPool := newInt(val)
	sb.Spc.Object.Spec.MaxPools = maxPool
	return sb
}

// newInt returns a pointer to the int value.
func newInt(val int) *int {
	newVal := val
	return &newVal
}

// Build returns the SPC object built by this builder.
func (sb *Builder) Build() *SPC {
	return sb.Spc
}

// NewListBuilder returns a new instance of ListBuilder object.
func NewListBuilder() *ListBuilder {
	return &ListBuilder{SpcList: &SPCList{ObjectList: &apis.StoragePoolClaimList{}}}
}

// WithUIDs builds a list of StoragePoolClaims based on the provided pool UIDs
func (b *ListBuilder) WithUIDs(poolUIDs ...string) *ListBuilder {
	for _, uid := range poolUIDs {
		obj := &SPC{&apis.StoragePoolClaim{}}
		obj.Object.SetUID(types.UID(uid))
		b.SpcList.ObjectList.Items = append(b.SpcList.ObjectList.Items, *obj.Object)
	}
	return b
}

// WithList builds the list based on the provided *SPCList instances.
func (b *ListBuilder) WithList(pools *SPCList) *ListBuilder {
	if pools == nil {
		return b
	}
	b.SpcList.ObjectList.Items = append(b.SpcList.ObjectList.Items, pools.ObjectList.Items...)
	return b
}

// WithAPIList builds the list based on the provided *apis.CStorPoolList.
func (b *ListBuilder) WithAPIList(pools *apis.StoragePoolClaimList) *ListBuilder {
	if pools == nil {
		return b
	}
	for _, pool := range pools.Items {
		pool := pool //pin it
		b.SpcList.ObjectList.Items = append(b.SpcList.ObjectList.Items, pool)
	}
	return b
}

// List returns the list of csp instances that were built by this builder.
func (b *ListBuilder) List() *SPCList {
	return b.SpcList
}

// Len returns the length og SPCList.
func (l *SPCList) Len() int {
	return len(l.ObjectList.Items)
}

// IsEmpty returns false if the SPCList is empty.
func (l *SPCList) IsEmpty() bool {
	return len(l.ObjectList.Items) == 0
}

// GetPoolUIDs retuns the UIDs of the pools available in the list.
func (l *SPCList) GetPoolUIDs() []string {
	uids := []string{}
	for _, pool := range l.ObjectList.Items {
		uids = append(uids, string(pool.GetUID()))
	}
	return uids
}
