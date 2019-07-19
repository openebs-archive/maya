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

package padd

// PredicateFunc defines data-type for validation function
type PredicateFunc func(*PoolExpansion) bool

// IsVdevListSet method check if the VdevList field of PoolExpansion object is set.
func IsVdevListSet() PredicateFunc {
	return func(p *PoolExpansion) bool {
		return len(p.VdevList) != 0
	}
}

// IsPropertySet method check if the Property field of PoolExpansion object is set.
func IsPropertySet() PredicateFunc {
	return func(p *PoolExpansion) bool {
		return len(p.Property) != 0
	}
}

// IsPoolSet method check if the Pool field of PoolExpansion object is set.
func IsPoolSet() PredicateFunc {
	return func(p *PoolExpansion) bool {
		return len(p.Pool) != 0
	}
}

// IsDeviceTypeSet method check if the DeviceType field of PoolExpansion object is set.
func IsDeviceTypeSet() PredicateFunc {
	return func(p *PoolExpansion) bool {
		// If no device type is mentioned in command
		// ZFS consider it as a stripe type by itself
		// We don't need to provide `stripe` as a type in command
		// So If user has provided type `stripe`, we will ignore it
		if p.DeviceType == "stripe" {
			return false
		}
		return len(p.DeviceType) != 0
	}
}

// IsTypeSet method check if the Type field of PoolExpansion object is set.
func IsTypeSet() PredicateFunc {
	return func(p *PoolExpansion) bool {
		// If no device type is mentioned in command
		// ZFS consider it as a stripe type by itself
		// We don't need to provide `stripe` as a type in command
		// So If user has provided type `stripe`, we will ignore it
		if p.Type == "stripe" {
			return false
		}
		return len(p.Type) != 0
	}
}

// IsForcefullySet method check if the Forcefully field of PoolExpansion object is set.
func IsForcefullySet() PredicateFunc {
	return func(p *PoolExpansion) bool {
		return p.Forcefully
	}
}

// IsCommandSet method check if the Command field of PoolExpansion object is set.
func IsCommandSet() PredicateFunc {
	return func(p *PoolExpansion) bool {
		return len(p.Command) != 0
	}
}
