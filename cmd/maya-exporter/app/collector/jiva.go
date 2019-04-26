// Copyright © 2017-2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

	v1 "github.com/openebs/maya/pkg/stats/v1alpha1"
	"github.com/pkg/errors"

	"github.com/golang/glog"
)

// jiva implements the Volume interface. It exposes
// the metrics of a OpenEBS (Jiva) volume.
type jiva struct {
	// url is jiva controller's url
	url string
}

// Jiva returns jiva's instance
func Jiva(url *url.URL) *jiva {
	url.Path = "/v1/stats"
	return &jiva{
		url: url.String(),
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
		return &colErr{
			errors.Errorf("%v", err),
		}
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	glog.V(2).Info("Got response: ", string(body))
	err = json.Unmarshal(body, &obj)

	if err != nil {
		return err
	}

	resp.Body.Close()
	return nil
}

func (j *jiva) parse(volStats v1.VolumeStats, metrics *metrics) stats {
	var stats = stats{}
	if !volStats.Got {
		glog.Warningf("%s", "can't parse stats, controller may not be reachable")
		return stats
	}
	stats.got = true
	stats.casType = "jiva"
	stats.reads = parseFloat64(volStats.Reads, metrics)
	stats.writes = parseFloat64(volStats.Writes, metrics)
	stats.totalReadTime = parseFloat64(volStats.TotalReadTime, metrics)
	stats.totalWriteTime = parseFloat64(volStats.TotalWriteTime, metrics)
	stats.totalReadBlockCount = parseFloat64(volStats.TotalReadBlockCount, metrics)
	stats.totalWriteBlockCount = parseFloat64(volStats.TotalWriteBlockCount, metrics)

	stats.sectorSize = parseFloat64(volStats.SectorSize, metrics)

	uBlocks := parseFloat64(volStats.UsedBlocks, metrics)
	uBlocks = uBlocks * stats.sectorSize
	stats.logicalSize, _ = v1.DivideFloat64(uBlocks, v1.BytesToGB)
	aUsed := parseFloat64(volStats.UsedLogicalBlocks, metrics)
	aUsed = aUsed * stats.sectorSize
	stats.actualSize, _ = v1.DivideFloat64(aUsed, v1.BytesToGB)
	size := parseFloat64(volStats.Size, metrics)
	stats.size, _ = v1.DivideFloat64(size, v1.BytesToGB)
	stats.totalReplicaCount = parseFloat64(volStats.ReplicaCounter, metrics)
	stats.revisionCount = parseFloat64(volStats.RevisionCounter, metrics)
	stats.uptime = parseFloat64(volStats.UpTime, metrics)
	stats.name = volStats.Name
	stats.replicas = volStats.Replicas
	stats.status = volStats.TargetStatus
	url := j.url
	url = strings.TrimSuffix(url, port+endpoint)
	url = strings.TrimPrefix(url, protocol)
	stats.address = url
	stats.iqn = jivaIQN + volStats.Name

	return stats
}
