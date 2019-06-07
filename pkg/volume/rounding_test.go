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

package volume

import (
	"testing"
)

func TestByteCount(t *testing.T) {
	tests := map[string]struct {
		size           uint64
		expectedOutput string
	}{
		"2T in bytes": {
			size:           2199023255552,
			expectedOutput: "2T",
		},
		"10G in bytes": {
			size:           10737418240,
			expectedOutput: "10G",
		},
		"100G in bytes": {
			size:           107374182400,
			expectedOutput: "100G",
		},
		"512M in bytes": {
			size:           536870912,
			expectedOutput: "512M",
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			output := ByteCount(test.size)
			if test.expectedOutput != output {
				t.Fatalf("Test %q failed: expected {%s} got {%s}",
					name, test.expectedOutput, output)
			}
		})
	}
}
