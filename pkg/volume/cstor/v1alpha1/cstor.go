package v1alpha1

import (
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	volume "github.com/openebs/maya/pkg/client/volume/cstor/v1alpha1"
)

// cstor is used to invoke resize call
type cstor struct {
	IP  string
	Vol *v1alpha1.CASVolume
}

// Cstor return a pointer to cstor
func Cstor() *cstor {
	return &cstor{}
}

// Resize resizes a cstor volume
func (c *cstor) Resize() (*v1alpha1.CASVolume, error) {
	_, err := volume.ResizeVolume(c.IP, c.Vol.Name, c.Vol.Spec.Capacity)
	if err != nil {
		return nil, err
	}

	// we are returning the same struct that we received as input.
	// This would be modified when server replies back with some property of
	// resize volume
	return c.Vol, nil
}
