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

package vsnapshot

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

// GetCommand method get the Command field of VolumeSnapshot object.
func (v *VolumeSnapshot) GetCommand() string {
	return v.Command
}
