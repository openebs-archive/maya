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

package iscsi

import (
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/kubernetes/pkg/util/mount"
)

// UnmountAndDetachDisk unmounts the disk from the specified path
// and logs out of the iSCSI Volume
func UnmountAndDetachDisk(vol *v1alpha1.CSIVolume, path string) error {
	iscsiInfo := &iscsiDisk{
		VolName: vol.Spec.Volume.Volname,
		Portals: []string{vol.Spec.ISCSI.TargetPortal},
		Iqn:     vol.Spec.ISCSI.Iqn,
		lun:     vol.Spec.ISCSI.Lun,
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

// AttachAndMountDisk logs in to the iSCSI Volume
// and mounts the disk to the specified path
func AttachAndMountDisk(vol *v1alpha1.CSIVolume) (string, error) {
	if len(vol.Spec.Volume.MountPath) == 0 {
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
