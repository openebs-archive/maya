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

// ValidateFunc is typed function for spc validation functions.
type ValidateFunc func(*SPC) error

// ValidateFuncList holds a list of validate functions for spc
var ValidateFuncList = []ValidateFunc{
	IsValidPoolType,
	IsValidDiskType,
	IsValidMaxPool,
}

// Validate validates a spc on the available predicates
func Validate(spc *SPC) error {
	for _, v := range ValidateFuncList {
		err := v(spc)
		if err != nil {
			return errors.Wrapf(err, "validation failed for spc object %s", spc.Object.Name)
		}
	}
	return nil
}

// IsValidDiskType validates the disk types in spc.
func IsValidDiskType(spc *SPC) error {
	diskType := spc.Object.Spec.Type
	if !(diskType == "sparse" || diskType == "disk") {
		return errors.Errorf("specified type on spc %s is %s which is invalid", spc.Object.Name, diskType)
	}
	return nil
}

// IsValidMaxPool validates the max pool count in auto spc
func IsValidMaxPool(spc *SPC) error {
	spcName := spc.Object.Name
	if IsProvisioningAuto()(spc) {
		maxPools := spc.Object.Spec.MaxPools
		if IsMaxPoolNil(maxPools) {
			return errors.Errorf("maxpool value is nil for spc %s which is invalid", spcName)
		}
		if IsMaxPoolNonNegative(*maxPools) {
			return errors.Errorf("maxpool value is %v for spc %s which is invalid", maxPools, spcName)
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

// IsValidPoolType validates pooltype in spc.
func IsValidPoolType(spc *SPC) error {
	for _, node := range spc.Object.Spec.Nodes {
		if !supportedPool[node.PoolSpec.PoolType] {
			return errors.Errorf("pool type is %s for node %s in spc %s", node.PoolSpec.PoolType, node.Name, spc.Object.Name)
		}
	}
	return nil
}
