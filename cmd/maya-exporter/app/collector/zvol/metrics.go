package zvol

import (
	"strconv"
	"strings"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
)

type metrics struct {
	readBytes  *prometheus.GaugeVec
	writeBytes *prometheus.GaugeVec
	readCount  *prometheus.GaugeVec
	writeCount *prometheus.GaugeVec

	syncCount   *prometheus.GaugeVec
	syncLatency *prometheus.GaugeVec

	readLatency  *prometheus.GaugeVec
	writeLatency *prometheus.GaugeVec

	replicaStatus *prometheus.GaugeVec

	inflightIOCount   *prometheus.GaugeVec
	dispatchedIOCount *prometheus.GaugeVec

	rebuildCount       *prometheus.GaugeVec
	rebuildBytes       *prometheus.GaugeVec
	rebuildStatus      *prometheus.GaugeVec
	rebuildDoneCount   *prometheus.GaugeVec
	rebuildFailedCount *prometheus.GaugeVec

	zfsCommandErrorCounter prometheus.Gauge
}

type listMetrics struct {
	used      *prometheus.GaugeVec
	available *prometheus.GaugeVec

	zfsParseErrorCounter       prometheus.Gauge
	zfsListCommandErrorCounter prometheus.Gauge
}

type fields struct {
	name      string
	used      float64
	available float64
}

// newMetrics initializes fields of the metrics and returns its instance
func newListMetrics() listMetrics {
	return listMetrics{

		used: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "used_size",
				Help:      "Used size of pool and volume",
			},
			[]string{"name"},
		),

		available: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "available_size",
				Help:      "Available size of pool and volume",
			},
			[]string{"name"},
		),

		zfsParseErrorCounter: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "zfs_parse_error",
				Help:      "zfs parse error counter",
			},
		),

		zfsListCommandErrorCounter: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "zfs_list_command_error",
				Help:      "zfs command error counter",
			},
		),
	}
}

// newMetrics initializes fields of the metrics and returns its instance
func newMetrics() metrics {
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

		readCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "total_read_count",
				Help:      "Total read io count",
			},
			[]string{"vol", "pool"},
		),

		writeCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "total_write_count",
				Help:      "Total write io count",
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
			[]string{"vol", "pool"},
		),

		replicaStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "replica_status",
				Help:      `Status of replica (0, 1, 2, 3) = {"Offline", "Healthy", "Degraded", "Rebuilding"}`,
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

		zfsCommandErrorCounter: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "zfs_command_error",
				Help:      "zfs command error counter",
			},
		),
	}
}

func parseFloat64(e string, m *listMetrics) float64 {
	num, err := strconv.ParseFloat(e, 64)
	if err != nil {
		glog.Error("failed to parse, err: ", err)
		m.zfsParseErrorCounter.Inc()
	}
	return num
}

func listParser(stdout []byte, m *listMetrics) ([]fields, error) {
	list := make([]fields, 0)
	vols := strings.Split(string(stdout), "\n")
	for _, v := range vols {
		f := strings.Fields(v)
		if len(f) < 3 {
			break
		}
		vol := fields{
			name:      f[0],
			used:      parseFloat64(f[1], m),
			available: parseFloat64(f[2], m),
		}
		list = append(list, vol)
	}
	return list, nil
}
