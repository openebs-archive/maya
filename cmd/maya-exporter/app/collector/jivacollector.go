// Package collector is used to collect metrics by implementing
// prometheus.Collector interface. See function level comments
// for more details.
package collector

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/openebs/maya/types/v1"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
)

// NewJivaStatsExporter returns Jiva volume controller URL along with Path.
func NewJivaStatsExporter(volumeControllerURL *url.URL) *JivaStatsExporter {
	volumeControllerURL.Path = v1.StatsAPI
	return &JivaStatsExporter{
		VolumeControllerURL: volumeControllerURL.String(),
		Metrics:             *MetricsInitializer(),
	}
}

// gaugeList returns the list of the registered gauge variables
func (j *JivaStatsExporter) gaugesList() []prometheus.Gauge {
	return []prometheus.Gauge{
		j.Metrics.readIOPS,
		j.Metrics.writeIOPS,
		j.Metrics.readTimePS,
		j.Metrics.writeTimePS,
		j.Metrics.readBlockCountPS,
		j.Metrics.writeBlockCountPS,
		j.Metrics.actualUsed,
		j.Metrics.logicalSize,
		j.Metrics.sectorSize,
		j.Metrics.readLatency,
		j.Metrics.writeLatency,
		j.Metrics.avgReadBlockCountPS,
		j.Metrics.avgWriteBlockCountPS,
		j.Metrics.sizeOfVolume,
	}
}

