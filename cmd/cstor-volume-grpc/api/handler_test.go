package api

import (
	"fmt"
	"testing"
)

type TestUnixSock struct{}

//SendCommand for the dummy unix sock for the test program,
func (r TestUnixSock) SendCommand(cmd string) ([]string, error) {
	ret := []string{"OK " + cmd}
	return ret, nil
}

func TestRunVolumeCommand(t *testing.T) {

	var sock TestUnixSock
	ApiUnixSockVar = sock

	cases := map[string]struct {
		expectedError error
		test          *VolumeCommand
	}{
		"successSnapshotCreate": {
			expectedError: nil,
			test: &VolumeCommand{
				Command:  CmdSnapCreate,
				Volume:   "dummyvol1",
				Snapname: "dummysnap1",
				Status:   "creating",
			},
		},

		"successSnapshotDestroy": {
			expectedError: nil,
			test: &VolumeCommand{
				Command:  CmdSnapDestroy,
				Volume:   "dummyvol1",
				Snapname: "dummysnap1",
				Status:   "destroying",
			},
		},

		"failureInvalidCommand": {
			expectedError: fmt.Errorf("Invalid VolumeCommand : SNAPDOSOMETHING"),
			test: &VolumeCommand{
				Command:  "SNAPDOSOMETHING",
				Volume:   "dummyvol1",
				Snapname: "dummysnap1",
				Status:   "SNAPDOSOMETHING",
			},
		},
	}

	var s Server
	for i, c := range cases {
		t.Run(i, func(t *testing.T) {
			resp, obtainedErr := s.RunVolumeCommand(nil, c.test)

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
