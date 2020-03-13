// Copyright © 2017-2019 The OpenEBS Authors
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

package collector

import (
	"bytes"
	"encoding/json"
	"net"
	"strings"
	"sync"
	"time"

	v1 "github.com/openebs/maya/pkg/stats/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	"github.com/pkg/errors"
	"k8s.io/klog"
)

const (

	// SocketPath where istgt is listening
	SocketPath = "/var/run/istgt_ctl_sock"
	// HeaderPrefix is the prefix comes in the header from cstor.
	HeaderPrefix = "iSCSI Target Controller version"
	// EOF separates the strings from response which comes from the
	// cstor as the collection of metrics.
	EOF = "\r\n"
	// Footer is used to verify if all the response has collected.
	Footer = "OK IOSTATS"
	// Command is a command that is used to write over wire and get
	// the iostats from the cstor.
	Command = "IOSTATS"
	// BufSize is the size of response from cstor read at one time.
	BufSize = 1024
)

var (
	dialFunc = func(path string) (net.Conn, error) {
		return net.Dial("unix", path)
	}
)

// cstor implements the Exporter interface. It exposes
// the metrics of a OpenEBS (cstor) volume.
type cstor struct {
	sync.Mutex
	// conn is used as unix network connection
	conn       net.Conn
	socketPath string
}

// Cstor returns cstor's instance
func Cstor(path string) *cstor {
	return &cstor{
		socketPath: path,
	}
}

// initiateConnection tries to initiates the connection with the cstor
// over unix domain socket.
func (c *cstor) initiateConnection() error {
	conn, err := dialFunc(c.socketPath)
	if err != nil {
		return err
	}
	if conn != nil {
		c.conn = conn
		c.conn.SetDeadline(time.Now().Add(5 * time.Second))
		c.readHeader()
	}
	return nil
}

// writer writes the "IOSTATS\n" command on the socket. it returns
// error if the connection is closed or not available with socket.
func (c *cstor) writer() error {
	msg := Command + "\n"
	_, err := c.conn.Write([]byte(msg))
	if err != nil {
		return err
	}
	return nil
}

// reader reads the response from socket in the buffer and
// exits if the buffer contains the "IOSTATS\r\n" at the end.
func (c *cstor) reader() (string, error) {
	buf := make([]byte, BufSize)
	var (
		err           error
		n             int
		str, response string
		buffer        bytes.Buffer
	)
	// infinite for loop to collect all the chunks.
	for {
		n, err = c.conn.Read(buf[:])
		if err != nil {
			return "", err
		}
		// concatnat the chunks received and then compare if it has
		// Footer (" IOSTATS") + EOF ("\r\n") and exit, else continue
		// appending the chunks at the end of str.
		buffer.WriteString(string(buf[0:n]))
		str = buffer.String()
		// It's possible that chunks read from socket is less than
		// 7 or 12(slice comparison), so exporter may panic, this check ensures
		// that str of atleast length 12 has been read from the socket, if not
		// it continue to read again
		if len(str) < 12 {
			continue
		}
		// confirm whether all the chunks have been collected
		// exp: "IOSTATS(Command) <json response> OK IOSTATS(Footer+EOF)\r\n"
		if str[:7] == Command && str[len(str)-12:] == Footer+EOF {
			response = str
			break
		}
	}
	return response, nil
}

// ReadHeader only reads the header of the response from cstor
func (c *cstor) readHeader() error {
	buf := make([]byte, 1024)
	var (
		err    error
		n      int
		str    string
		buffer bytes.Buffer
	)
	// collect all the chunks ending with EOF ("\r\n").
	for {
		n, err = c.conn.Read(buf[:])
		if err != nil {
			klog.Error("Error in reading response, error : ", err)
			return err
		}
		// apend the chunks into str
		buffer.WriteString(string(buf[0:n]))
		str = buffer.String()
		if strings.HasPrefix(str, HeaderPrefix) && strings.HasSuffix(str, EOF) {
			break
		}
	}
	klog.V(2).Infof("Connection established with istgt, got header: %#v", str)
	return nil
}

// unmarshal the result into VolumeStats instances.
func (c *cstor) unmarshal(result string) (v1.VolumeStats, error) {
	metrics := v1.VolumeStats{}
	if err := json.Unmarshal([]byte(result), &metrics); err != nil {
		klog.Error("Error in unmarshalling, found error: ", err)
		return metrics, err
	}
	return metrics, nil
}

