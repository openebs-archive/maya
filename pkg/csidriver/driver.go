package driver

import (
	"sync"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	"github.com/golang/protobuf/ptypes/timestamp"
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

type mountPointInfo struct {
	VolumeID   string
	Path       string
	MountPoint string
}

type CSIDriver struct {
	config *Config
	ids    *identityServer
	ns     *nodeServer
	cs     *controllerServer

	cap []*csi.VolumeCapability_AccessMode
}

type Snapshot struct {
	Name         string              `json:"name"`
	Id           string              `json:"id"`
	VolID        string              `json:"volID"`
	Path         string              `json:"path"`
	CreationTime timestamp.Timestamp `json:"creationTime"`
	SizeBytes    int64               `json:"sizeBytes"`
	ReadyToUse   bool                `json:"readyToUse"`
}

var Volumes map[string]*v1alpha1.CSIVolumeInfo
var VolumeSnapshots map[string]Snapshot
var monitorLock sync.RWMutex
var mountPointInfoList []*mountPointInfo

func GetVolumeCapabilityAccessModes() []*csi.VolumeCapability_AccessMode {
	vc := []csi.VolumeCapability_AccessMode_Mode{
		csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
	}
	var vca []*csi.VolumeCapability_AccessMode
	for _, c := range vc {
		glog.Infof("Enabling volume access mode: %v", c.String())
		vca = append(vca, NewVolumeCapabilityAccessMode(c))
	}
	return vca
}

func NewVolumeCapabilityAccessMode(mode csi.VolumeCapability_AccessMode_Mode) *csi.VolumeCapability_AccessMode {
	return &csi.VolumeCapability_AccessMode{Mode: mode}
}

func init() {
	Volumes = map[string]*v1alpha1.CSIVolumeInfo{}
	VolumeSnapshots = map[string]Snapshot{}
	mountPointInfoList = make([]*mountPointInfo, 0, 5)
}

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
	go monitor()
	s := NewNonBlockingGRPCServer()
	s.Start(drvr.config.Endpoint, drvr.ids, drvr.cs, drvr.ns)
	s.Wait()
	return nil
}
