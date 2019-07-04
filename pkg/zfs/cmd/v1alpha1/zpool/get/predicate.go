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

package pget

// PredicateFunc defines data-type for validation function
type PredicateFunc func(*PoolGProperty) bool

// IsPropListSet method check if the PropList field of PoolGProperty object is set.
func IsPropListSet() PredicateFunc {
	return func(p *PoolGProperty) bool {
		return len(p.PropList) != 0
	}
}

// IsPoolSet method check if the Pool field of PoolGProperty object is set.
func IsPoolSet() PredicateFunc {
	return func(p *PoolGProperty) bool {
		return len(p.Pool) != 0
	}
}

// IsCommandSet method check if the Command field of PoolGProperty object is set.
func IsCommandSet() PredicateFunc {
	return func(p *PoolGProperty) bool {
		return len(p.Command) != 0
	}
}

// IsScriptedModeSet method check if the IsScriptedMode field of PoolGProperty object is set.
func IsScriptedModeSet() PredicateFunc {
	return func(p *PoolGProperty) bool {
		return p.IsScriptedMode
	}
}

// IsParsableModeSet method check if the IsParsableMode field of PoolGProperty object is set.
func IsParsableModeSet() PredicateFunc {
	return func(p *PoolGProperty) bool {
		return p.IsParsableMode
	}
}

// IsFieldListSet method check if the FieldList field of PoolGProperty object is set.
func IsFieldListSet() PredicateFunc {
	return func(p *PoolGProperty) bool {
		return len(p.FieldList) != 0
	}
}
