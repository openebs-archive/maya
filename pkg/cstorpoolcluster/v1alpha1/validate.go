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
	"github.com/pkg/errors"
)

// TODO: Improve validation predicate application for various cases.
// TODO: Think on return type of vaildation predicates -- bool and error return type.
// TODO: Also trace of error in case of predicate fails when we want to run all predicates even if any one fails -- e.g. having array of errors in builder object.

var (
	// supportedPool is a map holding the supported raid configurations.
	supportedPool = map[string]bool{
		string(apisv1alpha1.PoolTypeStripedCPV):  true,
		string(apisv1alpha1.PoolTypeMirroredCPV): true,
		string(apisv1alpha1.PoolTypeRaidzCPV):    true,
		string(apisv1alpha1.PoolTypeRaidz2CPV):   true,
	}
)

// ValidateFunc is typed function for cspc validation functions.
type ValidateFunc func(*CSPC) error

// ValidateFuncList holds a list of validate functions for cspc
var ValidateFuncList = []ValidateFunc{
	IsValidPoolType,
	IsValidDiskType,
	IsValidMaxPool,
}

// Validate validates a cspc on the available predicates
func Validate(cspc *CSPC) error {
	for _, v := range ValidateFuncList {
		err := v(cspc)
		if err != nil {
			return errors.Wrapf(err, "validation failed for cspc object %s", cspc.Object.Name)
		}
	}
	return nil
}

// IsValidDiskType validates the disk types in cspc.
func IsValidDiskType(cspc *CSPC) error {
	diskType := cspc.Object.Spec.Type
	if !(diskType == "sparse" || diskType == "disk") {
		return errors.Errorf("specified type on cspc %s is %s which is invalid", cspc.Object.Name, diskType)
	}
	return nil
}

// IsValidMaxPool validates the max pool count in auto cspc
func IsValidMaxPool(cspc *CSPC) error {
	cspcName := cspc.Object.Name
	if IsProvisioningAuto()(cspc) {
		maxPools := cspc.Object.Spec.MaxPools
		if IsMaxPoolNil(maxPools) {
			return errors.Errorf("maxpool value is nil for cspc %s which is invalid", cspcName)
		}
		if IsMaxPoolNonNegative(*maxPools) {
			return errors.Errorf("maxpool value is %v for cspc %s which is invalid", maxPools, cspcName)
		}
	}
	return nil
}

// IsMaxPoolNil returns true if passed maxpool pointer is not nil.
func IsMaxPoolNil(maxPool *int) bool {
	return maxPool == nil
}

// IsMaxPoolNonNegative returns true if passed argument de-referenced value is non negative
func IsMaxPoolNonNegative(maxPool int) bool {
	return maxPool < 0
}

// IsValidPoolType validates pooltype in cspc.
func IsValidPoolType(cspc *CSPC) error {
	for _, node := range cspc.Object.Spec.Nodes {
		if !supportedPool[string(node.PoolSpec.PoolType)] {
			return errors.Errorf("pool type is %s for node %s in cspc %s", node.PoolSpec.PoolType, node.Name, cspc.Object.Name)
		}
	}
	return nil
}
