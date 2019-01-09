package collector

import (
	"bytes"
	"encoding/json"
	"net"
	"strings"

	"github.com/golang/glog"
	v1 "github.com/openebs/maya/pkg/stats/v1alpha1"
	"github.com/pkg/errors"
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
	// conn is used as unix network connection
	conn net.Conn
}

// Cstor returns cstor's instance
func Cstor(path string) *cstor {
	var c = new(cstor)
	// ignore error, istgt may not be up. Connection
	// can be established later upon new requests.
	if err := c.initiateConnection(path); err != nil {
		glog.Warning("Can't connect to istgt, error: ", err)
	}
	return c
}

// initiateConnection tries to initiates the connection with the cstor
// over unix domain socket.
func (c *cstor) initiateConnection(path string) error {
	conn, err := dialFunc(path)
	if err != nil {
		return err
	}
	if conn != nil {
		c.conn = conn
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
			glog.Error("Error in reading response, error : ", err)
			return err
		}
		// apend the chunks into str
		buffer.WriteString(string(buf[0:n]))
		str = buffer.String()
		if strings.HasPrefix(str, HeaderPrefix) && strings.HasSuffix(str, EOF) {
			break
		}
	}
	glog.Infof("Connection established with istgt, got header: %#v", str)
	return nil
}

// removeItem removes the string passed as argument from the slice
func (c *cstor) removeItem(slice []string, str string) []string {
	for index, value := range slice {
		if value == str {
			slice = append(slice[:index], slice[index+1:]...)
		}
	}
	return slice
}

// unmarshal the result into VolumeStats instances.
func (c *cstor) unmarshal(result string) (v1.VolumeStats, error) {
	metrics := v1.VolumeStats{}
	if err := json.Unmarshal([]byte(result), &metrics); err != nil {
		glog.Error("Error in unmarshalling, found error: ", err)
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
	result = c.removeItem(result, Footer)
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
	var (
		err   error
		stats v1.VolumeStats
	)
	if c.conn == nil {
		// initiate the connection again if connection with the istgt closed
		// due to timeout or some other errors from istgt side.
		if err := c.initiateConnection(SocketPath); err != nil {
			return v1.VolumeStats{}, &colErr{
				errors.Wrap(err, "Can't connect to istgt"),
			}
		}
	}
	if err := c.writer(); err != nil {
		c.close()
		return v1.VolumeStats{}, &colErr{
			errors.Errorf("%v: closing connection", err),
		}
	}
	buf, err := c.reader()
	if err != nil {
		c.close()
		return v1.VolumeStats{}, &colErr{
			errors.Errorf("%v: closing connection", err),
		}
	}
	glog.Infof("Got response: %v", buf)
	resp := c.splitter(buf)
	if len(resp) == 0 {
		return v1.VolumeStats{}, errors.New("Got empty response from cstor")
	}

	// unmarshal the json response into metrics instances.
	if stats, err = c.unmarshal(resp); err != nil {
		return v1.VolumeStats{}, err
	}

	stats.Got = true
	return stats, nil
}

// parse can used to parse the json strings into the respective types.
func (c *cstor) parse(volStats v1.VolumeStats, metrics *metrics) stats {
	var stats = stats{}
	if !volStats.Got {
		glog.Warningf("%s", "can't parse, got empty stats, istgt may not be reachable")
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
	c.conn.Close()
	c.conn = nil
}