// counterList returns the list of registered counter variables
func (j *JivaStatsExporter) countersList() []prometheus.Collector {
	return []prometheus.Collector{
		j.Metrics.volumeUpTime,
		j.Metrics.connectionErrorCounter,
		j.Metrics.connectionRetryCounter,
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
func (j *JivaStatsExporter) Describe(ch chan<- *prometheus.Desc) {
	for _, gauge := range j.gaugesList() {
		gauge.Describe(ch)
	}

	for _, counter := range j.countersList() {
		counter.Describe(ch)
	}
}

// collector selects the container attached storage for the collection of
// metrics.Supported CAS are jiva and cstor.
func (j *JivaStatsExporter) collector() error {
	// collect the metrics from jiva controller and send it via channels
	if err := j.collect(); err != nil {
		j.Metrics.connectionErrorCounter.WithLabelValues(err.Error()).Inc()
		glog.Error("Error in collecting metrics, found error:", err)
		return errors.New("error in collecting metrics")
	}
	return nil
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

// Collect collects all the registered stats metrics from the Ope...nEBS volumes.
// It tries to reconnect with the volume if there is any error via a goroutine.
func (j *JivaStatsExporter) Collect(ch chan<- prometheus.Metric) {
	// no need to catch the error as exporter should work even if
	// there are failures in collecting the metrics due to connection
	// issues or anything else.
	_ = j.collector()
	// collect the metrics extracted by collect method
	for _, gauge := range j.gaugesList() {
		gauge.Collect(ch)
	}
	for _, counter := range j.countersList() {
		counter.Collect(ch)
	}
}

// getVolumeStats is used to decode the response from the Jiva controller
// response received by the client is in json format which then decoded
// and mapped to VolumeMetrics.
func (j *JivaStatsExporter) getVolumeStats(obj interface{}) error {
	httpClient := http.DefaultClient
	httpClient.Timeout = 1 * time.Second
	resp, err := httpClient.Get(j.VolumeControllerURL)

	if err != nil {
		glog.Error("could not retrieve OpenEBS Volume controller metrics: %v", err)
		return err
	}

	err = json.NewDecoder(resp.Body).Decode(obj)

	if err != nil {
		glog.Error("could not decode OpenEBS Volume controller metrics: %v", err)
		return err
	}

	defer resp.Body.Close()
	return nil
}

// collect is used to set the values gathered from OpenEBS volume
// controller to prometheus gauges. In this method two RestAPI
// calls made over a gap of 1 second.
// For exp : suppose we got reads = 10 at 3:30:00 PM in the
// first call and reads = 25 at 3:30:01 PM in the second call
// then reads per second = (25 - 10) = 15 . i.e, 15 is set to
// the read.
// This call happens as per the prometheus'c configuration.So
// if scrap interval in prometheus config is set to 5 seconds
// then this method makes two calls over a gap of one second
// to calculate the stats per second.
func (j *JivaStatsExporter) collect() error {
	var (
		initialMetrics, finalMetrics v1.VolumeMetrics
		m                            MetricsDiff
	)

	err := j.getVolumeStats(&initialMetrics)
	if err != nil {
		glog.Error("Could not decode: %v", err)
		return err
	}

	time.Sleep(1 * time.Second)
	err = j.getVolumeStats(&finalMetrics)
	if err != nil {
		glog.Error("Could not decode: %v", err)
		return err
	}

	m = j.parser(initialMetrics, finalMetrics)

	j.Metrics.readIOPS.Set(m.readIOPS)
	j.Metrics.readTimePS.Set(m.readTimePS)
	j.Metrics.readBlockCountPS.Set(m.readBlockCountPS)
	j.Metrics.writeIOPS.Set(m.writeIOPS)
	j.Metrics.writeTimePS.Set(m.writeTimePS)
	j.Metrics.writeBlockCountPS.Set(m.writeBlockCountPS)
	j.Metrics.sectorSize.Set(m.sectorSize)
	j.Metrics.logicalSize.Set(m.logicalSize)
	j.Metrics.actualUsed.Set(m.actualSize)
	j.Metrics.sizeOfVolume.Set(m.size)
	url := j.VolumeControllerURL
	url = strings.TrimSuffix(url, ":9501/v1/stats")
	url = strings.TrimPrefix(url, "http://")
	j.Metrics.volumeUpTime.WithLabelValues(finalMetrics.Name, "iqn.2016-09.com.openebs.jiva:"+finalMetrics.Name, url).Set(finalMetrics.UpTime)
	return nil
}

func (j *JivaStatsExporter) parser(m1, m2 v1.VolumeMetrics) MetricsDiff {
	metrics := MetricsDiff{}
	metrics.readIOPS, _ = v1.ParseAndSubstract(m1.ReadIOPS, m2.ReadIOPS)
	metrics.writeIOPS, _ = v1.ParseAndSubstract(m1.WriteIOPS, m2.WriteIOPS)
	metrics.readTimePS, _ = v1.ParseAndSubstract(m1.TotalReadTime, m2.TotalReadTime)
	metrics.writeTimePS, _ = v1.ParseAndSubstract(m1.TotalWriteTime, m2.TotalWriteTime)
	metrics.readBlockCountPS, _ = v1.ParseAndSubstract(m1.TotalReadBlockCount, m2.TotalReadBlockCount)
	metrics.writeBlockCountPS, _ = v1.ParseAndSubstract(m1.TotalWriteBlockCount, m2.TotalWriteBlockCount)

	if metrics.readIOPS != 0 {
		rLatency, _ := v1.DivideFloat64(metrics.readTimePS, metrics.readIOPS)
		rLatency, _ = v1.DivideFloat64(rLatency, v1.MicSec)
		metrics.readLatency = rLatency
		avgRBCountPS, _ := v1.DivideFloat64(metrics.readBlockCountPS, metrics.readIOPS)
		metrics.avgReadBlockCountPS, _ = v1.DivideFloat64(avgRBCountPS, v1.BytesToKB)
	} else {
		metrics.readLatency = 0
		metrics.avgReadBlockCountPS = 0
	}

	if metrics.writeIOPS != 0 {
		wLatency, _ := v1.DivideFloat64(metrics.writeTimePS, metrics.writeIOPS)
		wLatency, _ = v1.DivideFloat64(wLatency, v1.MicSec)
		metrics.writeLatency = wLatency
		avgWBCountPS, _ := v1.DivideFloat64(metrics.writeBlockCountPS, metrics.writeIOPS)
		metrics.avgWriteBlockCountPS, _ = v1.DivideFloat64(avgWBCountPS, v1.BytesToKB)
	} else {
		metrics.writeLatency = 0
		metrics.avgWriteBlockCountPS = 0
	}

	metrics.sectorSize, _ = strconv.ParseFloat(m2.SectorSize, 64)

	uBlocks, _ := strconv.ParseFloat(m2.UsedBlocks, 64)
	uBlocks = uBlocks * metrics.sectorSize
	metrics.logicalSize, _ = v1.DivideFloat64(uBlocks, v1.BytesToGB)
	aUsed, _ := strconv.ParseFloat(m2.UsedLogicalBlocks, 64)
	aUsed = aUsed * metrics.sectorSize
	metrics.actualSize, _ = v1.DivideFloat64(aUsed, v1.BytesToGB)
	size, _ := strconv.ParseFloat(m2.Size, 64)
	metrics.size, _ = v1.DivideFloat64(size, v1.BytesToGB)
	return metrics
}
