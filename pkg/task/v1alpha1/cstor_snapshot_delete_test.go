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

func TestCstorSnapshotDelete(t *testing.T) {
	tests := map[string]struct {
		ip       string
		volName  string
		snapName string
		isErr    bool
		errMsg   string
	}{
		"test 101": {ip: "", volName: "vol1", snapName: "", isErr: true, errMsg: "failed to delete cstor snapshot: missing ip address"},
		"test 102": {ip: "0.0.0.0", volName: "", snapName: "", isErr: true, errMsg: "failed to delete cstor snapshot: missing volume name"},

		// TODO
		// Move this test to integration test or something that is more manageable
		// make test times out with this
		//
		//"test 103": {ip: "0.0.0.0", volName: "vol1", snapName: "s1", isErr: true, errMsg: `Error when calling RunVolumeSnapDeleteCommand: rpc error: code = Unavailable desc = all SubConns are in TransientFailure, latest connection error: connection error: desc = "transport: Error while dialing dial tcp 1.1.1.1:7777: i/o timeout"`},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := Command()
			cmd = WithData(cmd, "ip", mock.ip)
			cmd = WithData(cmd, "ip", mock.ip)
			cmd = WithData(cmd, "volname", mock.volName)
			cmd = WithData(cmd, "snapname", mock.snapName)

			c := &cstorSnapshotDelete{&cstorSnapshotCommand{cmd}}

			result := c.Run()

			if mock.isErr {
				if mock.errMsg != result.Error().Error() {
					t.Fatalf("Test '%s' failed: expected error: %q actual error: %q", name, mock.errMsg, result.Error().Error())
				}
			}
		})
	}
}
