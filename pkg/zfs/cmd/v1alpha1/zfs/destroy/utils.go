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

package vdestroy

// SetDryRun method set the DryRun field of VolumeDestroy object.
func (v *VolumeDestroy) SetDryRun(DryRun bool) {
	v.DryRun = DryRun
}

// SetRecursive method set the Recursive field of VolumeDestroy object.
func (v *VolumeDestroy) SetRecursive(Recursive bool) {
	v.Recursive = Recursive
}

// SetName method set the Name field of VolumeDestroy object.
func (v *VolumeDestroy) SetName(Name string) {
	v.Name = Name
}

// SetCommand method set the Command field of VolumeDestroy object.
func (v *VolumeDestroy) SetCommand(Command string) {
	v.Command = Command
}

// GetDryRun method get the DryRun field of VolumeDestroy object.
func (v *VolumeDestroy) GetDryRun() bool {
	return v.DryRun
}

// GetRecursive method get the Recursive field of VolumeDestroy object.
func (v *VolumeDestroy) GetRecursive() bool {
	return v.Recursive
}

// GetName method get the Name field of VolumeDestroy object.
func (v *VolumeDestroy) GetName() string {
	return v.Name
}

// GetCommand method get the Command field of VolumeDestroy object.
func (v *VolumeDestroy) GetCommand() string {
	return v.Command
}
