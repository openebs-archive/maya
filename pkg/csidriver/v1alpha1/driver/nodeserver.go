// Copyright © 2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package driver

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/csidriver/v1alpha1/utils"
	iscsi "github.com/openebs/maya/pkg/iscsi/v1alpha1"

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
		devicePath string
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

	// A temporary TCP connection is made to the volume to check if its
	// reachable
	if err := utils.WaitForVolumeToBeReachable(vol.Spec.ISCSI.TargetPortal); err != nil {
		return nil,
			status.Error(codes.Internal, err.Error())
	}

verifyPublish:
	utils.VolumesListLock.Lock()
	// Check if the volume has already been published(mounted) or if the mount
	// is in progress
	if _, ok := utils.Volumes[volumeID]; ok {
		// The volume appears to be present in the inmomory list of volumes
		// which implies that either the mount operation is complete
		// or under progress.
		// Lets verify if the mount is already completed
		for _, info := range utils.Volumes {
			if info.Spec.Volume.DevicePath != "" {
				// Once the devicePath is set implies the volume mount has been
				// completed, a success response can be sent back
				utils.VolumesListLock.Unlock()
				return &csi.NodePublishVolumeResponse{}, nil
			}
		}
		// The mount appears to be under progress lets wait for 13 seconds and
		// reverify. 13s was decided based on the kubernetes timeout values
		// which is 15s. Lets reply to kubernetes before it reattempts a
		// duplicate request
		utils.VolumesListLock.Unlock()
		if !reVerified {
			time.Sleep(13 * time.Second)
			reVerified = true
			goto verifyPublish
		}
		// It appears that the mount will still take some more time,
		// lets convey the same to kubernetes. The message responded will be
		// added to the app description which has requested this volume

		return nil, status.Error(codes.Internal, "Mount under progress")
	}

	// This helps in cases when the node on which the volume was originally
	// mounted is down. When that node is down, kubelet would not have been able
	// to trigger an unpublish event on that node due to which when it comes up
	// it starts remounting that volume. If the node's CSIVolume CR is marked
	// for deletion that node will not reattempt to mount this volume again.
	//	if err = utils.DeleteOldCSIVolumeCR(vol); err != nil {
	//		utils.VolumesListLock.Unlock()
	//		return nil, status.Error(codes.Internal, err.Error())
	//	}

	// This CR creation will help iSCSI target(istgt) identify
	// the current owner node of the volume and accordingly the target will
	// allow only that node to login to the volume
	//	err = utils.CreateCSIVolumeCR(vol, ns.driver.config.NodeID, mountPath)
	//	if err != nil {
	//		utils.VolumesListLock.Unlock()
	//		return nil, status.Error(codes.Internal, err.Error())
	//	}
	utils.Volumes[volumeID] = vol
	utils.VolumesListLock.Unlock()

	// Permission is changed for the local directory when the volume is not
	// mounted on it. This helps to resolve cases when the CSI driver
	// Unmounts the volume to remount again in correct mode,
	// the app starts writing directly in the local directory.
	// As soon as the volume is mounted the permissions of this directory are
	// automatically changed to allow Reads and writes.
	// And as soon as it is unmounted permissions change
	// back to what we are setting over here.
	if err = utils.ChmodMountPath(vol.Spec.Volume.MountPath); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	// Login to the volume and attempt mount operation on the requested path
	if devicePath, err = iscsi.AttachAndMountDisk(vol); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Setting the devicePath in the volume spec is an indication that the mount
	// operation for the volume has been completed for the first time. This
	// helps in 2 ways:
	// 1) Duplicate nodePublish requests from kubernetes are responded with
	//    success response if this path is set
	// 2) The volumeMonitoring thread doesn't attemp remount unless this path is
	//    set
	utils.VolumesListLock.Lock()
	vol.Spec.Volume.DevicePath = devicePath
	utils.VolumesListLock.Unlock()

	return &csi.NodePublishVolumeResponse{}, nil
}

// NodeUnpublishVolume unpublishes(unmounts) the volume from the corresponding
// node from the given path
func (ns *NodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	var (
		err error
	)
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
	utils.VolumesListLock.Lock()
	delete(utils.Volumes, volumeID)
	utils.VolumesListLock.Unlock()
	// if node driver restarts before this step Kubelet will trigger the
	// NodeUnpublish command again so there is no need to worry that when this
	// driver restarts it will pick up the CSIVolume CR and start monitoring
	// mount point again.
	// If the node is down for some time, other node driver will first delete
	// this node's CSIVolume CR and then only will start its mount process.
	// If there is a case that this node comes up and CSIVolume CR is picked and
	// this node starts monitoring the mount point while the other node is also
	// trying to mount which appears to be a race condition but is not since
	// first of  all this CR will be marked for deletion when the other node
	// starts mounting. But lets say this node started monitoring and
	// immediately other node deleted this node's CR, in that case iSCSI
	// target(istgt) will pick up the new one and allow only that node to login,
	// so all the cases are handled
	if err = iscsi.UnmountAndDetachDisk(vol, req.GetTargetPath()); err != nil {
		return nil, status.Error(codes.Internal,
			err.Error())
	}
	// It is safe to delete the CSIVolume CR now since the volume has already
	// been unmounted and logged out
	//	err = utils.DeleteCSIVolumeCR(vol)
	//	if err != nil {
	//		return nil, status.Error(codes.Internal,
	//			err.Error())
	//	}
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
