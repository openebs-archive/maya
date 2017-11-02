// Package collector is used to collect metrics by implementing
// prometheus.Collector interface. See function level comments
// for more details.
package collector

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/openebs/maya/types/v1"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	bytesToGB = 1073741824
	bytesToMB = 1048567
	micSec    = 1000000
	bytesToKB = 1024
	minwidth  = 0
	maxwidth  = 0
	padding   = 3
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
		Name:      "write_block_count_per_second",
		Help:      "Write Block count of volume per second",
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
}

// collect is used to set the values gathered from OpenEBS volume controller
func (e *VolumeExporter) collect() error {
	var (
		metrics1 v1.VolumeMetrics
		metrics2 v1.VolumeMetrics
	)

	httpClient := http.DefaultClient
	httpClient.Timeout = 1 * time.Second
	resp, err := httpClient.Get(e.VolumeControllerURL)
	if err != nil {
		log.Printf("could not retrieve OpenEBS Volume controller metrics: %v", err)
		return err
	}

	err = json.NewDecoder(resp.Body).Decode(&metrics1)
	if err != nil {
		log.Printf("could not decode OpenEBS Volume controller metrics: %v", err)
		return err
	}

	err = json.NewDecoder(resp.Body).Decode(&metrics2)
	if err != nil {
		log.Printf("could not decode OpenEBS Volume controller metrics: %v", err)
		return err
	}

	iRIOPS, _ := strconv.ParseFloat(metrics1.ReadIOPS, 64)
	fRIOPS, _ := strconv.ParseFloat(metrics2.ReadIOPS, 64)
	rIOPS := fRIOPS - iRIOPS
	readIOPS.Set(rIOPS)

	iRTimePS, _ := strconv.ParseFloat(metrics1.TotalReadTime, 64)
	fRTimePS, _ := strconv.ParseFloat(metrics2.TotalReadTime, 64)
	rTimePS := fRTimePS - iRTimePS
	readTimePS.Set(rTimePS)

	iRBCountPS, _ := strconv.ParseFloat(metrics1.TotalReadBlockCount, 64)
	fRBCountPS, _ := strconv.ParseFloat(metrics2.TotalReadBlockCount, 64)
	rBCountPS := fRBCountPS - iRBCountPS
	readBlockCountPS.Set(rBCountPS)

	if readIOPS == nil {
		rLatency := float64(rTimePS / rIOPS)
		rLatency = float64(rLatency / micSec)
		readLatency.Set(rLatency)
		avgRBCountPS := float64(rBCountPS / rIOPS)
		avgRBCountPS = float64(rBCountPS / bytesToKB)
		avgReadBlockCountPS.Set(avgRBCountPS)
	} else {
		readLatency = nil
		avgReadBlockCountPS = nil
	}

	iWIOPS, _ := strconv.ParseFloat(metrics1.WriteIOPS, 64)
	fWIOPS, _ := strconv.ParseFloat(metrics2.WriteIOPS, 64)
	wIOPS := fWIOPS - iWIOPS
	writeIOPS.Set(wIOPS)

	iWTimePS, _ := strconv.ParseFloat(metrics1.TotalWriteTime, 64)
	fWTimePS, _ := strconv.ParseFloat(metrics2.TotalWriteTime, 64)
	wTimePS := fWTimePS - iWTimePS
	writeTimePS.Set(wTimePS)

	iWBCountPS, _ := strconv.ParseFloat(metrics1.TotalWriteBlockCount, 64)
	fWBCountPS, _ := strconv.ParseFloat(metrics2.TotalWriteBlockCount, 64)
	wBCountPS := fWBCountPS - iWBCountPS
	writeBlockCountPS.Set(wBCountPS)

	if writeIOPS == nil {
		wLatency := float64(wTimePS / wIOPS)
		wLatency = float64(wLatency / micSec)
		writeLatency.Set(wLatency)
		avgWBCountPS := float64(wBCountPS / wIOPS)
		avgWBCountPS = float64(avgWBCountPS / bytesToKB)
		avgWriteBlockCountPS.Set(avgWBCountPS)
	} else {
		writeLatency = nil
		avgWriteBlockCountPS = nil
	}

	sSize, _ := strconv.ParseFloat(metrics2.SectorSize, 64)
	sectorSize.Set(sSize)
	uBlocks, _ := strconv.ParseFloat(metrics2.UsedBlocks, 64)
	uBlocks = uBlocks * sSize
	logicalSize.Set(uBlocks / bytesToGB)
	aUsed, _ := strconv.ParseFloat(metrics2.UsedLogicalBlocks, 64)
	aUsed = aUsed * sSize
	actualUsed.Set(aUsed / bytesToGB)

	return nil
}
