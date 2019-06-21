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

package vrename

// PredicateFunc defines data-type for validation function
type PredicateFunc func(*VolumeRename) bool

// IsCreateParentSet method check if the CreateParent field of VolumeRename object is set.
func IsCreateParentSet() PredicateFunc {
	return func(v *VolumeRename) bool {
		return v.CreateParent
	}
}

// IsForceUnmountSet method check if the ForceUnmount field of VolumeRename object is set.
func IsForceUnmountSet() PredicateFunc {
	return func(v *VolumeRename) bool {
		return v.ForceUnmount
	}
}

// IsSourceSet method check if the Source field of VolumeRename object is set.
func IsSourceSet() PredicateFunc {
	return func(v *VolumeRename) bool {
		return len(v.Source) != 0
	}
}

// IsDestSet method check if the Dest field of VolumeRename object is set.
func IsDestSet() PredicateFunc {
	return func(v *VolumeRename) bool {
		return len(v.Dest) != 0
	}
}

// IsCommandSet method check if the Command field of VolumeRename object is set.
func IsCommandSet() PredicateFunc {
	return func(v *VolumeRename) bool {
		return len(v.Command) != 0
	}
}
