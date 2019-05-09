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
	"github.com/golang/glog"
	apisv1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	csp "github.com/openebs/maya/pkg/cstorpool/v1alpha3"
	cspc "github.com/openebs/maya/pkg/cstorpoolcluster/v1alpha1"
	disk "github.com/openebs/maya/pkg/disk/v1alpha2"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// OperationPredicate is typed function for operation builder predicates.
type OperationPredicate func(*OperationsBuilder) bool

// OperationsBuilder is the builder object for Operations object.
type OperationsBuilder struct {
	Operations       *Operations
	errs             []error
	validationErrs   []error
	validationChecks map[*OperationPredicate]string
	fastFail         bool
}

// Operations object includes client for different objects as well as CStorPoolCluster object.
// to carry out pool operations.
type Operations struct {
	SpcClient  *cspc.Kubeclient
	CspClient  *csp.Kubeclient
	DiskClient *disk.Kubeclient
	CspcObject *cspc.CSPC
}

// NewOperationsBuilder returns an empty instance of OperationsBuilder.
func NewOperationsBuilder() *OperationsBuilder {
	return &OperationsBuilder{
		Operations:       &Operations{},
		validationChecks: make(map[*OperationPredicate]string),
	}
}

// NewOperationsBuilderForObject returns OperationsBuilder by building with provided Operations object.
func NewOperationsBuilderForObject(op *Operations) *OperationsBuilder {
	return &OperationsBuilder{
		Operations:       op,
		validationChecks: make(map[*OperationPredicate]string),
	}
}

// Build returns Operations object.
func (ob *OperationsBuilder) Build() *Operations {
	return ob.Operations
}

// WithDefaults sets object clients to Operations object.
func (ob *OperationsBuilder) WithDefaults() *OperationsBuilder {
	ob.Operations.SpcClient = cspc.NewKubeClient()
	ob.Operations.CspClient = csp.KubeClient()
	ob.Operations.DiskClient = disk.NewKubeClient()
	return ob
}

// WithCStorPoolCluster sets CStorPoolCluster to operations object.
func (ob *OperationsBuilder) WithCStorPoolCluster(cspcObject *apisv1alpha1.CStorPoolCluster) *OperationsBuilder {
	ob.Operations.CspcObject = cspc.BuilderForAPIObject(cspcObject).Build()
	return ob
}

// WithFastFail sets fast fail.
func (ob *OperationsBuilder) WithFastFail() *OperationsBuilder {
	ob.fastFail = true
	return ob
}

// Validate validates the CSPC object encapsulated in  OperationsBuilder
func (ob *OperationsBuilder) Validate() error {
	for p, v := range ob.validationChecks {
		if !(*p)(ob) {
			if ob.errs != nil {
				return errors.Errorf("failed to validate: {%v}", ob.errs)
			}
			ob.validationErrs = append(ob.validationErrs, errors.New(v))
			if ob.fastFail {
				return errors.Errorf("validation failed:{%v}", ob.validationErrs)
			}
		}
	}
	if len(ob.validationErrs) > 0 {
		return errors.Errorf("validation failed for validation predicates:{%v}", ob.validationErrs)
	}
	return nil
}

// IsPoolTypeValid returns a OperationPredicate for validating pool type ( spec.poolSpec.poolTYpe) in cspc
func IsPoolTypeValid() OperationPredicate {
	return func(ob *OperationsBuilder) bool {
		return ob.IsPoolTypeValid()
	}
}

// IsPoolTypeValid returns true if the poolType ( spec.poolSpec.poolTYpe) is valid in cspc.
func (ob *OperationsBuilder) IsPoolTypeValid() bool {
	// TODO: Put this logic in cstorpoolcluster packages
	// Similarly at other places
	/*
		Example:
		return ob.Operations.CspcObject.IsPoolTypeValid
	*/
	return apisv1alpha1.SupportedPoolTypes[ob.Operations.CspcObject.GetPoolType()]
}

// IsMaxPoolNotNil returns a predicate to validate that max pool field is not nil
func IsMaxPoolNotNil() OperationPredicate {
	return func(ob *OperationsBuilder) bool {
		return ob.IsMaxPoolNotNil()
	}
}

// IsMaxPoolNotNil returns true if max pool field is not nil in cspc.
func (ob *OperationsBuilder) IsMaxPoolNotNil() bool {
	return ob.Operations.CspcObject.Object.Spec.MaxPools != nil
}

// IsTypeValid returns true to validate type in cspc.
// Note: Valid types are "sparse" and "disk"
func IsTypeValid() OperationPredicate {
	return func(ob *OperationsBuilder) bool {
		return ob.IsTypeValid()
	}
}

// IsTypeValid returns true if the type is valid in cspc
func (ob *OperationsBuilder) IsTypeValid() bool {
	CSPCObject := ob.Operations.CspcObject
	return CSPCObject.IsSparse() || CSPCObject.IsDisk()
}

// IsDiskNotRepeated returns a predicate to validate duplicate disk entry in cspc.
func IsDiskNotRepeated() OperationPredicate {
	return func(ob *OperationsBuilder) bool {
		return ob.IsDiskNotRepeated()
	}
}

