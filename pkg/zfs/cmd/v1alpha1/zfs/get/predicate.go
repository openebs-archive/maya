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

package vget

// PredicateFunc defines data-type for validation function
type PredicateFunc func(*VolumeGetProperty) bool

// IsProplistSet method check if the Proplist field of VolumeGetProperty object is set.
func IsProplistSet() PredicateFunc {
	return func(v *VolumeGetProperty) bool {
		return len(v.Proplist) != 0
	}
}

// IsDatasetSet method check if the Dataset field of VolumeGetProperty object is set.
func IsDatasetSet() PredicateFunc {
	return func(v *VolumeGetProperty) bool {
		return len(v.Dataset) != 0
	}
}

// IsSnapshotSet method check if the Snapshot field of VolumeGetProperty object is set.
func IsSnapshotSet() PredicateFunc {
	return func(v *VolumeGetProperty) bool {
		return len(v.Snapshot) != 0
	}
}

// IsCommandSet method check if the Command field of VolumeGetProperty object is set.
func IsCommandSet() PredicateFunc {
	return func(v *VolumeGetProperty) bool {
		return len(v.Command) != 0
	}
}

// IsScriptedModeSet method check if the IsScriptedMode field of VolumeGetProperty object is set.
func IsScriptedModeSet() PredicateFunc {
	return func(v *VolumeGetProperty) bool {
		return v.IsScriptedMode
	}
}

// IsParsableModeSet method check if the IsParsableMode field of VolumeGetProperty object is set.
func IsParsableModeSet() PredicateFunc {
	return func(v *VolumeGetProperty) bool {
		return v.IsParsableMode
	}
}

// IsFieldListSet method check if the FieldList field of VolumeGetProperty object is set.
func IsFieldListSet() PredicateFunc {
	return func(v *VolumeGetProperty) bool {
		return len(v.FieldList) != 0
	}
}
