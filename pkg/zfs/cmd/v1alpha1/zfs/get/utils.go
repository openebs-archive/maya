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

package vget

import "fmt"

// SetProperty method append the property to Proplist field of VolumeGetProperty object.
func (v *VolumeGetProperty) SetProperty(key, value string) {
	v.Proplist = append(v.Proplist, fmt.Sprintf("%s=%s", key, value))
}

// SetDataset method set the Dataset field of VolumeGetProperty object.
func (v *VolumeGetProperty) SetDataset(Dataset string) {
	v.Dataset = Dataset
}

// SetSnapshot method set the Snapshot field of VolumeGetProperty object.
func (v *VolumeGetProperty) SetSnapshot(Snapshot string) {
	v.Snapshot = Snapshot
}

// SetCommand method set the Command field of VolumeGetProperty object.
func (v *VolumeGetProperty) SetCommand(Command string) {
	v.Command = Command
}

// GetProplist method get the Proplist field of VolumeGetProperty object.
func (v *VolumeGetProperty) GetProplist() []string {
	return v.Proplist
}

// GetDataset method get the Dataset field of VolumeGetProperty object.
func (v *VolumeGetProperty) GetDataset() string {
	return v.Dataset
}

// GetSnapshot method get the Snapshot field of VolumeGetProperty object.
func (v *VolumeGetProperty) GetSnapshot() string {
	return v.Snapshot
}

// GetCommand method get the Command field of VolumeGetProperty object.
func (v *VolumeGetProperty) GetCommand() string {
	return v.Command
}
