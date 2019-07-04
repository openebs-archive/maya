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

package pset

// PredicateFunc defines data-type for validation function
type PredicateFunc func(*PoolSProperty) bool

// IsPropListSet method check if the PropList field of PoolSProperty object is set.
func IsPropListSet() PredicateFunc {
	return func(p *PoolSProperty) bool {
		return len(p.PropList) != 0
	}
}

// IsPoolSet method check if the Pool field of PoolSProperty object is set.
func IsPoolSet() PredicateFunc {
	return func(p *PoolSProperty) bool {
		return len(p.Pool) != 0
	}
}

// IsCommandSet method check if the Command field of PoolSProperty object is set.
func IsCommandSet() PredicateFunc {
	return func(p *PoolSProperty) bool {
		return len(p.Command) != 0
	}
}
