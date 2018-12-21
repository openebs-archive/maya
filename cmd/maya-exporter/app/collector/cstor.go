package collector

import (
	"bytes"
	"net"
	"strings"

	"github.com/pkg/errors"

	"github.com/golang/glog"
	v1 "github.com/openebs/maya/pkg/apis/openebs.io/stats"
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
	// BufSize is the size of response from cstor.
	BufSize = 1024
)

// cstor implements the Exporter interface. It exposes
// the metrics of a OpenEBS (cstor) volume.
type cstor struct {
	// conn is used as unix network connection
	conn net.Conn
	// stats is volume stats associated with
	// jiva (cas)
	stats stats
}

// Cstor returns cstor's instance
func Cstor() *cstor {
	return &cstor{
		stats: stats{},
	}
}

// InitiateConnection tries to initiates the connection with the cstor
// over unix domain socket. This function can not be unit tested (only
// negative cases are possible).
func (c *cstor) InitiateConnection(path string) error {
	conn, err := net.Dial("unix", path)
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
		if str[len(str)-12:] == Footer+EOF {
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
	glog.Info("Connection established with istgt, got header: %#v", str)
	return nil
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
		if err := c.InitiateConnection(SocketPath); err != nil {
			return v1.VolumeStats{}, &connErr{
				errors.Errorf("%s: %v", "can't initiate connection with istgt", err),
			}
		}
	}
	if err := c.writer(); err != nil {
		c.conn.Close()
		c.conn = nil
		return v1.VolumeStats{}, &connErr{
			errors.Errorf("%v: %v", err, "closing connection"),
		}
	}
	response, err := c.reader()
	if err != nil {
		c.conn.Close()
		c.conn = nil
		return v1.VolumeStats{}, &connErr{
			errors.Errorf("%v: %v", err, "closing connection"),
		}
	}
	glog.Infof("Got response: %v", response)
	response = splitter(response)
	if len(response) == 0 {
		return v1.VolumeStats{}, errors.New("Got empty response from cstor")
	}

	// unmarshal the json response into metrics instances.
	if stats, err = newResponse(response); err != nil {
		return v1.VolumeStats{}, err
	}

	stats.Got = true
	return stats, nil
}

// parse can used to parse the json strings into the respective types.
func (c *cstor) parse(volStats v1.VolumeStats) stats {
	if !volStats.Got {
		glog.Warningf("%s", "can't parse, got empty stats, istgt may not be reachable")
		return stats{}
	}
	c.stats.got = true
	c.stats.casType = "cstor"
	c.stats.reads, _ = volStats.Reads.Float64()
	c.stats.writes, _ = volStats.Writes.Float64()
	c.stats.totalReadBytes, _ = volStats.TotalReadBytes.Float64()
	c.stats.totalWriteBytes, _ = volStats.TotalWriteBytes.Float64()
	c.stats.sectorSize, _ = volStats.SectorSize.Float64()
	c.stats.totalReadTime, _ = volStats.TotalReadTime.Float64()
	c.stats.totalWriteTime, _ = volStats.TotalWriteTime.Float64()
	c.stats.totalReadBlockCount, _ = volStats.TotalReadBlockCount.Float64()
	c.stats.totalWriteBlockCount, _ = volStats.TotalWriteBlockCount.Float64()
	c.stats.uptime, _ = volStats.UpTime.Float64()
	c.stats.totalReplicaCount, _ = volStats.ReplicaCounter.Float64()
	c.stats.revisionCount, _ = volStats.RevisionCounter.Float64()
	aUsed, _ := volStats.UsedLogicalBlocks.Float64()
	aUsed = aUsed * c.stats.sectorSize
	c.stats.actualSize, _ = v1.DivideFloat64(aUsed, v1.BytesToGB)
	size, _ := volStats.Size.Float64()
	size, _ = v1.DivideFloat64(size, v1.BytesToGB)
	c.stats.size = size
	result := strings.Split(volStats.Iqn, ":")
	volName := result[1]
	c.stats.name = volName
	c.stats.replicas = volStats.Replicas
	c.stats.status = volStats.TargetStatus
	c.stats.iqn = volStats.Iqn
	c.stats.address = "127.0.0.1"
	return c.stats
}
