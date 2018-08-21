package cstor

import (
	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/client/generated/cstor-volume-grpc/v1alpha1"
)

// EngineOps exposes engine operations
type EngineOps interface {
	CreateSnapshot(volName, snapName, targetIP string) (*v1alpha1.VolumeCommand, error)
	DeleteSnapshot(volName, snapName, targetIP string) (*v1alpha1.VolumeCommand, error)
}

type CStorOps struct {
}

func (c *CStorOps) CreateSnapshot(volName, snapName, targetIP string) (*v1alpha1.VolumeCommand, error) {
	glog.Infof("cStor.CreateSnapshot called volName:%s,snapName:%s, ip:%s", volName, snapName, targetIP)
	return createSnapshot(volName, snapName, targetIP)
}

func (c *CStorOps) DeleteSnapshot(volName, snapName, targetIP string) (*v1alpha1.VolumeCommand, error) {
	glog.Infof("cStor.DreateSnapshot called volName:%s,snapName:%s, ip:%s", volName, snapName, targetIP)
	return destroySnapshot(volName, snapName, targetIP)
}
