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
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"text/tabwriter"
	"time"

	client "github.com/openebs/maya/pkg/client/jiva"
	"github.com/openebs/maya/types/v1"
	yaml "gopkg.in/yaml.v2"
)

const (
	http_timeout = 5 * time.Second
)

// SnapshotInfo stores the details of snapshot
type SnapshotInfo struct {
	Name     string
	Created  string
	Size     string
	Parent   string
	Children string
}

// CreateSnapshot creates a snapshot of volume by API request to m-apiserver
func CreateSnapshot(volName string, snapName string, namespace string) error {
	_, err := GetStatus()
	if err != nil {
		return err
	}

	var snap v1.SnapshotAPISpec

	snap.Kind = "VolumeSnapshot"
	snap.APIVersion = "v1"
	snap.Metadata.Name = snapName
	snap.Spec.VolumeName = volName

	//Marshal serializes the value provided into a YAML document
	yamlValue, _ := yaml.Marshal(snap)

	url := GetURL() + "/latest/snapshots/create/"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(yamlValue))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/yaml")
	req.Header.Set("namespace", namespace)

	c := &http.Client{
		Timeout: http_timeout,
	}
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	code := resp.StatusCode
	if err == nil && code != http.StatusOK {
		return fmt.Errorf(string(body))
	}
	if code != http.StatusOK {
		return fmt.Errorf("Server status error: %v ", http.StatusText(code))
	}
	return nil
}

// RevertSnapshot reverts a snapshot of volume by API request to m-apiserver
func RevertSnapshot(volName string, snapName string, namespace string) error {

	_, err := GetStatus()
	if err != nil {
		return err
	}

	var snap v1.SnapshotAPISpec

	snap.Kind = "VolumeSnapshot"
	snap.APIVersion = "v1"
	snap.Metadata.Name = snapName
	snap.Spec.VolumeName = volName

	//Marshal serializes the value provided into a YAML document
	yamlValue, _ := yaml.Marshal(snap)

	url := GetURL() + "/latest/snapshots/revert/"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(yamlValue))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/yaml")
	req.Header.Set("namespace", namespace)
	c := &http.Client{
		Timeout: http_timeout,
	}
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	code := resp.StatusCode

	if code != http.StatusOK {
		return fmt.Errorf("Server status error: %v", http.StatusText(code))
	}
	return nil
}

// ListSnapshot lists snapshots of volume by API request to m-apiserver
func ListSnapshot(volName string, namespace string) error {

	_, err := GetStatus()
	if err != nil {
		return err
	}

	url := GetURL() + "/latest/snapshots/list/" + volName

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("namespace", namespace)

	c := &http.Client{
		Timeout: timeoutVolumeDelete,
	}
	resp, err := c.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	code := resp.StatusCode
	if code != http.StatusOK {
		return fmt.Errorf("Server status error: %v", http.StatusText(code))
	}
	snapdisk, err := getInfo(body)
	if err != nil {
		return fmt.Errorf("Failed to get the snapshot info, found error - %v", err)
	}

	if len(snapdisk) == 1 || len(snapdisk) == 0 {
		fmt.Println("No snapshots available. \nUse `mayactl snapshot create --volname <vol-name> --snapname <snap-name>` to create snapshot")
		return nil
	}

	snapshotList := make(map[int]*SnapshotInfo)
	var i int

	for _, disk := range snapdisk {
		if !client.IsHeadDisk(disk.Name) {
			if len(disk.Children) > 1 {
				for _, value := range disk.Children {
					snapshotList[i] = &SnapshotInfo{
						Name:     client.TrimSnapshotName(disk.Name),
						Created:  disk.Created,
						Size:     disk.Size,
						Parent:   client.TrimSnapshotName(disk.Parent),
						Children: client.TrimSnapshotName(value),
					}
					i++
				}
			} else if len(disk.Children) == 0 {
				snapshotList[i] = &SnapshotInfo{
					Name:     client.TrimSnapshotName(disk.Name),
					Created:  disk.Created,
					Size:     disk.Size,
					Parent:   client.TrimSnapshotName(disk.Parent),
					Children: "NA",
				}
				i++
			} else {
				snapshotList[i] = &SnapshotInfo{
					Name:     client.TrimSnapshotName(disk.Name),
					Created:  disk.Created,
					Size:     disk.Size,
					Parent:   client.TrimSnapshotName(disk.Parent),
					Children: client.TrimSnapshotName(disk.Children[0]),
				}
				i++
			}
		}
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

func displayVolumeSnapshot(snapshotList map[int]*SnapshotInfo) error {
	const (
		snapshotTemplate = `
Snapshot Details: 
------------------
{{ printf "NAME\t CREATED AT\t SIZE\t PARENT\t CHILDREN" }}
{{ printf "-----\t -----------\t -----\t -------\t ---------" }} {{range $key, $value := .}}
{{ printf "%s\t" $value.Name }} {{ printf "%s\t" $value.Created }} {{ printf "%s\t" $value.Size }} {{ printf "%s\t" $value.Parent }} {{ printf "%s\t" $value.Children }} {{ end }}
`
	)

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
