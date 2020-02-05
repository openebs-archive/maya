package executor

import (
	upgrader "github.com/openebs/maya/pkg/upgrade/upgrader_new"
)

func (u *Upgrade) registerUpgrade(kind string, obj UpgradeOptions) *Upgrade {
	u.UpgradeMap[kind] = obj
	return u
}

// RegisterAll ...
func (u *Upgrade) RegisterAll() *Upgrade {
	u.registerUpgrade("cstorpoolinstance", RegisterCstorPoolInstance)
	u.registerUpgrade("cstorpoolcluster", RegisterCstorPoolCluster)
	// u.registerUpgrade("cstorVolume", RegisterCstorVolume)
	// u.registerUpgrade("jivaVolume", RegisterJivaVolume)
	return u
}

// RegisterCstorPoolInstance ....
func RegisterCstorPoolInstance(r *upgrader.ResourcePatch) upgrader.Upgrader {
	obj := upgrader.NewCSPIPatch(
		upgrader.WithCSPIResorcePatch(r),
	)
	return obj
}

// RegisterCstorPoolCluster ...
func RegisterCstorPoolCluster(r *upgrader.ResourcePatch) upgrader.Upgrader {
	obj := upgrader.NewCSPCPatch(
		upgrader.WithCSPCResorcePatch(r),
	)
	return obj
}
