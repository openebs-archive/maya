// Package collector is used to collect metrics by implementing
// prometheus.Collector interface. See function level comments
// for more details.
package collector

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	v1 "github.com/openebs/maya/pkg/apis/openebs.io/stats"
	"github.com/pkg/errors"

	"github.com/golang/glog"
)

// jiva implements the Volume interface. It exposes
// the metrics of a OpenEBS (Jiva) volume.
type jiva struct {
	// url is jiva controller's url
	url string
	// stats is volume statistics associated with
	// jiva (cas)
	stats stats
}

// Jiva returns jiva's instance
func Jiva(url *url.URL) *jiva {
	url.Path = "/v1/stats"
	return &jiva{
		url:   url.String(),
		stats: stats{},
	}
}

// getter get stats from jiva controller
func (j *jiva) get() (v1.VolumeStats, error) {
	var volStats v1.VolumeStats
	if err := j.getVolumeStats(&volStats); err != nil {
		return v1.VolumeStats{}, err
	}
	volStats.Got = true
	return volStats, nil
}

// getvolumeStats is used to get the response from the Jiva controller
// which then unmarshalled into the v1.VolumeStats structure.
func (j *jiva) getVolumeStats(obj *v1.VolumeStats) error {
	httpClient := http.DefaultClient
	httpClient.Timeout = 1 * time.Second
	resp, err := httpClient.Get(j.url)

	if err != nil {
		return &connErr{
			errors.Errorf("%v", err),
		}
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	glog.Info("Got response: ", string(body))
	err = json.Unmarshal(body, &obj)

	if err != nil {
		return err
	}

	resp.Body.Close()
	return nil
}

func (j *jiva) parse(volStats v1.VolumeStats) stats {
	if !volStats.Got {
		glog.Warningf("%s", "can't parse stats, controller may not be reachable")
		return stats{}
	}
	j.stats.got = true
	j.stats.casType = "jiva"
	j.stats.reads, _ = volStats.Reads.Float64()
	j.stats.writes, _ = volStats.Writes.Float64()
	j.stats.totalReadBytes, _ = volStats.TotalReadBytes.Float64()
	j.stats.totalWriteBytes, _ = volStats.TotalWriteBytes.Float64()
	j.stats.totalReadTime, _ = volStats.TotalReadTime.Float64()
	j.stats.totalWriteTime, _ = volStats.TotalWriteTime.Float64()
	j.stats.totalReadBlockCount, _ = volStats.TotalReadBlockCount.Float64()
	j.stats.totalWriteBlockCount, _ = volStats.TotalWriteBlockCount.Float64()

	j.stats.sectorSize, _ = volStats.SectorSize.Float64()

	uBlocks, _ := volStats.UsedBlocks.Float64()
	uBlocks = uBlocks * j.stats.sectorSize
	j.stats.logicalSize, _ = v1.DivideFloat64(uBlocks, v1.BytesToGB)
	aUsed, _ := volStats.UsedLogicalBlocks.Float64()
	aUsed = aUsed * j.stats.sectorSize
	j.stats.actualSize, _ = v1.DivideFloat64(aUsed, v1.BytesToGB)
	size, _ := volStats.Size.Float64()
	j.stats.size, _ = v1.DivideFloat64(size, v1.BytesToGB)
	j.stats.totalReplicaCount, _ = volStats.ReplicaCounter.Float64()
	j.stats.revisionCount, _ = volStats.RevisionCounter.Float64()
	j.stats.uptime, _ = volStats.UpTime.Float64()
	j.stats.name = volStats.Name
	j.stats.replicas = volStats.Replicas
	j.stats.status = volStats.TargetStatus
	url := j.url
	url = strings.TrimSuffix(url, ":9501/v1/stats")
	url = strings.TrimPrefix(url, "http://")
	j.stats.address = url
	j.stats.iqn = "iqn.2016-09.com.openebs.jiva:" + volStats.Name

	return j.stats
}
