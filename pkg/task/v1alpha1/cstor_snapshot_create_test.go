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

func TestCstorSnapshotCreate(t *testing.T) {
	tests := map[string]struct {
		ip       string
		volName  string
		snapName string
		isErr    bool
		errMsg   string
	}{
		"test 101": {"", "vol1", "", true, "failed to create cstor snapshot: missing ip address"},
		"test 102": {"1.1.1.1", "", "", true, "failed to create cstor snapshot: missing volume name"},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := Command()
			cmd = WithData(cmd, "ip", mock.ip)
			cmd = WithData(cmd, "volname", mock.volName)
			cmd = WithData(cmd, "snapname", mock.snapName)

			c := &cstorSnapshotCreate{&cstorSnapshotCommand{cmd}}

			result := c.Run()

			if mock.isErr {
				if mock.errMsg != result.Error().Error() {
					t.Fatalf("Test '%s' failed: expected error: %q actual error: %q", name, mock.errMsg, result.Error().Error())
				}
			}
		})
	}
}
