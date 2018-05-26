// Package collector is used to collect metrics by implementing
// prometheus.Collector interface. See function level comments
// for more details.
package collector

import (
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/openebs/maya/types/v1"
	"github.com/prometheus/client_golang/prometheus"
)

// SocketPath is path from where connection has to be created.
const SocketPath = "/var/run/istgt_ctl_sock"

// A gauge is a metric that represents a single numerical value that can
// arbitrarily go up and down.

// Gauges are typically used for measured values like temperatures or current
// memory usage, but also "counts" that can go up and down, like the number of
// running goroutines.

// GaugeOpts is the alias for Opts, which is used to create diffent type of
// metrics.

// All the stats exposed from jiva will be collected by the GaugeOpts.
var (
	actualUsed = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "OpenEBS",
		Name:      "actual_used",
		Help:      "Actual volume size used",
	})

	logicalSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "OpenEBS",
		Name:      "logical_size",
		Help:      "Logical size of volume",
	})

	sectorSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "OpenEBS",
		Name:      "sector_size",
		Help:      "sector size of volume",
	})

	readIOPS = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "OpenEBS",
		Name:      "read_iops",
		Help:      "Read Input/Outputs on Volume",
	})

	readTimePS = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "OpenEBS",
		Name:      "read_time_per_second",
		Help:      "Read time on volume per second",
	})

	readBlockCountPS = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "OpenEBS",
		Name:      "read_block_count_per_second",
		Help:      "Read Block count of volume per second",
	})

	writeIOPS = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "OpenEBS",
		Name:      "write_iops",
		Help:      "Write Input/Outputs on Volume per second",
	})

	writeTimePS = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "OpenEBS",
		Name:      "write_time_per_second",
		Help:      "Write time on volume per second",
	})

	writeBlockCountPS = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "OpenEBS",
		Name:      "write_block_count_per_second",
		Help:      "Write Block count of volume per second",
	})

	readLatency = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "OpenEBS",
		Name:      "read_latency",
		Help:      "Read Latency count of volume",
	})

	writeLatency = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "OpenEBS",
		Name:      "write_latency",
		Help:      "Write Latency count of volume",
	})

	avgReadBlockCountPS = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "OpenEBS",
		Name:      "avg_read_block_count_per_second",
		Help:      "Average Read Block count of volume per second",
	})

	avgWriteBlockCountPS = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "OpenEBS",
		Name:      "avg_write_block_count_per_second",
		Help:      "Average Write Block count of volume per second",
	})

	sizeOfVolume = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "OpenEBS",
		Name:      "size_of_volume",
		Help:      "Size of the volume requested",
	})

	volumeUpTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "OpenEBS",
		Name:      "volume_uptime",
		Help:      "Time since volume has registered",
	},
		[]string{"volName", "iqn", "portal"},
	)
)

// Collector is the interface implemented by anything that can be used by
// Prometheus to collect metrics. A Collector has to be registered for
// collection of  metrics. Basically it has two methods Describe and Collect.

// VolumeExporter implements the prometheus.Collector interface. It exposes
// the metrics of a OpenEBS (Jiva) volume.
type VolumeExporter struct {
	VolumeControllerURL string
	Conn                net.Conn
}

