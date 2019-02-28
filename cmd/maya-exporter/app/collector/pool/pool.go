package pool

import (
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/util"
	zpool "github.com/openebs/maya/pkg/zpool/v1alpha1"
	zvol "github.com/openebs/maya/pkg/zvol/v1alpha1"
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
		metrics: Metrics(),
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
		if !zpool.IsAvailable(str) {
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
		p.volumeStatus,
		p.rebuildStatus,
		p.inflightIOCount,
		p.rebuildDoneCount,
		p.dispatchedIOCount,
		p.rebuildFailedCount,
		p.usedCapacityPercent,
		p.commandErrorCounter,
	}
}

// zpoolGaugeVec returns list of Gauge vectors (prometheus's type)
// related to zpool in which values will be set.
// NOTE: Please donot edit the order, add new metrics at the end of
// the list
func (p *pool) zpoolGaugeVec() []prometheus.Gauge {
	return []prometheus.Gauge{
		p.size,
		p.status,
		p.usedCapacity,
		p.freeCapacity,
		p.usedCapacityPercent,
	}
}

// zfsGaugeVec returns list of zfs Gauge vectors (prometheus's type)
// in which values will be set.
// NOTE: Please donot edit the order, add new metrics at the end
// of the list
func (p *pool) zvolGaugeVec() []*prometheus.GaugeVec {
	return []*prometheus.GaugeVec{
		p.syncCount,
		p.readBytes,
		p.writeBytes,
		p.syncLatency,
		p.readLatency,
		p.writeLatency,
		p.rebuildCount,
		p.rebuildBytes,
		p.inflightIOCount,
		p.rebuildDoneCount,
		p.dispatchedIOCount,
		p.rebuildFailedCount,
		p.volumeStatus,
		p.rebuildStatus,
	}
}

// Describe is implementation of Describe method of prometheus.Collector
// interface.
func (p *pool) Describe(ch chan<- *prometheus.Desc) {
	for _, col := range p.collectors() {
		col.Describe(ch)
	}
}

func (p *pool) get() (zvol.Stats, zpool.Stats, error) {
	p.Lock()
	defer p.Unlock()
	var (
		err                    error
		stdoutZFS, stdoutZpool []byte
		timeout                = 5 * time.Second
		zvolStats, zpoolStats  = zvol.Stats{}, zpool.Stats{}
	)

	stdoutZFS, err = zvol.Run(timeout, runner, "stats")
	if err != nil {
		glog.Errorf("Failed to get zfs stats, error: %v, stdout: %v", err)
		return zvolStats, zpoolStats, err
	}

	zvolStats, err = zvol.StatsParser(stdoutZFS)
	if err != nil {
		glog.Errorln("Failed to parse zfs stats command, error: "+err.Error(), "stdout: "+string(stdoutZFS))
		return zvolStats, zpoolStats, err
	}

	stdoutZpool, err = zpool.Run(timeout, runner, "list", "-Hp")
	if err != nil {
		glog.Errorf("Failed to get zpool stats, error: %v, stdout: %v", err)
		return zvolStats, zpoolStats, err
	}

	zpoolStats, err = zpool.ListParser(stdoutZpool)
	if err != nil {
		glog.Errorln("Failed to parse zpool list command, error: "+err.Error(), "stdout: "+string(stdoutZpool))
		return zvolStats, zpoolStats, err
	}

	return zvolStats, zpoolStats, nil
}

// Collect is implementation of prometheus's prometheus.Collector interface
func (p *pool) Collect(ch chan<- prometheus.Metric) {

	poolStats := statsFloat64{}
	zvolStats, zpoolStats, err := p.get()
	if err != nil {
		p.incErrorCounter(err)
		return
	}

	poolStats.parse(zpoolStats, p)
	p.setZVolStats(zvolStats)
	p.setZPoolStats(poolStats)
	for _, col := range p.collectors() {
		col.Collect(ch)
	}
}

func (p *pool) incErrorCounter(err error) {
	p.commandErrorCounter.Inc()
}

func (p *pool) setZPoolStats(stats statsFloat64) {
	for index, col := range p.zpoolGaugeVec() {
		items := stats.List()
		col.Set(items[index])
	}
}

func (p *pool) setZVolStats(stats zvol.Stats) {
	for _, vol := range stats.Volumes {
		poolName, volname := split(vol.Name)
		items := zvol.StatsList(vol)
		for index, col := range p.zvolGaugeVec() {
			col.WithLabelValues(volname, poolName).Set(items[index])
		}
	}
}

func split(str string) (string, string) {
	newStr := strings.Split(str, "/")
	return newStr[0], newStr[1]
}
