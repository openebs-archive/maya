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

package cstorpoolspecs

import (
	apisv1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
)

// PS is a wrapper over poolspec api
// object. It provides build, validations and other common
// logic to be used by various feature specific callers.
type PS struct {
	object *apisv1alpha1.PoolSpec
}

// PSList is a wrapper over poolspec api
// object. It provides build, validations and other common
// logic to be used by various feature specific callers.
type PSList struct {
	items []apisv1alpha1.PoolSpec
}

// Len returns the number of items present
// in the PSList
func (c *PSList) Len() int {
	return len(c.items)
}

// IsStripePoolSpec returns true if pool spec is stripe pool or else it will
// return false
func IsStripePoolSpec(poolSpec *apisv1alpha1.PoolSpec) bool {
	if len(poolSpec.RaidGroups[0].Type) != 0 {
		return apisv1alpha1.PoolType(poolSpec.RaidGroups[0].Type) == apisv1alpha1.PoolStriped
	}
	return apisv1alpha1.PoolType(poolSpec.PoolConfig.DefaultRaidGroupType) == apisv1alpha1.PoolStriped
}
