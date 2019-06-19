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

package zfs

// PredicateFunc defines data-type for validation function
type PredicateFunc func(*PoolOffline) bool

// IsForceOfflineSet method check if the ForceOffline field of PoolOffline object is set.
func IsForceOfflineSet() PredicateFunc {
	return func(p *PoolOffline) bool {
		return p.ForceOffline
	}
}

// IsTemporarySet method check if the isTemporary field of PoolOffline object is set.
func IsTemporarySet() PredicateFunc {
	return func(p *PoolOffline) bool {
		return p.isTemporary
	}
}

// IsPoolSet method check if the Pool field of PoolOffline object is set.
func IsPoolSet() PredicateFunc {
	return func(p *PoolOffline) bool {
		return len(p.Pool) != 0
	}
}

// IsDevicelistSet method check if the Devicelist field of PoolOffline object is set.
func IsDevicelistSet() PredicateFunc {
	return func(p *PoolOffline) bool {
		return len(p.Devicelist) != 0
	}
}

// IsCommandSet method check if the Command field of PoolOffline object is set.
func IsCommandSet() PredicateFunc {
	return func(p *PoolOffline) bool {
		return len(p.Command) != 0
	}
}
