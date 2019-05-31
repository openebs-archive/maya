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

package zvol

import (
	"strings"
	"sync"

	"github.com/golang/glog"
	col "github.com/openebs/maya/cmd/maya-exporter/app/collector"
	types "github.com/openebs/maya/pkg/exec"
	zvol "github.com/openebs/maya/pkg/zvol/v1alpha1"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

// volume implements prometheus.Collector interface
type volume struct {
	sync.Mutex
	*metrics
	request bool
	runner  types.Runner
}

// New returns new instance of pool
func New(runner types.Runner) col.Collector {
	return &volume{
		metrics: newMetrics().withReadBytes().
			withWriteBytes().
			withReadCount().
			withWriteCount().
			withSyncCount().
			withSyncLatency().
			withReadLatency().
			withWriteLatency().
			withReplicaStatus().
			withinflightIOCount().
			withDispatchedIOCount().
			withRebuildCount().
			withRebuildBytes().
			withRebuildStatus().
			withRebuildDone().
			withFailedRebuild().
			withCommandErrorCounter().
			withParseErrorCounter().
			withRequestRejectCounter().
			withNoDatasetAvailableErrorCounter().
			withInitializeLibuzfsClientErrorCounter(),
		runner: runner,
	}
}

func (v *volume) isRequestInProgress() bool {
	return v.request
}

func (v *volume) setRequestToFalse() {
	v.Lock()
	v.request = false
	v.Unlock()
}

// collectors returns the list of the collectors
func (v *volume) collectors() []prometheus.Collector {
	return []prometheus.Collector{
		v.syncCount,
		v.readCount,
		v.writeCount,
		v.readBytes,
		v.writeBytes,
		v.syncLatency,
		v.readLatency,
		v.writeLatency,
		v.rebuildCount,
		v.rebuildBytes,
		v.replicaStatus,
		v.rebuildStatus,
		v.inflightIOCount,
		v.rebuildDoneCount,
		v.dispatchedIOCount,
		v.rebuildFailedCount,
		v.zfsCommandErrorCounter,
		v.zfsStatsParseErrorCounter,
		v.zfsStatsRejectRequestCounter,
	}
}

// gaugeVec returns list of zfs Gauge vectors (prometheus's type)
// in which values will be set.
// NOTE: Please donot edit the order, add new metrics at the end
// of the list
func (v *volume) gaugeVec() []*prometheus.GaugeVec {
	return []*prometheus.GaugeVec{
		v.syncCount,
		v.readCount,
		v.writeCount,
		v.readBytes,
		v.writeBytes,
		v.syncLatency,
		v.readLatency,
		v.writeLatency,
		v.rebuildCount,
		v.rebuildBytes,
		v.inflightIOCount,
		v.rebuildDoneCount,
		v.dispatchedIOCount,
		v.rebuildFailedCount,
		v.replicaStatus,
		v.rebuildStatus,
	}
}

// Describe is implementation of Describe method of prometheus.Collector
// interface.
func (v *volume) Describe(ch chan<- *prometheus.Desc) {
	for _, col := range v.collectors() {
		col.Describe(ch)
	}
}

func (v *volume) checkError(stdout []byte, ch chan<- prometheus.Metric) error {
	if zvol.IsNotInitialized(string(stdout)) {
		v.zfsStatsInitializeLibuzfsClientErrorCounter.Inc()
		v.zfsStatsInitializeLibuzfsClientErrorCounter.Collect(ch)
		return errors.New(zvol.InitializeLibuzfsClientErr.String())
	}

	if zvol.IsNoDataSetAvailable(string(stdout)) {
		v.zfsStatsNoDataSetAvailableErrorCounter.Inc()
		v.zfsStatsNoDataSetAvailableErrorCounter.Collect(ch)
		return errors.New(zvol.NoDataSetAvailable.String())
	}
	return nil
}

func (v *volume) get(ch chan<- prometheus.Metric) (zvol.Stats, error) {
	var (
		err    error
		stdout []byte
		stats  = zvol.Stats{}
	)

	glog.V(2).Info("Run zfs stats command")
	stdout, err = zvol.Run(v.runner)
	if err != nil {
		v.zfsCommandErrorCounter.Inc()
		v.zfsCommandErrorCounter.Collect(ch)
		return stats, err
	}

	err = v.checkError(stdout, ch)
	if err != nil {
		return stats, err
	}

	glog.V(2).Infof("Parse stdout of zfs stats command, got stdout: %v", string(stdout))
	stats, err = zvol.StatsParser(stdout)
	if err != nil {
		v.zfsStatsParseErrorCounter.Inc()
		v.zfsStatsParseErrorCounter.Collect(ch)
		return stats, err
	}

	return stats, nil
}

// Collect is implementation of prometheus's prometheus.Collector interface
func (v *volume) Collect(ch chan<- prometheus.Metric) {
	v.Lock()
	if v.isRequestInProgress() {
		v.zfsStatsRejectRequestCounter.Inc()
		v.Unlock()
		v.zfsStatsRejectRequestCounter.Collect(ch)
		return

	}

	v.request = true
	v.Unlock()

	zvolStats, err := v.get(ch)
	if err != nil {
		v.setRequestToFalse()
		return
	}

	glog.V(2).Infof("Got zfs stats: %#v", zvolStats)
	v.setZVolStats(zvolStats)
	for _, col := range v.collectors() {
		col.Collect(ch)
	}
	v.setRequestToFalse()
}

func (v *volume) setZVolStats(stats zvol.Stats) {
	for _, vol := range stats.Volumes {
		s := strings.Split(vol.Name, "/")
		poolName, volname := s[0], s[1]
		items := zvol.StatsList(vol)
		for index, col := range v.gaugeVec() {
			col.WithLabelValues(volname, poolName).Set(items[index])
		}
	}
}
