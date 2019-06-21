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

package pdestroy

// PredicateFunc defines data-type for validation function
type PredicateFunc func(*PoolDestroy) bool

// IsPoolSet method check if the Pool field of PoolDestroy object is set.
func IsPoolSet() PredicateFunc {
	return func(p *PoolDestroy) bool {
		return len(p.Pool) != 0
	}
}

// IsForcefullySet method check if the Forcefully field of PoolDestroy object is set.
func IsForcefullySet() PredicateFunc {
	return func(p *PoolDestroy) bool {
		return p.Forcefully
	}
}

// IsCommandSet method check if the Command field of PoolDestroy object is set.
func IsCommandSet() PredicateFunc {
	return func(p *PoolDestroy) bool {
		return len(p.Command) != 0
	}
}
