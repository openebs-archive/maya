/*
Copyright 2017 The OpenEBS Authors.

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

package volume

import (
	"fmt"
	"time"

	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/client/mapiserver"
	"github.com/openebs/maya/types/v1"

	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"
)

var (
	volumeStatsCommandHelpText = `
This command queries the statisics of a volume.

Usage: mayactl volume stats --volname <vol> [-size <size>]
`
)

const statsTemplate = ` 
Portal Details :
---------------
IQN     :   {{.IQN}}
Volume  :   {{.Volume}}
Portal  :   {{.Portal}}
Size    :   {{.Size}}

Performance Stats :
--------------------
{{ printf "r/s\t w/s\t r(MB/s)\t w(MB/s)\t rLat(ms)\t wLat(ms)" }}
{{ printf "----\t ----\t --------\t --------\t ---------\t ---------" }}
{{ printf "%d\t" .ReadIOPS }} {{ printf "%d\t" .WriteIOPS }} {{ printf "%.3f\t" .ReadThroughput }} {{ printf "%.3f\t" .WriteThroughput }} {{ printf "%.3f\t" .ReadLatency }} {{printf "%.3f\t" .WriteLatency }}

Capacity Stats :
---------------
{{ printf "LOGICAL(GB)\t USED(GB)" }}
{{ printf "------------\t ---------" }}
{{ printf "%.3f\t" .LogicalSize }} {{ printf "%.3f\t" .ActualUsed }}
`

// ReplicaStats keep info about the replicas.
type ReplicaStats struct {
	Replica         string
	Status          string
	DataUpdateIndex string
}

// NewCmdVolumeStats displays the runtime statistics of volume
func NewCmdVolumeStats() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "stats",
		Short:   "Displays the runtime statisics of Volume",
		Long:    volumeStatsCommandHelpText,
		Example: ` mayactl volume stats --volname=vol`,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(options.Validate(cmd, false, false, true), util.Fatal)
			util.CheckErr(options.runVolumeStats(cmd), util.Fatal)
		},
	}

	cmd.Flags().StringVarP(&options.volName, "volname", "", options.volName,
		"unique volume name.")
	return cmd
}

func (c *CmdVolumeOptions) runVolumeStats(cmd *cobra.Command) error {
	statsi, err := mapiserver.VolumeStats(c.volName, c.namespace)
	if err != nil {
		return fmt.Errorf("Volume not found")
	}
	time.Sleep(time.Second)
	statsf, err := mapiserver.VolumeStats(c.volName, c.namespace)
	if err != nil {
		return fmt.Errorf("Volume not found")
	}

	stats, err := processStats(statsi, statsf)
	if err != nil {
		return err
	}

	return print(statsTemplate, stats)
}

func processStats(stats1, stats2 v1alpha1.VolumeMetrics) (stats v1alpha1.StatsJSON, err error) {
	if len(stats1) != len(stats2) {
		return stats, fmt.Errorf("Invalid Response")
	}

	statsi, statsf := make(map[string]v1alpha1.MetricsFamily), make(map[string]v1alpha1.MetricsFamily)
	for i := 0; i < len(stats1); i++ {
		if len(stats1[i].Metric) == 0 {
			statsi[stats1[i].Name] = v1alpha1.MetricsFamily{}
		} else {
			statsi[stats1[i].Name] = stats1[i].Metric[0]
		}

		if len(stats2[i].Metric) == 0 {
			statsf[stats2[i].Name] = v1alpha1.MetricsFamily{}
		} else {
			statsf[stats2[i].Name] = stats2[i].Metric[0]
		}
	}

	// Calculate Read stats
	stats.ReadIOPS = int64(getValue("openebs_read", statsf) - getValue("openebs_read", statsi))
	rTimePS := getValue("openebs_read_time", statsf) - getValue("openebs_read_time", statsi)
	stats.ReadThroughput = getValue("openebs_read_block_count", statsf) - getValue("openebs_read_block_count", statsi)
	stats.ReadLatency, _ = v1.DivideFloat64(rTimePS, float64(stats.ReadIOPS))
	stats.AvgReadBlockSize, _ = v1.DivideInt64(int64(stats.ReadThroughput), stats.ReadIOPS)
	stats.AvgReadBlockSize = stats.AvgReadBlockSize / v1.BytesToKB
	stats.ReadThroughput = stats.ReadLatency / v1.BytesToMB

	// Calculate Write stats
	stats.WriteIOPS = int64(getValue("openebs_write", statsf) - getValue("openebs_write", statsi))
	wTimePS := getValue("openebs_write_time", statsf) - getValue("openebs_write_time", statsi)
	stats.WriteThroughput = getValue("openebs_write_block_count", statsf) - getValue("openebs_write_block_count", statsi)
	stats.WriteLatency, _ = v1.DivideFloat64(wTimePS, float64(stats.WriteIOPS))
	stats.AvgWriteBlockSize, _ = v1.DivideInt64(int64(stats.WriteThroughput), stats.WriteIOPS)
	stats.AvgWriteBlockSize = stats.AvgWriteBlockSize / v1.BytesToKB
	stats.WriteThroughput = stats.WriteLatency / v1.BytesToMB

	stats.SectorSize = getValue("openebs_sector_size", statsf)
	stats.LogicalSize = (getValue("openebs_logical_size", statsf) * stats.SectorSize) / v1.BytesToGB
	stats.ActualUsed = (getValue("openebs_actual_used", statsf) * stats.SectorSize) / v1.BytesToGB

	stats.Size = fmt.Sprintf("%f", getValue("openebs_size_of_volume", statsf))

	if val, p := statsf["openebs_volume_uptime"]; p {
		for _, v := range val.Label {
			if v.Name == "iqn" {
				stats.IQN = v.Value
			} else if v.Name == "portal" {
				stats.Portal = v.Value
			} else if v.Name == "volName" {
				stats.Volume = v.Value
			} else if v.Name == "castype" {
				stats.CASType = v.Value
			}
		}
	}

	return stats, nil
}

func getValue(key string, m map[string]v1alpha1.MetricsFamily) float64 {
	if val, p := m[key]; p {
		return val.Gauge.Value
	}
	return 0
}
