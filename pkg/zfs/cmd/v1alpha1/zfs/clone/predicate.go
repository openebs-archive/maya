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

package vclone

// PredicateFunc defines data-type for validation function
type PredicateFunc func(*VolumeClone) bool

// IsSnapshotSet method check if the Snapshot field of VolumeClone object is set.
func IsSnapshotSet() PredicateFunc {
	return func(v *VolumeClone) bool {
		return len(v.Snapshot) != 0
	}
}

// IsTargetDatasetSet method check if the TargetDataset field of VolumeClone object is set.
func IsTargetDatasetSet() PredicateFunc {
	return func(v *VolumeClone) bool {
		return len(v.TargetDataset) != 0
	}
}

// IsSourceDatasetSet method check if the SourceDataset field of VolumeClone object is set.
func IsSourceDatasetSet() PredicateFunc {
	return func(v *VolumeClone) bool {
		return len(v.SourceDataset) != 0
	}
}

// IsPropertySet method check if the Property field of VolumeClone object is set.
func IsPropertySet() PredicateFunc {
	return func(v *VolumeClone) bool {
		return len(v.Property) != 0
	}
}

// IsCreateParentSet method check if the CreateParent field of VolumeClone object is set.
func IsCreateParentSet() PredicateFunc {
	return func(v *VolumeClone) bool {
		return v.CreateParent
	}
}

// IsCommandSet method check if the Command field of VolumeClone object is set.
func IsCommandSet() PredicateFunc {
	return func(v *VolumeClone) bool {
		return len(v.Command) != 0
	}
}
