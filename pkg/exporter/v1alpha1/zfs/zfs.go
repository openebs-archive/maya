package zfs

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/openebs/maya/pkg/util"
)

const (
	Zfs = "./zfs"
)

var (
	ZVolStatus           map[string]float64 = map[string]float64{"Offline": 0, "Healthy": 1, "Degraded": 2, "Rebuilding": 3}
	ZVolRebuildingStatus map[string]float64 = map[string]float64{"INIT": 0, "DONE": 1, "SNAP REBUILD INPROGRESS": 2, "ACTIVE DATASET REBUILD INPROGRESS": 3, "ERRORED  ": 4, "FAILED": 5, "UNKNOWN": 6}
)

type Stats struct {
	Volumes []Volume `json:"stats"`
}

type Volume struct {
	Name          string `json:"name"`
	Status        string `json:"status"`
	RebuildStatus string `json:"rebuildStatus"`

	//  Size               float64
	SyncCount  float64 `json:"syncCount"`
	ReadBytes  float64 `json:"readByte"`
	WriteBytes float64 `json:"writeByte"`
	//  LogicalUsed        float64 `json:"s,string"`
	SyncLatency        float64 `json:"syncLatency"`
	ReadLatency        float64 `json:"readLatency"`
	WriteLatency       float64 `json:"writeLatency"`
	RebuildCount       float64 `json:"rebuildCnt"`
	RebuildBytes       float64 `json:"rebuildBytes"`
	InflightIOCount    float64 `json:"inflightIOCnt"`
	RebuildDoneCount   float64 `json:"rebuildDoneCnt"`
	DispatchedIOCount  float64 `json:"dispatchedIOCnt"`
	RebuildFailedCount float64 `json:"rebuildFailedCnt"`
}

func Run(runner util.Runner, args ...string) ([]byte, error) {
	out, err := runner.RunCombinedOutput(Zfs, args...)
	if err != nil {
		return out, err
	}
	return out, nil
}

func StatsParser(output []byte) (Stats, error) {
	var stats = Stats{}
	if err := json.NewDecoder(bytes.NewReader(output)).Decode(&stats); err != nil {
		return stats, err
	}
	if !isVolumeExist(stats.Volumes) {
		return stats, errors.New("Got empty pool/volume name")
	}
	return stats, nil
}

func isVolumeExist(volumes []Volume) bool {
	for _, vol := range volumes {
		if vol.Name == "" {
			return false
		}
	}
	return true
}

func StatsList(vol Volume) []float64 {
	return []float64{
		vol.SyncCount,
		vol.ReadBytes,
		vol.WriteBytes,
		vol.SyncLatency,
		vol.ReadLatency,
		vol.WriteLatency,
		vol.RebuildCount,
		vol.RebuildBytes,
		vol.InflightIOCount,
		vol.RebuildDoneCount,
		vol.DispatchedIOCount,
		vol.RebuildFailedCount,
		ZVolStatus[vol.Status],
		ZVolRebuildingStatus[vol.RebuildStatus],
	}
}
