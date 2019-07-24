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
