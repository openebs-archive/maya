package driver

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/golang/glog"

	"golang.org/x/net/context"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type nodeServer struct {
	driver *CSIDriver
}

func NewNodeServer(d *CSIDriver) *nodeServer {
	return &nodeServer{
		driver: d,
	}
}

func (ns *nodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	var (
		err        error
		reVerified bool
		retries    int
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

	vol, err := getVolumeDetails(volumeID, mountPath, req.Readonly, mountOptions)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	//Check if volume is ready to server IOs
checkVolumeStatus:
	volStatus, err := getVolStatus(volumeID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	} else if volStatus == "Healthy" || volStatus == "Degraded" {
		logrus.Infof("Volume is ready to accept IOs")
	} else if retries >= 6 {
		return nil,
			status.Error(codes.Internal, "Volume is not ready: Replicas yet to connect to controller")
	} else {
		time.Sleep(2 * time.Second)
		retries++
		goto checkVolumeStatus
	}

	// Check if the volume has already been published
verifyPublish:
	monitorLock.Lock()
	if _, ok := Volumes[volumeID]; ok {
		for _, info := range mountPointInfoList {
			if info.VolumeID == volumeID {
				monitorLock.Unlock()
				return &csi.NodePublishVolumeResponse{}, nil
			}
		}
		monitorLock.Unlock()
		if !reVerified {
			time.Sleep(13 * time.Second)
			reVerified = true
			goto verifyPublish
		}
		return nil, status.Error(codes.Internal, "Mount under progress")
	}
	deleteOldCSIVolumeInfoCR(vol)
	err = createCSIVolumeInfoCR(vol, ns.driver.config.NodeID, mountPath)
	if err != nil {
		monitorLock.Unlock()
		return nil, status.Error(codes.Internal, err.Error())
	}
	Volumes[volumeID] = vol
	monitorLock.Unlock()

	chmodMountPath(vol.Spec.MountPath)
	devicePath, _ := mountDisk(vol)

	mountPointInfo := &mountPointInfo{
		VolumeID:   volumeID,
		Path:       devicePath,
		MountPoint: mountPath,
	}

	monitorLock.Lock()
	mountPointInfoList = append(mountPointInfoList, mountPointInfo)
	monitorLock.Unlock()

	return &csi.NodePublishVolumeResponse{}, nil
}

func (ns *nodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
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
	vol, ok := Volumes[volumeID]
	if !ok {
		return &csi.NodeUnpublishVolumeResponse{}, nil
	}
	deleteCSIVolumeInfoCR(vol)
	monitorLock.Lock()
	removePathFromList(req.GetTargetPath())
	unmount(vol, targetPath)
	monitorLock.Unlock()
	glog.V(4).Infof("hostpath: volume %s/%s has been unmounted.",
		targetPath, volumeID)

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (ns *nodeServer) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	return &csi.NodeStageVolumeResponse{}, nil
}

func (ns *nodeServer) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	return &csi.NodeUnstageVolumeResponse{}, nil
}

func (ns *nodeServer) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	return &csi.NodeGetInfoResponse{
		NodeId:            ns.driver.config.NodeID,
		MaxVolumesPerNode: 1,
	}, nil
}

func (ns *nodeServer) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	return nil, nil
}

func (ns *nodeServer) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
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
func (ns *nodeServer) NodeGetVolumeStats(ctx context.Context, in *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
