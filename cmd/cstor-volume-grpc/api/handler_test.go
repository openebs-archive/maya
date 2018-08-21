package api

import (
	"fmt"
	"testing"

	"github.com/openebs/maya/pkg/client/generated/cstor-volume-grpc/v1alpha1"
)

type TestUnixSock struct{}

//SendCommand for the dummy unix sock for the test program,
func (r TestUnixSock) SendCommand(cmd string) ([]string, error) {
	ret := []string{"OK " + cmd}
	return ret, nil
}

func TestRunVolumeCommand(t *testing.T) {

	var sock TestUnixSock
	APIUnixSockVar = sock

	cases := map[string]struct {
		expectedError error
		test          *v1alpha1.VolumeSnapRequest
	}{
		"successSnapshotCreate": {
			expectedError: nil,
			test: &v1alpha1.VolumeSnapRequest{
				Command:  CmdSnapCreate,
				Volume:   "dummyvol1",
				Snapname: "dummysnap1",
			},
		},

		"successSnapshotDestroy": {
			expectedError: nil,
			test: &v1alpha1.VolumeSnapRequest{
				Command:  CmdSnapDestroy,
				Volume:   "dummyvol1",
				Snapname: "dummysnap1",
			},
		},

		"failureInvalidCommand": {
			expectedError: fmt.Errorf("Invalid VolumeCommand : SNAPDOSOMETHING"),
			test: &v1alpha1.VolumeSnapRequest{
				Command:  "SNAPDOSOMETHING",
				Volume:   "dummyvol1",
				Snapname: "dummysnap1",
			},
		},
	}

	var s Server
	for i, c := range cases {
		t.Run(i, func(t *testing.T) {
			resp, obtainedErr := s.RunVolumeSnapCommand(nil, c.test)

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
