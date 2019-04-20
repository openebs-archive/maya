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
)

// New returns an error with the supplied message.
// New also records the stack trace at the point it was called.
func New(message string) error {
	return &err{
		msg:   message,
		stack: callers(),
	}
}

// Errorf formats according to a format specifier and returns the string
// as a value that satisfies error.
// Errorf also records the stack trace at the point it was called.
func Errorf(format string, args ...interface{}) error {
	return &err{
		msg:   fmt.Sprintf(format, args...),
		stack: callers(),
	}
}

// Merge annotates list of errors with a new message.
// and errors messages. length of errs is 0 Merge returns nil.
func Merge(errs []error, message string) error {
	if len(errs) == 0 {
		return nil
	}
	for _, err := range errs {
		message += "\n  -  " + err.Error()
	}
	return &err{
		msg:   message,
		stack: callers(),
	}
}

// Mergef annotateslist of errors with with given format specifier.
// and errors messages. length of errs is 0 Mergef returns nil.
func Mergef(errs []error, format string, args ...interface{}) error {
	if len(errs) == 0 {
		return nil
	}
	message := fmt.Sprintf(format, args...)
	for _, e := range errs {
		message += "\n  -  " + e.Error()
	}
	return &err{
		msg:   message,
		stack: callers(),
	}

}

// Wrap annotates err with a new message.
// If err is nil, Wrap returns nil.
func Wrap(e error, message string) error {
	message = "  --  " + message
	return &wrapper{message, e}
}

// Wrapf annotates err with the format specifier.
// If err is nil, Wrapf returns nil.
func Wrapf(err error, format string, args ...interface{}) error {
	message := "  --  " + fmt.Sprintf(format, args...)
	return &wrapper{message, err}
}
