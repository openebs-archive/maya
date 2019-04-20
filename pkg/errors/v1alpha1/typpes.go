/*
Copyright 2019 The OpenEBS Authors

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

package v1alpha1

import (
	"fmt"
	"runtime"

	"github.com/pkg/errors"
)

// stack represents a stack of program counters.
type stack []uintptr

// callers returns stack of caller function
func callers() *stack {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	var st stack = pcs[0:n]
	return &st
}

// err implements error interface that has a message and stack
type err struct {
	msg string
	*stack
}

// Error is implementation of error interface
func (e *err) Error() string { return e.msg }

// Format is implementation of Formater interface
func (e *err) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprint(s, "error(s) were found: "+e.msg)
			for i, pc := range *e.stack {
				if i < 1 {
					f := errors.Frame(pc)
					fmt.Fprintf(s, "\n%+v", f)
				}
			}
			return
		}
		fallthrough
	case 's', 'q':
		fmt.Fprint(s, "error(s) were found: "+e.msg)
	}
}

// err implements error interface that has a message and error
type wrapper struct {
	msg string
	error
}

// Error is implementation of error interface
func (w *wrapper) Error() string { return w.msg }

// Format is implementation of Formater interface
func (w *wrapper) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%+v\n", w.error)
			fmt.Fprint(s, w.msg)
			return
		}
		fallthrough
	case 's', 'q':
		fmt.Fprintf(s, "%s\n", w.error)
		fmt.Fprint(s, w.msg)
	}
}
