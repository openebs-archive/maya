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
)

type PoolTopology struct {
	children	string		`json:"vdev_children"`
	vdev_tree	PoolVdevTree	`json:"vdev_tree"`
}

type PoolVdevTree struct {
	vdev_type	string		`json:"type"`
	topvdev		[]Vdev		`json:"children"`
	readcache	[]Vdev		`json:"l2cache"`
	spares		[]Vdev		`json:"spares"`
}

type Vdev struct {
	vdev_type	string		`json:"type"`
	path		string		`json:"path"`
	is_log		string		`json:"is_log"`
	is_spare	string		`json:"is_spare"`
	vdev		[]Vdev		`json:"children"`
}

func ZpoolDump() (PoolTopology, error) {
	var t PoolTopology
	out, err := RunnerVar.RunCombinedOutput(PoolOperator, "dump")
	if err != nil {
		return t, err;
	}
	err = json.Unmarshal(out, &t)
	return t, err
}

