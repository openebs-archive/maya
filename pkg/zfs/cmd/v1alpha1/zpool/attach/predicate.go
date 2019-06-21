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

package pattach

// PredicateFunc defines data-type for validation function
type PredicateFunc func(*PoolAttach) bool

// IsPropertySet method check if the Property field of PoolAttach object is set.
func IsPropertySet() PredicateFunc {
	return func(p *PoolAttach) bool {
		return len(p.Property) != 0
	}
}

// IsForcefullySet method check if the Forcefully field of PoolAttach object is set.
func IsForcefullySet() PredicateFunc {
	return func(p *PoolAttach) bool {
		return p.Forcefully
	}
}

// IsDeviceSet method check if the Device field of PoolAttach object is set.
func IsDeviceSet() PredicateFunc {
	return func(p *PoolAttach) bool {
		return len(p.Device) != 0
	}
}

// IsNewDeviceSet method check if the NewDevice field of PoolAttach object is set.
func IsNewDeviceSet() PredicateFunc {
	return func(p *PoolAttach) bool {
		return len(p.NewDevice) != 0
	}
}

// IsPoolSet method check if the Pool field of PoolAttach object is set.
func IsPoolSet() PredicateFunc {
	return func(p *PoolAttach) bool {
		return len(p.Pool) != 0
	}
}

// IsCommandSet method check if the Command field of PoolAttach object is set.
func IsCommandSet() PredicateFunc {
	return func(p *PoolAttach) bool {
		return len(p.Command) != 0
	}
}
