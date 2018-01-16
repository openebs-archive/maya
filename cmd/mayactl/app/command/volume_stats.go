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
	"errors"
	"fmt"
	"html/template"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	client "github.com/openebs/maya/pkg/client/jiva"
	"github.com/openebs/maya/pkg/util"
	"github.com/openebs/maya/types/v1"
	"github.com/spf13/cobra"
)

var (
	volumeStatsCommandHelpText = `
	Usage: maya volume stats <vol> [-size <size>]

	This command queries the stats of the volume.

	Volume stats options:
	-json
	Displays the stats in json format.

	`
)

// CmdVolumeStatsOptions is used to store the value of flags used in the cli
type CmdVolumeStatsOptions struct {
	json    string
	volName string
}

// NewCmdVolumeCreate creates a new OpenEBS Volume
func NewCmdVolumeStats() *cobra.Command {
	options := CmdVolumeStatsOptions{}

	cmd := &cobra.Command{
		Use:     "stats",
		Short:   "Displays the runtime statisics of Volume",
		Long:    volumeStatsCommandHelpText,
		Example: ` maya volume stats --volname=vol -j=json`,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(options.Validate(cmd), util.Fatal)
			util.CheckErr(options.RunVolumeStats(cmd), util.Fatal)
		},
	}

	cmd.Flags().StringVarP(&options.volName, "volname", "n", options.volName,
		"unique volume name.")
	cmd.Flags().StringVarP(&options.json, "json", "j", options.json, "display output in JSON.")
	return cmd
}

// Validate varifies whether a volume name has been provided or not followed by
// stats command, it returns nil and proceeds to execute the command if there is
// no error and returns the error if it is missing.
func (c *CmdVolumeStatsOptions) Validate(cmd *cobra.Command) error {
	if c.volName == "" {
		return errors.New("--volname is missing. Please specify a volume name created")
	}
	return nil
}

// RunVolumeStats runs stats command and display the outputs in standard
// I/O or in json format.
func (c *CmdVolumeStatsOptions) RunVolumeStats(cmd *cobra.Command) error {
	fmt.Println("Executing volume stats...")

	var (
		err, err1, err3 error
		err2, err4      int
		status          v1.VolStatus
		stats1, stats2  v1.VolumeMetrics
		repStatus       string
		statusArray     []string //keeps track of the replica's status such as IP, Status and Revision counter.
	)

	annotation := &Annotations{}
	err = annotation.GetVolAnnotations(c.volName)
	if err != nil {
		fmt.Println("Can't get annotation, found error ", err)
	}

	if err != nil || annotation == nil {
		fmt.Println(err)
		return nil
	}

	if annotation.ControllerStatus != "Running" {
		fmt.Println("Volume not reachable")
		return nil
	}

	//replicaCount := 0
	replicaStatus := strings.Split(annotation.ReplicaStatus, ",")
	for _, repStatus = range replicaStatus {
		if repStatus == "Pending" {
			statusArray = append(statusArray, "Unknown")
			statusArray = append(statusArray, "Unknown")
			statusArray = append(statusArray, "Unknown")
		}
	}

	replicas := strings.Split(annotation.Replicas, ",")
	for _, replica := range replicas {
		replicaClient := client.ReplicaClient{}
		errCode1, err := replicaClient.GetVolumeStats(replica+":9502", &status)
		if err != nil {
			if errCode1 == 500 || strings.Contains(err.Error(), "EOF") {
				statusArray = append(statusArray, replica)
				statusArray = append(statusArray, "waiting")
				statusArray = append(statusArray, "Unknown")
			} else {
				statusArray = append(statusArray, replica)
				statusArray = append(statusArray, "Offline")
				statusArray = append(statusArray, "Unknown")
			}
		} else {
			statusArray = append(statusArray, replica)
			statusArray = append(statusArray, "Online")
			statusArray = append(statusArray, status.RevisionCounter)
		}
	}

	controllerClient := client.ControllerClient{}
	err2, err1 = controllerClient.GetVolumeStats(annotation.ClusterIP+":9501", &stats1)
	if err1 != nil {
		if (err2 == 500) || (err2 == 503) || err1 != nil {
			fmt.Println("Volume not Reachable\n", err1)
			return nil
		}
	} else {
		time.Sleep(1 * time.Second)
		err4, err3 = controllerClient.GetVolumeStats(annotation.ClusterIP+":9501", &stats2)
		if err3 != nil {
			if err4 == 500 || err4 == 503 || err3 != nil {
				fmt.Println("Volume not Reachable\n", err3)
				return nil
			}
		} else {
			err := annotation.DisplayStats(c, statusArray, stats1, stats2)
			if err != nil {
				fmt.Println("Can't display stats\n", err)
				return nil
			}
		}
	}
	return nil
}

