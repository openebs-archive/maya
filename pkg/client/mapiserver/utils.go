package mapiserver

import (
	"net"
	"os"
	"sort"
	"time"
)

// MAPIAddr stores address of mapi server if passed through flag
var MAPIAddr string

// Initialize func sets the env variable with local ip address
func Initialize() {
	mapiaddr := os.Getenv("MAPI_ADDR")
	if mapiaddr == "" {
		mapiaddr = getDefaultAddr()
		os.Setenv("MAPI_ADDR", mapiaddr)
	}
}

// GetURL returns the mapi server address
func GetURL() string {
	if MAPIAddr != "" {
		return "http://" + MAPIAddr + ":5656"
	}
	return os.Getenv("MAPI_ADDR")
}

// GetConnectionStatus return the status of the connecion
func GetConnectionStatus() string {
	_, err := GetStatus()
	if err != nil {
		return "not reachable"
	}
	return "running"
}

func getDefaultAddr() string {
	env := "127.0.0.1"
	host, _ := os.Hostname()
	addrs, _ := net.LookupIP(host)
	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			if ipv4.String() != "127.0.0.1" {
				env = ipv4.String()
				break
			}
		}
	}
	return "http://" + env + ":5656"
}

// SortSnapshotDisksByDateTime orders the snapshot disks with respect to date and time
func SortSnapshotDisksByDateTime(snapshotDisks []SnapshotInfo) {
	sort.SliceStable(snapshotDisks, func(i, j int) bool {
		t1, _ := time.Parse(time.RFC3339, snapshotDisks[i].Created)
		t2, _ := time.Parse(time.RFC3339, snapshotDisks[j].Created)
		return t1.Before(t2)
	})
}

// ChangeDateFormatToUnixDate changes the created date from RFC3339 format to UnixDate format
func ChangeDateFormatToUnixDate(snapshotDisks []SnapshotInfo) error {
	for index := range snapshotDisks {
		created, err := time.Parse(time.RFC3339, snapshotDisks[index].Created)
		if err != nil {
			return err
		}
		snapshotDisks[index].Created = created.Format(time.UnixDate)
	}
	return nil
}
