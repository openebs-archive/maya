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

// A gauge is a metric that represents a single numerical value that can
// arbitrarily go up and down.

// Gauges are typically used for measured values like temperatures or current
// memory usage, but also "counts" that can go up and down, like the number of
// running goroutines.

// GaugeOpts is the alias for Opts, which is used to create diffent type of
// metrics.

// All the stats exposed from jiva will be collected by the GaugeOpts.
var (
	usedLogicalBlocks = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "OpenEBS",
		Name:      "used_logical_blocks",
		Help:      "Used Logical Blocks of volume",
	})

	usedBlocks = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "OpenEBS",
		Name:      "used_blocks",
		Help:      "Used Blocks of volume",
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

	totalReadTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "OpenEBS",
		Name:      "total_read_time",
		Help:      "Total Read time on volume",
	})

	totalReadBlockCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "OpenEBS",
		Name:      "total_read_block_count",
		Help:      "Total Read Block count of volume",
	})

	writeIOPS = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "OpenEBS",
		Name:      "write_iops",
		Help:      "Write Input/Outputs on Volume",
	})

	totalWriteTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "OpenEBS",
		Name:      "total_write_time",
		Help:      "Total Write time on volume",
	})

	totalWriteBlockCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "OpenEBS",
		Name:      "total_write_block_count",
		Help:      "Total Write Block count of volume",
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
	totalReadTime.Describe(ch)
	totalReadBlockCount.Describe(ch)
	writeIOPS.Describe(ch)
	totalWriteTime.Describe(ch)
	totalWriteBlockCount.Describe(ch)
	usedLogicalBlocks.Describe(ch)
	usedBlocks.Describe(ch)
	sectorSize.Describe(ch)
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
	totalReadTime.Collect(ch)
	totalReadBlockCount.Collect(ch)
	writeIOPS.Collect(ch)
	totalWriteTime.Collect(ch)
	totalWriteBlockCount.Collect(ch)
	usedLogicalBlocks.Collect(ch)
	usedBlocks.Collect(ch)
	sectorSize.Collect(ch)
}

// collect is used to set the values gathered from OpenEBS volume controller
func (e *VolumeExporter) collect() error {
	var metrics v1.VolumeMetrics

	httpClient := http.DefaultClient
	httpClient.Timeout = 1 * time.Second
	resp, err := httpClient.Get(e.VolumeControllerURL)
	if err != nil {
		log.Printf("could not retrieve OpenEBS Volume controller metrics: %v", err)
		return err
	}

	err = json.NewDecoder(resp.Body).Decode(&metrics)
	if err != nil {
		log.Printf("could not decode OpenEBS Volume controller metrics: %v", err)
		return err
	}

	rIOPS, _ := strconv.ParseFloat(metrics.ReadIOPS, 64)
	readIOPS.Set(rIOPS)
	totRTime, _ := strconv.ParseFloat(metrics.TotalReadTime, 64)
	totalReadTime.Set(totRTime)
	totRBCount, _ := strconv.ParseFloat(metrics.TotalReadBlockCount, 64)
	totalReadBlockCount.Set(totRBCount)

	wIOPS, _ := strconv.ParseFloat(metrics.WriteIOPS, 64)
	writeIOPS.Set(wIOPS)
	totWTime, _ := strconv.ParseFloat(metrics.TotalWriteTime, 64)
	totalWriteTime.Set(totWTime)
	totWBCount, _ := strconv.ParseFloat(metrics.TotalWriteBlockCount, 64)
	totalWriteBlockCount.Set(totWBCount)

	uLBlocks, _ := strconv.ParseFloat(metrics.UsedLogicalBlocks, 64)
	usedLogicalBlocks.Set(uLBlocks)
	uBlocks, _ := strconv.ParseFloat(metrics.UsedBlocks, 64)
	usedBlocks.Set(uBlocks)
	sSize, _ := strconv.ParseFloat(metrics.SectorSize, 64)
	sectorSize.Set(sSize)

	return nil
}
