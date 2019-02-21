package pool

import (
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/exporter/v1alpha1/zfs"
	"github.com/openebs/maya/pkg/exporter/v1alpha1/zpool"
	"github.com/openebs/maya/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
)

// pool implements prometheus.Collector interface
type pool struct {
	metrics
}

var (
	Runner util.Runner
)

func InitVar() {
	Runner = util.RealRunner{}
}

func New() *pool {
	return &pool{
		metrics: Metrics(),
	}
}

// GetInitStatus run zpool binary to verify whether zpool container
// has started.
func (p *pool) GetInitStatus() {

	for {
		_, err := Runner.RunCombinedOutput("zpool", "status")
		if err != nil {
			glog.Warningf("Failed to get zpool status, error: %v, retry after 2s", err)
			time.Sleep(2 * time.Second)
			continue
		}
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

func (p *pool) zpoolGaugeVec() []prometheus.Gauge {
	return []prometheus.Gauge{
		p.size,
		p.status,
		p.usedCapacity,
		p.freeCapacity,
		p.usedCapacityPercent,
	}
}

func (p *pool) zfsGaugeVec() []*prometheus.GaugeVec {
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

func (p *pool) Describe(ch chan<- *prometheus.Desc) {
	for _, col := range p.collectors() {
		col.Describe(ch)
	}
}

func (p *pool) get() (zfs.Stats, zpool.Stats, error) {
	var (
		err                    error
		zfsStats               = zfs.Stats{}
		zpoolStats             = zpool.Stats{}
		stdoutZFS, stdoutZpool []byte
	)

	stdoutZFS, err = zfs.Run(Runner, "stats")
	if err != nil {
		glog.Errorf("Failed to get zfs stats, error: %v", err)
		return zfs.Stats{}, zpool.Stats{}, err
	}

	zfsStats, err = zfs.StatsParser(stdoutZFS)
	if err != nil {
		glog.Errorf("Failed to parse zfs stats command, stdout: %#v, error: %v", string(stdoutZFS), err)
		return zfsStats, zpoolStats, &colErr{err}
	}

	stdoutZpool, err = zpool.Run(Runner, "list", "-Hp")
	if err != nil {
		glog.Errorf("Failed to get zpool stats, error: %v", err)
		return zfsStats, zpoolStats, err
	}

	zpoolStats, err = zpool.ListParser(stdoutZpool)
	if err != nil {
		glog.Errorf("Failed to parse zpool list command, stdout: %#v, error: %v", string(stdoutZpool), err)
		return zfsStats, zpoolStats, &colErr{err}
	}

	return zfsStats, zpoolStats, nil

}

func (p *pool) Collect(ch chan<- prometheus.Metric) {

	poolStats := statsFloat64{}
	zfsStats, zpoolStats, err := p.get()
	if err != nil {
		p.incErrorCounter(err)
		return
	}

	poolStats.parse(zpoolStats, p)
	p.setZFSStats(zfsStats)
	p.setZPoolStats(poolStats)
	for _, col := range p.collectors() {
		col.Collect(ch)
	}
}

func (p *pool) incErrorCounter(err error) {
	if _, ok := err.(*colErr); ok {
		p.commandErrorCounter.Inc()
	}
}

func (p *pool) setZPoolStats(stats statsFloat64) {
	for index, col := range p.zpoolGaugeVec() {
		items := stats.List()
		col.Set(items[index])
	}
}

func (p *pool) setZFSStats(stats zfs.Stats) {
	for _, vol := range stats.Volumes {
		poolName, volname := split(vol.Name)
		items := zfs.StatsList(vol)
		for index, col := range p.zfsGaugeVec() {
			col.WithLabelValues(volname, poolName).Set(items[index])
		}
	}
}

func split(str string) (string, string) {
	newStr := strings.Split(str, "/")
	return newStr[0], newStr[1]
}
