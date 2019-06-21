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

package vrename

// SetCreateParent method set the CreateParent field of VolumeRename object.
func (v *VolumeRename) SetCreateParent(CreateParent bool) {
	v.CreateParent = CreateParent
}

// SetForceUnmount method set the ForceUnmount field of VolumeRename object.
func (v *VolumeRename) SetForceUnmount(ForceUnmount bool) {
	v.ForceUnmount = ForceUnmount
}

// SetSource method set the Source field of VolumeRename object.
func (v *VolumeRename) SetSource(Source string) {
	v.Source = Source
}

// SetDest method set the Dest field of VolumeRename object.
func (v *VolumeRename) SetDest(Dest string) {
	v.Dest = Dest
}

// SetCommand method set the Command field of VolumeRename object.
func (v *VolumeRename) SetCommand(Command string) {
	v.Command = Command
}

// GetCreateParent method get the CreateParent field of VolumeRename object.
func (v *VolumeRename) GetCreateParent() bool {
	return v.CreateParent
}

// GetForceUnmount method get the ForceUnmount field of VolumeRename object.
func (v *VolumeRename) GetForceUnmount() bool {
	return v.ForceUnmount
}

// GetSource method get the Source field of VolumeRename object.
func (v *VolumeRename) GetSource() string {
	return v.Source
}

// GetDest method get the Dest field of VolumeRename object.
func (v *VolumeRename) GetDest() string {
	return v.Dest
}

// GetCommand method get the Command field of VolumeRename object.
func (v *VolumeRename) GetCommand() string {
	return v.Command
}
