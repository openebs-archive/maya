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
type PredicateFunc func(*PoolExport) bool

// IsAllPoolSet method check if the AllPool field of PoolExport object is set.
func IsAllPoolSet() PredicateFunc {
	return func(p *PoolExport) bool {
		return p.AllPool
	}
}

// IsForcefullySet method check if the Forcefully field of PoolExport object is set.
func IsForcefullySet() PredicateFunc {
	return func(p *PoolExport) bool {
		return p.Forcefully
	}
}

// IsPoolListSet method check if the PoolList field of PoolExport object is set.
func IsPoolListSet() PredicateFunc {
	return func(p *PoolExport) bool {
		return len(p.PoolList) != 0
	}
}

// IsCommandSet method check if the Command field of PoolExport object is set.
func IsCommandSet() PredicateFunc {
	return func(p *PoolExport) bool {
		return len(p.Command) != 0
	}
}
