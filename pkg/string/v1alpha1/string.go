// Copyright © 2018-2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package strings

import (
	lib_strings "strings"
)

// List holds the array of strings
type List struct {
	items []string
}

// MakeList will make list of string slices
func MakeList(entry ...string) *List {
	l := &List{items: []string{}}
	l.items = append(l.items, entry...)
	return l
}

// Contains will return true if it has matching string
func (l *List) Contains(search string) bool {
	for _, item := range l.items {
		if lib_strings.Contains(item, search) {
			return true
		}
	}
	return false
}
