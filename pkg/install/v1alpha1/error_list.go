/*
Copyright 2018 The OpenEBS Authors

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

type errorList struct {
	errors []error
}

// addError adds an error to error list
func (l *errorList) addError(err error) []error {
	if err == nil {
		return l.errors
	}

	l.errors = append(l.errors, err)
	return l.errors
}

// addErrors adds a list of errors to error list
func (l *errorList) addErrors(errs []error) []error {
	for _, err := range errs {
		l.addError(err)
	}

	return l.errors
}
