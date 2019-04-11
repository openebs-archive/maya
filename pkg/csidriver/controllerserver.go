package driver

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

const (
	deviceID           = "deviceID"
	provisionRoot      = "/tmp/"
	snapshotRoot       = "/tmp/"
	maxStorageCapacity = tib
	timeout            = 60 * time.Second
)

type controllerServer struct {
	driver *CSIDriver
	cscap  []*csi.ControllerServiceCapability
}

var (
	supportedAccessMode = &csi.VolumeCapability_AccessMode{
		Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
	}
)

func NewControllerServer(d *CSIDriver) *controllerServer {
	return &controllerServer{
		driver: d,
		cscap:  AddControllerServiceCaps(),
	}
}

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

func (cs *controllerServer) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	logrus.Infof("Create Volume")
	if err := cs.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		logrus.Infof("invalid create volume req: %v", req)
		return nil, err
	}

	if len(req.GetName()) == 0 {
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
	if exVol, err := getVolumeByName(req.GetName()); err == nil {
		capacity, _ := strconv.ParseInt(exVol.Spec.Capacity, 10, 64)
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
				fmt.Sprintf("Volume with the same name: %s but with different size already exist", req.GetName()))
	}
	vol, _ := provisionVolume(req)
	csivol := generateCSIVolInfoFromCASVolume(vol)

	volumeID := req.GetName()
	Volumes[volumeID] = csivol
	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      volumeID,
			CapacityBytes: req.GetCapacityRange().GetRequiredBytes(),
			VolumeContext: map[string]string{
				"volname":        req.GetName(),
				"iqn":            vol.Spec.Iqn,
				"targetPortal":   vol.Spec.TargetPortal,
				"lun":            "0",
				"iscsiInterface": "default",
				"portals":        vol.Spec.TargetPortal,
			},
		},
	}, nil
}

func (cs *controllerServer) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	logrus.Infof("Delete Volume")

	// Check arguments
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument,
			"Volume ID missing in request")
	}

	if err := cs.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		logrus.Infof("invalid delete volume req: %v", req)
		return nil, err
	}
	volumeID := req.VolumeId
	logrus.Infof("deleting volume %s", volumeID)
	pv, err := fetchPVDetails(volumeID)
	if err != nil {
		return nil, err
	}
	pvcNamespace := pv.Spec.ClaimRef.Namespace
	DeleteVolume(volumeID, pvcNamespace)
	delete(Volumes, volumeID)
	return &csi.DeleteVolumeResponse{}, nil
}

func (cs *controllerServer) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	return cs.ValidateVolumeCapabilities(ctx, req)
}

func (cs *controllerServer) ControllerGetCapabilities(ctx context.Context,
	req *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {

	resp := &csi.ControllerGetCapabilitiesResponse{
		Capabilities: cs.cscap,
	}

	return resp, nil
}

func (cs *controllerServer) ControllerExpandVolume(ctx context.Context,
	req *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	return nil, nil
}

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
func (cs *controllerServer) ValidateControllerServiceRequest(c csi.ControllerServiceCapability_RPC_Type) error {
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

func (cs *controllerServer) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *controllerServer) DeleteSnapshot(ctx context.Context, req *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *controllerServer) ListSnapshots(ctx context.Context, req *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *controllerServer) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *controllerServer) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *controllerServer) GetCapacity(ctx context.Context, req *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *controllerServer) ListVolumes(ctx context.Context, req *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
