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

package vsnapshotsend

// SetSnapshot method set the Snapshot field of VolumeSnapshotSend object.
func (v *VolumeSnapshotSend) SetSnapshot(Snapshot string) {
	v.Snapshot = Snapshot
}

// SetDataset method set the Dataset field of VolumeSnapshotSend object.
func (v *VolumeSnapshotSend) SetDataset(Dataset string) {
	v.Dataset = Dataset
}

// SetTarget method set the Target field of VolumeSnapshotSend object.
func (v *VolumeSnapshotSend) SetTarget(Target string) {
	v.Target = Target
}

// SetDedup method set the Dedup field of VolumeSnapshotSend object.
func (v *VolumeSnapshotSend) SetDedup(Dedup bool) {
	v.Dedup = Dedup
}

// SetLastSnapshot method set the LastSnapshot field of VolumeSnapshotSend object.
func (v *VolumeSnapshotSend) SetLastSnapshot(LastSnapshot string) {
	v.LastSnapshot = LastSnapshot
}

// SetDryRun method set the DryRun field of VolumeSnapshotSend object.
func (v *VolumeSnapshotSend) SetDryRun(DryRun bool) {
	v.DryRun = DryRun
}

// SetEnableCompression method set the EnableCompression field of VolumeSnapshotSend object.
func (v *VolumeSnapshotSend) SetEnableCompression(EnableCompression bool) {
	v.EnableCompression = EnableCompression
}

// SetCommand method set the Command field of VolumeSnapshotSend object.
func (v *VolumeSnapshotSend) SetCommand(Command string) {
	v.Command = Command
}

// GetSnapshot method get the Snapshot field of VolumeSnapshotSend object.
func (v *VolumeSnapshotSend) GetSnapshot() string {
	return v.Snapshot
}

// GetDataset method get the Dataset field of VolumeSnapshotSend object.
func (v *VolumeSnapshotSend) GetDataset() string {
	return v.Dataset
}

// GetTarget method get the Target field of VolumeSnapshotSend object.
func (v *VolumeSnapshotSend) GetTarget() string {
	return v.Target
}

// GetDedup method get the Dedup field of VolumeSnapshotSend object.
func (v *VolumeSnapshotSend) GetDedup() bool {
	return v.Dedup
}

// GetLastSnapshot method get the LastSnapshot field of VolumeSnapshotSend object.
func (v *VolumeSnapshotSend) GetLastSnapshot() string {
	return v.LastSnapshot
}

// GetDryRun method get the DryRun field of VolumeSnapshotSend object.
func (v *VolumeSnapshotSend) GetDryRun() bool {
	return v.DryRun
}

// GetEnableCompression method get the EnableCompression field of VolumeSnapshotSend object.
func (v *VolumeSnapshotSend) GetEnableCompression() bool {
	return v.EnableCompression
}

// GetCommand method get the Command field of VolumeSnapshotSend object.
func (v *VolumeSnapshotSend) GetCommand() string {
	return v.Command
}
