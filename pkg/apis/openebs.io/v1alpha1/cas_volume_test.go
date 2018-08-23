package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetField(t *testing.T) {
	tests := map[string]struct {
		fieldName      string
		expectedOutput string
		Volume         CASVolume
	}{
		"Fetching CasType when CasType is Jiva": {
			fieldName: "CasType",
			Volume: CASVolume{
				Spec: CASVolumeSpec{
					CasType: "jiva",
				},
			},
			expectedOutput: "jiva",
		},
		"Fetching CasType when CasType is cstor": {
			fieldName: "CasType",
			Volume: CASVolume{
				Spec: CASVolumeSpec{
					CasType: "cstor",
				},
			},
			expectedOutput: "cstor",
		},
		"Fetching CasType when CasType is none": {
			fieldName: "CasType",
			Volume: CASVolume{
				Spec: CASVolumeSpec{
					CasType: "",
				},
			},
			expectedOutput: "jiva",
		},
		"Fetching ClusterIP from openebs.io/cluster-ips": {
			fieldName: "ClusterIP",
			Volume: CASVolume{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"openebs.io/cluster-ips": "192.168.100.1",
					},
				},
			},
			expectedOutput: "192.168.100.1",
		},
		"Fetching ClusterIP from vsm.openebs.io/cluster-ips": {
			fieldName: "ClusterIP",
			Volume: CASVolume{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"vsm.openebs.io/cluster-ips": "192.168.100.1",
					},
				},
			},
			expectedOutput: "192.168.100.1",
		},
		"Fetching ClusterIP when both keys are not present": {
			fieldName: "ClusterIP",
			Volume: CASVolume{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			expectedOutput: "",
		},
		"Fetching Controller status from openebs.io/controller-status": {
			fieldName: "ControllerStatus",
			Volume: CASVolume{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"openebs.io/controller-status": "running",
					},
				},
			},
			expectedOutput: "running",
		},
		"Fetching Controller status from vsm.openebs.io/controller-status": {
			fieldName: "ControllerStatus",
			Volume: CASVolume{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"vsm.openebs.io/controller-status": "running",
					},
				},
			},
			expectedOutput: "running",
		},
		"Fetching Controller status when both keys are not present": {
			fieldName: "ControllerStatus",
			Volume: CASVolume{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			expectedOutput: "",
		},
		"Fetching IQN": {
			fieldName: "IQN",
			Volume: CASVolume{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
				Spec: CASVolumeSpec{
					Iqn: "iqn.2016-09.com.openebs.cstor:default-testclaim7",
				},
			},
			expectedOutput: "iqn.2016-09.com.openebs.cstor:default-testclaim7",
		},
		"Fetching VolumeInfo": {
			fieldName: "VolumeName",
			Volume: CASVolume{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
					Name:        "default-testclaim",
				},
			},
			expectedOutput: "default-testclaim",
		},
		"Fetching TargetPortal": {
			fieldName: "TargetPortal",
			Volume: CASVolume{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
				Spec: CASVolumeSpec{
					TargetPortal: "10.63.247.173:3260",
				},
			},
			expectedOutput: "10.63.247.173:3260",
		},
		"Fetching VolumeSize": {
			fieldName: "VolumeSize",
			Volume: CASVolume{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
				Spec: CASVolumeSpec{
					Capacity: "5G",
				},
			},
			expectedOutput: "5G",
		},
		"Fetching ReplicaCount": {
			fieldName: "ReplicaCount",
			Volume: CASVolume{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
				Spec: CASVolumeSpec{
					Replicas: "3",
				},
			},
			expectedOutput: "3",
		},
		"Fetching ReplicaStatus from openebs.io/replica-status": {
			fieldName: "ReplicaStatus",
			Volume: CASVolume{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"openebs.io/replica-status": "running, running, running",
					},
				},
			},
			expectedOutput: "running, running, running",
		},
		"Fetching ReplicaStatus from vsm.openebs.io/replica-status": {
			fieldName: "ReplicaStatus",
			Volume: CASVolume{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"vsm.openebs.io/replica-status": "running, running, running",
					},
				},
			},
			expectedOutput: "running, running, running",
		},
		"Fetching ReplicaStatus when no key is present": {
			fieldName: "ReplicaStatus",
			Volume: CASVolume{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			expectedOutput: "",
		},
		"Fetching ReplicaIP from openebs.io/replica-ips": {
			fieldName: "ReplicaIP",
			Volume: CASVolume{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"openebs.io/replica-ips": "10.60.0.11, 10.60.1.16, 10.60.2.10",
					},
				},
			},
			expectedOutput: "10.60.0.11, 10.60.1.16, 10.60.2.10",
		},
		"Fetching ReplicaIP from vsm.openebs.io/replica-ips": {
			fieldName: "ReplicaIP",
			Volume: CASVolume{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"vsm.openebs.io/replica-ips": "10.60.0.11, 10.60.1.16, 10.60.2.10",
					},
				},
			},
			expectedOutput: "10.60.0.11, 10.60.1.16, 10.60.2.10",
		},
		"Fetching ReplicaIP when no key is present": {
			fieldName: "ReplicaIP",
			Volume: CASVolume{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			expectedOutput: "",
		},
		"Fetching Invalid key": {
			fieldName: "test",
			Volume: CASVolume{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			expectedOutput: "N/A",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.Volume.GetField(tt.fieldName)
			if got != tt.expectedOutput {
				t.Fatalf("Test: %v Expected: %v but got: %v", name, tt.expectedOutput, got)
			}
		})
	}
}
