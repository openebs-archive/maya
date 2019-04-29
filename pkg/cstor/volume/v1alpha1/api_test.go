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

	"github.com/openebs/maya/pkg/client/generated/cstor-volume-mgmt/v1alpha1"
)

type TestUnixSock struct{}

//SendCommand for the dummy unix sock for the test program,
func (r TestUnixSock) SendCommand(cmd string) ([]string, error) {
	ret := []string{"OK " + cmd}
	return ret, nil
}

func TestRunVolumeSnapCreateCommand(t *testing.T) {

	var sock TestUnixSock
	APIUnixSockVar = sock

	cases := map[string]struct {
		expectedError error
		test          *v1alpha1.VolumeSnapCreateRequest
	}{
		"successSnapshotCreate": {
			expectedError: nil,
			test: &v1alpha1.VolumeSnapCreateRequest{
				Version:  ProtocolVersion,
				Volume:   "dummyvol1",
				Snapname: "dummysnap1",
			},
		},
	}

	var s Server
	for i, c := range cases {
		t.Run(i, func(t *testing.T) {
			resp, obtainedErr := s.RunVolumeSnapCreateCommand(nil, c.test)

			if c.expectedError != obtainedErr {
				// XXX: this can be written in a more compact way. but keeping it this way
				//  as it is easy to understand this way.
				if c.expectedError != nil && obtainedErr != nil &&
					(c.expectedError.Error() == obtainedErr.Error()) {
					//got the expected error

				} else {
					t.Fatalf("Expected: %v, Got: %v, resp.Status: %v",
						c.expectedError, obtainedErr, resp.Status)
				}
			}
		})
	}
}

func TestRunVolumeSnapDeleteCommand(t *testing.T) {

	var sock TestUnixSock
	APIUnixSockVar = sock

	cases := map[string]struct {
		expectedError error
		test          *v1alpha1.VolumeSnapDeleteRequest
	}{
		"successSnapshotDestroy": {
			expectedError: nil,
			test: &v1alpha1.VolumeSnapDeleteRequest{
				Version:  ProtocolVersion,
				Volume:   "dummyvol1",
				Snapname: "dummysnap1",
			},
		},
	}

	var s Server
	for i, c := range cases {
		t.Run(i, func(t *testing.T) {
			resp, obtainedErr := s.RunVolumeSnapDeleteCommand(nil, c.test)

			if c.expectedError != obtainedErr {
				// XXX: this can be written in a more compact way. but keeping it this way
				//  as it is easy to understand this way.
				if c.expectedError != nil && obtainedErr != nil &&
					(c.expectedError.Error() == obtainedErr.Error()) {
					//got the expected error

				} else {
					t.Fatalf("Expected: %v, Got: %v, resp.Status: %v",
						c.expectedError, obtainedErr, resp.Status)
				}
			}
		})
	}
}
