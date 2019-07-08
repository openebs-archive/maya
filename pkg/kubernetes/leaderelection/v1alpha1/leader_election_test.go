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

package v1alpha1

import (
	"testing"
)

func Test_sanitizeName(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output string
	}{
		{
			"requires no change",
			"test-driver",
			"test-driver",
		},
		{
			"has characters that should be replaced",
			"test!driver/foo",
			"test-driver-foo",
		},
		{
			"has trailing space",
			"driver\\",
			"driver-X",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			output := sanitizeName(test.input)
			if output != test.output {
				t.Logf("expected name: %q", test.output)
				t.Logf("actual name: %q", output)
				t.Errorf("unexpected santized name")
			}
		})
	}
}