// DisplayStats displays the volume stats as standard output and in json format.
// By defaault it displays in standard output but if  flag json is passed it
// displays stats in json format.
func (a *Annotations) DisplayStats(c *CmdVolumeStatsOptions, statusArray []string, stats1 v1.VolumeMetrics, stats2 v1.VolumeMetrics) error {

	var (
		err                  error
		ReadLatency          int64
		WriteLatency         int64
		AvgReadBlockCountPS  int64
		AvgWriteBlockCountPS int64
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
		IQN:    a.Iqn,
		Volume: c.volName,
		Portal: a.TargetPortal,
		Size:   a.VolSize,
	}

	if c.json == "json" {

		stat1 := v1.StatsJSON{

			IQN:    a.Iqn,
			Volume: c.volName,
			Portal: a.TargetPortal,
			Size:   a.VolSize,

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

		data, err := json.MarshalIndent(stat1, "", "\t")
		if err != nil {
			fmt.Println("Can't Marshal the data ", err)
		}

		os.Stdout.Write(data)
		fmt.Println()

	} else {

		// Printing using template
		tmpl, err1 := template.New("test").Parse("IQN     : {{.IQN}}\nVolume  : {{.Volume}}\nPortal  : {{.Portal}}\nSize    : {{.Size}}")
		err = err1
		if err != nil {
			fmt.Println("Can't Parse the template ", err)
		}
		err = tmpl.Execute(os.Stdout, annotation)
		if err != nil {
			fmt.Println("Can't execute the template ", err)
		}

		replicaCount, err := strconv.Atoi(a.ReplicaCount)
		if err != nil {
			fmt.Println("Can't convert to int, found error", err)
		}
		// Printing in tabular form
		q := tabwriter.NewWriter(os.Stdout, v1.MinWidth, v1.MaxWidth, v1.Padding, ' ', tabwriter.AlignRight|tabwriter.Debug)
		fmt.Fprintf(q, "\n\nReplica\tStatus\tDataUpdateIndex\t\n")
		fmt.Fprintf(q, "\t\t\t\n")
		for i := 0; i < (3 * replicaCount); i += 3 {
			fmt.Fprintf(q, "%s\t%s\t%s\t\n", statusArray[i], statusArray[i+1], statusArray[i+2])
		}
		q.Flush()

		w := tabwriter.NewWriter(os.Stdout, v1.MinWidth, v1.MaxWidth, v1.Padding, ' ', tabwriter.AlignRight|tabwriter.Debug)
		fmt.Println("\n----------- Performance Stats -----------\n")
		fmt.Fprintf(w, "r/s\tw/s\tr(MB/s)\tw(MB/s)\trLat(ms)\twLat(ms)\t\n")
		fmt.Fprintf(w, "%d\t%d\t%.3f\t%.3f\t%.3f\t%.3f\t\n", readIOPS, writeIOPS, float64(rThroughput)/v1.BytesToMB, float64(wThroughput)/v1.BytesToMB, float64(ReadLatency)/v1.MicSec, float64(WriteLatency)/v1.MicSec)
		w.Flush()

		x := tabwriter.NewWriter(os.Stdout, v1.MinWidth, v1.MaxWidth, v1.Padding, ' ', tabwriter.AlignRight|tabwriter.Debug)
		fmt.Println("\n------------ Capacity Stats -------------\n")
		fmt.Fprintf(x, "Logical(GB)\tUsed(GB)\t\n")
		fmt.Fprintf(x, "%.3f\t%.3f\t\n", logicalSize/v1.BytesToGB, actualUsed/v1.BytesToGB)
		x.Flush()
	}
	return err
}
