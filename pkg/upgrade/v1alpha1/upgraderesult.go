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
	apis "github.com/openebs/maya/pkg/apis/openebs.io/upgrade/v1alpha1"
)

type upgradeResult struct {
	// upgrade result object
	object apis.UpgradeResult
}

type urList struct {
	// list of upgrade results
	items []upgradeResult
}

// ListBuilder enables building
// an instance of urList
type listBuilder struct {
	list *urList
}

// ListBuilder returns a new instance
// of listBuilder
func ListBuilder() *listBuilder {
	return &listBuilder{list: &urList{}}
}

// WithAPIList builds the list of ur
// instances based on the provided
// ur api instances
func (b *listBuilder) WithAPIList(list *apis.UpgradeResultList) *listBuilder {
	if list == nil {
		return b
	}
	for _, c := range list.Items {
		b.list.items = append(b.list.items, upgradeResult{object: c})
	}
	return b
}

// List returns the list of ur
// instances that was built by this
// builder
func (b *listBuilder) List() *urList {
	return b.list
}
