// Copyright © 2017-2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1

//OsCommand is for operating system related commands for block devices
type OsCommand struct {
	Command string
	Flag    string
}

//BlockDeviceInfo exposes the json output of lsblk:"blockdevices"
type BlockDeviceInfo struct {
	Blockdevices []Blockdevice `json:"blockdevices"`
}

// Blockdevice has block disk fields
type Blockdevice struct {
	Name       string        `json:"name"`               //block device name
	Majmin     string        `json:"maj:min"`            //major and minor block device number
	Rm         string        `json:"rm"`                 //is device removable
	Size       string        `json:"size"`               //size of device
	Ro         string        `json:"ro"`                 //is device read-only
	Type       string        `json:"type"`               //is device disk or partition
	Mountpoint string        `json:"mountpoint"`         //block device mountpoint
	Children   []Blockdevice `json:"children,omitempty"` //Blockdevice ...
}
