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

package v1alpha1

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"

	"github.com/openebs/maya/pkg/exec"
)

// ZVolStatus is zvol's status
type ZVolStatus string

// ZVolRebuildingStatus is zvol's rebuilding status
type ZVolRebuildStatus string

const (
	// Binary represents zfs binary
	Binary = "zfs"
	// StatusOffline ...
	StatusOffline ZVolStatus = "Offline"
	// StatusHealthy ...
	StatusHealthy ZVolStatus = "Healthy"
	// StatusDegraded ...
	StatusDegraded ZVolStatus = "Degraded"
	// StatusRebuilding ...
	StatusRebuilding ZVolStatus = "Rebuilding"
	// NoDataSetAvailable ...
	NoDataSetAvailable ZVolStatus = "no datasets available"
	// InitializeLibuzfsClientErr ...
	InitializeLibuzfsClientErr ZVolStatus = "failed to initialize libuzfs client"
	// RebuildStatusInit ...
	RebuildStatusInit ZVolRebuildStatus = "INIT"
	// RebuildStatusDone ...
	RebuildStatusDone ZVolRebuildStatus = "DONE"
	// RebuildStatusErrored ...
	RebuildStatusErrored ZVolRebuildStatus = "ERRORED  "
	// RebuildStatusFailed ...
	RebuildStatusFailed ZVolRebuildStatus = "FAILED"
	// RebuildStatusUnknown ...
	RebuildStatusUnknown ZVolRebuildStatus = "UNKNOWN"
	// RebuildStatusInProgress ...
	RebuildStatusInProgress ZVolRebuildStatus = "SNAP REBUILD INPROGRESS"
	// RebuildStatusActiveDataSetInProgress ...
	RebuildStatusActiveDataSetInProgress ZVolRebuildStatus = "ACTIVE DATASET REBUILD INPROGRESS"
)

var (
	// Status is mapping of the  zvol status with values
	Status = map[ZVolStatus]float64{
		StatusOffline:    0,
		StatusHealthy:    1,
		StatusDegraded:   2,
		StatusRebuilding: 3,
	}
	// RebuildingStatus is mapping of rebuilding status of zvol with values
	RebuildingStatus = map[ZVolRebuildStatus]float64{
		RebuildStatusInit:                    0,
		RebuildStatusDone:                    1,
		RebuildStatusInProgress:              2,
		RebuildStatusActiveDataSetInProgress: 3,
		RebuildStatusErrored:                 4,
		RebuildStatusFailed:                  5,
		RebuildStatusUnknown:                 6,
	}
)

// Stats represents list of volume
type Stats struct {
	Volumes []Volume `json:"stats"`
}

// Volume represents the volume's various stats
type Volume struct {
	// Name contains name of pool appened with volume name.
	// It's of the form "<pool name>/<volume name>"
	Name          string            `json:"name"`
	Status        ZVolStatus        `json:"status"`        // Status of volume
	RebuildStatus ZVolRebuildStatus `json:"rebuildStatus"` // RebuildStatus of volume

	SyncCount   float64 `json:"syncCount"`   // Total Sync processed on this volume
	ReadCount   float64 `json:"readCount"`   // Total Reads
	WriteCount  float64 `json:"writeCount"`  // Total Writes
	ReadBytes   float64 `json:"readByte"`    // Total Reads in bytes
	WriteBytes  float64 `json:"writeByte"`   // Total Writes in bytes
	SyncLatency float64 `json:"syncLatency"` // Latency involved in processing sync io's

	ReadLatency        float64 `json:"readLatency"`      // Latency involved in processing read io's
	WriteLatency       float64 `json:"writeLatency"`     // Latency involved in processing write io's
	RebuildCount       float64 `json:"rebuildCnt"`       // Total rebuild processed
	RebuildBytes       float64 `json:"rebuildBytes"`     // Total rebuild in bytes
	InflightIOCount    float64 `json:"inflightIOCnt"`    // Total IO's processing currently
	RebuildDoneCount   float64 `json:"rebuildDoneCnt"`   // Total no of rebuilds done
	DispatchedIOCount  float64 `json:"dispatchedIOCnt"`  // Total IO's dispatched to disk
	RebuildFailedCount float64 `json:"rebuildFailedCnt"` // Total no of failed rebuilds
}

func (z ZVolStatus) String() string {
	return string(z)
}

// Run is wrapper over RunCommandWithTimeoutContext for zfs commands
func Run(runner exec.Runner) ([]byte, error) {
	out, err := runner.RunCommandWithTimeoutContext()
	if err != nil {
		return out, err
	}
	return out, nil
}

func checkError(str string) error {
	if IsNotInitialized(str) {
		return errors.New(string(InitializeLibuzfsClientErr))
	}

	if IsNoDataSetAvailable(str) {
		return errors.New(string(NoDataSetAvailable))
	}
	return nil
}

// StatsParser parses the json response of zfs stats command.
func StatsParser(stdout []byte) (Stats, error) {
	var stats = Stats{}
	if err := checkError(string(stdout)); err != nil {
		return Stats{}, err
	}

	if err := json.NewDecoder(bytes.NewReader(stdout)).Decode(&stats); err != nil {
		return stats, err
	}

	if isNotPresent(stats.Volumes) {
		return stats, errors.New("Got empty pool/volume name")
	}

	return stats, nil
}

// IsNotInitialized check whether libuzfs is initialized or not
func IsNotInitialized(str string) bool {
	return strings.Contains(str, string(InitializeLibuzfsClientErr))
}

// IsNoDataSetAvailable check whether dataset available or not
func IsNoDataSetAvailable(str string) bool {
	return strings.Contains(str, string(NoDataSetAvailable))
}

func isNotPresent(vol []Volume) bool {
	if len(vol) == 0 {
		return true
	}
	for _, v := range vol {
		if v.Name == "" {
			return true
		}
	}
	return false
}

// StatsList returns the list of stats
// NOTE: Please donot edit the order, add new stats
// at the end of the list.
func StatsList(v Volume) []float64 {
	return []float64{
		v.SyncCount,
		v.ReadCount,
		v.WriteCount,
		v.ReadBytes,
		v.WriteBytes,
		v.SyncLatency,
		v.ReadLatency,
		v.WriteLatency,
		v.RebuildCount,
		v.RebuildBytes,
		v.InflightIOCount,
		v.RebuildDoneCount,
		v.DispatchedIOCount,
		v.RebuildFailedCount,
		Status[v.Status],
		RebuildingStatus[v.RebuildStatus],
	}
}
