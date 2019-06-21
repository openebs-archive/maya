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

package vrollback

// PredicateFunc defines data-type for validation function
type PredicateFunc func(*VolumeRollback) bool

// IsDestroySet method check if the Destroy field of VolumeRollback object is set.
func IsDestroySet() PredicateFunc {
	return func(v *VolumeRollback) bool {
		return v.Destroy
	}
}

// IsForceUnmountSet method check if the ForceUnmount field of VolumeRollback object is set.
func IsForceUnmountSet() PredicateFunc {
	return func(v *VolumeRollback) bool {
		return v.ForceUnmount
	}
}

// IsDestroySnapSet method check if the DestroySnap field of VolumeRollback object is set.
func IsDestroySnapSet() PredicateFunc {
	return func(v *VolumeRollback) bool {
		return v.DestroySnap
	}
}

// IsSnapshotSet method check if the Snapshot field of VolumeRollback object is set.
func IsSnapshotSet() PredicateFunc {
	return func(v *VolumeRollback) bool {
		return len(v.Snapshot) != 0
	}
}

// IsCommandSet method check if the Command field of VolumeRollback object is set.
func IsCommandSet() PredicateFunc {
	return func(v *VolumeRollback) bool {
		return len(v.Command) != 0
	}
}
