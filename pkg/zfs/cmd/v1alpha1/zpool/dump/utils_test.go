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
