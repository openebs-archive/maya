// Copyright © 2018-2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"testing"
)

func TestCstorVolumeResize(t *testing.T) {
	tests := map[string]struct {
		ip       string
		volName  string
		capacity string
		isErr    bool
		errMsg   string
	}{
		"test 101": {"", "vol1", "", true, "failed to resize the cstor volume: 'missing ip address'"},
		"test 102": {"0.0.0.0", "", "", true, "failed to resize the cstor volume: 'missing volume name'"},
		"test 103": {"0.0.0.0", "pvc-21312-321312-321321-31231", "", true, "failed to resize the cstor volume: 'missing volume capacity'"},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := Command()
			cmd = WithData(cmd, "ip", mock.ip)
			cmd = WithData(cmd, "volname", mock.volName)
			cmd = WithData(cmd, "capacity", mock.capacity)

			c := &cstorVolumeResize{&cstorVolumeCommand{cmd}}

			result := c.Run()

			if mock.isErr {
				if mock.errMsg != result.Error().Error() {
					t.Fatalf("Test '%s' failed: expected error: %q actual error: %q", name, mock.errMsg, result.Error().Error())
				}
			}
		})
	}
}
