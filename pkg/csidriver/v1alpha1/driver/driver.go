package driver

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/csidriver/v1alpha1/config"
	"github.com/openebs/maya/pkg/csidriver/v1alpha1/utils"
)

const (
	// DriverName defines the name that is used in Kubernetes and the CSI
	// system for the canonical, official name of this plugin
	DriverName = "openebs-csi.openebs.io"
)

// CSIDriver defines a common data structure for drivers
type CSIDriver struct {
	config *config.Config
	ids    *IdentityServer
	ns     *NodeServer
	cs     *ControllerServer

	cap []*csi.VolumeCapability_AccessMode
}

// GetVolumeCapabilityAccessModes adds the access modes on which the volume can
// be exposed
func GetVolumeCapabilityAccessModes() []*csi.VolumeCapability_AccessMode {
	vc := []csi.VolumeCapability_AccessMode_Mode{
		csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
	}
	var vca []*csi.VolumeCapability_AccessMode
	for _, c := range vc {
		glog.Infof("Enabling volume access mode: %v", c.String())
		vca = append(vca, newVolumeCapabilityAccessMode(c))
	}
	return vca
}

func newVolumeCapabilityAccessMode(mode csi.VolumeCapability_AccessMode_Mode) *csi.VolumeCapability_AccessMode {
	return &csi.VolumeCapability_AccessMode{Mode: mode}
}

// New creates and returns an object of driver instance
// based on the type of driver passed
func New(driverConfig *config.Config) *CSIDriver {
	drvr := &CSIDriver{
		config: driverConfig,
		cap:    GetVolumeCapabilityAccessModes(),
	}
	switch driverConfig.PluginType {
	case "controller":
		drvr.cs = NewControllerServer(drvr)
	case "node":
		utils.FetchAndUpdateVolInfos(driverConfig.NodeID)
		// Start monitor goroutine to monitor the mounted paths,
		// If a path goes down/ becomes read only (in case of RW mount points), this
		// thread will fetch the path and relogin or remount
		go utils.MonitorMounts()
		drvr.ns = NewNodeServer(drvr)
	}
	drvr.ids = NewIdentityServer(drvr)
	return drvr

}

// Run starts the CSI plugin by communication over the given endpoint
func (drvr *CSIDriver) Run() error {
	// Initialize and start listening on grpc server
	s := utils.NewNonBlockingGRPCServer()
	s.Start(drvr.config.Endpoint, drvr.ids, drvr.cs, drvr.ns)
	s.Wait()
	return nil
}
