package driver

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/csidriver/v1alpha1/utils"
	"github.com/openebs/maya/pkg/iscsi"

	"golang.org/x/net/context"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NodeServer defines the structure of the csi node driver
type NodeServer struct {
	driver *CSIDriver
}

// NewNodeServer returns a new object of type NodeServer
func NewNodeServer(d *CSIDriver) *NodeServer {
	return &NodeServer{
		driver: d,
	}
}

// NodePublishVolume publishes(mounts) the volume at the corresponding node at
// a given path
func (ns *NodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	var (
		err        error
		reVerified bool
	)
	logrus.Infof("NodepublishVolume")

	if req.GetVolumeCapability() == nil {
		return nil, status.Error(codes.InvalidArgument,
			"Volume capability missing in request")
	}
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument,
			"Volume ID missing in request")
	}
	mountPath := req.GetTargetPath()
	volumeID := req.GetVolumeId()
	mountOptions := req.GetVolumeCapability().GetMount().GetMountFlags()

	vol, err := utils.GetVolumeDetails(volumeID, mountPath, req.Readonly, mountOptions)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	//Check if volume is ready to serve IOs,
	//info is fetched from the cstorvolume CR
	if err := utils.WaitForVolumeToBeReady(volumeID); err != nil {
		return nil,
			status.Error(codes.Internal, err.Error())
	}

	if err := utils.WaitForVolumeToBeReachable(vol.Spec.TargetPortal); err != nil {
		return nil,
			status.Error(codes.Internal, err.Error())
	}

	// Check if the volume has already been published
verifyPublish:
	utils.MonitorLock.Lock()
	if _, ok := utils.Volumes[volumeID]; ok {
		for _, info := range utils.Volumes {
			if info.Spec.Volname == volumeID {
				utils.MonitorLock.Unlock()
				return &csi.NodePublishVolumeResponse{}, nil
			}
		}
		utils.MonitorLock.Unlock()
		if !reVerified {
			time.Sleep(13 * time.Second)
			reVerified = true
			goto verifyPublish
		}
		return nil, status.Error(codes.Internal, "Mount under progress")
	}

	if err = utils.DeleteOldCSIVolumeInfoCR(vol); err != nil {
		utils.MonitorLock.Unlock()
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = utils.CreateCSIVolumeInfoCR(vol, ns.driver.config.NodeID, mountPath)
	if err != nil {
		utils.MonitorLock.Unlock()
		return nil, status.Error(codes.Internal, err.Error())
	}
	utils.Volumes[volumeID] = vol
	utils.MonitorLock.Unlock()

	if err = utils.ChmodMountPath(vol.Spec.MountPath); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if _, err = iscsi.AttachAndMountDisk(vol); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &csi.NodePublishVolumeResponse{}, nil
}

// NodeUnpublishVolume unpublishes(unmounts) the volume from the corresponding
// node from the given path
func (ns *NodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	logrus.Infof("NodeUnpublishVolume")

	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument,
			"Volume ID missing in request")
	}
	if len(req.GetTargetPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument,
			"Target path missing in request")
	}
	targetPath := req.GetTargetPath()
	volumeID := req.GetVolumeId()
	vol, ok := utils.Volumes[volumeID]
	if !ok {
		return &csi.NodeUnpublishVolumeResponse{}, nil
	}
	err := utils.DeleteCSIVolumeInfoCR(vol)
	if err != nil {
		return nil, status.Error(codes.Internal,
			err.Error())
	}
	utils.MonitorLock.Lock()
	delete(utils.Volumes, volumeID)
	if err = iscsi.UnmountAndDetachDisk(vol, req.GetTargetPath()); err != nil {
		return nil, status.Error(codes.Internal,
			err.Error())
	}
	utils.MonitorLock.Unlock()
	glog.V(4).Infof("hostpath: volume %s/%s has been unmounted.",
		targetPath, volumeID)

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

// NodeStageVolume mounts the volume on the staging path
func (ns *NodeServer) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	return &csi.NodeStageVolumeResponse{}, nil
}

// NodeUnstageVolume unmounts the volume from the staging path
func (ns *NodeServer) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	return &csi.NodeUnstageVolumeResponse{}, nil
}

// NodeGetInfo returns the info of the corresponding node
func (ns *NodeServer) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	return &csi.NodeGetInfoResponse{
		NodeId:            ns.driver.config.NodeID,
		MaxVolumesPerNode: 1,
	}, nil
}

// NodeExpandVolume resizes the filesystem if required
// if ControllerExpandVolumeResponse returns true in
// node_expansion_required then FileSystemResizePending condition will be added
// to PVC and NodeExpandVolume operation will be queued on kubelet
func (ns *NodeServer) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	return nil, nil
}

// NodeGetCapabilities returns the capabilities supported by the node
func (ns *NodeServer) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	glog.V(5).Infof("Using default NodeGetCapabilities")

	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: []*csi.NodeServiceCapability{
			{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_UNKNOWN,
					},
				},
			},
		},
	}, nil
}

// NodeGetVolumeStats returns the volume capacity statistics
// available for the volume
func (ns *NodeServer) NodeGetVolumeStats(ctx context.Context, in *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
