/*
Copyright 2019 The OpenEBS Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package command

import (
	vclone "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zfs/clone"
	vcreate "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zfs/create"
	vdestroy "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zfs/destroy"
	vproperty "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zfs/property"
	vsnapshotrecv "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zfs/receive"
	vrename "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zfs/rename"
	vrollback "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zfs/rollback"
	vsnapshotsend "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zfs/send"
	vsnapshot "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zfs/snapshot"
	padd "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zpool/add"
	pattach "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zpool/attach"
	pclear "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zpool/clear"
	pcreate "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zpool/create"
	pdestroy "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zpool/destroy"
	pdetach "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zpool/detach"
	pexport "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zpool/export"
	pimport "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zpool/import"
	poffline "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zpool/offline"
	ponline "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zpool/online"
	pproperty "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zpool/property"
	premove "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zpool/remove"
	pstatus "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zpool/status"
)

// NewVolumeClone returns new instance of object VolumeClone
func NewVolumeClone() *vclone.VolumeClone {
	return &vclone.VolumeClone{}
}

// NewPoolProperty returns new instance of object PoolProperty
func NewPoolProperty() *pproperty.PoolProperty {
	return &pproperty.PoolProperty{}
}

// NewPoolOffline returns new instance of object PoolOffline
func NewPoolOffline() *poffline.PoolOffline {
	return &poffline.PoolOffline{}
}

// NewPoolExpansion returns new instance of object PoolExpansion
func NewPoolExpansion() *padd.PoolExpansion {
	return &padd.PoolExpansion{}
}

// NewPoolStatus returns new instance of object PoolStatus
func NewPoolStatus() *pstatus.PoolStatus {
	return &pstatus.PoolStatus{}
}

// NewPoolDestroy returns new instance of object PoolDestroy
func NewPoolDestroy() *pdestroy.PoolDestroy {
	return &pdestroy.PoolDestroy{}
}

// NewPoolDetach returns new instance of object PoolDetach
func NewPoolDetach() *pdetach.PoolDetach {
	return &pdetach.PoolDetach{}
}

// NewPoolRemove returns new instance of object PoolRemove
func NewPoolRemove() *premove.PoolRemove {
	return &premove.PoolRemove{}
}

// NewPoolClear returns new instance of object PoolClear
func NewPoolClear() *pclear.PoolClear {
	return &pclear.PoolClear{}
}

// NewPoolOnline returns new instance of object PoolOnline
func NewPoolOnline() *ponline.PoolOnline {
	return &ponline.PoolOnline{}
}

// NewPoolImport returns new instance of object PoolImport
func NewPoolImport() *pimport.PoolImport {
	return &pimport.PoolImport{}
}

// NewPoolAttach returns new instance of object PoolAttach
func NewPoolAttach() *pattach.PoolAttach {
	return &pattach.PoolAttach{}
}

// NewPoolExport returns new instance of object PoolExport
func NewPoolExport() *pexport.PoolExport {
	return &pexport.PoolExport{}
}

// NewPoolCreate returns new instance of object PoolCreate
func NewPoolCreate() *pcreate.PoolCreate {
	return &pcreate.PoolCreate{}
}

// NewVolumeProperty returns new instance of object VolumeProperty
func NewVolumeProperty() *vproperty.VolumeProperty {
	return &vproperty.VolumeProperty{}
}

// NewVolumeRollback returns new instance of object VolumeRollback
func NewVolumeRollback() *vrollback.VolumeRollback {
	return &vrollback.VolumeRollback{}
}

// NewVolumeDestroy returns new instance of object VolumeDestroy
func NewVolumeDestroy() *vdestroy.VolumeDestroy {
	return &vdestroy.VolumeDestroy{}
}

// NewVolumeRename returns new instance of object VolumeRename
func NewVolumeRename() *vrename.VolumeRename {
	return &vrename.VolumeRename{}
}

// NewVolumeSnapshot returns new instance of object VolumeSnapshot
func NewVolumeSnapshot() *vsnapshot.VolumeSnapshot {
	return &vsnapshot.VolumeSnapshot{}
}

// NewVolumeCreate returns new instance of object VolumeCreate
func NewVolumeCreate() *vcreate.VolumeCreate {
	return &vcreate.VolumeCreate{}
}

// NewVolumeSnapshotSend returns new instance of object VolumeSnapshotSend
func NewVolumeSnapshotSend() *vsnapshotsend.VolumeSnapshotSend {
	return &vsnapshotsend.VolumeSnapshotSend{}
}

// NewVolumeSnapshotRecv returns new instance of object VolumeSnapshotRecv
func NewVolumeSnapshotRecv() *vsnapshotrecv.VolumeSnapshotRecv {
	return &vsnapshotrecv.VolumeSnapshotRecv{}
}
