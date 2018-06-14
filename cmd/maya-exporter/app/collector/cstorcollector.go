package collector

import (
	"bytes"
	"encoding/json"
	"errors"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/openebs/maya/types/v1"
	"github.com/prometheus/client_golang/prometheus"
)

// Response is used to parse the json response from cstor
// It keeps the values of all the stats into the respctive fields
// TODO: Add fields for latency, used capacity, volume uptime etc.
type Response struct {
	Iqn             string `json:"iqn"`
	Writes          string `json:"writes"`
	Reads           string `json:"reads"`
	TotalReadBytes  string `json:"totalreadbytes"`
	TotalWriteBytes string `json:"totalwritebytes"`
	Size            string `json:"size"`
}

// NewCstorStatsExporter returns cstor's socket connection instance
// for the registration collectors with prometheus.
func NewCstorStatsExporter(conn net.Conn) *CstorStatsExporter {
	return &CstorStatsExporter{
		Conn:    conn,
		Metrics: *MetricsInitializer(),
	}
}

// gaugeList returns the list of the registered gauge variables
func (c *CstorStatsExporter) gaugesList() []prometheus.Gauge {
	return []prometheus.Gauge{
		c.Metrics.readIOPS,
		c.Metrics.writeIOPS,
		c.Metrics.readTimePS,
		c.Metrics.writeTimePS,
		c.Metrics.readBlockCountPS,
		c.Metrics.writeBlockCountPS,
		c.Metrics.actualUsed,
		c.Metrics.logicalSize,
		c.Metrics.sectorSize,
		c.Metrics.readLatency,
		c.Metrics.writeLatency,
		c.Metrics.avgReadBlockCountPS,
		c.Metrics.avgWriteBlockCountPS,
		c.Metrics.sizeOfVolume,
	}
}

