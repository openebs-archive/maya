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
	}
}

// gaugeVec returns list of Gauge vectors (prometheus's type)
// related to zpool in which values will be set.
// NOTE: Please donot edit the order, add new metrics at the end of
// the list
func (p *pool) gaugeVec() []prometheus.Gauge {
	return []prometheus.Gauge{
		p.size,
		p.status,
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

func (p *pool) get() (zpool.Stats, error) {
	p.Lock()
	defer p.Unlock()
	var (
		err         error
		stdoutZpool []byte
		timeout     = 5 * time.Second
		zpoolStats  = zpool.Stats{}
	)

	glog.V(2).Info("Run zpool list command")
	stdoutZpool, err = zpool.Run(timeout, runner, "list", "-Hp")
	if err != nil {
		glog.Errorf("Failed to get zpool stats, error: %v", err)
		return zpoolStats, err
	}

	glog.V(2).Infof("Parse stdout of zpool list command, stdout: %v", string(stdoutZpool))
	zpoolStats, err = zpool.ListParser(stdoutZpool)
	if err != nil {
		glog.Errorf("Failed to parse zpool list command, error: %v, stdout: %v", err, string(stdoutZpool))
		return zpoolStats, err
	}

	return zpoolStats, nil
}

// Collect is implementation of prometheus's prometheus.Collector interface
func (p *pool) Collect(ch chan<- prometheus.Metric) {

	poolStats := statsFloat64{}
	zpoolStats, err := p.get()
	if err != nil {
		p.incErrorCounter(err)
		return
	}

	glog.V(2).Infof("Got zpool stats: %#v", zpoolStats)
	poolStats.parse(zpoolStats, p)
	p.setZPoolStats(poolStats)
	for _, col := range p.collectors() {
		col.Collect(ch)
	}
}

func (p *pool) incErrorCounter(err error) {
	p.zpoolCommandErrorCounter.Inc()
}

func (p *pool) setZPoolStats(stats statsFloat64) {
	for index, col := range p.gaugeVec() {
		items := stats.List()
		col.Set(items[index])
	}
}
