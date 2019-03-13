package pool

import (
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/util"
	zpool "github.com/openebs/maya/pkg/zpool/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
)

// pool implements prometheus.Collector interface
type pool struct {
	sync.Mutex
	metrics
	request bool
}

var (
	// runner variable is used for executing binaries
	runner util.Runner
)

// InitVar initialize runner variable
func InitVar() {
	runner = util.RealRunner{}
}

// New returns new instance of pool
func New() *pool {
	return &pool{
		metrics: newMetrics(),
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

// GetInitStatus run zpool binary to verify whether zpool container
// has started.
func (p *pool) GetInitStatus(timeout time.Duration) {

	for {
		stdout, err := zpool.Run(timeout, runner, "status")
		if err != nil {
			glog.Warningf("Failed to get zpool status, error: %v, pool container may be initializing,  retry after 2s", err)
			time.Sleep(2 * time.Second)
			continue
		}
		str := string(stdout)
		if zpool.IsNotAvailable(str) {
			glog.Warning("No pool available, pool must be creating, retry after 3s")
			time.Sleep(3 * time.Second)
			continue
		}
		glog.Info("\n", string(stdout))
		break
	}
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
		p.zpoolListparseErrorCounter,
		p.noPoolAvailableErrorCounter,
		p.incompleteOutputErrorCounter,
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

func (p *pool) getZpoolStats(ch chan<- prometheus.Metric) (zpool.Stats, error) {
	var (
		err         error
		stdoutZpool []byte
		timeout     = 30 * time.Second
		zpoolStats  = zpool.Stats{}
	)

	stdoutZpool, err = zpool.Run(timeout, runner, "list", "-Hp")
	if err != nil {
		p.zpoolCommandErrorCounter.Inc()
		p.zpoolCommandErrorCounter.Collect(ch)
		return zpoolStats, err
	}

	glog.V(2).Infof("Parse stdout of zpool list command, stdout: %v", string(stdoutZpool))
	zpoolStats, err = zpool.ListParser(stdoutZpool)
	if err != nil {
		if err.Error() == string(zpool.NoPoolAvailable) {
			p.noPoolAvailableErrorCounter.Inc()
			p.noPoolAvailableErrorCounter.Collect(ch)
		} else {
			p.incompleteOutputErrorCounter.Inc()
			p.incompleteOutputErrorCounter.Collect(ch)
		}

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
	poolStats.parse(zpoolStats, p, ch)
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
