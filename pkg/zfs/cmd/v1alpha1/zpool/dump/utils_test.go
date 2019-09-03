// Copyright © 2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pstatus

import (
	"testing"
)

func TestGetDiskStripPath(t *testing.T) {
	tests := map[string]struct {
		devPath, stripPath string
	}{
		"Test 1": {
			devPath:   "/dev/disk/by-id/scsi-0Google_PersistentDisk_persistent-disk-0-part23",
			stripPath: "/dev/disk/by-id/scsi-0Google_PersistentDisk_persistent-disk-0",
		},
		"Test 2": {
			devPath:   "/dev/disk/by-id/scsi-0Google_PersistentDisk_persistent-disk-0p12k",
			stripPath: "/dev/disk/by-id/scsi-0Google_PersistentDisk_persistent-disk-0p12k",
		},
		"Test 3": {
			devPath:   "/dev/xvdlps3",
			stripPath: "/dev/xvdlps",
		},
		"Test 4": {
			devPath:   "/dev/xvdlps3k",
			stripPath: "/dev/xvdlps3k",
		},
		"Test 5": {
			devPath:   "/dev/hdvdas2",
			stripPath: "/dev/hdvdas",
		},
		"Test 6": {
			devPath:   "/dev/hdvdas2k",
			stripPath: "/dev/hdvdas2k",
		},
		"Test 7": {
			devPath:   "/dev/sda1",
			stripPath: "/dev/sda",
		},
		"Test 8": {
			devPath:   "/dev/disk/by-id/scsi-0Google_PersistentDisk_persistent-disk-0p21k",
			stripPath: "/dev/disk/by-id/scsi-0Google_PersistentDisk_persistent-disk-0p21k",
		},
		"Test 9": {
			devPath:   "/dev/disk/by-id/scsi-0Google_PersistentDisk_persistent-disk-0p21",
			stripPath: "/dev/disk/by-id/scsi-0Google_PersistentDisk_persistent-disk-0",
		},
		"Test 10": {
			devPath:   "/dev/xvdlps32",
			stripPath: "/dev/xvdlps",
		},
	}

	for k, v := range tests {
		v := v
		k := k
		t.Run(k, func(t *testing.T) {
			stripPath := getDiskStripPath(v.devPath)
			if stripPath != v.stripPath {
				t.Fatalf("%v failed, Expected %v but got %v for devPath:%v", k, v.stripPath, stripPath, v.devPath)
			}
		})
	}
}
