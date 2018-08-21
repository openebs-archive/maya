package cstor

import (
	"github.com/openebs/maya/pkg/client/generated/cstor-volume-grpc/v1alpha1"
)

// EngineOps exposes engine operations
type EngineOps interface {
	CreateSnapshot(volName, snapName, targetIP string) (*v1alpha1.VolumeSnapResponse, error)
	DeleteSnapshot(volName, snapName, targetIP string) (*v1alpha1.VolumeSnapResponse, error)
}

type CStorOps struct {
}

func (c *CStorOps) CreateSnapshot(volName, snapName, targetIP string) (*v1alpha1.VolumeSnapResponse, error) {
	return createSnapshot(volName, snapName, targetIP)
}

func (c *CStorOps) DeleteSnapshot(volName, snapName, targetIP string) (*v1alpha1.VolumeSnapResponse, error) {
	return destroySnapshot(volName, snapName, targetIP)
}
