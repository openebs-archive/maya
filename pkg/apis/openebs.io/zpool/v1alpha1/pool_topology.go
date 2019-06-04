/*
Copyright 2018 The OpenEBS Authors.

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

package zpool

import (
	"encoding/json"
	"github.com/openebs/maya/pkg/util"
)

// PoolOperator is the name of the tool that makes pool-related operations.
const (
	PoolOperator = "zpool"
)

// Topology contains the topology strucure of disks used in backend
type Topology struct {
	// Number of top-level children in topology (doesnt include spare/l2cache)
	ChildrenCount int `json:"vdev_children,omitempty"`

	// Root of vdev topology
	VdevTree VdevTree `json:"vdev_tree,omitempty"`
}

// VdevTree contains the tree strucure of disks used in backend
type VdevTree struct {
	// root for Root vdev, Raid type in case of non-level 0 vdev,
	// and file/disk in case of level-0 vdev
	VdevType string `json:"type,omitempty"`

	// top-level vdev topology
	Topvdev []Vdev `json:"children,omitempty"`

	// list of read-cache devices
	Readcache []Vdev `json:"l2cache,omitempty"`

	// list of spare devices
	Spares []Vdev `json:"spares,omitempty"`
}

// Vdev relates to a logical or physical disk in backend
type Vdev struct {
	// root for Root vdev, Raid type in case of non-level 0 vdev,
	// and file/disk in case of level-0 vdev
	VdevType string `json:"type,omitempty"`

	// Path of the disk or sparse file
	Path string `json:"path,omitempty"`

	// 0 means not write-cache device, 1 means write-cache device
	IsLog int `json:"is_log,omitempty"`

	// 0 means not spare device, 1 means spare device
	IsSpare int `json:"is_spare,omitempty"`

	// child vdevs of the logical disk or null for physical disk/sparse
	Children []Vdev `json:"children,omitempty"`
}

// Dump runs 'zpool dump' command and unmarshal the output in above schema
func Dump() (Topology, error) {
	var t Topology
	runnerVar := util.RealRunner{}
	out, err := runnerVar.RunCombinedOutput(PoolOperator, "dump")
	if err != nil {
		return t, err
	}
	err = json.Unmarshal(out, &t)
	return t, err
}
