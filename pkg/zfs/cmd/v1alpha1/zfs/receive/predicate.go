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

package vsnapshotrecv

// PredicateFunc defines data-type for validation function
type PredicateFunc func(*VolumeSnapshotRecv) bool

// IsSnapshotSet method check if the Snapshot field of VolumeSnapshotRecv object is set.
func IsSnapshotSet() PredicateFunc {
	return func(v *VolumeSnapshotRecv) bool {
		return len(v.Snapshot) != 0
	}
}

// IsDatasetSet method check if the Dataset field of VolumeSnapshotRecv object is set.
func IsDatasetSet() PredicateFunc {
	return func(v *VolumeSnapshotRecv) bool {
		return len(v.Dataset) != 0
	}
}

// IsTargetSet method check if the Target field of VolumeSnapshotRecv object is set.
func IsTargetSet() PredicateFunc {
	return func(v *VolumeSnapshotRecv) bool {
		return len(v.Target) != 0
	}
}

// IsDedupSet method check if the Dedup field of VolumeSnapshotRecv object is set.
func IsDedupSet() PredicateFunc {
	return func(v *VolumeSnapshotRecv) bool {
		return v.Dedup
	}
}

// IsLastSnapshotSet method check if the LastSnapshot field of VolumeSnapshotRecv object is set.
func IsLastSnapshotSet() PredicateFunc {
	return func(v *VolumeSnapshotRecv) bool {
		return len(v.LastSnapshot) != 0
	}
}

// IsDryRunSet method check if the DryRun field of VolumeSnapshotRecv object is set.
func IsDryRunSet() PredicateFunc {
	return func(v *VolumeSnapshotRecv) bool {
		return v.DryRun
	}
}

// IsEnableCompressionSet method check if the EnableCompression field of VolumeSnapshotRecv object is set.
func IsEnableCompressionSet() PredicateFunc {
	return func(v *VolumeSnapshotRecv) bool {
		return v.EnableCompression
	}
}

// IsCommandSet method check if the Command field of VolumeSnapshotRecv object is set.
func IsCommandSet() PredicateFunc {
	return func(v *VolumeSnapshotRecv) bool {
		return len(v.Command) != 0
	}
}
