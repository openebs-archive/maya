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

import (
	"testing"
)

func TestJivaVolumeDelete(t *testing.T) {
	tests := map[string]struct {
		url   string
		iserr bool
	}{
		"test 101": {url: "", iserr: true},
		"test 102": {url: "http://", iserr: true},
		"test 103": {url: "http://1.1.1.1", iserr: true},
		"test 104": {url: "http://1.1.1.1:1010", iserr: true},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			j := &jivaVolumeDelete{
				cmd: Command(),
			}
			j.cmd = WithData(j.cmd, "url", mock.url)
			result := j.Run()

			if mock.iserr && result.Error() == nil {
				t.Fatalf("Test '%s' failed: expected error: actual no error: result '%s'", name, result)
			}
		})
	}
}
