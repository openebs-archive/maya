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
	"strings"
	"testing"

	fakeclientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned/fake"
	"github.com/openebs/maya/pkg/client/generated/cstor-volume-mgmt/v1alpha1"
)

type TestUnixSock struct{}

//SendCommand for the dummy unix sock for the test program,
func (r TestUnixSock) SendCommand(cmd string) ([]string, error) {
	ret := []string{}
	if strings.Contains(cmd, "SNAPCREATE") || strings.Contains(cmd, "SNAPDESTROY") {
		ret = append(ret, "OK "+cmd)
	} else if strings.Contains(cmd, "SNAPLIST") {
		ret = append(ret, "SNAPLIST {\"snapshots\":[{\"replica_id\":\"E89CDC9473E86A938C51F048DE8341E5\",\"snapshot\":[{\"name\":\"pvc-e4e29de2-52d2-11ea-b66e-42010a9a0080_snap2_1582091556611449485\",\"properties\":{\"io.openebs:volname\":\"pvc-e4e29de2-52d2-11ea-b66e-42010a9a0080-cstor-pool-7sqf\",\"io.openebs:livenesstimestamp\":\"1582113150\",\"io.openebs:poolname\":\"cstor-pool-7sqf\",\"refcompressratio\":\"108\",\"logicalreferenced\":\"723259904\",\"compressratio\":\"108\",\"used\":\"6144\",\"available\":\"9667151360\",\"referenced\":\"668972544\",\"creation\":\"1582091556\",\"createtxg\":\"732\",\"refquota\":\"0\",\"refreservation\":\"0\",\"guid\":\"2099668115632562841\",\"unique\":\"6144\",\"objsetid\":\"18\",\"userrefs\":\"0\",\"defer_destroy\":\"0\",\"written\":\"321503744\",\"type\":\"3\",\"useraccounting\":\"0\",\"volsize\":\"10737418240\",\"volblocksize\":\"4096\"}},{\"name\":\"pvc-e4e29de2-52d2-11ea-b66e-42010a9a0080_snap1_1582088954475131276\",\"properties\":{\"io.openebs:volname\":\"pvc-e4e29de2-52d2-11ea-b66e-42010a9a0080-cstor-pool-7sqf\",\"io.openebs:livenesstimestamp\":\"1582113150\",\"io.openebs:poolname\":\"cstor-pool-7sqf\",\"refcompressratio\":\"114\",\"logicalreferenced\":\"396610560\",\"compressratio\":\"114\",\"used\":\"104960\",\"available\":\"9667151360\",\"referenced\":\"347573760\",\"creation\":\"1582088954\",\"createtxg\":\"275\",\"refquota\":\"0\",\"refreservation\":\"0\",\"guid\":\"4881578479253596844\",\"unique\":\"104960\",\"objsetid\":\"125\",\"userrefs\":\"0\",\"defer_destroy\":\"0\",\"written\":\"347573760\",\"type\":\"3\",\"useraccounting\":\"0\",\"volsize\":\"10737418240\",\"volblocksize\":\"4096\"}}]}]}")
	}
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

	s := Server{
		Client: fakeclientset.NewSimpleClientset(),
	}
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

	s := Server{
		Client: fakeclientset.NewSimpleClientset(),
	}
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
