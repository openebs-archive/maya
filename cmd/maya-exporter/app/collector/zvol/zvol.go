package zvol

import (
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/util"
	zvol "github.com/openebs/maya/pkg/zvol/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
)

// volume implements prometheus.Collector interface
type volume struct {
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
func New() *volume {
	return &volume{
		metrics: newMetrics(),
	}
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
		v.volumeStatus,
		v.rebuildStatus,
		v.inflightIOCount,
		v.rebuildDoneCount,
		v.dispatchedIOCount,
		v.rebuildFailedCount,
		v.zfsCommandErrorCounter,
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
		v.volumeStatus,
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

func (v *volume) get() (zvol.Stats, error) {
	v.Lock()
	defer v.Unlock()
	var (
		err     error
		stdout  []byte
		timeout = 5 * time.Second
		stats   = zvol.Stats{}
	)

	glog.V(2).Info("Run zfs stats command")
	stdout, err = zvol.Run(timeout, runner, "stats")
	if err != nil {
		glog.Errorf("Failed to get zfs stats, error: %v", err)
		return stats, err
	}

	glog.V(2).Infof("Parse stdout of zfs stats command, got stdout: %v", string(stdout))
	stats, err = zvol.StatsParser(stdout)
	if err != nil {
		glog.Errorf("Failed to parse zfs stats command, error: %v, stdout: %v", err, string(stdout))
		return stats, err
	}

	return stats, nil
}

// Collect is implementation of prometheus's prometheus.Collector interface
func (v *volume) Collect(ch chan<- prometheus.Metric) {

	zvolStats, err := v.get()
	if err != nil {
		v.incErrorCounter(err)
		return
	}

	glog.V(2).Infof("Got zfs stats: %#v", zvolStats)
	v.setZVolStats(zvolStats)
	for _, col := range v.collectors() {
		col.Collect(ch)
	}
}

func (v *volume) incErrorCounter(err error) {
	v.zfsCommandErrorCounter.Inc()
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
