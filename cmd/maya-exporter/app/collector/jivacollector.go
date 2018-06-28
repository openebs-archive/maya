// Package collector is used to collect metrics by implementing
// prometheus.Collector interface. See function level comments
// for more details.
package collector

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/openebs/maya/types/v1"

	"github.com/golang/glog"
)

// NewJivaStatsExporter returns Jiva volume controller URL along with Path.
func NewJivaStatsExporter(volumeControllerURL *url.URL, casType string) *VolumeStatsExporter {
	volumeControllerURL.Path = "v1/stats"
	return &VolumeStatsExporter{
		CASType: casType,
		Jiva: Jiva{
			VolumeControllerURL: volumeControllerURL.String(),
		},
		Metrics: *MetricsInitializer(casType),
	}
}

// collector selects the container attached storage for the collection of
// metrics.Supported CAS are jiva and cstor.
func (j *Jiva) collector(v *VolumeStatsExporter) error {
	// set the metrics from jiva controller and send it via channels
	if err := j.set(v); err != nil {
		v.connectionErrorCounter.WithLabelValues(err.Error()).Inc()
		return errors.New("error in collecting metrics")
	}
	return nil
}

// getVolumeStats is used to decode the response from the Jiva controller
// response received by the client is in json format which then decoded
// and mapped to VolumeMetrics.
func (j *Jiva) getVolumeStats(obj *v1.VolumeStats) error {
	httpClient := http.DefaultClient
	httpClient.Timeout = 1 * time.Second
	resp, err := httpClient.Get(j.VolumeControllerURL)

	if err != nil {
		glog.Error("could not retrieve OpenEBS Volume controller metrics: %v", err)
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Error(err.Error())
		return err
	}
	glog.Info("Got response: ", string(body))
	err = json.Unmarshal(body, &obj)

	if err != nil {
		glog.Error("could not decode OpenEBS Volume controller metrics: %#v", err)
		return errors.New("Error in unmarshalling the json response")
	}

	defer resp.Body.Close()
	return nil
}

// set is used to set the values gathered from Jiva volume
// controller to prometheus gauges and counters.
func (j *Jiva) set(v *VolumeStatsExporter) error {
	var (
		// JSON response from jiva controller
		volStatsJSON v1.VolumeStats
		// parse JSON response into appropriate type.
		volStats VolumeStats
	)

	err := j.getVolumeStats(&volStatsJSON)
	if err != nil {
		return err
	}
	volStats = j.parser(volStatsJSON)

	v.reads.Set(volStats.reads)
	v.totalReadTime.Set(volStats.totalReadTime)
	v.writes.Set(volStats.writes)
	v.totalWriteTime.Set(volStats.totalWriteTime)
	v.totalReadBlockCount.Set(volStats.totalReadBlockCount)
	v.totalWriteBlockCount.Set(volStats.totalWriteBlockCount)
	v.sectorSize.Set(volStats.sectorSize)
	v.logicalSize.Set(volStats.logicalSize)
	v.actualUsed.Set(volStats.actualSize)
	v.sizeOfVolume.Set(volStats.size)
	url := j.VolumeControllerURL
	url = strings.TrimSuffix(url, ":9501/v1/stats")
	url = strings.TrimPrefix(url, "http://")
	v.volumeUpTime.WithLabelValues(volStatsJSON.Name, "iqn.2016-09.com.openebs.jiva:"+volStatsJSON.Name, url).Set(volStatsJSON.UpTime)
	return nil
}

func (j *Jiva) parser(stats v1.VolumeStats) VolumeStats {
	volStats := VolumeStats{}
	volStats.reads, _ = stats.Reads.Float64()
	volStats.writes, _ = stats.Writes.Float64()
	volStats.totalReadTime, _ = stats.TotalReadTime.Float64()
	volStats.totalWriteTime, _ = stats.TotalWriteTime.Float64()
	volStats.totalReadBlockCount, _ = stats.TotalReadBlockCount.Float64()
	volStats.totalWriteBlockCount, _ = stats.TotalWriteBlockCount.Float64()

	volStats.sectorSize, _ = stats.SectorSize.Float64()

	uBlocks, _ := stats.UsedBlocks.Float64()
	uBlocks = uBlocks * volStats.sectorSize
	volStats.logicalSize, _ = v1.DivideFloat64(uBlocks, v1.BytesToGB)
	aUsed, _ := stats.UsedLogicalBlocks.Float64()
	aUsed = aUsed * volStats.sectorSize
	volStats.actualSize, _ = v1.DivideFloat64(aUsed, v1.BytesToGB)
	size, _ := stats.Size.Float64()
	volStats.size, _ = v1.DivideFloat64(size, v1.BytesToGB)
	return volStats
}
