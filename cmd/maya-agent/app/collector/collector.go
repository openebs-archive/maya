package collector

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type openEBSMetrics struct {
	ReadIOPS            string `json:"ReadIOPS"`
	TotalReadTime       string `json:"TotalReadTime"`
	TotalReadBlockCount string `json:"TotalReadBlockCount"`

	WriteIOPS            string `json:"WriteIOPS"`
	TotalWriteTime       string `json:"TotalWriteTime"`
	TotalWriteBlockCount string `json:"TotatWriteBlockCount"`

	UsedLogicalBlocks string `json:"UsedLogicalBlocks"`
	UsedBlocks        string `json:"UsedBlocks"`
	SectorSize        string `json:"SectorSize"`
}

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

// Exporter implements the prometheus.Collector interface. It exposes the metrics
// of a NATS node.
type OpenEBSExporter struct {
	OpenEBSControllerURL string
}

// NewExporter instantiates a new NATS Exporter.
func NewExporter(openEBSControllerURL *url.URL) *OpenEBSExporter {
	openEBSControllerURL.Path = "/v1/stats"
	return &OpenEBSExporter{
		OpenEBSControllerURL: openEBSControllerURL.String(),
	}
}

// Describe describes all the registered stats metrics from the NATS node.
func (e *OpenEBSExporter) Describe(ch chan<- *prometheus.Desc) {
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

// Collect collects all the registered stats metrics from the NATS node.
func (e *OpenEBSExporter) Collect(ch chan<- prometheus.Metric) {
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

func (e *OpenEBSExporter) collect() error {
	var metrics openEBSMetrics

	httpClient := http.DefaultClient
	httpClient.Timeout = 1 * time.Second
	resp, err := httpClient.Get(e.OpenEBSControllerURL)
	if err != nil {
		log.Printf("could not retrieve OpenEBS controller metrics: %v", err)
		return err
	}

	err = json.NewDecoder(resp.Body).Decode(&metrics)
	if err != nil {
		log.Printf("could not decode OpenEBS controller metrics: %v", err)
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
