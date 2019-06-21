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

package vcreate

import "fmt"

// SetName method set the Name field of VolumeCreate object.
func (v *VolumeCreate) SetName(Name string) {
	v.Name = Name
}

// SetSize method set the Size field of VolumeCreate object.
func (v *VolumeCreate) SetSize(Size string) {
	v.Size = Size
}

// SetBlockSize method set the BlockSize field of VolumeCreate object.
func (v *VolumeCreate) SetBlockSize(BlockSize string) {
	v.BlockSize = BlockSize
}

// SetProperty method set the Property field of VolumeCreate object.
func (v *VolumeCreate) SetProperty(key, value string) {
	v.Property = append(v.Property, fmt.Sprintf("%s=%s", key, value))
}

// SetReservation method set the Reservation field of VolumeCreate object.
func (v *VolumeCreate) SetReservation(Reservation bool) {
	v.Reservation = Reservation
}

// SetCreateParent method set the CreateParent field of VolumeCreate object.
func (v *VolumeCreate) SetCreateParent(CreateParent bool) {
	v.CreateParent = CreateParent
}

// SetCommand method set the Command field of VolumeCreate object.
func (v *VolumeCreate) SetCommand(Command string) {
	v.Command = Command
}
