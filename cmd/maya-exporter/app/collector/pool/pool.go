package pool

import (
	"github.com/openebs/maya/pkg/exporter/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
)

// pool implements prometheus.Collector interface
type pool struct {
	metrics
}

func New() *pool {
	return &pool{
		metrics: Metrics(),
	}
}

// collectors returns the list of the collectors
func (p *pool) collectors() []prometheus.Collector {
	return []prometheus.Collector{
		p.reads,
		p.writes,
		p.status,
		p.capacity,
		p.syncCount,
		p.readBytes,
		p.writeBytes,
		p.syncLatency,
		p.readLatency,
		p.writeLatency,
		p.usedCapacity,
		p.freeCapacity,
		p.rebuildCount,
		p.rebuildBytes,
		p.rebuildStatus,
		p.replicaStatus,
		p.inflightIOCount,
		p.rebuildDoneCount,
		p.dispatchedIOCount,
		p.rebuildFailedCount,
		p.usedCapacityPercent,
		p.connectionErrorCounter,
		p.connectionRetryCounter,
	}
}

func (p *pool) Describe(ch chan<- *prometheus.Desc) {
	for _, col := range p.collectors() {
		col.Describe(ch)
	}
}

func (p *pool) Collect(ch chan<- prometheus.Metric) {
	p.get()
	p.set()
	for _, col := range p.collectors() {
		col.Collect(ch)
	}
}

func (p *pool) setError(err error) {
	p.connectionErrorCounter.Inc()
	if _, ok := err.(*v1alpha1.CollectorError); ok {
		p.connectionRetryCounter.Inc()
	}
}

func (p *pool) set() {
}

func (p *pool) get() {
}
