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

package zfs

import "fmt"

// SetProperty method set the Property field of VolumeSnapshot object.
func (v *VolumeSnapshot) SetProperty(key, value string) {
	v.Property = append(v.Property, fmt.Sprintf("%s=%s", key, value))
}

// SetRecursive method set the Recursive field of VolumeSnapshot object.
func (v *VolumeSnapshot) SetRecursive(Recursive bool) {
	v.Recursive = Recursive
}

// SetSnapshot method set the Snapshot field of VolumeSnapshot object.
func (v *VolumeSnapshot) SetSnapshot(Snapshot string) {
	v.Snapshot = Snapshot
}

// SetDataset method set the Dataset field of VolumeSnapshot object.
func (v *VolumeSnapshot) SetDataset(Dataset string) {
	v.Dataset = Dataset
}

// SetOpType method set the OpType field of VolumeSnapshot object.
func (v *VolumeSnapshot) SetOpType(OpType SnapOp) {
	v.OpType = OpType
}

// SetTarget method set the Target field of VolumeSnapshot object.
func (v *VolumeSnapshot) SetTarget(Target string) {
	v.Target = Target
}

// SetDedup method set the Dedup field of VolumeSnapshot object.
func (v *VolumeSnapshot) SetDedup(Dedup bool) {
	v.Dedup = Dedup
}

// SetLastSnapshot method set the LastSnapshot field of VolumeSnapshot object.
func (v *VolumeSnapshot) SetLastSnapshot(LastSnapshot string) {
	v.LastSnapshot = LastSnapshot
}

// SetDryRun method set the DryRun field of VolumeSnapshot object.
func (v *VolumeSnapshot) SetDryRun(DryRun bool) {
	v.DryRun = DryRun
}

// SetEnableCompression method set the EnableCompression field of VolumeSnapshot object.
func (v *VolumeSnapshot) SetEnableCompression(EnableCompression bool) {
	v.EnableCompression = EnableCompression
}

// SetCommand method set the Command field of VolumeSnapshot object.
func (v *VolumeSnapshot) SetCommand(Command string) {
	v.Command = Command
}

// GetProperty method get the Property field of VolumeSnapshot object.
func (v *VolumeSnapshot) GetProperty() []string {
	return v.Property
}

// GetRecursive method get the Recursive field of VolumeSnapshot object.
func (v *VolumeSnapshot) GetRecursive() bool {
	return v.Recursive
}

// GetSnapshot method get the Snapshot field of VolumeSnapshot object.
func (v *VolumeSnapshot) GetSnapshot() string {
	return v.Snapshot
}

// GetDataset method get the Dataset field of VolumeSnapshot object.
func (v *VolumeSnapshot) GetDataset() string {
	return v.Dataset
}

// GetOpType method get the OpType field of VolumeSnapshot object.
func (v *VolumeSnapshot) GetOpType() SnapOp {
	return v.OpType
}

// GetTarget method get the Target field of VolumeSnapshot object.
func (v *VolumeSnapshot) GetTarget() string {
	return v.Target
}

// GetDedup method get the Dedup field of VolumeSnapshot object.
func (v *VolumeSnapshot) GetDedup() bool {
	return v.Dedup
}

// GetLastSnapshot method get the LastSnapshot field of VolumeSnapshot object.
func (v *VolumeSnapshot) GetLastSnapshot() string {
	return v.LastSnapshot
}

// GetDryRun method get the DryRun field of VolumeSnapshot object.
func (v *VolumeSnapshot) GetDryRun() bool {
	return v.DryRun
}

// GetEnableCompression method get the EnableCompression field of VolumeSnapshot object.
func (v *VolumeSnapshot) GetEnableCompression() bool {
	return v.EnableCompression
}

// GetCommand method get the Command field of VolumeSnapshot object.
func (v *VolumeSnapshot) GetCommand() string {
	return v.Command
}
