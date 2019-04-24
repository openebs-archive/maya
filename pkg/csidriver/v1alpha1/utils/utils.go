package utils

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/golang/glog"
	"github.com/kubernetes-csi/csi-lib-utils/protosanitizer"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	m_k8s_client "github.com/openebs/maya/pkg/client/k8s"
	iscsi "github.com/openebs/maya/pkg/iscsi/v1alpha1"
	"google.golang.org/grpc"
	api_core_v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/util/mount"
)

var (
	// MAPIServerEndpoint is the address to connect to m-apiserver to send
	// volume related requests
	MAPIServerEndpoint string
	// OpenEBSNamespace is where all the OpenEBS related pods are running and
	// CSIVolInfo as to be placed
	OpenEBSNamespace string

	// Volumes contains the list of volumes created in case of controller plugin
	// and list of volumes attached to this node in node plugin
	Volumes map[string]*v1alpha1.CSIVolumeInfo

	// MonitorLock is required to protect the above Volumes list
	MonitorLock sync.RWMutex
)

const (
	timeout = 60 * time.Second
)

func init() {

	OpenEBSNamespace = os.Getenv("OPENEBS_NAMESPACE")
	if OpenEBSNamespace == "" {
		logrus.Fatalf("OPENEBS_NAMESPACE environment variable not set")
	}

	MAPIServiceName := os.Getenv("OPENEBS_MAPI_SVC")
	if MAPIServiceName == "" {
		logrus.Fatalf("OPENEBS_MAPI_SVC environment variable not set")
	}

	kc, err := m_k8s_client.NewK8sClient(OpenEBSNamespace)
	if err != nil {
		logrus.Fatalf(err.Error())
	}
	svc, err := kc.GetService(MAPIServiceName, metav1.GetOptions{})
	if err != nil {
		logrus.Fatalf(err.Error())
	}

	svcIP := svc.Spec.ClusterIP
	svcPort := strconv.FormatInt(int64(svc.Spec.Ports[0].Port), 10)
	MAPIServerEndpoint = "http://" + svcIP + ":" + svcPort

	Volumes = map[string]*v1alpha1.CSIVolumeInfo{}

}

