/*
Copyright 2019 The OpenEBS Authors.

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

package v1alpha2

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	ndmapis "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	zpool "github.com/openebs/maya/pkg/apis/openebs.io/zpool/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	testKey = "testJSON"
)

func testExecuteDumpCommand(cspi *apis.CStorPoolInstance) (zpool.Topology, error) {
	valuesFromEnv := strings.TrimSpace(os.Getenv(testKey))
	if valuesFromEnv == "" {
		return zpool.Topology{}, errors.Errorf("failed to read the testJSON")
	}
	topology := zpool.Topology{}
	err := json.Unmarshal([]byte(valuesFromEnv), &topology)
	if err != nil {
		return zpool.Topology{}, errors.Errorf("failed to unmarshal json output error: %v", err)
	}
	return topology, nil
}

func TestGetPathForBDevFromBlockDevice(t *testing.T) {
	testBDPaths := map[string]struct {
		bd         *ndmapis.BlockDevice
		linksCount int
	}{
		"Blockdevice with only dev links": {
			bd: &ndmapis.BlockDevice{
				Spec: ndmapis.DeviceSpec{
					DevLinks: []ndmapis.DeviceDevLink{
						ndmapis.DeviceDevLink{Kind: "one",
							Links: []string{
								"/dev/by-uid/path1",
								"/dev/by-uid/path2",
							},
						},
						ndmapis.DeviceDevLink{Kind: "two", Links: []string{"/dev/by-uid/path1"}},
					},
					Path: "/dev/sda",
				},
			},
			linksCount: 4,
		},
		"Blockdevice with only path links": {
			bd: &ndmapis.BlockDevice{
				Spec: ndmapis.DeviceSpec{
					Path: "/dev/sda",
				},
			},
			linksCount: 1,
		},
	}
	for name, test := range testBDPaths {
		// pin it
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			paths := getPathForBDevFromBlockDevice(test.bd)
			if len(paths) != test.linksCount {
				t.Errorf("test: %s failed expected device links %d but got %d",
					name,
					test.linksCount,
					len(paths),
				)
			}
		})
	}
}

func TestIsResilveringInProgress(t *testing.T) {
	tests := map[string]struct {
		jsonValues          string
		cspi                *apis.CStorPoolInstance
		path                string
		expectedResilvering bool
		executeFunc         func(cspi *apis.CStorPoolInstance) (zpool.Topology, error)
	}{
		"Resilvering In progress": {
			//In jsonValues persistentDisk_sai-disk1 is replaced with
			//persistentDisk_sai-disk3
			jsonValues: `{"name":"cstor-e2a5f5ff-10d8-11ea-bf3b-42010a8001ff","state":0,"pool_guid":5658042564790239978,"vdev_children":1,"vdev_tree":{"type":"root","id":0,"guid":5658042564790239978,"vdev_stats":[1394692727542,7,0,4287558656,10670309376,10670309376,0,0,0,99018,1522020,0,0,0,0,796413952,21988136448,0,0,0,0,0,0,0,0,0,0],"scan_stats":[2,1,1574834545,0,4266596352,723547136,0,723547136,0,723547136,1574834545,0,0],"children":[{"type":"replacing","id":0,"guid":8250149513645842791,"whole_disk":0,"vdev_stats":[302482414371,7,0,0,0,0,10722213888,0,0,67394,841082,0,0,0,0,668213248,11584425984,0,0,0,0,0,0,0,0,128688640,0],"scan_stats":[2,1,1574834545,0,4266596352,723547136,0,723547136,0,723547136,1574834545,0,0],"children":[{"type":"disk","id":0,"guid":12875869339379092712,"path":"/dev/disk/by-id/scsi-0Google_PersistentDisk_sai-disk1","devid":"scsi-0Google_PersistentDisk_sai-disk1","phys_path":"pci-0000:00:03.0-scsi-0:0:2:0","whole_disk":1,"vdev_stats":[1394692769652,7,0,0,0,0,10726932480,0,0,67387,688754,0,0,0,0,667729920,10403712512,0,0,0,0,0,0,0,0,0,0],"scan_stats":[2,1,1574834545,0,4266596352,723547136,0,723547136,0,723547136,1574834545,0,0]},{"type":"disk","id":1,"guid":12895321627809657814,"path":"/dev/disk/by-id/scsi-0Google_PersistentDisk_sai-disk3","devid":"scsi-0Google_PersistentDisk_sai-disk3","phys_path":"pci-0000:00:03.0-scsi-0:0:4:0","whole_disk":1,"DTL":236,"create_txg":4,"com.delphix:vdev_zap_leaf":235,"vdev_stats":[302656212847,7,0,0,0,0,10726932480,0,0,7,152328,0,0,0,0,483328,1180713472,0,0,0,0,0,0,0,0,723547136,0],"scan_stats":[2,1,1574834545,0,4266596352,723547136,0,723547136,0,723547136,1574834545,0,0],"resilver_txg":243}]},{"type":"disk","id":1,"guid":10525056882961272669,"path":"/dev/disk/by-id/scsi-0Google_PersistentDisk_sai-disk4","devid":"scsi-0Google_PersistentDisk_sai-disk4","phys_path":"pci-0000:00:03.0-scsi-0:0:5:0","whole_disk":1,"vdev_stats":[1394692809521,7,0,0,0,0,10726932480,0,0,31624,680938,0,0,0,0,128200704,10403710464,0,0,0,0,0,0,0,0,0,0],"scan_stats":[2,1,1574834545,0,4266596352,723547136,0,723547136,0,723547136,1574834545,0,0]}]}}`,
			cspi: &apis.CStorPoolInstance{
				ObjectMeta: metav1.ObjectMeta{Name: "pool1"},
			},
			expectedResilvering: true,
			executeFunc:         testExecuteDumpCommand,
			path:                "/dev/disk/by-id/scsi-0Google_PersistentDisk_sai-disk3",
		},
		"Without any stats": {
			jsonValues: ``,
			cspi: &apis.CStorPoolInstance{
				ObjectMeta: metav1.ObjectMeta{Name: "pool1"},
			},
			expectedResilvering: true,
			executeFunc:         testExecuteDumpCommand,
			path:                "path1",
		},
		"Invalid json": {
			jsonValues: `{"name": "pool1","pool_guid":123,"vdev_stats":{]}`,
			cspi: &apis.CStorPoolInstance{
				ObjectMeta: metav1.ObjectMeta{Name: "pool1"},
			},
			expectedResilvering: true,
			executeFunc:         testExecuteDumpCommand,
			path:                "path1",
		},
		"Without Scan Stats": {
			jsonValues: `{"vdev_children":1,"vdev_tree":{"type":"root","vdev_stats":[1394692727542,7,0,4287558656,10670309376,10670309376,0,0,0,99018,1522020,0,0,0,0,796413952,21988136448,0,0,0,0,0,0,0,0,0,0],"children":[{"type":"disk","path":"/dev/disk/by-id/scsi-0Google_PersistentDisk_sai-disk4","whole_disk":1,"vdev_stats":[1394692809521,7,0,0,0,0,10726932480,0,0,31624,680938,0,0,0,0,128200704,10403710464,0,0,0,0,0,0,0,0,0,0]}]}}`,
			cspi: &apis.CStorPoolInstance{
				ObjectMeta: metav1.ObjectMeta{Name: "pool1"},
			},
			expectedResilvering: false,
			executeFunc:         testExecuteDumpCommand,
			path:                "/dev/disk/by-id/scsi-0Google_PersistentDisk_sai-disk4",
		},
		"Resilvering success": {
			jsonValues: `{"vdev_children":1,"vdev_tree":{"type":"root","vdev_stats":[1394692727542,7,0,4287558656,10670309376,10670309376,0,0,0,99018,1522020,0,0,0,0,796413952,21988136448,0,0,0,0,0,0,0,0,0,0],"scan_stats":[2,1,1574834545,0,4266596352,723547136,0,723547136,0,723547136,1574834545,0,0],"children":[{"type":"disk","path":"/dev/disk/by-id/scsi-0Google_PersistentDisk_sai-disk4","whole_disk":1,"vdev_stats":[302656212847,7,0,0,0,0,10726932480,0,0,7,152328,0,0,0,0,483328,1180713472,0,0,0,0,0,0,0,0,723547136,0],"scan_stats":[2,2,1574834545,0,4266596352,723547136,0,723547136,0,723547136,1574834545,0,0]}]}}`,
			cspi: &apis.CStorPoolInstance{
				ObjectMeta: metav1.ObjectMeta{Name: "pool1"},
			},
			expectedResilvering: false,
			executeFunc:         testExecuteDumpCommand,
			path:                "/dev/disk/by-id/scsi-0Google_PersistentDisk_sai-disk4",
		},
	}
	// Don't change test to run in parallel
	for name, test := range tests {
		// pin it
		name, test := name, test
		os.Setenv(testKey, test.jsonValues)
		isResilvering := isResilveringInProgress(test.executeFunc, test.cspi, test.path)
		if test.expectedResilvering != isResilvering {
			t.Errorf("test %s failed expected resilvering process: %t but got %t",
				name,
				test.expectedResilvering,
				isResilvering,
			)
		}
		os.Unsetenv(testKey)
	}
}
