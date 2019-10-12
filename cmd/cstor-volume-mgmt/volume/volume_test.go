// Copyright © 2017-2019 The OpenEBS Authors
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

package volume

import (
	"reflect"
	"testing"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func fakeStrToQuantity(capacity string) resource.Quantity {
	qntCapacity, _ := resource.ParseQuantity(capacity)
	//	fmt.Printf("Error: %v", err)
	return qntCapacity
}

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
					Capacity:          fakeStrToQuantity("5G"),
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

// TestCreateVolumeTarget is to test cStorVolume creation.
func TestCreateIstgtConf(t *testing.T) {
	testVolumeResource := map[string]struct {
		expectedErr bool
		test        *apis.CStorVolume
	}{
		"testcase1": {
			expectedErr: false,
			test: &apis.CStorVolume{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					Name: "testvol1",
					UID:  types.UID("abc"),
				},
				Spec: apis.CStorVolumeSpec{
					TargetIP:          "0.0.0.0",
					Capacity:          fakeStrToQuantity("5G"),
					Status:            "init",
					ReplicationFactor: 3,
					ConsistencyFactor: 2,
				},
			},
		},
		"testcase2": {
			expectedErr: true,
			test:        nil,
		},
	}
	for name, mock := range testVolumeResource {
		name := name // pin it
		mock := mock // pin it
		t.Run(name, func(t *testing.T) {
			_, err := CreateIstgtConf(mock.test)
			if mock.expectedErr && err == nil {
				t.Fatalf("test %q failed : expected error not to be nil but got %v", name, err)
			}
			if !mock.expectedErr && err != nil {
				t.Fatalf("test %q failed : expected error be nil but got %v", name, err)
			}
		})
	}
}

