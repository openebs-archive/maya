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
	"strings"

	"github.com/openebs/maya/pkg/util"
)

/*
* enum maintaing in cstor data plane side
* typedef enum dsl_scan_state {
*        DSS_NONE,
*        DSS_SCANNING,
*        DSS_FINISHED,
*        DSS_CANCELED,
*        DSS_NUM_STATES
*} dsl_scan_state_t;
 */

//TODO: Improve comments during review process

//PoolScanState states various pool scan states
type PoolScanState uint64

const (
	// PoolScanNone represents pool scanning is not yet started
	PoolScanNone PoolScanState = iota
	// PoolScanScanning represents pool is undergoing scanning
	PoolScanScanning
	// PoolScanFinished represents pool scanning is finished
	PoolScanFinished
	// PoolScanCanceled represents pool scan is aborted
	PoolScanCanceled
	// PoolScanNumOfStates holds value 4
	PoolScanNumOfStates
)

// PoolScanFunc holds various scanning functions
type PoolScanFunc uint64

const (
	// PoolScanFuncNone holds value 0
	PoolScanFuncNone PoolScanFunc = iota
	// PoolScanFuncScrub holds value 1
	PoolScanFuncScrub
	// PoolScanFuncResilver holds value 2 which states device under went resilvering
	PoolScanFuncResilver
	// PoolScanFuncStates holds value 3
	PoolScanFuncStates
)

const (
	// PoolOperator is the name of the tool that makes pool-related operations.
	PoolOperator = "zpool"
	// VdevScanProcessedIndex is index of scaned bytes on disk
	VdevScanProcessedIndex = 25
	// VdevScanStatsStateIndex represents the index of dataset scan state
	VdevScanStatsStateIndex = 1
	// VdevScanStatsScanFuncIndex point to index which inform whether device
	// under went resilvering or not
	VdevScanStatsScanFuncIndex = 0
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

	// vdev indetailed statistics
	VdevStats []uint64 `json:"vdev_stats,omitempty"`

	// ScanStats states replaced device scan state
	ScanStats []uint64 `json:"scan_stats,omitempty"`
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

	// 0 means partitioned disk, 1 means whole disk
	IsWholeDisk int `json:"whole_disk,omitempty"`

	// vdev indetailed statistics
	VdevStats []uint64 `json:"vdev_stats,omitempty"`

	ScanStats []uint64 `json:"scan_stats,omitempty"`

	// child vdevs of the logical disk or null for physical disk/sparse
	Children []Vdev `json:"children,omitempty"`
}

// VdevList is alias of list of Vdevs
type VdevList []Vdev

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

// GetVdevFromPath returns vdev if provided path exists in vdev topology
func (l VdevList) GetVdevFromPath(path string) (Vdev, bool) {
	for _, v := range l {
		if strings.EqualFold(path, v.Path) {
			return v, true
		}
		for _, p := range v.Children {
			if strings.EqualFold(path, p.Path) {
				return p, true
			}
			if vdev, r := VdevList(p.Children).GetVdevFromPath(path); r {
				return vdev, true
			}
		}
	}
	return Vdev{}, false
}
