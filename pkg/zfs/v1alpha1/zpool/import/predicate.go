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
type PredicateFunc func(*PoolImport) bool

// IsCachefileSet method check if the Cachefile field of PoolImport object is set.
func IsCachefileSet() PredicateFunc {
	return func(p *PoolImport) bool {
		return len(p.Cachefile) != 0
	}
}

// IsDirectorylistSet method check if the Directorylist field of PoolImport object is set.
func IsDirectorylistSet() PredicateFunc {
	return func(p *PoolImport) bool {
		return len(p.Directorylist) != 0
	}
}

// IsImportAllSet method check if the ImportAll field of PoolImport object is set.
func IsImportAllSet() PredicateFunc {
	return func(p *PoolImport) bool {
		return p.ImportAll
	}
}

// IsForceImportSet method check if the ForceImport field of PoolImport object is set.
func IsForceImportSet() PredicateFunc {
	return func(p *PoolImport) bool {
		return p.ForceImport
	}
}

// IsPropertySet method check if the Property field of PoolImport object is set.
func IsPropertySet() PredicateFunc {
	return func(p *PoolImport) bool {
		return len(p.Property) != 0
	}
}

// IsPoolSet method check if the Pool field of PoolImport object is set.
func IsPoolSet() PredicateFunc {
	return func(p *PoolImport) bool {
		return len(p.Pool) != 0
	}
}

// IsNewPoolSet method check if the NewPool field of PoolImport object is set.
func IsNewPoolSet() PredicateFunc {
	return func(p *PoolImport) bool {
		return len(p.NewPool) != 0
	}
}

// IsCommandSet method check if the Command field of PoolImport object is set.
func IsCommandSet() PredicateFunc {
	return func(p *PoolImport) bool {
		return len(p.Command) != 0
	}
}
