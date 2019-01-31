package collector

import (
	"github.com/golang/glog"
	v1 "github.com/openebs/maya/pkg/stats/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
)

// collector implements prometheus.Collector interface
type collector struct {
	Volume
	metrics
}

// New returns an instance of collector
func New(vol Volume) *collector {
	t := casType(vol)
	if t == "" {
		// only jiva and cstor are supported casType
		glog.Fatal("exiting...")
	}
	return &collector{
		vol,
		Metrics(t),
	}
}

func casType(vol Volume) string {
	switch t := vol.(type) {
	case *jiva:
		return "jiva"
	case *cstor:
		return "cstor"
	default:
		glog.Error("Unknown cas type: ", t)
		return ""
	}
}

// collectors returns the list of the collectors
func (c *collector) collectors() []prometheus.Collector {
	return []prometheus.Collector{
		c.reads,
		c.writes,
		c.totalReadBytes,
		c.totalWriteBytes,
		c.totalReadTime,
		c.totalWriteTime,
		c.totalReadBlockCount,
		c.totalWriteBlockCount,
		c.actualUsed,
		c.logicalSize,
		c.sectorSize,
		c.sizeOfVolume,
		c.volumeStatus,
		c.connectionErrorCounter,
		c.connectionRetryCounter,
		c.parseErrorCounter,
		c.totalReplicaCounter,
		c.degradedReplicaCounter,
		c.healthyReplicaCounter,
		c.volumeUpTime,
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector to the provided channel and returns once
// the last descriptor has been sent. The sent descriptors fulfill the
// consistency and uniqueness requirements described in the Desc
// documentation. (It is valid if one and the same Collector sends
// duplicate descriptors. Those duplicates are simply ignored. However,
// two different Collectors must not send duplicate descriptors.) This
// method idempotently sends the same descriptors throughout the
// lifetime of the Collector. If a Collector encounters an error while
// executing this method, it must send an invalid descriptor (created
// with NewInvalidDesc) to signal the error to the registry.
//
// Describe implements Describe method of prometheus.Collector interface.
func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	for _, col := range c.collectors() {
		col.Describe(ch)
	}
}

// Collect is called by the Prometheus registry when collecting
// metrics. The implementation sends each collected metric via the
// provided channel and returns once the last metric has been sent. The
// descriptor of each sent metric is one of those returned by
// Describe. Returned metrics that share the same descriptor must differ
// in their variable label values. This method may be called
// way. Blocking occurs at the expense of total performance of rendering
// concurrently and must therefore be implemented in a concurrency safe
// all registered metrics. Ideally, Collector implementations support
// concurrent readers.
//
// For exp: openebs_volume_uptime{"volName":"", "casType":""} 0
//          openebs_volume_uptime{"volName":"vol1", "casType":"jiva"} 30
//
// In the above case metrics are unique and not duplicated because of
// the labels. In our case we have several collectors such as gauges and
// counters. Each collectors are unique at any instance.
//
// Collect implements Collect method of prometheus.Collector interface.
func (c *collector) Collect(ch chan<- prometheus.Metric) {
	var (
		err         error
		volumeStats v1.VolumeStats
		stats       stats
	)

	metrics := &c.metrics
	if volumeStats, err = c.get(); err != nil {
		glog.Errorln(err)
		c.setError(err)
	}

	stats = c.parse(volumeStats, metrics)

	c.set(stats)

	// collect the metrics extracted by collect method
	for _, col := range c.collectors() {
		col.Collect(ch)
	}
}

func (c *collector) setError(err error) {
	c.connectionErrorCounter.Inc()
	if _, ok := err.(*colErr); ok {
		c.connectionRetryCounter.Inc()
	}
}

// set is used to set the values gathered from cas to
// prometheus gauges and counters.
func (c *collector) set(volStats stats) {
	//	var replicaAddress, replicaMode strings.Builder

	if !volStats.got {
		glog.Warningf("%s", "setting up empty values")
	}
	c.reads.Set(volStats.reads)
	c.totalReadTime.Set(volStats.totalReadTime)
	c.writes.Set(volStats.writes)
	c.totalWriteTime.Set(volStats.totalWriteTime)
	c.totalReadBytes.Set(volStats.totalReadBytes)
	c.totalWriteBytes.Set(volStats.totalWriteBytes)
	c.totalReadBlockCount.Set(volStats.totalReadBlockCount)
	c.totalWriteBlockCount.Set(volStats.totalWriteBlockCount)
	c.sectorSize.Set(volStats.sectorSize)
	c.logicalSize.Set(volStats.logicalSize)
	c.actualUsed.Set(volStats.actualSize)
	c.sizeOfVolume.Set(volStats.size)

	c.volumeUpTime.WithLabelValues(
		volStats.name,
		volStats.casType,
	).Set(volStats.uptime)

	//	volStats.buildStringof(&replicaAddress, &replicaMode)
	volStats.getReplicaCount()
	c.totalReplicaCounter.Set(volStats.totalReplicaCount)
	c.degradedReplicaCounter.Set(volStats.degradedReplicaCount)
	c.healthyReplicaCounter.Set(volStats.healthyReplicaCount)

	c.volumeStatus.Set(float64(volStats.getVolumeStatus()))

	return
}