// IsDiskNotRepeated returns true if there is a no duplicate disk entry in cspc.
func (ob *OperationsBuilder) IsDiskNotRepeated() bool {
	CSPCObject := ob.Operations.CspcObject
	return !CSPCObject.IsDiskRepeated()
}

// IsNodeNotRepeated returns a predicate to validate duplicate node entry in cspc.
func IsNodeNotRepeated() OperationPredicate {
	return func(ob *OperationsBuilder) bool {
		return ob.IsNodeNotRepeated()
	}
}

// IsNodeNotRepeated returns true if there is no duplicate node entries in cspc
func (ob *OperationsBuilder) IsNodeNotRepeated() bool {
	CSPCObject := ob.Operations.CspcObject
	nodeNames := CSPCObject.GetNodeNames()
	nodeCount := make(map[string]int)
	for _, nodeName := range nodeNames {
		nodeCount[nodeName]++
		if nodeCount[nodeName] > 0 {
			return true
		}
	}
	return false
}

// IsPoolTypeOnNodeValid returns a predicate to validate pool type in node spec in cspc.
func IsPoolTypeOnNodeValid() OperationPredicate {
	return func(ob *OperationsBuilder) bool {
		return ob.IsPoolTypeOnNodeValid()
	}
}

// IsPoolTypeOnNodeValid returns true if pool type on node spec is valid
func (ob *OperationsBuilder) IsPoolTypeOnNodeValid() bool {
	CSPCObject := ob.Operations.CspcObject

	for _, node := range CSPCObject.Object.Spec.Nodes {
		if !apisv1alpha1.SupportedPoolTypes[string(node.PoolSpec.PoolType)] {
			return false
		}
	}
	return true
}

// IsNodeDiskRelationValid returns a predicate to validate that disk is specified at correct node name slot
func IsNodeDiskRelationValid() OperationPredicate {
	return func(ob *OperationsBuilder) bool {
		return ob.IsNodeDiskRelationValid()
	}
}

// IsNodeDiskRelationValid returns true if disk belong the specified node in the cspc spec
func (ob *OperationsBuilder) IsNodeDiskRelationValid() bool {
	CSPCObject := ob.Operations.CspcObject

	for _, node := range CSPCObject.Object.Spec.Nodes {
		for _, diskGroup := range node.DiskGroups {
			for _, disk := range diskGroup.Disks {
				if !ob.IsDiskBelongToNode(node.Name, disk.Name) {
					return false
				}
			}
		}
	}
	return true
}

// IsDiskBelongToNode returns true if the provided disk belongs to provided node.
func (ob *OperationsBuilder) IsDiskBelongToNode(nodeName, diskName string) bool {
	diskObject, err := disk.NewKubeClient().Get(diskName, metav1.GetOptions{})
	if err != nil {
		glog.Errorf("error in getting disk object while validating is node disk valid:{%s}", err)
		return false
	}
	if diskObject.GetLabels()[string(apisv1alpha1.HostNameCPK)] == nodeName {
		return true
	}
	return false
}

// IsDiskCountValid returns predicate to validate if the disk count is valid for specified pool type.
func IsDiskCountValid() OperationPredicate {
	return func(ob *OperationsBuilder) bool {
		return ob.IsDiskCountValid()
	}
}

// IsDiskCountValid returns true if the disk count is valid for specified pool type.
func (ob *OperationsBuilder) IsDiskCountValid() bool {
	CSPCObject := ob.Operations.CspcObject
	for _, node := range CSPCObject.Object.Spec.Nodes {
		poolType := node.PoolSpec.PoolType
		for _, diskGroup := range node.DiskGroups {
			if poolType == apisv1alpha1.PoolStriped {
				if len(diskGroup.Disks) < 1 {
					return false
				}
			} else if !(len(diskGroup.Disks) == disk.DefaultDiskCount[string(poolType)]) {
				return false
			}
		}
	}
	return true
}

// IsDiskActive returns a predicate to validate if the specified disk is active.
func IsDiskActive() OperationPredicate {
	return func(ob *OperationsBuilder) bool {
		return ob.IsDiskActive()
	}
}

// IsDiskActive returns true if all the disk specified in cspc is active.
func (ob *OperationsBuilder) IsDiskActive() bool {
	CSPCObject := ob.Operations.CspcObject
	for _, node := range CSPCObject.Object.Spec.Nodes {
		for _, diskGroup := range node.DiskGroups {
			for _, diskDetails := range diskGroup.Disks {
				diskObj, err := disk.NewKubeClient().Get(diskDetails.Name, metav1.GetOptions{})
				if err != nil {
					glog.Errorf("error in getting disk object while validating is disk active:{%s}", err)
					return false
				}
				if !disk.BuilderForAPIObject(diskObj).Build().IsActive() {
					return false
				}
			}
		}
	}
	return true
}

// WithCheck method is used to add a validation predicate on OperationsBuilder
func (ob *OperationsBuilder) WithCheck(opPredicate OperationPredicate) *OperationsBuilder {
	ob.WithCheckf(opPredicate, "")
	return ob
}

// WithCheckf method is used to add a validation predicate on OperationsBuilder with message
func (ob *OperationsBuilder) WithCheckf(opPredicate OperationPredicate, msg string, args ...interface{}) *OperationsBuilder {
	ob.validationChecks[&opPredicate] = fmt.Sprintf(msg, args...)
	return ob
}
