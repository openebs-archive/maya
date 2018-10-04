package v1alpha1

import (
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	snapshot "github.com/openebs/maya/pkg/client/snapshot/cstor/v1alpha1"
)

// cstor is used to invoke Create call
// TODO: Convert this to implement interface
type cstor struct {
}

// Cstor return a pointer to cstor
// TODO: Cstor should return interface which implements all the current
// methods of cstor
func Cstor() *cstor {
	return &cstor{}
}

// Create creates a snapshot of cstor volume
func (c *cstor) Create(ip string, snap *v1alpha1.CASSnapshot) (*v1alpha1.CASSnapshot, error) {
	_, err := snapshot.CreateSnapshot(ip, snap.Spec.VolumeName, snap.Name)
	// If there is no err that means call was successful
	if err != nil {
		return nil, err
	}
	// we are returning the same struct that we received as input.
	// This would be modified when server replies back with some property of
	// created snapshot
	return snap, nil
}

// Delete deletes a snapshot of cstor volume
func (c *cstor) Delete(ip string, snap *v1alpha1.CASSnapshot) (*v1alpha1.CASSnapshot, error) {
	_, err := snapshot.DestroySnapshot(ip, snap.Spec.VolumeName, snap.Name)
	// If there is no err that means call was successful
	if err != nil {
		return nil, err
	}
	// we are returning the same struct that we received as input.
	// This would be modified when server replies back with some property of
	// created snapshot
	return snap, nil
}
