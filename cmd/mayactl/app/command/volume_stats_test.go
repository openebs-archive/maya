package command

import (
	"testing"

	"github.com/openebs/maya/types/v1"
)

func TestDisplayStats(t *testing.T) {
	var validStats = []struct {
		cmdOptions   *CmdVolumeStatsOptions
		status       []string
		initialStats v1.VolumeMetrics
		finalStats   v1.VolumeMetrics
		output       error
		replicaCount int
	}{
		{
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
			replicaCount: 2,
			output:       nil,
		},
		{
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
			replicaCount: 2,
			output:       nil,
		},
	}
	for _, c := range validStats {
		annotation := &Annotations{
			TargetPortal:     "10.99.73.74:3260",
			ClusterIP:        "10.99.73.74",
			Iqn:              "iqn.2016-09.com.openebs.jiva:vol1",
			ReplicaCount:     "2",
			ControllerStatus: "Running",
			ReplicaStatus:    "",
			VolSize:          "1G",
			ControllerIP:     "",
			Replicas:         "10.10.10.10,10.10.10.11",
		}
		out := annotation.DisplayStats(c.cmdOptions, c.status, c.initialStats, c.finalStats, c.replicaCount)

		t.Logf("Command Options passed is %v \n, status passed is %v\n, initial and final stats passed is %v \n,replica Count passed is %v\n, expected output %v\n, got %v\n", c.cmdOptions, c.status, c.finalStats, c.initialStats, c.replicaCount, c.output, out)
		if out != c.output {
			t.Errorf("DisplayStats(%v, %v, %v, %v, %v) => %v, expected output %v", c.cmdOptions, c.status, c.initialStats, c.finalStats, c.replicaCount, out, c.output)
		}
	}
}
