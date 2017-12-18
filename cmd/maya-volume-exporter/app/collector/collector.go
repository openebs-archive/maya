// Package collector is used to collect metrics by implementing
// prometheus.Collector interface. See function level comments
// for more details.
package collector

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/openebs/maya/types/v1"
	"github.com/prometheus/client_golang/prometheus"
)

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
func (e *VolumeExporter) Collect(ch chan<- prometheus.Metric) {
	if err := e.collect(); err != nil {
		return
	}

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
		log.Printf("could not retrieve OpenEBS Volume controller metrics: %v", err)
		return err
	}

	err = json.NewDecoder(resp.Body).Decode(obj)

	if err != nil {
		log.Printf("could not decode OpenEBS Volume controller metrics: %v", err)
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
		fmt.Printf("Could not decode: %v", err)
	}

	time.Sleep(1 * time.Second)
	err = e.getVolumeStats(&metrics2)
	if err != nil {
		fmt.Printf("Could not decode: %v", err)
	}

	// i and f is used for initial and final
	iRIOPS, _ := strconv.ParseInt(metrics1.ReadIOPS, 10, 64)
	fRIOPS, _ := strconv.ParseInt(metrics2.ReadIOPS, 10, 64)
	rIOPS, _ := v1.SubstractInt64(fRIOPS, iRIOPS)
	readIOPS.Set(float64(rIOPS))

	iRTimePS, _ := strconv.ParseInt(metrics1.TotalReadTime, 10, 64)
	fRTimePS, _ := strconv.ParseInt(metrics2.TotalReadTime, 10, 64)
	rTimePS, _ := v1.SubstractInt64(fRTimePS, iRTimePS)
	readTimePS.Set(float64(rTimePS))

	iRBCountPS, _ := strconv.ParseInt(metrics1.TotalReadBlockCount, 10, 64)
	fRBCountPS, _ := strconv.ParseInt(metrics2.TotalReadBlockCount, 10, 64)
	rBCountPS, _ := v1.SubstractInt64(fRBCountPS, iRBCountPS)
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

	iWIOPS, _ := strconv.ParseInt(metrics1.WriteIOPS, 10, 64)
	fWIOPS, _ := strconv.ParseInt(metrics2.WriteIOPS, 10, 64)
	wIOPS, _ := v1.SubstractInt64(fWIOPS, iWIOPS)
	writeIOPS.Set(float64(wIOPS))

	iWTimePS, _ := strconv.ParseInt(metrics1.TotalWriteTime, 10, 64)
	fWTimePS, _ := strconv.ParseInt(metrics2.TotalWriteTime, 10, 64)
	wTimePS, _ := v1.SubstractInt64(fWTimePS, iWTimePS)
	writeTimePS.Set(float64(wTimePS))

	iWBCountPS, _ := strconv.ParseInt(metrics1.TotalWriteBlockCount, 10, 64)
	fWBCountPS, _ := strconv.ParseInt(metrics2.TotalWriteBlockCount, 10, 64)
	wBCountPS, _ := v1.SubstractInt64(fWBCountPS, iWBCountPS)
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
