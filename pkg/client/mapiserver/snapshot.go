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

package mapiserver

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	client "github.com/openebs/maya/pkg/client/jiva"
	"github.com/openebs/maya/types/v1"
)

const (
	httpTimeout        = 5 * time.Second
	snapshotCreatePath = "/latest/snapshots/create/"
	snapshotRevertPath = "/latest/snapshots/revert/"
	snapshotListPath   = "/latest/snapshots/list/"
	snapshotTemplate   = `
Snapshot Details:
------------------
{{ printf "NAME\t CREATED AT\t SIZE(in MB)\t PARENT\t CHILDREN" }}
{{ printf "-----\t -----------\t ------------\t -------\t ---------" }} {{ range $key, $value := . }}
{{ printf "%s\t" $value.Name }} {{ printf "%s\t" $value.Created }} {{ printf "%s\t" $value.Size }} {{ printf "%s\t" $value.Parent }} {{ range $value.Children -}} {{printf "%s\n\t \t \t \t" . }} {{ end }} {{ end }}
`
)

// SnapshotInfo stores the details of snapshot
type SnapshotInfo struct {
	Name    string
	Created string
	Size    string
	// Parent keeps most recently saved version of current state of volume
	Parent string
	// Child keeps latest saved version of volume
	Children []string
}

// CreateSnapshot creates a snapshot of volume by API request to m-apiserver
func CreateSnapshot(volName string, snapName string, namespace string) error {
	snap := v1.VolumeSnapshot{
		TypeMeta: v1.TypeMeta{
			Kind:       "VolumeSnapshot",
			APIVersion: "v1",
		},
		Metadata: v1.ObjectMeta{
			Name: snapName,
		},
		Spec: v1.VolumeSnapshotSpec{
			VolumeName: volName,
		},
	}

	// Marshal serializes the values
	jsonValue, err := json.Marshal(snap)
	if err != nil {
		return err
	}
	_, err = postRequest(GetURL()+snapshotCreatePath, jsonValue, namespace, true)
	return err
}

// RevertSnapshot reverts a snapshot of volume by API request to m-apiserver
func RevertSnapshot(volName string, snapName string, namespace string) error {
	snap := v1.VolumeSnapshot{
		TypeMeta: v1.TypeMeta{
			Kind:       "VolumeSnapshot",
			APIVersion: "v1",
		},
		Metadata: v1.ObjectMeta{
			Name: snapName,
		},
		Spec: v1.VolumeSnapshotSpec{
			VolumeName: volName,
		},
	}

	// Marshal serializes the values
	jsonValue, err := json.Marshal(snap)
	if err != nil {
		return err
	}
	_, err = postRequest(GetURL()+snapshotRevertPath, jsonValue, namespace, false)
	return err
}

// ListSnapshot lists snapshots of volume by API request to m-apiserver
func ListSnapshot(volName string, namespace string) error {

	body, err := getRequest(GetURL()+snapshotListPath+volName, namespace, false)
	if err != nil {
		return err
	}
	snapdisk, err := getInfo(body)
	if err != nil {
		return fmt.Errorf("Failed to get the snapshot info, found error - %v", err)
	}

	if len(snapdisk) == 1 || len(snapdisk) == 0 {
		fmt.Println("No snapshots available. \nUse `mayactl snapshot create --volname <vol-name> --snapname <snap-name>` to create snapshot")
		return nil
	}

	snapshotList := make([]SnapshotInfo, len(snapdisk)-1)

	i := 0
	for _, disk := range snapdisk {
		if !client.IsHeadDisk(disk.Name) {
			size, _ := strconv.ParseFloat(disk.Size, 64)
			size = size / v1.BytesToMB
			snapshotList[i] = SnapshotInfo{
				Name:     client.TrimSnapshotName(disk.Name),
				Created:  disk.Created,
				Size:     fmt.Sprintf("%.4f", size),
				Parent:   client.TrimSnapshotName(disk.Parent),
				Children: client.TrimSnapshotNamesOfSlice(disk.Children),
			}
			i++
		}
	}
	SortSnapshotDisksByDateTime(snapshotList)
	err = ChangeDateFormatToUnixDate(snapshotList)
	if err != nil {
		return fmt.Errorf("Error changing date format to UnixDate, found error - %v", err)
	}

	err = displayVolumeSnapshot(snapshotList)
	if err != nil {
		fmt.Println("Error displaying snapshot list, found error - ", err)
		return err
	}

	return nil
}

// getInfo unmarshal http response body to DiskInfo struct
func getInfo(body []byte) (map[string]client.DiskInfo, error) {

	var s = make(map[string]client.DiskInfo)
	err := json.Unmarshal(body, &s)
	if err != nil {
		return nil, err
	}
	return s, err
}

// displayVolumeSnapshot displays the snapshot information in the snapshot template
func displayVolumeSnapshot(snapshotList []SnapshotInfo) error {

	tmpl := template.New("SnapshotList")
	tmpl = template.Must(tmpl.Parse(snapshotTemplate))

	w := tabwriter.NewWriter(os.Stdout, v1.MinWidth, v1.MaxWidth, v1.Padding, ' ', 0)
	err := tmpl.Execute(w, snapshotList)
	if err != nil {
		return err
	}
	w.Flush()

	return nil
}
