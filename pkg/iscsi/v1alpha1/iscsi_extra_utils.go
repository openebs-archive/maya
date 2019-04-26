package iscsi

import (
	"fmt"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"k8s.io/kubernetes/pkg/util/mount"
	"k8s.io/kubernetes/pkg/volume/util"
)

func getISCSIInfo(vol *v1alpha1.CSIVolume) (*iscsiDisk, error) {
	portal := portalMounter(vol.Spec.ISCSI.TargetPortal)
	var portals []string
	portals = append(portals, portal)

	chapDiscovery := false

	chapSession := false

	return &iscsiDisk{
		VolName:       vol.Spec.Volume.Volname,
		Portals:       portals,
		Iqn:           vol.Spec.ISCSI.Iqn,
		lun:           vol.Spec.ISCSI.Lun,
		Iface:         vol.Spec.ISCSI.IscsiInterface,
		chapDiscovery: chapDiscovery,
		chapSession:   chapSession}, nil
}

func getISCSIInfoFromPV(req *csi.NodePublishVolumeRequest) (*iscsiDisk, error) {
	volName := req.GetVolumeId()
	tp := req.GetVolumeContext()["targetPortal"]
	iqn := req.GetVolumeContext()["iqn"]
	lun := req.GetVolumeContext()["lun"]
	if tp == "" || iqn == "" || lun == "" {
		return nil, fmt.Errorf("iSCSI target information is missing")
	}

	//portalList := req.GetVolumeContext()["portals"]
	secretParams := req.GetVolumeContext()["secret"]
	secret := parseSecret(secretParams)

	portal := portalMounter(tp)
	var portals []string
	portals = append(portals, portal)

	iface := req.GetVolumeContext()["iscsiInterface"]
	initiatorName := req.GetVolumeContext()["initiatorName"]
	chapDiscovery := false
	if req.GetVolumeContext()["discoveryCHAPAuth"] == "true" {
		chapDiscovery = true
	}

	chapSession := false
	if req.GetVolumeContext()["sessionCHAPAuth"] == "true" {
		chapSession = true
	}

	return &iscsiDisk{
		VolName:       volName,
		Portals:       portals,
		Iqn:           iqn,
		lun:           lun,
		Iface:         iface,
		chapDiscovery: chapDiscovery,
		chapSession:   chapSession,
		secret:        secret,
		InitiatorName: initiatorName}, nil
}

func getISCSIDiskMounter(iscsiInfo *iscsiDisk, vol *v1alpha1.CSIVolume) *iscsiDiskMounter {

	return &iscsiDiskMounter{
		iscsiDisk:    iscsiInfo,
		fsType:       vol.Spec.Volume.FSType,
		readOnly:     vol.Spec.Volume.ReadOnly,
		mountOptions: vol.Spec.Volume.MountOptions,
		mounter:      &mount.SafeFormatAndMount{Interface: mount.New(""), Exec: mount.NewOsExec()},
		exec:         mount.NewOsExec(),
		targetPath:   vol.Spec.Volume.MountPath,
		deviceUtil:   util.NewDeviceHandler(util.NewIOHandler()),
	}
}

func getISCSIDiskUnmounter(req *csi.NodeUnpublishVolumeRequest) *iscsiDiskUnmounter {
	return &iscsiDiskUnmounter{
		iscsiDisk: &iscsiDisk{
			VolName: req.GetVolumeId(),
		},
		mounter: mount.New(""),
		exec:    mount.NewOsExec(),
	}
}
