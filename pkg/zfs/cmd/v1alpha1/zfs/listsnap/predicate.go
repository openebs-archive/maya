/*
Copyright 2020 The OpenEBS Authors.

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

package vlistsnap

// PredicateFunc defines data-type for validation function
type PredicateFunc func(*VolumeListSnapshot) bool

// IsDatasetSet method check if the Dataset field of VolumeListSnapshot object is set.
func IsDatasetSet() PredicateFunc {
	return func(v *VolumeListSnapshot) bool {
		return len(v.Dataset) != 0
	}
}
