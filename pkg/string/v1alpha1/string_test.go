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

package strings

import (
	"testing"
)

func TestContains(t *testing.T) {
	tests := map[string]struct {
		array     []string
		element   string
		isPresent bool
	}{
		"contains string - positive test case - element is present": {
			array:     []string{"hi", "hello"},
			element:   "hello",
			isPresent: true,
		},
		"contains string - positive test case - element is not present": {
			array:     []string{"there", "you", "go"},
			element:   "no",
			isPresent: false,
		},
		"contains string - boundary test case - similar elements but not same": {
			array:     []string{"hi there", "ok now"},
			element:   "hi there ",
			isPresent: false,
		},
	}

	for name, mock := range tests {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			isPresent := MakeList(mock.array...).Contains(mock.element)
			if mock.isPresent != isPresent {
				t.Fatalf("failed to test contains string: expected element '%t': actual element '%t'", mock.isPresent, isPresent)
			}
		})
	}
}
