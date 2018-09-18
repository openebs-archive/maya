package volume

import (
	"fmt"
	"testing"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

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
					Name: "testvol1",
					UID:  types.UID("abc"),
				},
				Spec: apis.CStorVolumeSpec{
					ISCSISpec: apis.ISCSISpec{
						TargetIPSpec: apis.TargetIPSpec{
							"0.0.0.0",
						},
					},
					CapacitySpec: apis.CapacitySpec{
						"5G",
					},
					Status:            "init",
					ReplicationFactor: 3,
					ConsistencyFactor: 2,
				},
			},
		},
	}
	FileOperatorVar = util.TestFileOperator{}
	UnixSockVar = util.TestUnixSock{}
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
		"Invalid-volumeResource": {
			expectedError: fmt.Errorf("Invalid volume resource"),
			test: &apis.CStorVolume{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					Name: "testvol1",
					UID:  types.UID(""),
				},
				Spec: apis.CStorVolumeSpec{
					ISCSISpec: apis.ISCSISpec{
						TargetIPSpec: apis.TargetIPSpec{
							"0.0.0.0",
						},
					},
					CapacitySpec: apis.CapacitySpec{
						"5G",
					},
					Status:            "init",
					ReplicationFactor: 3,
					ConsistencyFactor: 2,
				},
			},
		},
		"Invalid-cstorControllerIPEmpty": {
			expectedError: fmt.Errorf("targetIP cannot be empty"),
			test: &apis.CStorVolume{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					Name: "testvol1",
					UID:  types.UID("123"),
				},
				Spec: apis.CStorVolumeSpec{
					ISCSISpec: apis.ISCSISpec{
						TargetIPSpec: apis.TargetIPSpec{
							"",
						},
					},
					CapacitySpec: apis.CapacitySpec{
						"5G",
					},
					Status:            "init",
					ReplicationFactor: 3,
					ConsistencyFactor: 2,
				},
			},
		},

		"Invalid-volumeNameEmpty": {
			expectedError: fmt.Errorf("volumeName cannot be empty"),
			test: &apis.CStorVolume{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					Name: "",
					UID:  types.UID("123"),
				},
				Spec: apis.CStorVolumeSpec{
					ISCSISpec: apis.ISCSISpec{
						TargetIPSpec: apis.TargetIPSpec{
							"0.0.0.0",
						},
					},
					CapacitySpec: apis.CapacitySpec{
						"5G",
					},
					Status:            "init",
					ReplicationFactor: 3,
					ConsistencyFactor: 2,
				},
			},
		},
		"Invalid-volumeCapacityEmpty": {
			expectedError: fmt.Errorf("capacity cannot be empty"),
			test: &apis.CStorVolume{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					Name: "testvol1",
					UID:  types.UID("123"),
				},
				Spec: apis.CStorVolumeSpec{
					ISCSISpec: apis.ISCSISpec{
						TargetIPSpec: apis.TargetIPSpec{
							"0.0.0.0",
						},
					},
					CapacitySpec: apis.CapacitySpec{
						"",
					},
					Status:            "init",
					ReplicationFactor: 3,
					ConsistencyFactor: 2,
				},
			},
		},
		"Invalid-ReplicationFactorEmpty": {
			expectedError: fmt.Errorf("replicationFactor cannot be zero"),
			test: &apis.CStorVolume{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					Name: "testvol1",
					UID:  types.UID("123"),
				},
				Spec: apis.CStorVolumeSpec{
					ISCSISpec: apis.ISCSISpec{
						TargetIPSpec: apis.TargetIPSpec{
							"0.0.0.0",
						},
					},
					CapacitySpec: apis.CapacitySpec{
						"2G",
					},
					Status:            "init",
					ReplicationFactor: 0,
					ConsistencyFactor: 2,
				},
			},
		},
		"Invalid-ConsistencyFactorEmpty": {
			expectedError: fmt.Errorf("consistencyFactor cannot be zero"),
			test: &apis.CStorVolume{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					Name: "testvol1",
					UID:  types.UID("123"),
				},
				Spec: apis.CStorVolumeSpec{
					ISCSISpec: apis.ISCSISpec{
						TargetIPSpec: apis.TargetIPSpec{
							"0.0.0.0",
						},
					},
					CapacitySpec: apis.CapacitySpec{
						"2G",
					},
					Status:            "init",
					ReplicationFactor: 3,
					ConsistencyFactor: 0,
				},
			},
		},
		"Invalid-ReplicationFactorLessThanConsistencyFactor": {
			expectedError: fmt.Errorf("replicationFactor cannot be less than consistencyFactor"),
			test: &apis.CStorVolume{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					Name: "testvol1",
					UID:  types.UID("123"),
				},
				Spec: apis.CStorVolumeSpec{
					ISCSISpec: apis.ISCSISpec{
						TargetIPSpec: apis.TargetIPSpec{
							"0.0.0.0",
						},
					},
					CapacitySpec: apis.CapacitySpec{
						"2G",
					},
					Status:            "init",
					ReplicationFactor: 2,
					ConsistencyFactor: 3,
				},
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
