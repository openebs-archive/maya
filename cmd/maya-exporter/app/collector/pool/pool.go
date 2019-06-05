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

package pool

import (
	"errors"
	"sync"

	"github.com/golang/glog"
	col "github.com/openebs/maya/cmd/maya-exporter/app/collector"
	types "github.com/openebs/maya/pkg/exec"
	zpool "github.com/openebs/maya/pkg/zpool/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
)

// pool implements prometheus.Collector interface
type pool struct {
	sync.Mutex
	*metrics
	request bool
	runner  types.Runner
}

// New returns new instance of pool
func New(runner types.Runner) col.Collector {
	return &pool{
		metrics: newMetrics().
			withSize().
			withStatus().
			withUsedCapacity().
			withFreeCapacity().
			withUsedCapacityPercent().
			withParseErrorCounter().
			withRejectRequestCounter().
			withCommandErrorCounter().
			withNoPoolAvailableErrorCounter().
			withIncompleteOutputErrorCounter().
			withInitializeLibuzfsClientErrorCounter(),
		runner: runner,
	}
}

func (p *pool) isRequestInProgress() bool {
	return p.request
}

func (p *pool) setRequestToFalse() {
	p.Lock()
	p.request = false
	p.Unlock()
}

// collectors returns the list of the collectors
func (p *pool) collectors() []prometheus.Collector {
	return []prometheus.Collector{
		p.size,
		p.status,
		p.usedCapacity,
		p.freeCapacity,
		p.usedCapacityPercent,
		p.zpoolCommandErrorCounter,
		p.zpoolRejectRequestCounter,
		p.zpoolListParseErrorCounter,
		p.noPoolAvailableErrorCounter,
		p.inCompleteOutputErrorCounter,
		p.initializeLibUZFSClientErrorCounter,
	}
}

// gaugeVec returns list of Gauge vectors (prometheus's type)
// related to zpool in which values will be set.
// NOTE: Please donot edit the order, add new metrics at the end of
// the list
func (p *pool) gaugeVec() []prometheus.Gauge {
	return []prometheus.Gauge{
		p.size,
		p.usedCapacity,
		p.freeCapacity,
		p.usedCapacityPercent,
	}
}

// Describe is implementation of Describe method of prometheus.Collector
// interface.
func (p *pool) Describe(ch chan<- *prometheus.Desc) {
	for _, col := range p.collectors() {
		col.Describe(ch)
	}
}

func (p *pool) checkError(stdout []byte, ch chan<- prometheus.Metric) error {
	if zpool.IsNotInitialized(string(stdout)) {
		p.initializeLibUZFSClientErrorCounter.Inc()
		p.initializeLibUZFSClientErrorCounter.Collect(ch)
		return errors.New(zpool.InitializeLibuzfsClientErr.String())
	}

	if zpool.IsNotAvailable(string(stdout)) {
		p.noPoolAvailableErrorCounter.Inc()
		p.noPoolAvailableErrorCounter.Collect(ch)
		return errors.New(zpool.NoPoolAvailable.String())
	}
	return nil
}

func (p *pool) getZpoolStats(ch chan<- prometheus.Metric) (zpool.Stats, error) {
	var (
		err         error
		stdoutZpool []byte
		zpoolStats  = zpool.Stats{}
	)

	glog.V(2).Info("Run zpool list command")
	stdoutZpool, err = zpool.Run(p.runner)
	if err != nil {
		p.zpoolCommandErrorCounter.Inc()
		p.zpoolCommandErrorCounter.Collect(ch)
		return zpoolStats, err
	}

	err = p.checkError(stdoutZpool, ch)
	if err != nil {
		return zpoolStats, err
	}

	glog.V(2).Infof("Parse stdout of zpool list command, stdout: %v", string(stdoutZpool))
	zpoolStats, err = zpool.ListParser(stdoutZpool)
	if err != nil {
		p.inCompleteOutputErrorCounter.Inc()
		p.inCompleteOutputErrorCounter.Collect(ch)
		return zpoolStats, err
	}

	return zpoolStats, nil
}

// Collect is implementation of prometheus's prometheus.Collector interface
func (p *pool) Collect(ch chan<- prometheus.Metric) {

	p.Lock()
	if p.isRequestInProgress() {
		p.zpoolRejectRequestCounter.Inc()
		p.Unlock()
		p.zpoolRejectRequestCounter.Collect(ch)
		return
	}

	p.request = true
	p.Unlock()

	poolStats := statsFloat64{}
	zpoolStats, err := p.getZpoolStats(ch)
	if err != nil {
		p.setRequestToFalse()
		return
	}
	glog.V(2).Infof("Got zpool stats: %#v", zpoolStats)
	poolStats.parse(zpoolStats, p)
	p.setZPoolStats(poolStats, zpoolStats.Name)
	for _, col := range p.collectors() {
		col.Collect(ch)
	}
	p.setRequestToFalse()
}

func (p *pool) setZPoolStats(stats statsFloat64, name string) {
	items := stats.List()
	for index, col := range p.gaugeVec() {
		col.Set(items[index])
	}
	p.status.WithLabelValues(name).Set(stats.status)
}
