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
	"strings"
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
	Name    string
	Created string
	Size    string
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
		fmt.Println("Failed to get the snapshot info", err)
	}

	snapshotList := make(map[int]*SnapshotInfo)

	// out := make([]string, len(snapdisk)+1)

	// out[0] = "Name|Created At|Size"
	var i int

	for _, disk := range snapdisk {
		//	if !util.IsHeadDisk(disk.Name) {
		// out[i+1] = fmt.Sprintf("%s|%s|%s",
		// 	strings.TrimSuffix(strings.TrimPrefix(disk.Name, "volume-snap-"), ".img"),
		// 	disk.Created,
		// 	disk.Size)
		// i = i + 1
		snapshotList[i] = &SnapshotInfo{strings.TrimSuffix(strings.TrimPrefix(disk.Name, "volume-snap-"), ".img"), disk.Created, disk.Size}
		i++
	}
	err = displayVolumeSnapshot(snapshotList)
	if err != nil {
		fmt.Println("Error displaying snapshot list, found error - ", err)
		return err
	}

	// fmt.Println(util.FormatList(out))
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
{{ printf "NAME\t CREATED AT\t SIZE" }}
{{ printf "-----\t -----------\t -----" }} {{range $key, $value := .}}
{{ printf "%s\t" $value.Name }} {{ printf "%s\t" $value.Created }} {{ printf "%s\t" $value.Size }} {{ end }}
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
