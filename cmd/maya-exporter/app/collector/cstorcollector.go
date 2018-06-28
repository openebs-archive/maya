package collector

import (
	"bytes"
	"encoding/json"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/openebs/maya/types/v1"
)

// NewCstorStatsExporter returns cstor's socket connection instance
// for the registration collectors with prometheus.
func NewCstorStatsExporter(conn net.Conn, casType string) *VolumeStatsExporter {
	return &VolumeStatsExporter{
		CASType: casType,
		Cstor: Cstor{
			Conn: conn,
		},
		Metrics: *MetricsInitializer(casType),
	}
}

// collector makes call to set for the collection of metrics
// if the connection is available else retry to initiate
// connection again.
func (c *Cstor) collector(v *VolumeStatsExporter) error {

	if c.Conn == nil {
		// initiate the connection again if connection with the istgt closed
		// due to timeout or some other errors from istgt side.
		v.connectionRetryCounter.WithLabelValues("Connection closed from cstor, retry").Inc()
		if c.InitiateConnection(); c.Conn == nil {
			glog.Error("Error in initiating the connection")
			return errors.New("error in initiating connection with socket")
		}
	}
	// set the values of stats from cstor if cstor is reachable, else set nil to
	// the value of net.Conn in case of error so that new connection be created
	// after the new request from prometheus comes.
	if err := c.set(v); err != nil {
		v.connectionErrorCounter.WithLabelValues(err.Error()).Inc()
		glog.Error("Error in connection, closing the connection")
		c.Conn.Close()
		c.Conn = nil
		return err
	}

	return nil
}

// writer writes the "IOSTATS\n" command on the socket. it returns
// error if the connection is closed or not available with socket.
func (c *Cstor) writer() error {
	msg := Command + "\n"
	_, err := c.Conn.Write([]byte(msg))
	if err != nil {
		glog.Error("Write error:", err)
		return err
	}
	return nil
}

// reader reads the response from socket in the buffer and
// exits if the buffer contains the "IOSTATS\r\n" at the end.
func (c *Cstor) reader() (string, error) {
	buf := make([]byte, 256)
	var (
		err           error
		n             int
		str, response string
		buffer        bytes.Buffer
	)
	// infinite for loop to collect all the chunks.
	for {
		n, err = c.Conn.Read(buf[:])
		if err != nil {
			glog.Error("Error in reading response, found error : ", err)
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

// splitter extracts the JSON from the response :
// "IOSTATS  { \"iqn\": \"iqn.2017-08.OpenEBS.cstor:vol1\",
//	\"Writes\": \"0\", \"Reads\": \"0\", \"TotalWriteBytes\": \"0\",
//  \"TotalReadBytes\": \"0\", \"Size\": \"10737418240\" }\r\nOK IOSTATS\r\n"
func splitter(resp string) string {
	var result []string
	result = strings.Split(resp, EOF)
	result = v1.Remove(result, Footer)
	if len(result[0]) == 0 {
		return ""
	}
	res := strings.TrimPrefix(result[0], Command+"  ")
	return res
}

// newResponse unmarshal the JSON into Response instances.
func newResponse(result string) v1.VolumeStats {
	metrics := v1.VolumeStats{}
	if err := json.Unmarshal([]byte(result), &metrics); err != nil {
		glog.Error("Error in unmarshalling, found error: ", err)
	}
	glog.Infof("Parsed metrics : %+v", metrics)
	return metrics
}

// set make call to reader and writer to write the
// IOSTATS command over wire and then reads the response.

func (c *Cstor) set(v *VolumeStatsExporter) error {
	var (
		// aggregated response from cstor stored into response
		response string
		// split response (string) and remove header, footer
		// and store only JSON data.
		newResp v1.VolumeStats
		// parse JSON response (string) into appropriate type
		// (float64, int64 etc).JSON can only handle the data
		// upto 53 bits precision, so this needs to be converted
		// into string.
		volStats VolumeStats
		err      error
	)
	if err := c.writer(); err != nil {
		return err
	}
	response, err = c.reader()
	if err != nil {
		return err
	}
	response = splitter(response)
	if len(response) == 0 {
		glog.Error("Got empty response from cstor")
		return errors.New("Got empty response from cstor")
	}

	// unmarshal the json response into Metrics instances.
	newResp = newResponse(response)
	volStats = c.parser(newResp)
	v.reads.Set(volStats.reads)
	v.writes.Set(volStats.writes)
	v.totalReadBytes.Set(volStats.totalReadBytes)
	v.totalWriteBytes.Set(volStats.totalWriteBytes)
	v.sizeOfVolume.Set(volStats.size)
	volName := strings.TrimPrefix(newResp.Iqn, "iqn.2017-08.OpenEBS.cstor:")
	// currently volumeUpTime, portal address is not available
	// from the cstor.
	// TODO : Update the volumeUpTime from 0 to the exact value
	// and add portal address and remove hardcoded value.
	now := time.Now()
	v.volumeUpTime.WithLabelValues(volName, newResp.Iqn, "localhost").Set(float64(now.Second()))
	return nil
}

// Parser can used to parse the json strings into the respective types.
// TODO: Instead of using two parser methods make it
// a generic parser that can be used for both jiva and cstor.
func (c *Cstor) parser(stats v1.VolumeStats) VolumeStats {
	volStats := VolumeStats{}
	volStats.reads, _ = stats.Reads.Float64()
	volStats.writes, _ = stats.Writes.Float64()
	volStats.totalReadBytes, _ = stats.TotalReadBytes.Float64()
	volStats.totalWriteBytes, _ = stats.TotalWriteBytes.Float64()
	size, _ := stats.Size.Float64()
	size, _ = v1.DivideFloat64(size, v1.BytesToGB)
	volStats.size = size
	return volStats
}

// InitiateConnection tries to initiates the connection with the cstor
// over unix domain socket. This function can not be unit tested (only
// negative cases are possible).
func (c *Cstor) InitiateConnection() {
	conn, err := net.Dial("unix", SocketPath)
	if err != nil {
		glog.Errorln("Dial error :", err)
	}
	if conn != nil {
		c.Conn = conn
		glog.Info("Connection established")
		c.ReadHeader()
	}
	return
}

// ReadHeader only reads the header of the response from cstor
func (c *Cstor) ReadHeader() error {
	buf := make([]byte, 1024)
	var (
		err    error
		n      int
		str    string
		buffer bytes.Buffer
	)
	// collect all the chunks ending with EOF ("\r\n").
	for {
		n, err = c.Conn.Read(buf[:])
		if err != nil {
			glog.Error("Error in reading response, found error : ", err)
			return err
		}
		// apend the chunks into str
		buffer.WriteString(string(buf[0:n]))
		str = buffer.String()
		if strings.HasPrefix(str, HeaderPrefix) && strings.HasSuffix(str, EOF) {
			break
		}
	}
	glog.Infof("Got header: %#v", str)
	return nil
}
