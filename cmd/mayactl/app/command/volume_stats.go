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

package command

import (
	"fmt"
	"time"

	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/client/mapiserver"
	v1 "github.com/openebs/maya/types/v1"

	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"
)

var (
	volumeStatsCommandHelpText = `
This command queries the statisics of a volume.

Usage: mayactl volume stats --volname <vol> [-size <size>]
`
)

// statsTemplate is used for formatting the stats output
const statsTemplate = ` 
Portal Details :
---------------
Volume  :   {{.Volume}}
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

// convertMappedResponse converts array of metrics to map[string]MetricsFamily
func convertMappedResponse(rawMetrics v1alpha1.VolumeMetricsList) map[string]v1alpha1.MetricsFamily {
	newMetrics := make(map[string]v1alpha1.MetricsFamily)
	for _, metric := range rawMetrics {
		if len(metric.Metric) == 0 {
			newMetrics[metric.Name] = v1alpha1.MetricsFamily{}
		} else {
			newMetrics[metric.Name] = metric.Metric[0]
		}
	}
	return newMetrics
}

func (c *CmdVolumeOptions) runVolumeStats(cmd *cobra.Command) error {
	rawStatsInitial, err := mapiserver.VolumeStats(c.volName, c.namespace)
	if err != nil {
		return fmt.Errorf("Volume not found")
	}
	time.Sleep(time.Second)
	rawStatsFinal, err := mapiserver.VolumeStats(c.volName, c.namespace)
	if err != nil {
		return fmt.Errorf("Volume not found")
	}

	stats := processStats(convertMappedResponse(rawStatsInitial), convertMappedResponse(rawStatsFinal))

	return print(statsTemplate, stats)
}

// processStats calculates the figures from the final and initial response.
func processStats(statsInitial, statsFinal map[string]v1alpha1.MetricsFamily) (stats v1alpha1.StatsJSON) {

	// Calculate Read stats
	stats.ReadIOPS, _ = v1.SubstractInt64(int64(getValue("openebs_reads", statsFinal)), int64(getValue("openebs_reads", statsInitial)))
	rTimePS, _ := v1.SubstractFloat64(getValue("openebs_read_time", statsFinal), getValue("openebs_read_time", statsInitial))
	stats.ReadThroughput, _ = v1.SubstractFloat64(getValue("openebs_read_block_count", statsFinal), getValue("openebs_read_block_count", statsInitial))
	stats.ReadLatency, _ = v1.DivideFloat64(rTimePS, float64(stats.ReadIOPS))

	// Convert from nanosec to milliseconds
	stats.ReadLatency = stats.ReadLatency / v1.MicSec
	stats.AvgReadBlockSize, _ = v1.DivideInt64(int64(stats.ReadThroughput), stats.ReadIOPS)
	stats.ReadThroughput = stats.ReadThroughput / v1.BytesToMB
	stats.AvgReadBlockSize = stats.AvgReadBlockSize / v1.BytesToKB

	// Calculate Write stats
	stats.WriteIOPS, _ = v1.SubstractInt64(int64(getValue("openebs_writes", statsFinal)), int64(getValue("openebs_writes", statsInitial)))
	wTimePS, _ := v1.SubstractFloat64(getValue("openebs_write_time", statsFinal), getValue("openebs_write_time", statsInitial))
	stats.WriteThroughput, _ = v1.SubstractFloat64(getValue("openebs_write_block_count", statsFinal), getValue("openebs_write_block_count", statsInitial))
	stats.WriteLatency, _ = v1.DivideFloat64(wTimePS, float64(stats.WriteIOPS))
	// Convert from nanosec to milliseconds
	stats.WriteLatency = stats.WriteLatency / v1.MicSec
	stats.AvgWriteBlockSize, _ = v1.DivideInt64(int64(stats.WriteThroughput), stats.WriteIOPS)
	stats.WriteThroughput = stats.WriteThroughput / v1.BytesToMB
	stats.AvgWriteBlockSize = stats.AvgWriteBlockSize / v1.BytesToKB

	stats.SectorSize = getValue("openebs_sector_size", statsFinal)
	stats.LogicalSize = getValue("openebs_logical_size", statsFinal)
	stats.ActualUsed = getValue("openebs_actual_used", statsFinal)

	stats.Size = fmt.Sprintf("%f", getValue("openebs_size_of_volume", statsFinal))

	if val, p := statsFinal["openebs_volume_uptime"]; p {
		for _, v := range val.Label {
			if v.Name == "volName" {
				stats.Volume = v.Value
			} else if v.Name == "castype" {
				stats.CASType = v.Value
			}
		}
	}
	return stats
}

// getValue returns the value of the key if the key is present in map[string]MetricsFamily.
func getValue(key string, m map[string]v1alpha1.MetricsFamily) float64 {
	if val, p := m[key]; p {
		return val.Gauge.Value
	}
	return 0
}
