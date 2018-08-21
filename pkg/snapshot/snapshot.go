package snapshot

import (
	//	"fmt"

	//	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	m_k8s_client "github.com/openebs/maya/pkg/client/k8s"
)

type SnapshotOps interface {
	GetVolumeType(volName string) (volType string, err error)
}

// SnapshotOperation contains the options with respect to
// snapshot related operations
type SnapshotOperation struct {
	// k8sClient will make K8s API calls
	k8sClient *m_k8s_client.K8sClient
}

// NewSnapshotOperation returns an object of SnapshotOperation
func NewSnapshotOperation(namespace string) (SnapshotOps, error) {
	kc, err := m_k8s_client.NewK8sClient(namespace)
	if err != nil {
		return nil, err
	}
	return &SnapshotOperation{k8sClient: kc}, nil
}

// GetVolumeType returns the volume type i.e. jiva or cstor
// it will return in an error if the PV is not found or the
// volume type label is missing
func (sops *SnapshotOperation) GetVolumeType(volName string) (volType string, err error) {
	/*pv, err := sops.k8sClient.GetPV(volName, mach_apis_meta_v1.GetOptions{})
	if err != nil {
		return "", err
	}
	volType = pv.Annotations[string(v1alpha1.VolumeTypeKey)]

	if volType == "" {
		return "", fmt.Errorf("missing/empty annotation key '%s' for volume '%s'", v1alpha1.VolumeTypeKey, volName)
	}*/
	return
}
