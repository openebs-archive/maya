package driver

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/golang/glog"
	"github.com/kubernetes-csi/csi-lib-utils/protosanitizer"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	api_core_v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/util/mount"
)

func ParseEndpoint(ep string) (string, string, error) {
	if strings.HasPrefix(strings.ToLower(ep), "unix://") || strings.HasPrefix(strings.ToLower(ep), "tcp://") {
		s := strings.SplitN(ep, "://", 2)
		if s[1] != "" {
			return s[0], s[1], nil
		}
	}
	return "", "", fmt.Errorf("Invalid endpoint: %v", ep)
}

func logGRPC(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	glog.V(3).Infof("GRPC call: %s", info.FullMethod)
	glog.V(5).Infof("GRPC request: %s", protosanitizer.StripSecrets(req))
	resp, err := handler(ctx, req)
	if err != nil {
		glog.Errorf("GRPC error: %v", err)
	} else {
		glog.V(5).Infof("GRPC response: %s", protosanitizer.StripSecrets(resp))
	}
	return resp, err
}

func chmodMountPath(mountPath string) error {
	return os.Chmod(mountPath, 0000)
}
func getVolumeByID(volumeID string) (*v1alpha1.CSIVolumeInfo, error) {
	if Vol, ok := Volumes[volumeID]; ok {
		return Vol, nil
	}
	return nil,
		fmt.Errorf("volume id %s doesn't exit in the volumes list", volumeID)
}

func getVolumeByName(volName string) (*v1alpha1.CSIVolumeInfo, error) {
	for _, Vol := range Volumes {
		if Vol.Name == volName {
			return Vol, nil
		}
	}
	return nil,
		fmt.Errorf("volume name %s does not exit in the volumes list", volName)
}

func getSnapshotByName(name string) (Snapshot, error) {
	for _, snapshot := range VolumeSnapshots {
		if snapshot.Name == name {
			return snapshot, nil
		}
	}
	return Snapshot{},
		fmt.Errorf("snapshot name %s does not exit in the snapshots list", name)
}

func getVolumeDetails(volumeID, mountPath string, readOnly bool, mountOptions []string) (*v1alpha1.CSIVolumeInfo, error) {
	pv, err := fetchPVDetails(volumeID)
	if err != nil {
		return nil, err
	}
	vol := v1alpha1.CSIVolumeInfo{}
	cap := pv.Spec.Capacity[api_core_v1.ResourceName(api_core_v1.ResourceStorage)]
	for _, accessmode := range pv.Spec.AccessModes {
		vol.Spec.AccessModes = append(vol.Spec.AccessModes, string(accessmode))
	}
	vol.Spec.Volname = volumeID
	vol.Spec.Iqn = pv.Spec.CSI.VolumeAttributes["iqn"]
	vol.Spec.Lun = pv.Spec.CSI.VolumeAttributes["lun"]
	vol.Spec.IscsiInterface = pv.Spec.CSI.VolumeAttributes["iscsiInterface"]
	vol.Spec.TargetPortal = pv.Spec.CSI.VolumeAttributes["targetPortal"]
	vol.Spec.FSType = pv.Spec.CSI.FSType
	vol.Spec.Capacity = cap.String()
	vol.Spec.MountPath = mountPath
	vol.Spec.ReadOnly = readOnly
	vol.Spec.MountOptions = mountOptions
	return &vol, nil
}

func removePathFromList(mountpath string) error {
	var (
		indx int
		info *mountPointInfo
	)
	for indx, info = range mountPointInfoList {
		if info.Path == mountpath {
			mountPointInfoList = append(mountPointInfoList[:indx],
				mountPointInfoList[indx+1:]...)
			break
		}
	}
	return nil
}

func unmount(vol *v1alpha1.CSIVolumeInfo, path string) error {
	iscsiInfo := &iscsiDisk{
		VolName: vol.Name,
		Portals: []string{vol.Spec.TargetPortal},
		Iqn:     vol.Spec.Iqn,
		lun:     vol.Spec.Lun,
	}

	diskUnmounter := &iscsiDiskUnmounter{
		iscsiDisk: iscsiInfo,
		mounter:   &mount.SafeFormatAndMount{Interface: mount.New(""), Exec: mount.NewOsExec()},
		exec:      mount.NewOsExec(),
	}
	util := &ISCSIUtil{}
	err := util.DetachDisk(*diskUnmounter, path)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	return nil
}

func mountDisk(vol *v1alpha1.CSIVolumeInfo) (string, error) {
	if len(vol.Spec.MountPath) == 0 {
		return "", status.Error(codes.InvalidArgument, "Target path missing in request")
	}
	iscsiInfo, err := getISCSIInfo(vol)
	if err != nil {
		return "", status.Error(codes.Internal, err.Error())
	}
	diskMounter := getISCSIDiskMounter(iscsiInfo, vol)

	util := &ISCSIUtil{}
	devicePath, err := util.AttachDisk(*diskMounter)
	if err != nil {
		return "", status.Error(codes.Internal, err.Error())
	}
	return devicePath, err
}

func listContains(mountPath string, list []mount.MountPoint) (*mount.MountPoint, bool) {
	for _, info := range list {
		if info.Path == mountPath {
			return &info, true
		}
	}
	return nil, false
}

func monitor() {
	mounter := mount.New("")
	options := []string{"rw"}
	//options = append(options, "remount")
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			monitorLock.RLock()
			list, _ := mounter.List()
			for _, vol := range Volumes {
				path := vol.Spec.MountPath
				mountPoint, exists := listContains(path, list)
				if exists {
					for _, opts := range mountPoint.Opts {
						if opts == "ro" {
							logrus.Infof("MountPoint:%v IN RO MODE", mountPoint.Path)
							mounter.Unmount(path)
							err := mounter.Mount(mountPoint.Device, mountPoint.Path, "", options)
							logrus.Infof("ERR: %v", err)
							break
						} else if opts == "rw" {
							break
						}
					}
				} else {
					mountDisk(vol)
				}
			}
			monitorLock.RUnlock()
		}
	}
}

func generateCSIVolInfoFromCASVolume(vol *v1alpha1.CASVolume) *v1alpha1.CSIVolumeInfo {
	csivol := &v1alpha1.CSIVolumeInfo{}
	csivol.Spec.Volname = vol.Name
	csivol.Spec.Iqn = vol.Spec.Iqn
	csivol.Spec.Capacity = vol.Spec.Capacity
	csivol.Spec.TargetPortal = vol.Spec.TargetPortal
	csivol.Spec.Lun = strconv.FormatInt(int64(vol.Spec.Lun), 10)

	return csivol
}
