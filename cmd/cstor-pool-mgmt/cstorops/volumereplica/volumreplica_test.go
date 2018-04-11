package volumereplica

import (
	"fmt"
	"reflect"
	"testing"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestCheckValidVolumeReplica tests VolumeReplica related operations
func TestCheckValidVolumeReplica(t *testing.T) {
	testVolumeReplicaResource := map[string]struct {
		expectedError error
		test          *apis.CStorVolumeReplica
	}{
		"Valid-vol1Resource": {
			expectedError: nil,
			test: &apis.CStorVolumeReplica{
				ObjectMeta: metav1.ObjectMeta{Name: "VolumeReplicaResource0"},
				Spec: apis.CStorVolumeReplicaSpec{
					CStorControllerIP: "10.210.110.121",
					VolName:           "vol0",
					Capacity:          "100MB",
				},
			},
		},
		"Invalid-volNameEmpty": {
			expectedError: fmt.Errorf("Volume name cannot be empty"),
			test: &apis.CStorVolumeReplica{
				ObjectMeta: metav1.ObjectMeta{Name: "VolumeReplicaResource1"},
				Spec: apis.CStorVolumeReplicaSpec{
					CStorControllerIP: "10.210.110.121",
					VolName:           "",
					Capacity:          "100MB",
				},
			},
		},
		"Invalid-CapacityEmpty": {
			expectedError: fmt.Errorf("Capacity cannot be empty"),
			test: &apis.CStorVolumeReplica{
				ObjectMeta: metav1.ObjectMeta{Name: "VolumeReplicaResource2"},
				Spec: apis.CStorVolumeReplicaSpec{
					CStorControllerIP: "10.210.110.121",
					VolName:           "abcdefgh_Volume_2",
					Capacity:          "",
				},
			},
		},
	}

	for desc, ut := range testVolumeReplicaResource {
		Obtainederr := CheckValidVolumeReplica(ut.test)
		if Obtainederr != nil {
			if Obtainederr.Error() == ut.expectedError.Error() {
				return
			}
			t.Fatalf("Desc : %v, Expected error: %v, Got : %v",
				desc, ut.expectedError, Obtainederr)
		}

	}
}

// TestCheckValidPool tests pool related operations
func TestCreateVolumeReplicaBuilder(t *testing.T) {
	testVolumeReplicaResource := map[string]struct {
		expectedCmd []string
		test        *apis.CStorVolumeReplica
	}{
		"vol1Resource": {
			expectedCmd: []string{VolumeReplicaOperator + " create -s -v 100MB pool1/vol1"},
			test: &apis.CStorVolumeReplica{
				ObjectMeta: metav1.ObjectMeta{Name: "VolumeReplicaResource1"},
				Spec: apis.CStorVolumeReplicaSpec{
					CStorControllerIP: "10.210.110.121",
					VolName:           "vol1",
					Capacity:          "100MB",
				},
			},
		},

		"vol2Resource": {
			expectedCmd: []string{VolumeReplicaOperator + " create -s -v 100MB pool1/vol2"},
			test: &apis.CStorVolumeReplica{
				ObjectMeta: metav1.ObjectMeta{Name: "VolumeReplicaResource2"},
				Spec: apis.CStorVolumeReplicaSpec{
					CStorControllerIP: "10.210.110.121",
					VolName:           "vol2",
					Capacity:          "100MB",
				},
			},
		},
	}

	for desc, ut := range testVolumeReplicaResource {
		obtainedCmd := createVolumeBuilder(ut.test, "pool1/"+ut.test.Spec.VolName)
		if reflect.DeepEqual(ut.expectedCmd, obtainedCmd.Args) {
			t.Fatalf("desc:%v, Commands mismatch, expected:%v, Got:%v", desc,
				ut.expectedCmd, obtainedCmd.Args[0])
		}
	}
}
