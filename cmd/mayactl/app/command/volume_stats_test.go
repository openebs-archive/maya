package command

import (
	"testing"

	"github.com/openebs/maya/types/v1"
)

func TestDisplayStats(t *testing.T) {
	annotation := &Annotations{
		TargetPortal:     "10.99.73.74:3260",
		ClusterIP:        "10.99.73.74",
		Iqn:              "iqn.2016-09.com.openebs.jiva:vol1",
		ReplicaCount:     "2",
		ControllerStatus: "Running",
		ReplicaStatus:    "Running,Running",
		VolSize:          "1G",
		ControllerIP:     "",
		Replicas:         "10.10.10.10,10.10.10.11",
	}

	validStats := map[string]struct {
		cmdOptions   *CmdVolumeStatsOptions
		status       []string
		initialStats v1.VolumeMetrics
		finalStats   v1.VolumeMetrics
		output       error
		replicaCount int
	}{
		"StatsJSON": {
			cmdOptions: &CmdVolumeStatsOptions{
				json:    "json",
				volName: "vol1",
			},
			status: []string{
				"10.10.10.10",
				"Online",
				"1",
				"10.10.10.11",
				"Online",
				"1",
			},
			initialStats: v1.VolumeMetrics{
				Name:                 "vol1",
				ReadIOPS:             "0",
				ReplicaCounter:       2,
				RevisionCounter:      100,
				SectorSize:           "4096",
				Size:                 "1073741824",
				TotalReadBlockCount:  "3",
				TotalReadTime:        "10",
				TotalWriteTime:       "15",
				TotalWriteBlockCount: "10",
				UpTime:               13162.971420756,
				UsedBlocks:           "1048576",
				UsedLogicalBlocks:    "1048576",
				WriteIOPS:            "15",
			},
			finalStats: v1.VolumeMetrics{
				Name:                 "vol1",
				ReadIOPS:             "0",
				ReplicaCounter:       2,
				RevisionCounter:      100,
				SectorSize:           "4096",
				Size:                 "1073741824",
				TotalReadBlockCount:  "0",
				TotalReadTime:        "0",
				TotalWriteTime:       "0",
				TotalWriteBlockCount: "0",
				UpTime:               13170.971420756,
				UsedBlocks:           "1048576",
				UsedLogicalBlocks:    "1048576",
				WriteIOPS:            "20",
			},
			output: nil,
		},
		"StatsStd": {
			cmdOptions: &CmdVolumeStatsOptions{
				json:    "",
				volName: "vol1",
			},
			status: []string{
				"10.10.10.10",
				"Online",
				"1",
				"10.10.10.11",
				"Online",
				"1",
			},
			initialStats: v1.VolumeMetrics{
				Name:                 "vol1",
				ReadIOPS:             "0",
				ReplicaCounter:       6,
				RevisionCounter:      100,
				SectorSize:           "4096",
				Size:                 "1073741824",
				TotalReadBlockCount:  "3",
				TotalReadTime:        "10",
				TotalWriteTime:       "15",
				TotalWriteBlockCount: "10",
				UpTime:               13162.971420756,
				UsedBlocks:           "1048576",
				UsedLogicalBlocks:    "1048576",
				WriteIOPS:            "15",
			},
			finalStats: v1.VolumeMetrics{
				Name:                 "vol1",
				ReadIOPS:             "0",
				ReplicaCounter:       6,
				RevisionCounter:      100,
				SectorSize:           "4096",
				Size:                 "1073741824",
				TotalReadBlockCount:  "4",
				TotalReadTime:        "12",
				TotalWriteTime:       "16",
				TotalWriteBlockCount: "15",
				UpTime:               13170.971420756,
				UsedBlocks:           "1048576",
				UsedLogicalBlocks:    "1048576",
				WriteIOPS:            "20",
			},
			output: nil,
		},
		"ReadIOPSIsNotZero": {
			cmdOptions: &CmdVolumeStatsOptions{
				json:    "",
				volName: "vol1",
			},
			status: []string{
				"10.10.10.10",
				"Online",
				"1",
				"10.10.10.11",
				"Online",
				"1",
			},
			initialStats: v1.VolumeMetrics{
				Name:                 "vol1",
				ReadIOPS:             "0",
				ReplicaCounter:       6,
				RevisionCounter:      100,
				SectorSize:           "4096",
				Size:                 "1073741824",
				TotalReadBlockCount:  "3",
				TotalReadTime:        "10",
				TotalWriteTime:       "15",
				TotalWriteBlockCount: "10",
				UpTime:               13162.971420756,
				UsedBlocks:           "1048576",
				UsedLogicalBlocks:    "1048576",
				WriteIOPS:            "15",
			},
			finalStats: v1.VolumeMetrics{
				Name:                 "vol1",
				ReadIOPS:             "10",
				ReplicaCounter:       6,
				RevisionCounter:      100,
				SectorSize:           "4096",
				Size:                 "1073741824",
				TotalReadBlockCount:  "4",
				TotalReadTime:        "12",
				TotalWriteTime:       "16",
				TotalWriteBlockCount: "15",
				UpTime:               13170.971420756,
				UsedBlocks:           "1048576",
				UsedLogicalBlocks:    "1048576",
				WriteIOPS:            "20",
			},
			output: nil,
		},
		"WriteIOPSIsZero": {
			cmdOptions: &CmdVolumeStatsOptions{
				json:    "",
				volName: "vol1",
			},
			status: []string{
				"10.10.10.10",
				"Online",
				"1",
				"10.10.10.11",
				"Online",
				"1",
			},
			initialStats: v1.VolumeMetrics{
				Name:                 "vol1",
				ReadIOPS:             "0",
				ReplicaCounter:       6,
				RevisionCounter:      100,
				SectorSize:           "4096",
				Size:                 "1073741824",
				TotalReadBlockCount:  "3",
				TotalReadTime:        "10",
				TotalWriteTime:       "15",
				TotalWriteBlockCount: "10",
				UpTime:               13162.971420756,
				UsedBlocks:           "1048576",
				UsedLogicalBlocks:    "1048576",
				WriteIOPS:            "15",
			},
			finalStats: v1.VolumeMetrics{
				Name:                 "vol1",
				ReadIOPS:             "10",
				ReplicaCounter:       6,
				RevisionCounter:      100,
				SectorSize:           "4096",
				Size:                 "1073741824",
				TotalReadBlockCount:  "4",
				TotalReadTime:        "12",
				TotalWriteTime:       "16",
				TotalWriteBlockCount: "15",
				UpTime:               13170.971420756,
				UsedBlocks:           "1048576",
				UsedLogicalBlocks:    "1048576",
				WriteIOPS:            "15",
			},
			output: nil,
		},
	}
	for name, tt := range validStats {
		t.Run(name, func(t *testing.T) {
			if got := annotation.DisplayStats(tt.cmdOptions, tt.status, tt.initialStats, tt.finalStats); got != tt.output {
				t.Fatalf("DisplayStats(%v, %v, %v, %v) => %v, want %v", tt.cmdOptions, tt.status, tt.initialStats, tt.finalStats, got, tt.output)
			}
		})
	}

}
