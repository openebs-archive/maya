package driver

import (
	"sync"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
)

const (
	// DriverName defines the name that is used in Kubernetes and the CSI
	// system for the canonical, official name of this plugin
	DriverName = "dobs.csi.openebs.com"
)

const (
	kib    int64 = 1024
	mib    int64 = kib * 1024
	gib    int64 = mib * 1024
	gib100 int64 = gib * 100
	tib    int64 = gib * 1024
	tib100 int64 = tib * 100
)

// CSIDriver defines a common data structure for drivers
type CSIDriver struct {
	config *Config
	ids    *IdentityServer
	ns     *NodeServer
	cs     *ControllerServer

	cap []*csi.VolumeCapability_AccessMode
}

// Volumes contains the list of volumes created in case of controller plugin
// and list of volumes attached to this node in node plugin
var Volumes map[string]*v1alpha1.CSIVolumeInfo

// monitorLock is required to protect the above Volumes list
var monitorLock sync.RWMutex

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

func init() {
	Volumes = map[string]*v1alpha1.CSIVolumeInfo{}
}

// New creates and returns an object of driver instance
// based on the type of driver passed
func New(config *Config) *CSIDriver {
	drvr := &CSIDriver{
		config: config,
		cap:    GetVolumeCapabilityAccessModes(),
	}
	switch config.PluginType {
	case "controller":
		drvr.cs = NewControllerServer(drvr)
	case "node":
		fetchAndUpdateVolInfos(config.NodeID)
		drvr.ns = NewNodeServer(drvr)
	}
	drvr.ids = NewIdentityServer(drvr)
	return drvr

}

// Run starts the CSI plugin by communication over the given endpoint
func (drvr *CSIDriver) Run() error {
	// Start monitor goroutine to monitor the mounted paths,
	// If a path goes down/ becomes read only (in case of RW mount points), this
	// thread will fetch the path and relogin or remount
	go monitor()
	// Initialize and start listening on grpc server
	s := NewNonBlockingGRPCServer()
	s.Start(drvr.config.Endpoint, drvr.ids, drvr.cs, drvr.ns)
	s.Wait()
	return nil
}
