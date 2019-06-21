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

package vdestroy

// PredicateFunc defines data-type for validation function
type PredicateFunc func(*VolumeDestroy) bool

// IsDryRunSet method check if the DryRun field of VolumeDestroy object is set.
func IsDryRunSet() PredicateFunc {
	return func(v *VolumeDestroy) bool {
		return v.DryRun
	}
}

// IsRecursiveSet method check if the Recursive field of VolumeDestroy object is set.
func IsRecursiveSet() PredicateFunc {
	return func(v *VolumeDestroy) bool {
		return v.Recursive
	}
}

// IsNameSet method check if the Name field of VolumeDestroy object is set.
func IsNameSet() PredicateFunc {
	return func(v *VolumeDestroy) bool {
		return len(v.Name) != 0
	}
}

// IsCommandSet method check if the Command field of VolumeDestroy object is set.
func IsCommandSet() PredicateFunc {
	return func(v *VolumeDestroy) bool {
		return len(v.Command) != 0
	}
}
