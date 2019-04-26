package driver

import (
	"fmt"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/openebs/maya/pkg/csidriver/v1alpha1/utils"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

// ControllerServer defines the data structure of the controller driver
type ControllerServer struct {
	driver *CSIDriver
	cscap  []*csi.ControllerServiceCapability
}

var (
	// supportedAccessMode specifies the AccessModes that can
	// be supported by the volume
	supportedAccessMode = &csi.VolumeCapability_AccessMode{
		// Volume can only be published once as read/write on a single node, at
		// any given time.
		Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
	}
)

// NewControllerServer returns a new instance of controller server
func NewControllerServer(d *CSIDriver) *ControllerServer {
	return &ControllerServer{
		driver: d,
		cscap:  AddControllerServiceCaps(),
	}
}

// AddControllerServiceCaps adds controller service capabilities of the driver
func AddControllerServiceCaps() []*csi.ControllerServiceCapability {
	newCap := func(cap csi.ControllerServiceCapability_RPC_Type) *csi.ControllerServiceCapability {
		return &csi.ControllerServiceCapability{
			Type: &csi.ControllerServiceCapability_Rpc{
				Rpc: &csi.ControllerServiceCapability_RPC{
					Type: cap,
				},
			},
		}
	}

	var caps []*csi.ControllerServiceCapability
	for _, cap := range []csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT,
		csi.ControllerServiceCapability_RPC_LIST_SNAPSHOTS,
		csi.ControllerServiceCapability_RPC_LIST_VOLUMES,
	} {
		caps = append(caps, newCap(cap))
	}
	return caps

}

// CreateVolume dynamically provisions a volume on demand
func (cs *ControllerServer) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	logrus.Infof("Create Volume")
	if err := cs.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		logrus.Infof("invalid create volume req: %v", req)
		return nil, err
	}
	volName := req.GetName()

	if len(volName) == 0 {
		return nil, status.Error(codes.InvalidArgument,
			"Name missing in request")
	}
	caps := req.GetVolumeCapabilities()
	if caps == nil {
		return nil, status.Error(codes.InvalidArgument,
			"Volume Capabilities missing in request")
	}
	for _, cap := range caps {
		if cap.GetBlock() != nil {
			return nil, status.Error(codes.Unimplemented,
				"Block Volume not supported")
		}
	}

	// Verify if the volume has already been created
	if exVol, err := utils.GetVolumeByName(volName); err == nil {
		capacity, _ := strconv.ParseInt(exVol.Spec.Volume.Capacity, 10, 64)
		if capacity >= int64(req.GetCapacityRange().GetRequiredBytes()) {
			return &csi.CreateVolumeResponse{
				Volume: &csi.Volume{
					VolumeId:      exVol.Name,
					CapacityBytes: capacity,
					VolumeContext: req.GetParameters(),
				},
			}, nil
		}
		return nil,
			status.Error(codes.AlreadyExists,
				fmt.Sprintf("Volume with the same name: %s but with different size already exist",
					req.GetName()))
	}

	// Send volume creation request to m-apiserver
	vol, err := utils.ProvisionVolume(req)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	// Create a csi vol object from info returned by m-apiserver
	csivol := utils.GenerateCSIVolFromCASVolume(vol)

	volumeID := volName
	// Keep a local copy of the volume info to catch duplicate requests
	utils.Volumes[volumeID] = csivol
	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      volumeID,
			CapacityBytes: req.GetCapacityRange().GetRequiredBytes(),
			// VolumeContext is essential for publishing volumes at nodes,
			// for iscsi login, this will be stored in PV CR
			VolumeContext: map[string]string{
				"volname":        volName,
				"iqn":            vol.Spec.Iqn,
				"targetPortal":   vol.Spec.TargetPortal,
				"lun":            "0",
				"iscsiInterface": "default",
				"portals":        vol.Spec.TargetPortal,
			},
		},
	}, nil
}

// DeleteVolume deletes the specified volume
func (cs *ControllerServer) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	logrus.Infof("Delete Volume")

	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument,
			"Volume ID missing in request")
	}

	if err := cs.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		logrus.Infof("invalid delete volume req: %v", req)
		return nil, err
	}
	volumeID := req.VolumeId

	// This call is made just to fetch pvc namespace
	pv, err := utils.FetchPVDetails(volumeID)
	if err != nil {
		logrus.Infof("fetch Volume Failed, volID:%v %v", volumeID, err)
		return nil, err
	}
	pvcNamespace := pv.Spec.ClaimRef.Namespace

	//Send delete request to m-apiserver
	if err := utils.DeleteVolume(volumeID, pvcNamespace); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Remove entry from the volume list maintained
	delete(utils.Volumes, volumeID)
	return &csi.DeleteVolumeResponse{}, nil
}

// ValidateVolumeCapabilities validates the capabilities required to create a
// new volume
func (cs *ControllerServer) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	return cs.ValidateVolumeCapabilities(ctx, req)
}

// ControllerGetCapabilities fetches the controller capabilities
func (cs *ControllerServer) ControllerGetCapabilities(ctx context.Context,
	req *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {

	resp := &csi.ControllerGetCapabilitiesResponse{
		Capabilities: cs.cscap,
	}

	return resp, nil
}

// ControllerExpandVolume can be used to resize the previously provisioned
// volume
func (cs *ControllerServer) ControllerExpandVolume(ctx context.Context,
	req *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	return nil, nil
}

// validateCapabilities validates if the corresponding capability is supported
// by the driver
func validateCapabilities(caps []*csi.VolumeCapability) bool {
	vcaps := []*csi.VolumeCapability_AccessMode{supportedAccessMode}

	hasSupport := func(mode csi.VolumeCapability_AccessMode_Mode) bool {
		for _, m := range vcaps {
			if mode == m.Mode {
				return true
			}
		}
		return false
	}

	supported := false
	for _, cap := range caps {
		if hasSupport(cap.AccessMode.Mode) {
			supported = true
		} else {
			supported = false
		}
	}

	return supported
}

// ValidateControllerServiceRequest validates if the requested service is
// supported by the driver
func (cs *ControllerServer) ValidateControllerServiceRequest(c csi.ControllerServiceCapability_RPC_Type) error {
	if c == csi.ControllerServiceCapability_RPC_UNKNOWN {
		return nil
	}

	for _, cap := range cs.cscap {
		if c == cap.GetRpc().GetType() {
			return nil
		}
	}
	return status.Error(codes.InvalidArgument, fmt.Sprintf("%s", c))
}

// CreateSnapshot can be used to create a snapsnhot for a particular volumeID
// provided
func (cs *ControllerServer) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// DeleteSnapshot can be used to delete a particular snapshot of a specified
// volume
func (cs *ControllerServer) DeleteSnapshot(ctx context.Context, req *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// ListSnapshots lists all the snapshots for the volume specified via VolumeID
func (cs *ControllerServer) ListSnapshots(ctx context.Context, req *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// ControllerUnpublishVolume can be used to remove a previously attached volume
// from the specified node
func (cs *ControllerServer) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// ControllerPublishVolume can be used to attach the volume at the specified
// node
func (cs *ControllerServer) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// GetCapacity return the capacity of the the storage pool
func (cs *ControllerServer) GetCapacity(ctx context.Context, req *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// ListVolumes lists the info of all the OpenEBS volumes created via m-apiserver
func (cs *ControllerServer) ListVolumes(ctx context.Context, req *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
