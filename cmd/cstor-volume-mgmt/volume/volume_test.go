package volume

import (
	"fmt"
	"os"
	"testing"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type TestRunner struct{}

// RunCombinedOutput is to mock Real runner exec.
func (r TestRunner) RunCombinedOutput(command string, args ...string) ([]byte, error) {
	return []byte("success"), nil
}

// RunStdoutPipe is to mock real runner exec with stdoutpipe.
func (r TestRunner) RunStdoutPipe(command string, args ...string) ([]byte, error) {
	return []byte("successs"), nil
}

type TestFileOperator struct{}

//Write is to mock write operation for FileOperator interface
func (r TestFileOperator) Write(filename string, data []byte, perm os.FileMode) error {
	return nil
}

type TestUnixSock struct{}

//SendCommand for the real unix sock for the actual program,
func (r TestUnixSock) SendCommand(cmd string) error {
	return nil
}

// TestCreateVolume is to test cStorVolume creation.
func TestCreateVolume(t *testing.T) {
	testVolumeResource := map[string]struct {
		expectedError error
		test          *apis.CStorVolume
	}{
		"img1VolumeResource": {
			expectedError: nil,
			test: &apis.CStorVolume{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					UID: types.UID("abc"),
				},
				Spec: apis.CStorVolumeSpec{
					CStorControllerIP: "0.0.0.0",
					VolumeName:        "abc",
					VolumeID:          "abc",
					Capacity:          "5G",
					Status:            "init",
				},
				Status: apis.CStorVolumePhase{},
			},
		},
	}
	RunnerVar = TestRunner{}
	FileOperatorVar = TestFileOperator{}
	UnixSockVar = TestUnixSock{}
	obtainedErr := CreateVolume(testVolumeResource["img1VolumeResource"].test)
	if testVolumeResource["img1VolumeResource"].expectedError != obtainedErr {
		t.Fatalf("Expected: %v, Got: %v", testVolumeResource["img1VolumeResource"].expectedError, obtainedErr)
	}
}

// TestCheckValidVolume tests volume related operations.
func TestCheckValidVolume(t *testing.T) {
	testVolumeResource := map[string]struct {
		expectedError error
		test          *apis.CStorVolume
	}{
		"Invalid-volumeNameEmpty": {
			expectedError: fmt.Errorf("Volumename cannot be empty"),
			test: &apis.CStorVolume{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					UID: types.UID(""),
				},
				Spec: apis.CStorVolumeSpec{
					CStorControllerIP: "0.0.0.0",
					VolumeName:        "abc",
					VolumeID:          "abc",
					Capacity:          "5G",
					Status:            "init",
				},
				Status: apis.CStorVolumePhase{},
			},
		},
	}

	for desc, ut := range testVolumeResource {
		Obtainederr := CheckValidVolume(ut.test)
		if Obtainederr != nil {
			if Obtainederr.Error() != ut.expectedError.Error() {
				t.Fatalf("Desc : %v, Expected error: %v, Got : %v",
					desc, ut.expectedError, Obtainederr)
			}
		}

	}
}