// counterList returns the list of registered counter variables
func (c *CstorStatsExporter) countersList() []prometheus.Collector {
	return []prometheus.Collector{
		c.Metrics.volumeUpTime,
		c.Metrics.connectionErrorCounter,
		c.Metrics.connectionRetryCounter,
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector to the provided channel and returns once
// the last descriptor has been sent. The sent descriptors fulfill the
// consistency and uniqueness requirements described in the Desc
// documentation. (It is valid if one and the same Collector sends
// duplicate descriptors. Those duplicates are simply ignored. However,
// two different Collectors must not send duplicate descriptors.) This
// method idempotently sends the same descriptors throughout the
// lifetime of the Collector. If a Collector encounters an error while
// executing this method, it must send an invalid descriptor (created
// with NewInvalidDesc) to signal the error to the registry.

// Describe describes all the registered stats metrics from the OpenEBS volumes.
func (c *CstorStatsExporter) Describe(ch chan<- *prometheus.Desc) {
	for _, gauge := range c.gaugesList() {
		gauge.Describe(ch)
	}

	for _, counter := range c.countersList() {
		counter.Describe(ch)
	}
}

// Collect is called by the Prometheus registry when collecting
// metrics. The implementation sends each collected metric via the
// provided channel and returns once the last metric has been sent. The
// descriptor of each sent metric is one of those returned by
// Describe. Returned metrics that share the same descriptor must differ
// in their variable label values. This method may be called
// way. Bloc	king occurs at the expense of total performance of rendering
// concurrently and must therefore be implemented in a concurrency safe
// all registered metrics. Ideally, Collector implementations support
// concurrent readers.

// Collect collects all the registered stats metrics from the OpenEBS volumes.
// It tries to reconnect with the volume if there is any error via a goroutine.
func (c *CstorStatsExporter) Collect(ch chan<- prometheus.Metric) {
	// no need to catch the error as exporter should work even if
	// there are failures in collecting the metrics due to connection
	// issues or anything else. This also makes collector unit testable.
	_ = c.collector()
	// collect the metrics extracted by collect method
	for _, gauge := range c.gaugesList() {
		gauge.Collect(ch)
	}
	for _, counter := range c.countersList() {
		counter.Collect(ch)
	}
}

// collector selects the container attached storage for the collection of
// metrics.Supported CAS are jiva and cstor.
func (c *CstorStatsExporter) collector() error {

	if c.Conn == nil {
		// initiate the connection again if connection with the istgt closed
		// due to timeout or some other errors from istgt side.
		c.Metrics.connectionRetryCounter.WithLabelValues("Connection closed from cstor, retry").Inc()
		c.InitiateConnection()
		if c.Conn == nil {
			glog.Error("Error in initiating the connection")
			return errors.New("error in initiating connection with socket")
		}

	}
	// collect the metrics from cstor if cstor is reachable, else set nil to
	// the value of net.Conn in case of error so that new connection be created
	// after the new request from prometheus comes.
	if err := c.collect(); err != nil {
		c.Metrics.connectionErrorCounter.WithLabelValues(err.Error()).Inc()
		glog.Error("Error in connection, closing the connection")
		c.Conn.Close()
		c.Conn = nil
		return err
	}

	return nil
}

// writer writes the "IOSTATS\n" command on the socket. it returns
// error if the connection is closed or not available with socket.
func (c *CstorStatsExporter) writer() error {
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
func (c *CstorStatsExporter) reader() (string, error) {
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
//	\"writes\": \"0\", \"reads\": \"0\", \"totalwritebytes\": \"0\",
//  \"totalreadbytes\": \"0\", \"size\": \"10737418240\" }\r\nOK IOSTATS\r\n"
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
func newResponse(result string) Response {
	metrics := Response{}
	if err := json.Unmarshal([]byte(result), &metrics); err != nil {
		glog.Error("Error in unmarshalling, found error: ", err)
	}
	glog.Infof("Parsed metrics : %+v", metrics)
	return metrics
}

// collect make call to reader and writer to write the
// IOSTATS command over wire and then reads the response.
// There are two consecutive calls for read and write over
// a gap of 1 second to get the I/O stats per second.
// For exp : suppose we got reads = 10 at 3:30:00 PM in the
// first call and reads = 25 at 3:30:01 PM in the second call
// then reads per second = (25 - 10) = 15 .i.e, 15 is set to
// the read.
// This call happens as per the prometheus'c configuration.So
// if scrap interval in prometheus config is set to 5 seconds
// then this method makes two calls over a gap of one second
// to calculate the stats per second.
func (c *CstorStatsExporter) collect() error {
	var (
		initialResponse, finalResponse string
		initialMetrics, finalMetrics   Response
		metrics                        MetricsDiff
		err                            error
	)
	if err := c.writer(); err != nil {
		return err
	}
	initialResponse, err = c.reader()
	if err != nil {
		return err
	}
	time.Sleep(1 * time.Second)
	if err = c.writer(); err != nil {
		return err
	}
	finalResponse, err = c.reader()
	if err != nil {
		return err
	}
	// Response comes in the format given below :
	// "IOSTATS  { \"iqn\": \"iqn.2017-08.OpenEBS.cstor:vol1\",
	//	\"writes\": \"0\", \"reads\": \"0\", \"totalwritebytes\": \"0\",
	//  \"totalreadbytes\": \"0\", \"size\": \"10737418240\" }\r\nOK IOSTATS\r\n"
	// so to unmarshal the json part we need to split
	// the string and remove the command and footer.
	initialResponse = splitter(initialResponse)
	finalResponse = splitter(finalResponse)
	if len(initialResponse) == 0 || len(finalResponse) == 0 {
		glog.Error("Got empty response from cstor")
		return errors.New("Got empty response from cstor")
	}

	// unmarshal the json response into Metrics instances.
	initialMetrics = newResponse(initialResponse)
	finalMetrics = newResponse(finalResponse)

	// since json can only parse the data upto 53 bit precision, we need
	// to convert uint64 into string and then send in json.
	// parser() parse the json data into MetricsDiff instance.
	// Ref : https://stackoverflow.com/questions/209869/what-is-the-accepted-way-to-send-64-bit-values-over-json
	metrics = c.parser(initialMetrics, finalMetrics)
	c.Metrics.readIOPS.Set(metrics.readIOPS)
	c.Metrics.writeIOPS.Set(metrics.writeIOPS)
	c.Metrics.readBlockCountPS.Set(metrics.readBlockCountPS)
	c.Metrics.writeBlockCountPS.Set(metrics.writeBlockCountPS)
	c.Metrics.sizeOfVolume.Set(metrics.size)
	volName := strings.TrimPrefix(finalMetrics.Iqn, "iqn.2017-08.OpenEBS.cstor:")
	// currently volumeUpTime is not available from the cstor.
	// TODO : Update the volumeUpTime from 0 to the exact value
	c.Metrics.volumeUpTime.WithLabelValues(volName, finalMetrics.Iqn, "localhost").Set(0)
	return nil
}

// Parser can used to parse the json strings into the respective types.
// TODO: Remove the ParseAndSubstract instead parse the data
// into another structure with appropriate field types.
func (c *CstorStatsExporter) parser(m1, m2 Response) MetricsDiff {
	metrics := MetricsDiff{}
	metrics.readIOPS, _ = v1.ParseAndSubstract(m1.Reads, m2.Reads)
	metrics.writeIOPS, _ = v1.ParseAndSubstract(m1.Writes, m2.Writes)
	rThput, _ := v1.ParseAndSubstract(m1.TotalReadBytes, m2.TotalReadBytes)
	metrics.readBlockCountPS = (rThput / v1.BytesToMB)
	wThput, _ := v1.ParseAndSubstract(m1.TotalWriteBytes, m2.TotalWriteBytes)
	metrics.writeBlockCountPS = (wThput / v1.BytesToMB)
	size, _ := strconv.ParseFloat(m1.Size, 64)
	size, _ = v1.DivideFloat64(size, v1.BytesToGB)
	metrics.size = size
	return metrics
}

// InitiateConnection tries to initiates the connection with the cstor
// over unix domain socket. This function can not be unit tested (only
// negative cases are possible).
func (c *CstorStatsExporter) InitiateConnection() {
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
func (c *CstorStatsExporter) ReadHeader() error {
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