func parseEndpoint(ep string) (string, string, error) {
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

// ChmodMountPath removes all permission from the folder if volume is not
// mounted on it
func ChmodMountPath(mountPath string) error {
	return os.Chmod(mountPath, 0000)
}

// WaitForVolumeToBeReachable keeps the mounts on hold until the volume is
// reachable
func WaitForVolumeToBeReachable(targetPortal string) error {
	var (
		retries int
		err     error
		conn    net.Conn
	)

	for {
		if conn, err = net.Dial("tcp", targetPortal); err == nil {
			conn.Close()
			logrus.Infof("Volume is reachable to create connections")
			return nil
		}
		time.Sleep(2 * time.Second)
		retries++
		if retries >= 6 {
			return fmt.Errorf("iSCSI Target not reachable, TargetPortal %v, err:%v",
				targetPortal, err)
		}
	}

}

// WaitForVolumeToBeReady retrieves the volume info from cstorVolume CR and
// waits until consistency factor is met for connected replicas
func WaitForVolumeToBeReady(volumeID string) error {
	var retries int
checkVolumeStatus:
	volStatus, err := getVolStatus(volumeID)
	if err != nil {
		return err
	} else if volStatus == "Healthy" || volStatus == "Degraded" {
		logrus.Infof("Volume is ready to accept IOs")
	} else if retries >= 6 {
		return fmt.Errorf("Volume is not ready: Replicas yet to connect to controller")
	} else {
		time.Sleep(2 * time.Second)
		retries++
		goto checkVolumeStatus
	}
	return nil
}

// GetVolumeByName fetches the volume from Volumes list based on th input name
func GetVolumeByName(volName string) (*v1alpha1.CSIVolumeInfo, error) {
	for _, Vol := range Volumes {
		if Vol.Name == volName {
			return Vol, nil
		}
	}
	return nil,
		fmt.Errorf("volume name %s does not exit in the volumes list", volName)
}

// GetVolumeDetails returns a new instance of csiVolumeInfo filled with the
// VolumeAttributes fetched from the corresponding PV and some additional info
// required for remounting
func GetVolumeDetails(volumeID, mountPath string, readOnly bool, mountOptions []string) (*v1alpha1.CSIVolumeInfo, error) {
	pv, err := FetchPVDetails(volumeID)
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

func listContains(mountPath string, list []mount.MountPoint) (*mount.MountPoint, bool) {
	for _, info := range list {
		if info.Path == mountPath {
			mntInfo := info
			return &mntInfo, true
		}
	}
	return nil, false
}

// MonitorMounts makes sure that all the volumes present in the inmemory list
// with the driver are mounted with the original mount options
func MonitorMounts() {
	mounter := mount.New("")
	//options = append(options, "remount")
	ticker := time.NewTicker(5 * time.Second)
	volMonitorMap := make(map[string]bool)
	for {
		select {
		case <-ticker.C:
			MonitorLock.RLock()
			// Get list of mount paths present with the node
			list, _ := mounter.List()
			for _, vol := range Volumes {
				_, volMonitor := volMonitorMap["vol.Spec.Volname"]
				if volMonitor == true {
					continue
				}
				path := vol.Spec.MountPath
				mountPoint, exists := listContains(path, list)
				go func() {
					MonitorLock.Lock()
					volMonitorMap[vol.Spec.Volname] = true
					MonitorLock.Unlock()
					verifyAndRemount(exists, vol, mountPoint, path)
					MonitorLock.Lock()
					delete(volMonitorMap, vol.Spec.Volname)
					MonitorLock.Unlock()

				}()
			}
			MonitorLock.RUnlock()
		}
	}
}

// WaitForVolumeReadyAndReachable waits until the volume is ready to accept IOs
// and is reachable
func WaitForVolumeReadyAndReachable(vol *v1alpha1.CSIVolumeInfo) {
	for {
		if err := WaitForVolumeToBeReady(vol.Spec.Volname); err == nil {
			logrus.Info(err)
			break
		}
		if err := WaitForVolumeToBeReachable(vol.Spec.TargetPortal); err == nil {
			logrus.Info(err)
			break
		}
	}
}

func verifyAndRemount(exists bool, vol *v1alpha1.CSIVolumeInfo, mountPoint *mount.MountPoint, path string) {
	mounter := mount.New("")
	options := []string{"rw"}
	if exists {
		for _, opts := range mountPoint.Opts {
			if opts == "ro" {
				logrus.Infof("MountPoint:%v IN RO MODE", mountPoint.Path)
				mounter.Unmount(path)
				WaitForVolumeReadyAndReachable(vol)
				err := mounter.Mount(mountPoint.Device,
					mountPoint.Path, "", options)
				logrus.Infof("ERR: %v", err)
				break
			} else if opts == "rw" {
				break
			}
		}
	} else {
		WaitForVolumeReadyAndReachable(vol)
		iscsi.AttachAndMountDisk(vol)
	}
}

// GenerateCSIVolInfoFromCASVolume returns an instance of CSIVolInfo
func GenerateCSIVolInfoFromCASVolume(vol *v1alpha1.CASVolume) *v1alpha1.CSIVolumeInfo {
	csivol := &v1alpha1.CSIVolumeInfo{}
	csivol.Spec.Volname = vol.Name
	csivol.Spec.Iqn = vol.Spec.Iqn
	csivol.Spec.Capacity = vol.Spec.Capacity
	csivol.Spec.TargetPortal = vol.Spec.TargetPortal
	csivol.Spec.Lun = strconv.FormatInt(int64(vol.Spec.Lun), 10)

	return csivol
}
