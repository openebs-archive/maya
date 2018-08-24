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
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	client "github.com/openebs/maya/pkg/client/jiva"
	"github.com/openebs/maya/pkg/client/mapiserver"
	"github.com/openebs/maya/pkg/util"
	"github.com/openebs/maya/types/v1"
	"github.com/spf13/cobra"
)

var (
	volumeStatsCommandHelpText = `
This command queries the statisics of a volume.

Usage: mayactl volume stats --volname <vol> [-size <size>]
`
)

const (
	controllerStatusOk = "running"
)

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
		Example: ` mayactl volume stats --volname=vol -j=json`,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(options.Validate(cmd, false, false, true), util.Fatal)
			util.CheckErr(options.RunVolumeStats(cmd), util.Fatal)
		},
	}

	cmd.Flags().StringVarP(&options.volName, "volname", "", options.volName,
		"unique volume name.")
	cmd.Flags().StringVarP(&options.json, "json", "j", options.json, "display output in JSON.")
	return cmd
}

// RunVolumeStats runs stats command and display the outputs in standard
// I/O or in json format.
func (c *CmdVolumeOptions) RunVolumeStats(cmd *cobra.Command) error {
	fmt.Println("Executing volume stats...")
	var (
		status         v1.VolStatus
		stats1, stats2 v1.VolumeMetrics
	)
	volumeInfo := &v1alpha1.CASVolume{}
	// Filling the volumeInfo structure with response from mayapi server
	err := volumeInfo.FetchVolumeInfo(mapiserver.GetURL()+listVolumeAPIPath+c.volName, c.volName, c.namespace)
	if err != nil {
		return nil
	}

	controllerStatus := strings.Split(volumeInfo.GetControllerStatus(), ",")
	for i := range controllerStatus {
		if controllerStatus[i] != controllerStatusOk {
			fmt.Printf("Unable to fetch volume details, Volume controller's status is '%s'.\n", controllerStatus)
			return nil
		}
	}

	replicas := strings.Split(volumeInfo.GetReplicaIP(), ",")
	replicaStatus := strings.Split(volumeInfo.GetReplicaStatus(), ",")
	replicaStats := make(map[int]*ReplicaStats)
	for i, replica := range replicas {
		replicaClient := client.ReplicaClient{}
		respStatus, err := replicaClient.GetVolumeStats(replica+v1.ReplicaPort, &status)
		if err != nil {
			if respStatus == 500 || strings.Contains(err.Error(), "EOF") {
				replicaStats[i] = &ReplicaStats{replica, replicaStatus[i], "Unknown"}
			} else {
				replicaStats[i] = &ReplicaStats{replica, replicaStatus[i], "Unknown"}
			}
		} else {
			replicaStats[i] = &ReplicaStats{replica, replicaStatus[i], status.RevisionCounter}
		}
	}

	controllerClient := client.ControllerClient{}
	// Fetching volume stats from replica controller
	respStatus, err := controllerClient.GetVolumeStats(volumeInfo.GetClusterIP()+v1.ControllerPort, v1.StatsAPI, &stats1)
	if err != nil {
		if (respStatus == 500) || (respStatus == 503) || err != nil {
			fmt.Println("Volume not Reachable\n", err)
			return nil
		}
	} else {
		time.Sleep(1 * time.Second)
		respStatus, err := controllerClient.GetVolumeStats(volumeInfo.GetClusterIP()+v1.ControllerPort, v1.StatsAPI, &stats2)
		if err != nil {
			if respStatus == 500 || respStatus == 503 || err != nil {
				fmt.Println("Volume not Reachable\n", err)
				return nil
			}
		} else {
			err := displayStats(volumeInfo, c, replicaStats, stats1, stats2)
			if err != nil {
				fmt.Println("Can't display stats\n", err)
				return nil
			}
		}
	}
	return nil
}

