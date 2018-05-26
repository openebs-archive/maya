package collector

import (
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/openebs/maya/types/v1"
)

// NewStatsExporter returns cstor's socket connection instance.
func NewStatsExporter(conn net.Conn) *VolumeExporter {
	return &VolumeExporter{
		Conn: conn,
	}
}

// writer writes the "IOSTATS\n" command on the socket. it returns
// error if the connection is closed or not available with socket.
func (e *VolumeExporter) writer() error {
	msg := "IOSTATS\n"
	_, err := e.Conn.Write([]byte(msg))
	if err != nil {
		glog.Info("Write error:", err)
		return err
	}
	return nil
}

// reader reads the response from socket in the buffer and
// exits if the buffer contains the "IOSTATS\r\n" at the end.
func (e *VolumeExporter) reader() (string, error) {
	buf := make([]byte, 1024)
	var (
		err           error
		n             int
		str, response string
	)
	// infinite for loop to collect all the chunks.
	for {
		n, err = e.Conn.Read(buf[:])
		if err != nil {
			glog.Info("Error in reading response, found error : ", err)
			return "", err
		}
		// concatnat the chunks received and then compare if it
		// has " IOSTATS\r\n" and exit, else continue appending
		// the chunks at the end of str for further comparison.
		str = str + string(buf[0:n])
		if str[len(str)-10:] == " IOSTATS\r\n" {
			response = str
			break
		}
	}
	glog.Infof("Client Got : %#v", response)
	return response, nil
}

// split converts the response which is a string of (key, value)
// pairs and removes header. It returns the map of (key, value)
// at the end.
func split(resp string) map[string]string {
	m := make(map[string]string)
	var result []string
	result = strings.Split(resp, "\r\n")
	result = v1.Remove(result, "OK IOSTATS")
	if strings.HasPrefix(resp, "iSCSI Target") {
		result = strings.Split(result[1], " ")
	}
	if strings.HasPrefix(resp, "IOSTATS  ") {
		result = strings.Split(result[0], " ")
	}
	result = v1.Remove(result, "IOSTATS")
	result = v1.Remove(result, "")
	for _, pair := range result {
		z := strings.Split(pair, "=")
		m[z[0]] = z[1]
	}
	return m
}

// collect make call to reader and writer and parses the response
// into the respective metrics variable.
func collect(e *VolumeExporter) error {
	var (
		resp1, resp2 string
		err          error
	)
	glog.Infof("Started collecting metrics from volume")
	if err := e.writer(); err != nil {
		return err
	}
	resp1, err = e.reader()
	if err != nil {
		return err
	}
	time.Sleep(1 * time.Second)
	if err = e.writer(); err != nil {
		return err
	}
	resp2, err = e.reader()
	if err != nil {
		return err
	}
	m1 := split(resp1)
	m2 := split(resp2)
	rIOPS, _ := v1.ParseAndSubstract(m1["Reads"], m2["Reads"])
	readIOPS.Set(float64(rIOPS))
	wIOPS, _ := v1.ParseAndSubstract(m1["Writes"], m2["Writes"])
	writeIOPS.Set(float64(wIOPS))
	rThput, _ := v1.ParseAndSubstract(m1["TotalReadBytes"], m2["TotalReadBytes"])
	readBlockCountPS.Set(float64(rThput) / v1.BytesToMB)
	wThput, _ := v1.ParseAndSubstract(m1["TotalWriteBytes"], m2["TotalWriteBytes"])
	writeBlockCountPS.Set(float64(wThput) / v1.BytesToMB)
	size, _ := strconv.ParseFloat(m1["Size"], 64)
	size, _ = v1.DivideFloat64(size, v1.BytesToGB)
	sizeOfVolume.Set(size)
	return nil
}
