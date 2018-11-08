package volume

import (
	"fmt"
	"reflect"
	"testing"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// TestCreateVolumeTarget is to test cStorVolume creation.
func TestCreateVolumeTarget(t *testing.T) {
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
					TargetIP:          "0.0.0.0",
					Capacity:          "5G",
					Status:            "init",
					ReplicationFactor: 3,
					ConsistencyFactor: 2,
				},
			},
		},
	}
	FileOperatorVar = util.TestFileOperator{}
	UnixSockVar = util.TestUnixSock{}
	obtainedErr := CreateVolumeTarget(testVolumeResource["img1VolumeResource"].test)
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
					TargetIP:          "0.0.0.0",
					Capacity:          "5G",
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
					TargetIP:          "",
					Capacity:          "5G",
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
					TargetIP:          "0.0.0.0",
					Capacity:          "5G",
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
					TargetIP:          "0.0.0.0",
					Capacity:          "",
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
					TargetIP:          "0.0.0.0",
					Capacity:          "2G",
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
					TargetIP:          "0.0.0.0",
					Capacity:          "2G",
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
					TargetIP:          "0.0.0.0",
					Capacity:          "2G",
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

func TestExtractReplicaStatusFromJSON(t *testing.T) {
	type args struct {
		str string
	}
	tests := map[string]struct {
		str     string
		volName string
		resp    []string
		wantErr bool
	}{
		"two replicas with one HEALTHY and one BAD status": {
			`{
				"Replica status":[
				   {
					  "pvc-c7f1a961-e0e3-11e8-b49d-42010a800233":[
						 {
							"replica":-5523611450015704000,
							"status":"HEALTHY",
							"checkpointed_io_seq":0,
							"inflight_read":0,
							"inflight_write":0,
							"inflight_sync":0,
							"connected since(in seconds)":1275
						 },
						 {
							"replica":-23523553,
							"status":"BAD",
							"checkpointed_io_seq":0,
							"inflight_read":0,
							"inflight_write":0,
							"inflight_sync":0,
							"connected since(in seconds)":1375
						 }
					  ]
				   }
				]
			 }`,
			"pvc-c7f1a961-e0e3-11e8-b49d-42010a800233",
			[]string{"HEALTHY", "BAD"},
			false,
		},
		"One replica with HEALTHY status": {
			`{
				"Replica status":[
				   {
					  "pvc-c7f1a961-e0e3-11e8-b49d-42010a800233":[
						 {
							"replica":-5523611450015704000,
							"status":"HEALTHY",
							"checkpointed_io_seq":0,
							"inflight_read":0,
							"inflight_write":0,
							"inflight_sync":0,
							"connected since(in seconds)":1275
						 }
					  ]
				   }
				]
			 }`,
			"pvc-c7f1a961-e0e3-11e8-b49d-42010a800233",
			[]string{"HEALTHY"},
			false,
		},
		"Missing replicas": {
			`{
				"Replica status":[{"pvc-389710cf-e351-11e8-b49d-42010a800233":[]}]
			 }`,
			"pvc-c7f1a961-e0e3-11e8-b49d-42010a800233",
			nil,
			false,
		},
		"Invalid input": {
			`{
				"Replica status":[asd]
			 }`,
			"pvc-c7f1a961-e0e3-11e8-b49d-42010a800233",
			nil,
			true,
		},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := extractReplicaStatusFromJSON(mock.volName, mock.str)
			if (err != nil) != mock.wantErr {
				t.Errorf("extractReplicaStatusFromJSON() error = %v, wantErr %v", err != nil, mock.wantErr)
			}
			if !reflect.DeepEqual(got, mock.resp) {
				t.Errorf("extractReplicaStatusFromJSON() = %v, want %v", got, mock.resp)
			}
		})
	}
}