// splitter extracts the JSON from the response :
// "IOSTATS  { \"iqn\": \"iqn.2017-08.OpenEBS.cstor:vol1\",
//	\"Writes\": \"0\", \"Reads\": \"0\", \"TotalWriteBytes\": \"0\",
//  \"TotalReadBytes\": \"0\", \"Size\": \"10737418240\" }\r\nOK IOSTATS\r\n"
func (c *cstor) splitter(resp string) string {
	var result []string
	result = strings.Split(resp, EOF)
	result = util.RemoveItemFromSlice(result, Footer)
	if len(result[0]) == 0 {
		return ""
	}
	res := strings.TrimPrefix(result[0], Command+"  ")
	return res
}

// getter writes query on the socket for the getting stats
// if the connection is available else retry to initiate
// connection again.
func (c *cstor) get() (v1.VolumeStats, error) {
	// locking ensures only one request is being processed
	// at a time and hence ensure that there is no fd leak.
	// because if we create a new connection for each request
	// there will be fd leak.
	c.Lock()
	defer c.Unlock()
	var (
		err   error
		stats v1.VolumeStats
	)

	klog.V(2).Info("Initiate connection")
	if err := c.initiateConnection(); err != nil {
		return v1.VolumeStats{}, &colErr{
			errors.Wrap(err, "Can't connect to istgt"),
		}
	}
	defer c.close()

	klog.V(2).Info("Request istgt to get volume stats")
	c.conn.SetDeadline(time.Now().Add(5 * time.Second))

	if err := c.writer(); err != nil {
		return v1.VolumeStats{}, &colErr{
			errors.Errorf("%v: closing connection", err),
		}
	}

	klog.V(2).Info("Read response from istgt")
	c.conn.SetDeadline(time.Now().Add(5 * time.Second))

	buf, err := c.reader()
	if err != nil {
		klog.Errorf("Got response: %v", buf)
		return v1.VolumeStats{}, &colErr{
			errors.Errorf("%v: closing connection", err),
		}
	}

	resp := c.splitter(buf)
	if len(resp) == 0 {
		return v1.VolumeStats{}, errors.New("Got empty response from cstor")
	}

	// unmarshal the json response into metrics instances.
	if stats, err = c.unmarshal(resp); err != nil {
		klog.Errorf("Got response: %v", buf)
		return v1.VolumeStats{}, err
	}

	stats.Got = true
	return stats, nil
}

// parse can used to parse the json strings into the respective types.
func (c *cstor) parse(volStats v1.VolumeStats, metrics *metrics) stats {
	var stats = stats{}
	if !volStats.Got {
		klog.Warningf("%s", "can't parse, got empty stats, istgt may not be reachable")
		return stats
	}
	stats.got = true
	stats.casType = "cstor"
	stats.reads = parseFloat64(volStats.Reads, metrics)
	stats.writes = parseFloat64(volStats.Writes, metrics)
	stats.totalReadBytes = parseFloat64(volStats.TotalReadBytes, metrics)
	stats.totalWriteBytes = parseFloat64(volStats.TotalWriteBytes, metrics)
	stats.sectorSize = parseFloat64(volStats.SectorSize, metrics)
	stats.totalReadTime = parseFloat64(volStats.TotalReadTime, metrics)
	stats.totalWriteTime = parseFloat64(volStats.TotalWriteTime, metrics)
	stats.totalReadBlockCount = parseFloat64(volStats.TotalReadBlockCount, metrics)
	stats.totalWriteBlockCount = parseFloat64(volStats.TotalWriteBlockCount, metrics)
	stats.uptime = parseFloat64(volStats.UpTime, metrics)
	stats.totalReplicaCount = parseFloat64(volStats.ReplicaCounter, metrics)
	stats.revisionCount = parseFloat64(volStats.RevisionCounter, metrics)
	aUsed := parseFloat64(volStats.UsedLogicalBlocks, metrics)
	aUsed = aUsed * stats.sectorSize
	stats.actualSize, _ = v1.DivideFloat64(aUsed, v1.BytesToGB)
	size := parseFloat64(volStats.Size, metrics)
	size, _ = v1.DivideFloat64(size, v1.BytesToGB)
	stats.size = size
	result := strings.Split(volStats.Iqn, ":")
	volName := result[1]
	stats.name = volName
	stats.replicas = volStats.Replicas
	stats.status = volStats.TargetStatus
	stats.iqn = volStats.Iqn
	stats.address = "127.0.0.1"
	return stats
}

func (c *cstor) close() {
	// exporter may get concurrent requests, so it's possible that
	// one of the connection may have been closed but other one is
	// still going to read or write. This ensures that read and
	// write both are closed on fd's (shutdown) and then finally close the connection.
	if conn, ok := c.conn.(*net.UnixConn); ok {
		conn.CloseRead()
	}
	if conn, ok := c.conn.(*net.UnixConn); ok {
		conn.CloseWrite()
	}
	c.conn.Close()
	c.conn = nil
}
