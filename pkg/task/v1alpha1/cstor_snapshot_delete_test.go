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
		"test 101": {ip: "", volName: "vol1", snapName: "", isErr: true, errMsg: "missing ip address: failed to delete cstor snapshot"},
		"test 102": {ip: "1.1.1.1", volName: "", snapName: "", isErr: true, errMsg: "missing volume name: failed to delete cstor snapshot"},
		"test 103": {ip: "1.1.1.1", volName: "vol1", snapName: "s1", isErr: true, errMsg: `error when calling RunVolumeSnapDeleteCommand: rpc error: code = Unavailable desc = all SubConns are in TransientFailure, latest connection error: connection error: desc = "transport: Error while dialing dial tcp 1.1.1.1:7777: i/o timeout"`},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
<<<<<<< HEAD
			c := &cstorSnapshotDelete{
				cmd: Command(),
			}
			c.cmd = WithData(c.cmd, "ip", mock.ip)
			c.cmd = WithData(c.cmd, "volname", mock.volName)
			c.cmd = WithData(c.cmd, "snapname", mock.snapName)
			result := c.Run()

			if mock.isErr {
				if mock.errMsg != result.Error().Error() {
					t.Fatalf("Test '%s' failed: expected error: %q actual error: %q", name, mock.errMsg, result.Error().Error())
				}
=======
			c := &cstorSnapshotCreate{
				cmd: Command(),
			}
			c.cmd = WithData(c.cmd, "url", mock.url)
			c.cmd = WithData(c.cmd, "url", mock.url)
			c.cmd = WithData(c.cmd, "url", mock.url)
			result := c.Run()

			if mock.iserr && result.Error() == nil {
				t.Fatalf("Test '%s' failed: expected error: actual no error: result '%s'", name, result)
>>>>>>> 346aee0a... temp commit
			}
		})
	}
}
