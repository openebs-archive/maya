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

package task

import (
	"fmt"
	"strconv"
	"strings"
)

// TaskResultVerify helps in verifying specific data from the task's result
//
// NOTE:
//  A TaskResult is the result obtained after this task's execution.
type TaskResultVerify struct {
	// Count is the key used to hold the count related info
	Count string `json:"count"`
	// Split splits the result using this separator. This will split the task
	// result into an array of strings. A comma, space, colon, etc can be a valid
	// value of Split.
	//
	// ```yaml
	// split: `,`
	// ```
	//
	// This property can be used along with the other properties e.g. Count.
	Split string `json:"split"`
}

// TaskResultVerifyError represent task result verification related error
type TaskResultVerifyError struct {
	err string
}

func (e *TaskResultVerifyError) Error() string {
	return e.err
}

type taskResultVerifyExecutor struct {
	// property is the name of the task result property whose value is being
	// verified
	property string
	// actual represents the actual value which will be verified against the
	// expected
	actual string
	// expected represents the expected value
	expected TaskResultVerify
}

func newTaskResultVerifyExecutor(property, actual string, expected TaskResultVerify) *taskResultVerifyExecutor {
	return &taskResultVerifyExecutor{
		property: property,
		actual:   actual,
		expected: expected,
	}
}

// calculateCount gets the length of the array i.e. number of elements present
// in the value. Each of these elements are separated by the split separator.
func calculateCount(value, split string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimSuffix(value, split)
	valArr := strings.Split(value, split)
	return strconv.Itoa(len(valArr))
}

// isCount is count verification w.r.t task result property
func (t *taskResultVerifyExecutor) isCount() (ok bool, err error) {
	ok = true

	if len(t.expected.Count) == 0 {
		// no need to verify
		return
	}

	if len(t.actual) == 0 {
		ok = false
		err = &TaskResultVerifyError{fmt.Sprintf("%s's expected count: '%s' actual: '%s'", t.property, t.expected.Count, t.actual)}
	}

	count := t.actual
	// check if count needs to be calculated from the actual value
	if len(t.expected.Split) != 0 {
		count = calculateCount(t.actual, t.expected.Split)
	}

	if count != t.expected.Count {
		ok = false
		err = &TaskResultVerifyError{fmt.Sprintf("%s's expected count: '%s' actual count: '%s' actual: '%s'", t.property, t.expected.Count, count, t.actual)}
	}

	return
}

// verify verifies the task result property
//
// It supports count verification currently
func (t *taskResultVerifyExecutor) verify() (bool, error) {
	return t.isCount()
}
