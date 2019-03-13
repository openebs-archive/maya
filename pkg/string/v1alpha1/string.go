package strings

import (
	lib_strings "strings"
)

type list struct {
	items []string
}

// List will make list of sstring slices
func List(entry ...string) *list {
	l := &list{items: []string{}}
	l.items = append(l.items, entry...)
	return l
}

func (l *list) Contains(search string) bool {
	for _, item := range l.items {
		if lib_strings.Contains(item, search) {
			return true
		}
	}
	return false
}