// displayStats displays the volume stats as standard output and in json format.
// By default it displays in standard output format, if flag json has passed
// displays stats in json format.
func displayStats(v *v1alpha1.CASVolume, c *CmdVolumeOptions, replicaStats map[int]*ReplicaStats, stats1 v1.VolumeMetrics, stats2 v1.VolumeMetrics) error {

	var (
		ReadLatency          int64
		WriteLatency         int64
		AvgReadBlockCountPS  int64
		AvgWriteBlockCountPS int64
	)

	const (
		portalTemplate = `
Portal Details :
---------------
IQN     :   {{.IQN}}
Volume  :   {{.Volume}}
Portal  :   {{.Portal}}
Size    :   {{.Size}}

`
		replicaTemplate = `
Replica Stats :
----------------
{{ printf "REPLICA\t STATUS\t DATAUPDATEINDEX" }}
{{ printf "--------\t -------\t ----------------" }} {{range $key, $value := .}}
{{ printf "%s\t" $value.Replica }} {{ printf "%s\t" $value.Status }} {{ printf "%s\t" $value.DataUpdateIndex }} {{end}}

`
		performanceTemplate = `
Performance Stats :
--------------------
{{ printf "r/s\t w/s\t r(MB/s)\t w(MB/s)\t rLat(ms)\t wLat(ms)" }}
{{ printf "----\t ----\t --------\t --------\t ---------\t ---------" }}
{{ printf "%d\t" .ReadIOPS }} {{ printf "%d\t" .WriteIOPS }} {{ printf "%.3f\t" .ReadThroughput }} {{ printf "%.3f\t" .WriteThroughput }} {{ printf "%.3f\t" .ReadLatency }} {{printf "%.3f\t" .WriteLatency }}

`
		capicityTemplate = `
Capacity Stats :
---------------
{{ printf "LOGICAL(GB)\t USED(GB)" }}
{{ printf "------------\t ---------" }}
{{ printf "%.3f\t" .LogicalSize }} {{ printf "%.3f\t" .ActualUsed }}
`
	)

	// 10 and 64 represents decimal and bits respectively
	iReadIOPS, _ := strconv.ParseInt(stats1.ReadIOPS, 10, 64) // Initial
	fReadIOPS, _ := strconv.ParseInt(stats2.ReadIOPS, 10, 64) // Final
	readIOPS, _ := v1.SubstractInt64(fReadIOPS, iReadIOPS)

	iReadTimePS, _ := strconv.ParseInt(stats1.TotalReadTime, 10, 64)
	fReadTimePS, _ := strconv.ParseInt(stats2.TotalReadTime, 10, 64)
	readTimePS, _ := v1.SubstractInt64(fReadTimePS, iReadTimePS)

	iReadBlockCountPS, _ := strconv.ParseInt(stats1.TotalReadBlockCount, 10, 64)
	fReadBlockCountPS, _ := strconv.ParseInt(stats2.TotalReadBlockCount, 10, 64)
	readBlockCountPS, _ := v1.SubstractInt64(fReadBlockCountPS, iReadBlockCountPS)

	rThroughput := readBlockCountPS
	if readIOPS != 0 {
		ReadLatency, _ = v1.DivideInt64(readTimePS, readIOPS)
		AvgReadBlockCountPS, _ = v1.DivideInt64(readBlockCountPS, readIOPS)
	} else {
		ReadLatency = 0
		AvgReadBlockCountPS = 0
	}

	iWriteIOPS, _ := strconv.ParseInt(stats1.WriteIOPS, 10, 64)
	fWriteIOPS, _ := strconv.ParseInt(stats2.WriteIOPS, 10, 64)
	writeIOPS, _ := v1.SubstractInt64(fWriteIOPS, iWriteIOPS)

	iWriteTimePS, _ := strconv.ParseInt(stats1.TotalWriteTime, 10, 64)
	fWriteTimePS, _ := strconv.ParseInt(stats2.TotalWriteTime, 10, 64)
	writeTimePS, _ := v1.SubstractInt64(fWriteTimePS, iWriteTimePS)

	iWriteBlockCountPS, _ := strconv.ParseInt(stats1.TotalWriteBlockCount, 10, 64)
	fWriteBlockCountPS, _ := strconv.ParseInt(stats2.TotalWriteBlockCount, 10, 64)
	writeBlockCountPS, _ := v1.SubstractInt64(fWriteBlockCountPS, iWriteBlockCountPS)

	wThroughput := writeBlockCountPS
	if writeIOPS != 0 {
		WriteLatency, _ = v1.DivideInt64(writeTimePS, writeIOPS)
		AvgWriteBlockCountPS, _ = v1.DivideInt64(writeBlockCountPS, writeIOPS)
	} else {
		WriteLatency = 0
		AvgWriteBlockCountPS = 0
	}

	sectorSize, _ := strconv.ParseFloat(stats2.SectorSize, 64) // Sector Size

	logicalSize, _ := strconv.ParseFloat(stats2.UsedBlocks, 64) // Logical Size
	logicalSize = logicalSize * sectorSize

	actualUsed, _ := strconv.ParseFloat(stats2.UsedLogicalBlocks, 64) // Actual Used
	actualUsed = actualUsed * sectorSize

	annotation := v1.Annotation{
		IQN:    v.Spec.Iqn,
		Volume: c.volName,
		Portal: v.Spec.TargetPortal,
		Size:   v.Spec.Capacity,
	}

	stat1 := v1.StatsJSON{

		IQN:    v.GetIQN(),
		Volume: v.GetVolumeName(),
		Portal: v.GetTargetPortal(),
		Size:   v.GetVolumeSize(),

		ReadIOPS:  readIOPS,
		WriteIOPS: writeIOPS,

		ReadThroughput:  float64(rThroughput) / v1.BytesToMB, // bytes to MB
		WriteThroughput: float64(wThroughput) / v1.BytesToMB,

		ReadLatency:  float64(ReadLatency) / v1.MicSec, // Microsecond
		WriteLatency: float64(WriteLatency) / v1.MicSec,

		AvgReadBlockSize:  AvgReadBlockCountPS / v1.BytesToKB, // Bytes to KB
		AvgWriteBlockSize: AvgWriteBlockCountPS / v1.BytesToKB,

		SectorSize:  sectorSize,
		ActualUsed:  actualUsed / v1.BytesToGB,
		LogicalSize: logicalSize / v1.BytesToGB,
	}

	if c.json == "json" {

		data, err := json.MarshalIndent(stat1, "", "\t")
		if err != nil {
			fmt.Println("Can't Marshal the data ", err)
		}

		os.Stdout.Write(data)
		fmt.Println()

	} else {
		tmpl, err := template.New("VolumeStats").Parse(portalTemplate)
		if err != nil {
			fmt.Println("Error in parsing portal template, found error : ", err)
			return nil
		}
		err = tmpl.Execute(os.Stdout, annotation)
		if err != nil {
			fmt.Println("Error in executing portal template, found error :", err)
			return nil
		}

		replicaCount, err := strconv.Atoi(v.GetReplicaCount())
		if err != nil {
			fmt.Println("Can't convert to int, found error", err)
			return nil
		}
		// This case will occur only if user has manually specified zero replica.
		if replicaCount == 0 {
			fmt.Println("None of the replicas are running, please check the volume pod's status by running [kubectl describe pod -l=openebs/replica --all-namespaces] or try again later.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, v1.MinWidth, v1.MaxWidth, v1.Padding, ' ', 0)
		// Updating the templates
		tmpl, err = template.New("ReplicaStats").Parse(replicaTemplate)
		if err != nil {
			fmt.Println("Error in parsing replica template, found error : ", err)
			return nil
		}
		err = tmpl.Execute(w, replicaStats)
		if err != nil {
			fmt.Println("Error in executing replica template, found error :", err)
			return nil
		}
		w.Flush()

		tmpl, err = template.New("PerformanceStats").Parse(performanceTemplate)
		if err != nil {
			fmt.Println("Error in parsing performance template, found error : ", err)
			return nil
		}
		err = tmpl.Execute(w, stat1)
		if err != nil {
			fmt.Println("Error in executing performance template, found error :", err)
			return nil
		}
		w.Flush()

		tmpl, err = template.New("CapacityStats").Parse(capicityTemplate)
		if err != nil {
			fmt.Println("Error in parsing capacity template, found error : ", err)
			return nil
		}
		err = tmpl.Execute(w, stat1)
		if err != nil {
			fmt.Println("Error in executing capacity template, found error :", err)
			return nil
		}
		w.Flush()
	}
	return nil
}
