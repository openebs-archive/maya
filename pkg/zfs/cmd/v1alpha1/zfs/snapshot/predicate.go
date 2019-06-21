/*
Copyright 2019 The OpenEBS Authors.

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

package vsnapshot

// PredicateFunc defines data-type for validation function
type PredicateFunc func(*VolumeSnapshot) bool

// IsPropertySet method check if the Property field of VolumeSnapshot object is set.
func IsPropertySet() PredicateFunc {
	return func(v *VolumeSnapshot) bool {
		return len(v.Property) != 0
	}
}

// IsRecursiveSet method check if the Recursive field of VolumeSnapshot object is set.
func IsRecursiveSet() PredicateFunc {
	return func(v *VolumeSnapshot) bool {
		return v.Recursive
	}
}

// IsSnapshotSet method check if the Snapshot field of VolumeSnapshot object is set.
func IsSnapshotSet() PredicateFunc {
	return func(v *VolumeSnapshot) bool {
		return len(v.Snapshot) != 0
	}
}

// IsDatasetSet method check if the Dataset field of VolumeSnapshot object is set.
func IsDatasetSet() PredicateFunc {
	return func(v *VolumeSnapshot) bool {
		return len(v.Dataset) != 0
	}
}

// IsCommandSet method check if the Command field of VolumeSnapshot object is set.
func IsCommandSet() PredicateFunc {
	return func(v *VolumeSnapshot) bool {
		return len(v.Command) != 0
	}
}
