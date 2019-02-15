package pool

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Replica struct {
	Name          string `json:"name"`
	Status        string `json:"status"`
	RebuildStatus string `json:"rebuildStatus"`

	//	Size               float64
	Reads      float64 `json:"reads,string"`
	Writes     float64 `json:"writes,string"`
	SyncCount  float64 `json:"syncCount,string"`
	ReadBytes  float64 `json:"readByte,string"`
	WriteBytes float64 `json:"writeByte,string"`
	//	LogicalUsed        float64 `json:"s,string"`
	SyncLatency        float64 `json:"syncLatency,string"`
	ReadLatency        float64 `json:"readLatency,string"`
	WriteLatency       float64 `json:"writeLatency,string"`
	RebuildCount       float64 `json:"rebuildCnt,string"`
	RebuildBytes       float64 `json:"rebuildBytes,string"`
	InflightIOCount    float64 `json:"inflightIOCnt,string"`
	RebuildDoneCount   float64 `json:"rebuildDoneCnt,string"`
	DispatchedIOCount  float64 `json:"dispathedIOCnt,string"`
	RebuildFailedCount float64 `json:"rebuildFailedCnt,string"`
}

// metrics keeps all the volume related stats values into the respective fields.
type metrics struct {
	reads      *prometheus.GaugeVec
	writes     *prometheus.GaugeVec
	readBytes  *prometheus.GaugeVec
	writeBytes *prometheus.GaugeVec

	syncCount   *prometheus.GaugeVec
	syncLatency *prometheus.GaugeVec

	readLatency  *prometheus.GaugeVec
	writeLatency *prometheus.GaugeVec

	status        *prometheus.GaugeVec
	replicaStatus *prometheus.GaugeVec

	inflightIOCount   *prometheus.GaugeVec
	dispatchedIOCount *prometheus.GaugeVec

	rebuildCount       *prometheus.GaugeVec
	rebuildBytes       *prometheus.GaugeVec
	rebuildStatus      *prometheus.GaugeVec
	rebuildDoneCount   *prometheus.GaugeVec
	rebuildFailedCount *prometheus.GaugeVec

	capacity            *prometheus.GaugeVec
	usedCapacity        *prometheus.GaugeVec
	freeCapacity        *prometheus.GaugeVec
	usedCapacityPercent *prometheus.GaugeVec

	connectionRetryCounter prometheus.Gauge
	connectionErrorCounter prometheus.Gauge
}

func Metrics() metrics {
	return metrics{
		reads: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "reads",
				Help:      "Total no of read IO's on replica",
			},
			[]string{"vol", "pool"},
		),

		writes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "writes",
				Help:      "Total no of write IO's on replica",
			},
			[]string{"vol", "pool"},
		),

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

		status: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "pool_status",
				Help:      `Status of pool (1, 2, 3, 4)= {"OFFLINE", "HEALTHY", "DEGRADED", "ONLINE"}`,
			},
			[]string{"vol", "pool"},
		),

		replicaStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "replica_status",
				Help:      `Status of pool (1, 2, 3, 4)= {"OFFLINE", "HEALTHY", "DEGRADED", "ONLINE"}`,
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
				Help:      "Status of rebuild on replica",
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

		capacity: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "pool_capacity",
				Help:      "capacity of pool",
			},
			[]string{"vol", "pool"},
		),

		usedCapacity: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "used_pool_capacity",
				Help:      "Capacity used by pool",
			},
			[]string{"vol", "pool"},
		),

		freeCapacity: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "free_pool_capacity",
				Help:      "Free capacity in pool",
			},
			[]string{"vol", "pool"},
		),

		usedCapacityPercent: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "used_pool_capacity_percent",
				Help:      "Capacity used by pool in percent",
			},
			[]string{"vol", "pool"},
		),

		connectionRetryCounter: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "connection_retry",
				Help:      "Connection retry counter",
			},
		),

		connectionErrorCounter: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "connection_error",
				Help:      "Connection error counter",
			},
		),
	}
}
