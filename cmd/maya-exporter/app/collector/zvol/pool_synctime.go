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
	"os"
	"sync"

	col "github.com/openebs/maya/cmd/maya-exporter/app/collector"
	types "github.com/openebs/maya/pkg/exec"
	zpool "github.com/openebs/maya/pkg/zpool/v1alpha1"
	zvol "github.com/openebs/maya/pkg/zvol/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/klog"
)

// poolMetrics implements prometheus.Collector interface
type poolMetrics struct {
	sync.Mutex
	*poolSyncMetrics
	request bool
	runner  types.Runner
}

// NewPoolSyncMetric returns new instance of poolMetrics
func NewPoolSyncMetric(runner types.Runner) col.Collector {
	return &poolMetrics{
		poolSyncMetrics: newPoolMetrics().
			withZpoolLastSyncTime().
			withZpoolStateUnknown().
			withRequestRejectCounter().
			withzpoolLastSyncTimeCommandError(),
		runner: runner,
	}
}

func (p *poolMetrics) isRequestInProgress() bool {
	return p.request
}

func (p *poolMetrics) setRequestToFalse() {
	p.Lock()
	p.request = false
	p.Unlock()
}

// collectors returns the list of the collectors
func (p *poolMetrics) collectors() []prometheus.Collector {
	return []prometheus.Collector{
		p.zpoolLastSyncTime,
		p.zpoolStateUnknown,
		p.zpoolLastSyncTimeCommandError,
	}
}

// Describe is implementation of Describe method of prometheus.Collector
// interface.
func (p *poolMetrics) Describe(ch chan<- *prometheus.Desc) {
	for _, col := range p.collectors() {
		col.Describe(ch)
	}
}

func (p *poolMetrics) checkError(stdout []byte) *poolfields {

	if zvol.IsNoDataSetAvailable(string(stdout)) || zpool.IsNotAvailable(string(stdout)) {
		pool := poolfields{
			name:                          os.Getenv("HOSTNAME"),
			zpoolLastSyncTime:             zpool.ZpoolLastSyncCommandErrorOrUnknownUnset,
			zpoolStateUnknown:             zpool.ZpoolLastSyncCommandErrorOrUnknownSet,
			zpoolLastSyncTimeCommandError: zpool.ZpoolLastSyncCommandErrorOrUnknownUnset,
		}
		return &pool
	}
	return nil
}

func (p *poolMetrics) get() *poolfields {
	p.Lock()
	defer p.Unlock()

	klog.V(2).Info("Run zfs get io.openebs:livenesstimestamp")
	stdout, err := zvol.Run(p.runner)
	if err != nil {
		pool := poolfields{
			name:                          os.Getenv("HOSTNAME"),
			zpoolLastSyncTime:             zpool.ZpoolLastSyncCommandErrorOrUnknownUnset,
			zpoolStateUnknown:             zpool.ZpoolLastSyncCommandErrorOrUnknownUnset,
			zpoolLastSyncTimeCommandError: zpool.ZpoolLastSyncCommandErrorOrUnknownSet,
		}
		return &pool
	}

	if pool := p.checkError(stdout); pool != nil {
		return pool
	}
	klog.V(2).Infof("Parse stdout of zfs get io.openebs:livenesstimestamp command, got stdout: \n%v", string(stdout))
	pool := poolMetricParser(stdout)

	return pool
}

// Collect is implementation of prometheus's prometheus.Collector interface
func (p *poolMetrics) Collect(ch chan<- prometheus.Metric) {
	p.Lock()
	if p.isRequestInProgress() {
		p.cspiRequestRejectCounter.Inc()
		p.Unlock()
		p.cspiRequestRejectCounter.Collect(ch)
		return
	}
	p.request = true
	p.Unlock()

	pool := p.get()

	if pool == nil {
		p.setRequestToFalse()
		return
	}

	klog.V(2).Infof("Got zfs pool last sync time: %#v", pool)
	p.setPoolStats(pool)
	for _, col := range p.collectors() {
		col.Collect(ch)
	}
	p.setRequestToFalse()

}

func (p *poolMetrics) setPoolStats(poolSyncTime *poolfields) {
	p.zpoolLastSyncTime.WithLabelValues(poolSyncTime.name).Set(poolSyncTime.zpoolLastSyncTime)
	p.zpoolLastSyncTimeCommandError.WithLabelValues(poolSyncTime.name).Set(poolSyncTime.zpoolLastSyncTimeCommandError)
	p.zpoolStateUnknown.WithLabelValues(poolSyncTime.name).Set(poolSyncTime.zpoolStateUnknown)

}
