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

package vclone

// SetSnapshot method set the Snapshot field of VolumeClone object.
func (v *VolumeClone) SetSnapshot(Snapshot string) {
	v.Snapshot = Snapshot
}

// SetTargetDataset method set the TargetDataset field of VolumeClone object.
func (v *VolumeClone) SetTargetDataset(TargetDataset string) {
	v.TargetDataset = TargetDataset
}

// SetSourceDataset method set the SourceDataset field of VolumeClone object.
func (v *VolumeClone) SetSourceDataset(SourceDataset string) {
	v.SourceDataset = SourceDataset
}

// SetProperty method append the Property to VolumeClone object's property.
func (v *VolumeClone) SetProperty(key, value string) {
	v.Property = append(v.Property, "%s=%s", key, value)
}

// SetCreateParent method set the CreateParent field of VolumeClone object.
func (v *VolumeClone) SetCreateParent(CreateParent bool) {
	v.CreateParent = CreateParent
}

// SetCommand method set the Command field of VolumeClone object.
func (v *VolumeClone) SetCommand(Command string) {
	v.Command = Command
}
