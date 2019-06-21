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

package vcreate

// PredicateFunc defines data-type for validation function
type PredicateFunc func(*VolumeCreate) bool

// IsNameSet method check if the Name field of VolumeCreate object is set.
func IsNameSet() PredicateFunc {
	return func(v *VolumeCreate) bool {
		return len(v.Name) != 0
	}
}

// IsSizeSet method check if the Size field of VolumeCreate object is set.
func IsSizeSet() PredicateFunc {
	return func(v *VolumeCreate) bool {
		return len(v.Size) != 0
	}
}

// IsBlockSizeSet method check if the BlockSize field of VolumeCreate object is set.
func IsBlockSizeSet() PredicateFunc {
	return func(v *VolumeCreate) bool {
		return len(v.BlockSize) != 0
	}
}

// IsPropertySet method check if the Property field of VolumeCreate object is set.
func IsPropertySet() PredicateFunc {
	return func(v *VolumeCreate) bool {
		return len(v.Property) != 0
	}
}

// IsReservationSet method check if the Reservation field of VolumeCreate object is set.
func IsReservationSet() PredicateFunc {
	return func(v *VolumeCreate) bool {
		return v.Reservation
	}
}

// IsCreateParentSet method check if the CreateParent field of VolumeCreate object is set.
func IsCreateParentSet() PredicateFunc {
	return func(v *VolumeCreate) bool {
		return v.CreateParent
	}
}

// IsCommandSet method check if the Command field of VolumeCreate object is set.
func IsCommandSet() PredicateFunc {
	return func(v *VolumeCreate) bool {
		return len(v.Command) != 0
	}
}
