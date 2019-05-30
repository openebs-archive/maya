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

package pool

import (
	"encoding/json"
	"github.com/golang/glog"
)

// PoolTopology contains the topology strucure of disks used in backend
type PoolTopology struct {
	Children  int          `json:"vdev_children,omitempty"`
	Vdev_tree PoolVdevTree `json:"vdev_tree,omitempty"`
}

// PoolVdevTree contains the tree strucure of disks used in backend
type PoolVdevTree struct {
	Vdev_type string `json:"type,omitempty"`
	Topvdev   []Vdev `json:"children,omitempty"`
	Readcache []Vdev `json:"l2cache,omitempty"`
	Spares    []Vdev `json:"spares,omitempty"`
}

// Vdev relates to a logical or physical disk in backend
type Vdev struct {
	Vdev_type string `json:"type,omitempty"`
	Path      string `json:"path,omitempty"`
	Is_log    int    `json:"is_log,omitempty"`
	Is_spare  int    `json:"is_spare,omitempty"`
	Vdev      []Vdev `json:"children,omitempty"`
}

// ZpoolDump runs 'zpool dump' command and unmarshal the output in above schema
func ZpoolDump() (PoolTopology, error) {
	var t PoolTopology
	out, err := RunnerVar.RunCombinedOutput(PoolOperator, "dump")
	if err != nil {
		glog.Errorf("error in zpool dump output: %v", err)
		return t, err
	}
	err = json.Unmarshal(out, &t)
	return t, err
}
