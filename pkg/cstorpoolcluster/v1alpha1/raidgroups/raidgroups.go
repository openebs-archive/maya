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

package raidgroups

import (
	apisv1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
)

// RG is a wrapper over raidgroup api
// object. It provides build, validations and other common
// logic to be used by various feature specific callers.
type RG struct {
	object *apisv1alpha1.RaidGroup
}

// RGList is a wrapper over raidgroup api
// object. It provides build, validations and other common
// logic to be used by various feature specific callers.
type RGList struct {
	items []apisv1alpha1.RaidGroup
}

// Len returns the number of items present
// in the PSList
func (c *RGList) Len() int {
	return len(c.items)
}
