package mapiserver

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"sort"
	"time"
)

type requestType string

const (
	// M-apiserver request types
	get    requestType = "GET"
	post   requestType = "POST"
	delete requestType = "DELETE"

	// volume request constants
	volumePath           = "/latest/volumes/"
	volumeRequestTimeout = 5 * time.Second
)

// MAPIAddr stores address of mapi server if passed through flag
var MAPIAddr string

// MAPIAddrPort stores port number of mapi server if passed through flag
var MAPIAddrPort string

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
	if len(MAPIAddr) > 0 {
		return "http://" + MAPIAddr + ":" + MAPIAddrPort
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

// getDefaultAddr returns the local ip address
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

// serverRequest is a request function to perform various request operations like GET,POST,DELETE to mapi service
func serverRequest(method requestType, payLoad []byte, url, namespace string) ([]byte, int, error) {
	if len(url) == 0 {
		return nil, http.StatusInternalServerError, errors.New("Invalid URL")
	}

	req, err := http.NewRequest(string(method), url, bytes.NewBuffer(payLoad))
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	if method == post {
		req.Header.Add("Content-Type", "application/json")
	}

	if len(namespace) > 0 {
		req.Header.Set("namespace", namespace)
	}

	c := &http.Client{
		Timeout: volumeRequestTimeout,
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return body, resp.StatusCode, nil
}
