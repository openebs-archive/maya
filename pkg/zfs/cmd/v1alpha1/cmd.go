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
	vget "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zfs/get"
	vlistsnap "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zfs/listsnap"
	vsnapshotrecv "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zfs/receive"
	vrename "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zfs/rename"
	vrollback "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zfs/rollback"
	vsnapshotsend "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zfs/send"
	vset "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zfs/set"
	vsnapshot "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zfs/snapshot"
	padd "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zpool/add"
	pattach "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zpool/attach"
	pclear "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zpool/clear"
	pcreate "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zpool/create"
	pdestroy "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zpool/destroy"
	pdetach "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zpool/detach"
	pdump "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zpool/dump"
	pexport "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zpool/export"
	pget "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zpool/get"
	pimport "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zpool/import"
	plabelclear "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zpool/labelclear"
	poffline "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zpool/offline"
	ponline "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zpool/online"
	premove "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zpool/remove"
	preplace "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zpool/replace"
	pset "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zpool/set"
	pstatus "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/zpool/status"
)

// NewVolumeClone returns new instance of object VolumeClone
func NewVolumeClone() *vclone.VolumeClone {
	return &vclone.VolumeClone{}
}

// NewPoolSetProperty returns new instance of object PoolSProperty
func NewPoolSetProperty() *pset.PoolSProperty {
	return &pset.PoolSProperty{}
}

// NewPoolGetProperty returns new instance of object PoolGProperty
func NewPoolGetProperty() *pget.PoolGProperty {
	return &pget.PoolGProperty{}
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

// NewVolumeGetProperty returns new instance of object VolumeGetProperty
func NewVolumeGetProperty() *vget.VolumeGetProperty {
	return &vget.VolumeGetProperty{}
}

// NewVolumeListSnapshot returns new instance of object VolumeListSnapshot
func NewVolumeListSnapshot() *vlistsnap.VolumeListSnapshot {
	return &vlistsnap.VolumeListSnapshot{}
}

// NewVolumeSetProperty returns new instance of object VolumeSetProperty
func NewVolumeSetProperty() *vset.VolumeSetProperty {
	return &vset.VolumeSetProperty{}
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

// NewPoolLabelClear returns new instance of object PoolLabelClear
func NewPoolLabelClear() *plabelclear.PoolLabelClear {
	return &plabelclear.PoolLabelClear{}
}

// NewPoolDiskReplace returns new instance of object PoolDiskReplace
func NewPoolDiskReplace() *preplace.PoolDiskReplace {
	return &preplace.PoolDiskReplace{}
}

// NewPoolDump returns new instance of object PoolDump
func NewPoolDump() *pdump.PoolDump {
	return &pdump.PoolDump{}
}