// NewExporter returns Jiva volume controller URL along with Path.
func NewExporter(volumeControllerURL *url.URL) *VolumeExporter {
	volumeControllerURL.Path = "/v1/stats"
	return &VolumeExporter{
		VolumeControllerURL: volumeControllerURL.String(),
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
func (e *VolumeExporter) Describe(ch chan<- *prometheus.Desc) {
	readIOPS.Describe(ch)
	readTimePS.Describe(ch)
	readBlockCountPS.Describe(ch)
	writeIOPS.Describe(ch)
	writeTimePS.Describe(ch)
	writeBlockCountPS.Describe(ch)
	actualUsed.Describe(ch)
	logicalSize.Describe(ch)
	sectorSize.Describe(ch)
	readLatency.Describe(ch)
	writeLatency.Describe(ch)
	avgReadBlockCountPS.Describe(ch)
	avgWriteBlockCountPS.Describe(ch)
	sizeOfVolume.Describe(ch)
	volumeUpTime.Describe(ch)
}

// Collect is called by the Prometheus registry when collecting
// metrics. The implementation sends each collected metric via the
// provided channel and returns once the last metric has been sent. The
// descriptor of each sent metric is one of those returned by
// Describe. Returned metrics that share the same descriptor must differ
// in their variable label values. This method may be called
// concurrently and must therefore be implemented in a concurrency safe
// way. Blocking occurs at the expense of total performance of rendering
// all registered metrics. Ideally, Collector implementations support
// concurrent readers.

// Collect collects all the registered stats metrics from the OpenEBS volumes.
// It tries to reconnect with the volume if there is any error via a goroutine.
func (e *VolumeExporter) Collect(ch chan<- prometheus.Metric) {
	// verify if controller url is not empty
	if len(e.VolumeControllerURL) != 0 {
		// collect the metrics from jiva controller and send it via channels
		if err := e.collect(); err != nil {
			glog.Error("Error in collecting metrics, found error:", err)
			return
		}
	}
	// verify if net.Conn is not nil and proceed to collect metrics from
	// socket.
	if e.Conn != nil {
		if err := collect(e); err != nil {
			glog.Info("Error in connection, retrying")
			go e.Retry()
		}
	}

	// collect the metrics extracted by collect methods called above
	readIOPS.Collect(ch)
	readTimePS.Collect(ch)
	readBlockCountPS.Collect(ch)
	writeIOPS.Collect(ch)
	writeTimePS.Collect(ch)
	writeBlockCountPS.Collect(ch)
	actualUsed.Collect(ch)
	logicalSize.Collect(ch)
	sectorSize.Collect(ch)
	readLatency.Collect(ch)
	writeLatency.Collect(ch)
	avgReadBlockCountPS.Collect(ch)
	avgWriteBlockCountPS.Collect(ch)
	sizeOfVolume.Collect(ch)
	volumeUpTime.Collect(ch)
}

// getVolumeStats is used to decode the response from the Jiva controller
// response received by the client is in json format which then decoded
// and mapped to VolumeMetrics.
func (e *VolumeExporter) getVolumeStats(obj interface{}) error {
	httpClient := http.DefaultClient
	httpClient.Timeout = 1 * time.Second
	resp, err := httpClient.Get(e.VolumeControllerURL)

	if err != nil {
		glog.Infof("could not retrieve OpenEBS Volume controller metrics: %v", err)
		return err
	}

	err = json.NewDecoder(resp.Body).Decode(obj)

	if err != nil {
		glog.Infof("could not decode OpenEBS Volume controller metrics: %v", err)
		return err
	}

	defer resp.Body.Close()
	return nil
}

// collect is used to set the values gathered from OpenEBS volume controller
// to prometheus gauges.
func (e *VolumeExporter) collect() error {
	var (
		metrics1 v1.VolumeMetrics
		metrics2 v1.VolumeMetrics
	)

	err := e.getVolumeStats(&metrics1)
	if err != nil {
		glog.Infof("Could not decode: %v", err)
	}

	time.Sleep(1 * time.Second)
	err = e.getVolumeStats(&metrics2)
	if err != nil {
		glog.Infof("Could not decode: %v", err)
	}

	// i and f is used for initial and final
	rIOPS, _ := v1.ParseAndSubstract(metrics1.ReadIOPS, metrics2.ReadIOPS)
	readIOPS.Set(float64(rIOPS))

	rTimePS, _ := v1.ParseAndSubstract(metrics1.TotalReadTime, metrics2.TotalReadTime)
	readTimePS.Set(float64(rTimePS))

	rBCountPS, _ := v1.ParseAndSubstract(metrics1.TotalReadBlockCount, metrics2.TotalReadBlockCount)
	readBlockCountPS.Set(float64(rBCountPS))

	if rIOPS != 0 {
		rLatency, _ := v1.DivideInt64(rTimePS, rIOPS)
		rLatency, _ = v1.DivideInt64(rLatency, v1.MicSec)
		readLatency.Set(float64(rLatency))
		avgRBCountPS, _ := v1.DivideInt64(rBCountPS, rIOPS)
		avgRBCountPS, _ = v1.DivideInt64(rBCountPS, v1.BytesToKB)
		avgReadBlockCountPS.Set(float64(avgRBCountPS))
	} else {
		readLatency.Set(0)
		avgReadBlockCountPS.Set(0)
	}

	wIOPS, _ := v1.ParseAndSubstract(metrics1.WriteIOPS, metrics2.WriteIOPS)
	writeIOPS.Set(float64(wIOPS))

	wTimePS, _ := v1.ParseAndSubstract(metrics1.TotalWriteTime, metrics2.TotalWriteTime)
	writeTimePS.Set(float64(wTimePS))

	wBCountPS, _ := v1.ParseAndSubstract(metrics1.TotalWriteBlockCount, metrics2.TotalWriteBlockCount)
	writeBlockCountPS.Set(float64(wBCountPS))

	if wIOPS != 0 {
		wLatency, _ := v1.DivideInt64(wTimePS, wIOPS)
		wLatency, _ = v1.DivideInt64(wLatency, v1.MicSec)
		writeLatency.Set(float64(wLatency))
		avgWBCountPS, _ := v1.DivideInt64(wBCountPS, wIOPS)
		avgWBCountPS, _ = v1.DivideInt64(avgWBCountPS, v1.BytesToKB)
		avgWriteBlockCountPS.Set(float64(avgWBCountPS))
	} else {
		writeLatency.Set(0)
		avgWriteBlockCountPS.Set(0)
	}

	sSize, _ := strconv.ParseFloat(metrics2.SectorSize, 64)
	sectorSize.Set(sSize)
	uBlocks, _ := strconv.ParseFloat(metrics2.UsedBlocks, 64)
	uBlocks = uBlocks * sSize
	lSize, _ := v1.DivideFloat64(uBlocks, v1.BytesToGB)
	logicalSize.Set(lSize)
	aUsed, _ := strconv.ParseFloat(metrics2.UsedLogicalBlocks, 64)
	aUsed = aUsed * sSize
	aSize, _ := v1.DivideFloat64(aUsed, v1.BytesToGB)
	actualUsed.Set(aSize)
	size, _ := strconv.ParseInt(metrics1.Size, 10, 64)
	size, _ = v1.DivideInt64(size, v1.BytesToGB)
	sizeOfVolume.Set(float64(size))
	url := e.VolumeControllerURL
	url = strings.TrimSuffix(url, ":9501/v1/stats")
	url = strings.TrimPrefix(url, "http://")
	volumeUpTime.WithLabelValues(metrics1.Name, "iqn.2016-09.com.openebs.jiva:"+metrics1.Name, url).Set(metrics1.UpTime)
	return nil
}

// Retry tries to initiates the connection with the socket. It retries
// untill the connection is established.
func (e *VolumeExporter) Retry() {
	var (
		i   int
		err error
	)

retry:
	e.Conn, err = net.Dial("unix", SocketPath)
	if err != nil {
		glog.Errorln("Dial error :", err)
		glog.Info("Sleep for 5 second and then retry initiating connection.")
		time.Sleep(5 * time.Second)
		for {
			i++
			glog.Info("Retrying to connect to the server, retry count :", i)
			goto retry
		}
	}
	glog.Info("Connection established")
}
