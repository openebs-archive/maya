package pool

import (
	"strconv"

	"github.com/golang/glog"
	zpool "github.com/openebs/maya/pkg/zpool/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
)

type metrics struct {
	readBytes  *prometheus.GaugeVec
	writeBytes *prometheus.GaugeVec

	syncCount   *prometheus.GaugeVec
	syncLatency *prometheus.GaugeVec

	readLatency  *prometheus.GaugeVec
	writeLatency *prometheus.GaugeVec

	volumeStatus *prometheus.GaugeVec

	inflightIOCount   *prometheus.GaugeVec
	dispatchedIOCount *prometheus.GaugeVec

	rebuildCount       *prometheus.GaugeVec
	rebuildBytes       *prometheus.GaugeVec
	rebuildStatus      *prometheus.GaugeVec
	rebuildDoneCount   *prometheus.GaugeVec
	rebuildFailedCount *prometheus.GaugeVec

	size                prometheus.Gauge
	status              prometheus.Gauge
	usedCapacity        prometheus.Gauge
	freeCapacity        prometheus.Gauge
	usedCapacityPercent prometheus.Gauge

	parseErrorCounter   prometheus.Gauge
	commandErrorCounter prometheus.Gauge
}

type statsFloat64 struct {
	status              float64
	size                float64
	used                float64
	free                float64
	usedCapacityPercent float64
}

// List returns list of type float64 of various stats
// NOTE: Please donot change the order, add the new stats
// at the end of the list.
func (s *statsFloat64) List() []float64 {
	return []float64{
		s.size,
		s.status,
		s.used,
		s.free,
		s.usedCapacityPercent,
	}
}

func parseFloat64(e string, m *metrics) float64 {
	num, err := strconv.ParseFloat(e, 64)
	if err != nil {
		glog.Error("failed to parse, err: ", err)
		m.parseErrorCounter.Inc()
	}
	return num
}

func (s *statsFloat64) parse(stats zpool.Stats, p *pool) {
	s.size = parseFloat64(stats.Size, &p.metrics)
	s.used = parseFloat64(stats.Used, &p.metrics)
	s.free = parseFloat64(stats.Free, &p.metrics)
	s.status = zpool.Status[stats.Status]
	s.usedCapacityPercent = parseFloat64(stats.UsedCapacityPercent, &p.metrics)
}

// Metrics initializes fields of the metrics and returns its instance
func Metrics() metrics {
	return metrics{
		readBytes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "total_read_bytes",
				Help:      "Total read in bytes",
			},
			[]string{"vol", "pool"},
		),

		writeBytes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "total_write_bytes",
				Help:      "Total write in bytes",
			},
			[]string{"vol", "pool"},
		),

		syncCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "sync_count",
				Help:      "Total no of sync on replica",
			},
			[]string{"vol", "pool"},
		),

		syncLatency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "sync_latency",
				Help:      "Sync latency on replica",
			},
			[]string{"vol", "pool"},
		),

		readLatency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "read_latency",
				Help:      "Read latency on replica",
			},
			[]string{"vol", "pool"},
		),

		writeLatency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "write_latency",
				Help:      "Write latency on replica",
			},
			[]string{"volName", "castype"},
		),

		volumeStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "volume_status",
				Help:      `Status of volume (0, 1, 2, 3) = {"Offline", "Healthy", "Degraded", "Rebuilding"}`,
			},
			[]string{"vol", "pool"},
		),

		inflightIOCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "inflight_io_count",
				Help:      "Inflight IO's count",
			},
			[]string{"vol", "pool"},
		),

		dispatchedIOCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "dispatched_io_count",
				Help:      "Dispatched IO's count",
			},
			[]string{"vol", "pool"},
		),

		rebuildCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "rebuild_count",
				Help:      "Rebuild count",
			},
			[]string{"vol", "pool"},
		),

		rebuildBytes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "rebuild_bytes",
				Help:      "Rebuild bytes",
			},
			[]string{"vol", "pool"},
		),

		rebuildStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "rebuild_status",
				Help:      `Status of rebuild on replica (0, 1, 2, 3, 4, 5, 6)= {"INIT", "DONE", "SNAP REBUILD INPROGRESS", "ACTIVE DATASET REBUILD INPROGRESS", "ERRORED", "FAILED", "UNKNOWN"}`,
			},
			[]string{"vol", "pool"},
		),

		rebuildDoneCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "total_rebuild_done",
				Help:      "Total no of rebuild done on replica",
			},
			[]string{"vol", "pool"},
		),

		rebuildFailedCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "total_failed_rebuild",
				Help:      "Total no of failed rebuilds on replica",
			},
			[]string{"vol", "pool"},
		),

		size: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "pool_size",
				Help:      "Size of pool",
			},
		),

		status: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "pool_status",
				Help:      `Status of pool (0, 1, 2, 3, 4, 5, 6)= {"Offline", "Online", "Degraded", "Faulted", "Removed", "Unavail", "NoPoolsAvailable"}`,
			},
		),

		usedCapacity: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "used_pool_capacity",
				Help:      "Capacity used by pool",
			},
		),

		freeCapacity: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "free_pool_capacity",
				Help:      "Free capacity in pool",
			},
		),

		usedCapacityPercent: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "used_pool_capacity_percent",
				Help:      "Capacity used by pool in percent",
			},
		),

		parseErrorCounter: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "parse_error_total",
				Help:      "Total no of parsing errors",
			},
		),

		commandErrorCounter: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "command_error",
				Help:      "Command error counter (zfs/zpool)",
			},
		),
	}
}
