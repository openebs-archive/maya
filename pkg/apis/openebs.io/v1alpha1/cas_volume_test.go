package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetCASType(t *testing.T) {
	tests := map[string]struct {
		expectedOutput string
		Volume         CASVolume
	}{
		"Fetching CasType when CasType is Jiva": {
			Volume: CASVolume{
				Spec: CASVolumeSpec{
					CasType: "jiva",
				},
			},
			expectedOutput: "jiva",
		},
		"Fetching CasType when CasType is cstor": {
			Volume: CASVolume{
				Spec: CASVolumeSpec{
					CasType: "cstor",
				},
			},
			expectedOutput: "cstor",
		},
		"Fetching CasType when CasType is none": {
			Volume: CASVolume{
				Spec: CASVolumeSpec{
					CasType: "",
				},
			},
			expectedOutput: "jiva",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.Volume.GetCASType()
			if got != tt.expectedOutput {
				t.Fatalf("Test: %v Expected: %v but got: %v", name, tt.expectedOutput, got)
			}
		})
	}
}

func TestGetClusterIP(t *testing.T) {
	tests := map[string]struct {
		expectedOutput string
		Volume         CASVolume
	}{
		"Fetching ClusterIP from openebs.io/cluster-ips": {
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
			Volume: CASVolume{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			expectedOutput: "",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.Volume.GetClusterIP()
			if got != tt.expectedOutput {
				t.Fatalf("Test: %v Expected: %v but got: %v", name, tt.expectedOutput, got)
			}
		})
	}
}

func TestGetControllerStatus(t *testing.T) {
	tests := map[string]struct {
		expectedOutput string
		Volume         CASVolume
	}{
		"Fetching Controller status from openebs.io/controller-status": {
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
			Volume: CASVolume{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			expectedOutput: "",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.Volume.GetControllerStatus()
			if got != tt.expectedOutput {
				t.Fatalf("Test: %v Expected: %v but got: %v", name, tt.expectedOutput, got)
			}
		})
	}
}

func TestGetIQN(t *testing.T) {
	tests := map[string]struct {
		expectedOutput string
		Volume         CASVolume
	}{
		"Fetching IQN": {
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
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.Volume.GetIQN()
			if got != tt.expectedOutput {
				t.Fatalf("Test: %v Expected: %v but got: %v", name, tt.expectedOutput, got)
			}
		})
	}
}

func TestGetVolumeName(t *testing.T) {
	tests := map[string]struct {
		expectedOutput string
		Volume         CASVolume
	}{
		"Fetching VolumeInfo": {
			Volume: CASVolume{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
					Name:        "default-testclaim",
				},
			},
			expectedOutput: "default-testclaim",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.Volume.GetVolumeName()
			if got != tt.expectedOutput {
				t.Fatalf("Test: %v Expected: %v but got: %v", name, tt.expectedOutput, got)
			}
		})
	}
}

func TestGetTargetPortal(t *testing.T) {
	tests := map[string]struct {
		expectedOutput string
		Volume         CASVolume
	}{
		"Fetching TargetPortal": {
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
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.Volume.GetTargetPortal()
			if got != tt.expectedOutput {
				t.Fatalf("Test: %v Expected: %v but got: %v", name, tt.expectedOutput, got)
			}
		})
	}
}

func TestGetVolumeSize(t *testing.T) {
	tests := map[string]struct {
		expectedOutput string
		Volume         CASVolume
	}{
		"Fetching VolumeSize": {
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
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.Volume.GetVolumeSize()
			if got != tt.expectedOutput {
				t.Fatalf("Test: %v Expected: %v but got: %v", name, tt.expectedOutput, got)
			}
		})
	}
}

func TestGetReplicaCount(t *testing.T) {
	tests := map[string]struct {
		expectedOutput string
		Volume         CASVolume
	}{
		"Fetching ReplicaCount": {
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
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.Volume.GetReplicaCount()
			if got != tt.expectedOutput {
				t.Fatalf("Test: %v Expected: %v but got: %v", name, tt.expectedOutput, got)
			}
		})
	}
}

func TestGetReplicaStatus(t *testing.T) {
	tests := map[string]struct {
		expectedOutput string
		Volume         CASVolume
	}{
		"Fetching ReplicaStatus from openebs.io/replica-status": {
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
			Volume: CASVolume{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			expectedOutput: "",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.Volume.GetReplicaStatus()
			if got != tt.expectedOutput {
				t.Fatalf("Test: %v Expected: %v but got: %v", name, tt.expectedOutput, got)
			}
		})
	}
}

func TestGetReplicaIP(t *testing.T) {
	tests := map[string]struct {
		expectedOutput string
		Volume         CASVolume
	}{
		"Fetching ReplicaIP from openebs.io/replica-ips": {
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
			Volume: CASVolume{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			expectedOutput: "",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.Volume.GetReplicaIP()
			if got != tt.expectedOutput {
				t.Fatalf("Test: %v Expected: %v but got: %v", name, tt.expectedOutput, got)
			}
		})
	}
}
