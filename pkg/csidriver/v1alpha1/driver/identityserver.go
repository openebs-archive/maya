package driver

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/version"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// IdentityServer defines the structure for Identity Plugin
type IdentityServer struct {
	driver *CSIDriver
}

// NewIdentityServer created and returns an instance of IdentityServer object
func NewIdentityServer(d *CSIDriver) *IdentityServer {
	return &IdentityServer{
		driver: d,
	}
}

// GetPluginInfo returns the version and name of the plugin
func (ids *IdentityServer) GetPluginInfo(ctx context.Context, req *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	glog.V(5).Infof("Using default GetPluginInfo")

	if ids.driver.config.DriverName == "" {
		return nil, status.Error(codes.Unavailable, "Driver name not configured")
	}

	if ids.driver.config.Version == "" {
		return nil, status.Error(codes.Unavailable, "Driver is missing version")
	}

	return &csi.GetPluginInfoResponse{
		Name:          ids.driver.config.DriverName,
		VendorVersion: version.GetVersion(),
	}, nil
}

// Probe can be used to check whether the plugin is running or not
func (ids *IdentityServer) Probe(ctx context.Context, req *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	return &csi.ProbeResponse{}, nil
}

// GetPluginCapabilities returns the capabilities of the plugin
// Currently it reports whether the plugin has the ability of serving the
// Controller interface. Controller interface methods are called depending
// on whether this method returns the capability or not
func (ids *IdentityServer) GetPluginCapabilities(ctx context.Context, req *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	glog.V(5).Infof("Using default capabilities")
	return &csi.GetPluginCapabilitiesResponse{
		Capabilities: []*csi.PluginCapability{
			{
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						Type: csi.PluginCapability_Service_CONTROLLER_SERVICE,
					},
				},
			},
			{
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						Type: csi.PluginCapability_Service_VOLUME_ACCESSIBILITY_CONSTRAINTS,
					},
				},
			},
		},
	}, nil
}
