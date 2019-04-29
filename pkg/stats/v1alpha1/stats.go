// Copyright © 2018-2019 The OpenEBS Authors
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

package stats

import "encoding/json"

// ReplicaMode is the mode of replica.In jiva it can be either RO
// or RW and HEALTHY or DEGRADED for cstor respectively
type ReplicaMode string
type TargetMode string

const (
	// BytesToGB used to convert bytes to GB
	BytesToGB = 1073741824
	// BytesToMB used to convert bytes to MB
	BytesToMB = 1048567
	// BytesToKB used to convert bytes to KB
	BytesToKB = 1024
	// MicSec used to convert to microsec to second
	MicSec = 1000000
)

// VolumeStats is used to store the collected metrics
// TODO: Make this generic, so that it can be used by mayactl
// and other components of maya.
type VolumeStats struct {
	Got bool
	// Iqn is unique iSCSI qualified name, it is used to
	// configure iscsi initiator and target.
	Iqn string `json:"iqn"`
	// Reads is the total no of read io's that's been
	// read from the volume.
	Reads json.Number `json:"ReadIOPS"`
	// TotalReadTime is the total time taken to read from
	// the device.
	TotalReadTime json.Number `json:"TotalReadTime"`
	// TotalReadBlockCount is total no of block that has been
	// read from the volume.
	TotalReadBlockCount json.Number `json:"TotalReadBlockCount"`
	// TotalReadBytes is the total size of the read io's in byte.
	TotalReadBytes json.Number `json:"TotalReadBytes"`
	// Writes is the total no of write io's that's been
	// written to the volume.
	Writes json.Number `json:"WriteIOPS"`
	// TotalWriteTime is the total time taken to write to
	// the volume.
	TotalWriteTime json.Number `json:"TotalWriteTime"`
	// TotalWriteBlockCount is total no of block that has been
	// written to the volume.
	TotalWriteBlockCount json.Number `json:"TotalWriteBlockCount"`
	// TotalWriteBytes is the total size of the write io's
	// in byte.
	TotalWriteBytes json.Number `json:"TotalWriteBytes"`
	// UsedLogicalBlocks is the no of logical blocks that is
	// used by volume.
	UsedLogicalBlocks json.Number `json:"UsedLogicalBlocks"`
	// UsedBlocks is the no of physical blocks used by volume.
	// (each block is a physical sector)
	UsedBlocks json.Number `json:"UsedBlocks"`
	// SectorSize minimum storage unit of a hard drive.
	// Most disk partitioning schemes are designed to have
	// files occupy an integral number of sectors regardless
	// of the file's actual size. Files that do not fill a whole
	// sector will have the remainder of their last sector filled
	// with zeroes. In practice, operating systems typically
	// operate on blocks of data, which may span multiple sectors.
	// (source: https://en.wikipedia.org/wiki/Disk_sector)
	SectorSize json.Number `json:"SectorSize"`
	// Size is the size of the volume created.
	Size json.Number `json:"Size"`
	// RevisionCounter is the no of times io's have been done
	// on volume.
	RevisionCounter json.Number `json:"RevisionCounter"`
	// ReplicaCounter is the no of replicas connected to target
	// or controller (istgt or jiva controller)
	ReplicaCounter json.Number `json:"ReplicaCounter"`
	// Uptime is the time since target is up.
	UpTime json.Number `json:"UpTime"`
	// Name is the name of the volume given while creation
	Name string `json:"Name"`
	// Replicas keeps the details about the replicas connected
	// to the target.
	Replicas []Replica `json:"Replicas"`
	// Target status is the status of the target (RW/RO)
	TargetStatus TargetMode `json:"Status"`
}

// Replica is used to store the info about the replicas
// connected to the target.
type Replica struct {
	// Address is the address of the replica
	Address string `json:"Address"`
	// Mode is the mode of replica.
	Mode ReplicaMode `json:"Mode"`
}