// TestCheckValidVolume tests volume related operations.
func TestCheckValidVolume(t *testing.T) {
	testVolumeResource := map[string]struct {
		expectedError bool
		test          *apis.CStorVolume
	}{
		"Invalid-volumeResource": {
			expectedError: true,
			test: &apis.CStorVolume{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					Name: "testvol1",
					UID:  types.UID(""),
				},
				Spec: apis.CStorVolumeSpec{
					TargetIP:          "0.0.0.0",
					Capacity:          fakeStrToQuantity("5G"),
					Status:            "init",
					ReplicationFactor: 3,
					ConsistencyFactor: 2,
				},
			},
		},
		"Invalid-cstorControllerIPEmpty": {
			expectedError: true,
			test: &apis.CStorVolume{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					Name: "testvol1",
					UID:  types.UID("123"),
				},
				Spec: apis.CStorVolumeSpec{
					TargetIP:          "",
					Capacity:          fakeStrToQuantity("5G"),
					Status:            "init",
					ReplicationFactor: 3,
					ConsistencyFactor: 2,
				},
			},
		},

		"Invalid-volumeNameEmpty": {
			expectedError: true,
			test: &apis.CStorVolume{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					Name: "",
					UID:  types.UID("123"),
				},
				Spec: apis.CStorVolumeSpec{
					TargetIP:          "0.0.0.0",
					Capacity:          fakeStrToQuantity("5G"),
					Status:            "init",
					ReplicationFactor: 3,
					ConsistencyFactor: 2,
				},
			},
		},
		"Invalid-volumeCapacityEmpty": {
			expectedError: true,
			test: &apis.CStorVolume{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					Name: "testvol1",
					UID:  types.UID("123"),
				},
				Spec: apis.CStorVolumeSpec{
					TargetIP:          "0.0.0.0",
					Capacity:          fakeStrToQuantity(""),
					Status:            "init",
					ReplicationFactor: 3,
					ConsistencyFactor: 2,
				},
			},
		},
		"Invalid-volumeCapacity": {
			expectedError: true,
			test: &apis.CStorVolume{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					Name: "testvol1",
					UID:  types.UID("123"),
				},
				Spec: apis.CStorVolumeSpec{
					TargetIP:          "0.0.0.0",
					Capacity:          fakeStrToQuantity("1B"),
					Status:            "init",
					ReplicationFactor: 3,
					ConsistencyFactor: 2,
				},
			},
		},
		"Invalid-ReplicationFactorEmpty": {
			expectedError: true,
			test: &apis.CStorVolume{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					Name: "testvol1",
					UID:  types.UID("123"),
				},
				Spec: apis.CStorVolumeSpec{
					TargetIP:          "0.0.0.0",
					Capacity:          fakeStrToQuantity("2G"),
					Status:            "init",
					ReplicationFactor: 0,
					ConsistencyFactor: 2,
				},
			},
		},
		"Invalid-ConsistencyFactorEmpty": {
			expectedError: true,
			test: &apis.CStorVolume{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					Name: "testvol1",
					UID:  types.UID("123"),
				},
				Spec: apis.CStorVolumeSpec{
					TargetIP:          "0.0.0.0",
					Capacity:          fakeStrToQuantity("2G"),
					Status:            "init",
					ReplicationFactor: 3,
					ConsistencyFactor: 0,
				},
			},
		},
		"Invalid-ReplicationFactorLessThanConsistencyFactor": {
			expectedError: true,
			test: &apis.CStorVolume{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					Name: "testvol1",
					UID:  types.UID("123"),
				},
				Spec: apis.CStorVolumeSpec{
					TargetIP:          "0.0.0.0",
					Capacity:          fakeStrToQuantity("2G"),
					Status:            "init",
					ReplicationFactor: 2,
					ConsistencyFactor: 3,
				},
			},
		},
		"invalid desired replication factor": {
			expectedError: true,
			test: &apis.CStorVolume{
				ObjectMeta: v1.ObjectMeta{
					Name: "testvol1",
					UID:  types.UID("123456"),
				},
				Spec: apis.CStorVolumeSpec{
					TargetIP:                 "0.0.0.0",
					Capacity:                 fakeStrToQuantity("5G"),
					Status:                   "init",
					ReplicationFactor:        3,
					ConsistencyFactor:        2,
					DesiredReplicationFactor: 0,
				},
				VersionDetails: apis.VersionDetails{
					Status: apis.VersionStatus{
						Current: "1.3.0",
					},
				},
			},
		},
		"empty desired replication factor for old volume": {
			expectedError: false,
			test: &apis.CStorVolume{
				ObjectMeta: v1.ObjectMeta{
					Name: "testvol1",
					UID:  types.UID("123456"),
				},
				Spec: apis.CStorVolumeSpec{
					TargetIP:                 "0.0.0.0",
					Capacity:                 fakeStrToQuantity("5G"),
					Status:                   "init",
					ReplicationFactor:        3,
					ConsistencyFactor:        2,
					DesiredReplicationFactor: 0,
				},
				VersionDetails: apis.VersionDetails{
					Status: apis.VersionStatus{
						Current: "1.2.0",
					},
				},
			},
		},
		"invalid desiredreplicationfactor/replicationfactor": {
			expectedError: true,
			test: &apis.CStorVolume{
				ObjectMeta: v1.ObjectMeta{
					Name: "testvol1",
					UID:  types.UID("123456"),
				},
				Spec: apis.CStorVolumeSpec{
					TargetIP:                 "0.0.0.0",
					Capacity:                 fakeStrToQuantity("5G"),
					Status:                   "init",
					ReplicationFactor:        4,
					ConsistencyFactor:        2,
					DesiredReplicationFactor: 3,
				},
				VersionDetails: apis.VersionDetails{
					Status: apis.VersionStatus{
						Current: "1.3.0",
					},
				},
			},
		},
		"valid resource": {
			expectedError: false,
			test: &apis.CStorVolume{
				ObjectMeta: v1.ObjectMeta{
					Name: "testvol1",
					UID:  types.UID("123456"),
				},
				Spec: apis.CStorVolumeSpec{
					TargetIP:                 "0.0.0.0",
					Capacity:                 fakeStrToQuantity("5G"),
					Status:                   "init",
					ReplicationFactor:        3,
					ConsistencyFactor:        2,
					DesiredReplicationFactor: 3,
				},
			},
		},
	}

	for name, mock := range testVolumeResource {
		name := name
		mock := mock
		t.Run(name, func(t *testing.T) {
			err := CheckValidVolume(mock.test)
			if mock.expectedError && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil but got :%v",
					name,
					err,
				)
			}
			if !mock.expectedError && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestExtractReplicaStatusFromJSON(t *testing.T) {
	type args struct {
		str string
	}
	tests := map[string]struct {
		str     string
		resp    *apis.CVStatus
		wantErr bool
	}{
		"two replicas with one HEALTHY and one BAD status": {
			`{
				"volumeStatus":[
				   {
						"name" : "pvc-c7f1a961-e0e3-11e8-b49d-42010a800233",
						"status": "Healthy",
						"replicaStatus" : [
						{
							"replicaId":"5523611450015704000",
							"mode":"HEALTHY",
							"checkpointedIOSeq":"0",
							"inflightRead":"0",
							"inflightWrite":"0",
							"inflightSync":"0",
							"upTime":1275
						},
						{
							"replicaId":"23523553",
							"mode":"BAD",
							"checkpointedIOSeq":"0",
							"inflightRead":"0",
							"inflightWrite":"0",
							"inflightSync":"0",
							"upTime":1375
						}
					  ]
				   }
				]
			 }`,
			&apis.CVStatus{
				Name:   "pvc-c7f1a961-e0e3-11e8-b49d-42010a800233",
				Status: "Healthy",
				ReplicaStatuses: []apis.ReplicaStatus{
					{
						ID:                "5523611450015704000",
						Mode:              "HEALTHY",
						CheckpointedIOSeq: "0",
						InflightRead:      "0",
						InflightWrite:     "0",
						InflightSync:      "0",
						UpTime:            1275,
					},
					{
						ID:                "23523553",
						Mode:              "BAD",
						CheckpointedIOSeq: "0",
						InflightRead:      "0",
						InflightWrite:     "0",
						InflightSync:      "0",
						UpTime:            1375,
					},
				},
			},
			false,
		},
		"incorrect value in replicaId": {
			`{
				"volumeStatus":[
				   {
						"name" : "pvc-c7f1a961-e0e3-11e8-b49d-42010a800233",
						"status": "Healthy",
						"replicaStatus" : [
						{
							"replicaId":5523611450015704000,
							"mode":"HEALTHY",
							"checkpointedIOSeq":"0",
							"inflightRead":"0",
							"inflightWrite":"0",
							"inflightSync":"0",
							"upTime":1275
						},
					  ]
				   }
				]
			 }`,
			&apis.CVStatus{
				Name:   "pvc-c7f1a961-e0e3-11e8-b49d-42010a800233",
				Status: "Healthy",
				ReplicaStatuses: []apis.ReplicaStatus{
					{
						ID:                "5523611450015704000",
						Mode:              "HEALTHY",
						CheckpointedIOSeq: "0",
						InflightRead:      "0",
						InflightWrite:     "0",
						InflightSync:      "0",
						UpTime:            1275,
					},
				},
			},
			true,
		},
		"valid single replica healthy status": {
			`{
				"volumeStatus":[
				   {
						"name" : "pvc-c7f1a961-e0e3-11e8-b49d-42010a800233",
						"status": "Healthy",
						"replicaStatus" : [
						{
							"replicaId":5523611450015704000,
							"Mode":"HEALTHY",
							"checkpointedIOSeq":"0",
							"Address" : "192.168.1.23",
							"inflightRead":"0",
							"inflightWrite":"0",
							"inflightSync":"0",
							"upTime":1275
						},
					  ]
				   }
				]
			 }`,
			&apis.CVStatus{
				Name:   "pvc-c7f1a961-e0e3-11e8-b49d-42010a800233",
				Status: "Healthy",
				ReplicaStatuses: []apis.ReplicaStatus{
					{
						ID:                "5523611450015704000",
						Mode:              "HEALTHY",
						CheckpointedIOSeq: "0",
						InflightRead:      "0",
						InflightWrite:     "0",
						InflightSync:      "0",
						UpTime:            1275,
					},
				},
			},
			true,
		},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := extractReplicaStatusFromJSON(mock.str)
			if err != nil {
				if !mock.wantErr {
					t.Errorf("extractReplicaStatusFromJSON() error = %v, wantErr %v", err != nil, mock.wantErr)
				}
			} else {
				if !reflect.DeepEqual(got, mock.resp) {
					t.Errorf("extractReplicaStatusFromJSON() = %v, want %v", got, mock.resp)
				}
			}
		})
	}
}
