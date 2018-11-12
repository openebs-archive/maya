package collector

import (
	"net"

	"github.com/openebs/maya/types/v1"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// SocketPath is path from where connection has to be created.
	SocketPath = "/var/run/istgt_ctl_sock"
	// HeaderPrefix is the prefix comes in the header from cstor.
	HeaderPrefix = "iSCSI Target Controller version"
	// EOF separates the strings from response which comes from the
	// cstor as the collection of metrics.
	EOF = "\r\n"
	// Footer is used to verify if all the response has collected.
	Footer = "OK IOSTATS"
	// Command is a command that is used to write over wire and get
	// the iostats from the cstor.
	Command = "IOSTATS"
	// BufSize is the size of response from cstor.
	BufSize = 256
)

// Exporter interface defines the interfaces that has methods to be
// implemented by the CstorStatsExporter and JivaStatsExporter.
type Exporter interface {
	Collect()
	Parse()
	Set()
}

// Parse interface defines the method that to be implemented by the
// CstorStatsExporter and JivaStatsExporter. parse() is used to parse
// the response into the Metrics struct.
type Parse interface {
	parser()
}

// Collect interface defines the the method that to be implemented by
// the CstorStatsExporter and JivaStatsExporter. collector() is used
// to collect the metrics from the Jiva and Cstor.
type Collect interface {
	collector()
}

// Set interface defines the method set() which is used to set the
// values to the gauges and counters.
type Set interface {
	set()
}

// VolumeStatsExporter inherits the properties of cstor and jiva,
// these properties includes metrics of the volumes.
type VolumeStatsExporter struct {
	CASType string
	Cstor
	Jiva
	Metrics
}

// Collector is the interface implemented by struct that can be used by
// Prometheus to collect metrics. A Collector has to be registered for
// collection of  metrics. Basically it has two methods Describe and Collect.

// Cstor implements the prometheus.Collector interface. It exposes
// the metrics of a OpenEBS (cstor) volume.
type Cstor struct {
	Conn net.Conn
}

// Jiva implements the prometheus.Collector interface. It exposes
// the metrics of a OpenEBS (Jiva) volume.
type Jiva struct {
	VolumeControllerURL string
}

// A gauge is a metric that represents a single numerical value that can
// arbitrarily go up and down.

// Gauges are typically used for measured values like temperatures or current
// memory usage, but also "counts" that can go up and down, like the number of
// running goroutines.

// GaugeOpts is the alias for Opts, which is used to create diffent type of
// metrics.

// GaugeVec is a Collector that bundles a set of Gauges that all share the same
// Desc, but have different values for their variable labels. This is used if
// you want to count the same thing partitioned by various dimensions
// (e.g. number of operations queued, partitioned by user and operation
// type). Create instances with NewGaugeVec.

// CounterVec is a Collector that bundles a set of Counters that all share the
// same Desc, but have different values for their variable labels. This is used
// if you want to count the same thing partitioned by various dimensions
// (e.g. number of HTTP requests, partitioned by response code and
// method). Create instances with NewCounterVec.
//

// Metrics keeps all the volume related stats values into the respective fields.
type Metrics struct {
	actualUsed             prometheus.Gauge
	logicalSize            prometheus.Gauge
	sectorSize             prometheus.Gauge
	reads                  prometheus.Gauge
	totalReadTime          prometheus.Gauge
	totalReadBlockCount    prometheus.Gauge
	totalReadBytes         prometheus.Gauge
	writes                 prometheus.Gauge
	totalWriteTime         prometheus.Gauge
	totalWriteBlockCount   prometheus.Gauge
	totalWriteBytes        prometheus.Gauge
	sizeOfVolume           prometheus.Gauge
	volumeUpTime           *prometheus.CounterVec
	connectionRetryCounter *prometheus.CounterVec
	connectionErrorCounter *prometheus.CounterVec
	replicaCounter         *prometheus.GaugeVec
}

// VolumeStats keep the values of read/write I/O's and
// other volume statistics per second.
type VolumeStats struct {
	reads                float64
	writes               float64
	totalReadBlockCount  float64
	totalReadBytes       float64
	totalWriteBlockCount float64
	totalWriteBytes      float64
	totalReadTime        float64
	totalWriteTime       float64
	size                 float64
	sectorSize           float64
	logicalSize          float64
	actualSize           float64
	uptime               float64
	revisionCount        float64
	replicaCount         float64
	name                 string
	replicas             []v1.Replica
	status               string
}

// MetricsInitializer returns the Metrics instance used for registration
// of exporter while instantiating JivaStatsExporter and
// CstorStatsExporter.
func MetricsInitializer(casType string) *Metrics {
	return &Metrics{
		actualUsed: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "actual_used",
				Help:      "Actual volume size used",
			}),

		logicalSize: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "logical_size",
				Help:      "Logical size of volume",
			}),

		sizeOfVolume: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "size_of_volume",
				Help:      "Size of the volume requested",
			}),

		sectorSize: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "sector_size",
				Help:      "sector size of volume",
			}),

		totalReadBytes: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "total_read_bytes",
				Help:      "Total read bytes",
			}),

		reads: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "reads",
				Help:      "Read Input/Outputs on Volume",
			}),

		totalReadTime: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "read_time",
				Help:      "Read time on volume",
			}),

		totalReadBlockCount: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "read_block_count",
				Help:      "Read Block count of volume",
			}),

		totalWriteBytes: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "total_write_bytes",
				Help:      "Total write bytes",
			}),

		writes: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "writes",
				Help:      "Write Input/Outputs on Volume",
			}),

		totalWriteTime: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "write_time",
				Help:      "Write time on volume",
			}),

		totalWriteBlockCount: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "write_block_count",
				Help:      "Write Block count of volume",
			}),

		volumeUpTime: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "openebs",
				Name:      "volume_uptime",
				Help:      "Time since volume has registered",
			},
			[]string{"volName", "iqn", "portal", "castype", "status"},
		),

		connectionRetryCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "openebs",
				Name:      "connection_retry_total",
				Help:      "Total no of connection retry requests",
			},
			[]string{"err"},
		),

		connectionErrorCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "openebs",
				Name:      "connection_error_total",
				Help:      "Total no of connection errors",
			},
			[]string{"err"},
		),

		replicaCounter: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "replica_count",
				Help:      "Total no of replicas",
			},
			[]string{"address", "mode"},
		),
	}
}

// gaugeList returns the list of the registered gauge variables
func (v *VolumeStatsExporter) gaugesList() []prometheus.Gauge {
	return []prometheus.Gauge{
		v.reads,
		v.writes,
		v.totalReadBytes,
		v.totalWriteBytes,
		v.totalReadTime,
		v.totalWriteTime,
		v.totalReadBlockCount,
		v.totalWriteBlockCount,
		v.actualUsed,
		v.logicalSize,
		v.sectorSize,
		v.sizeOfVolume,
	}
}

// counterList returns the list of registered counter variables
func (v *VolumeStatsExporter) countersList() []prometheus.Collector {
	return []prometheus.Collector{
		v.volumeUpTime,
		v.connectionErrorCounter,
		v.connectionRetryCounter,
		v.replicaCounter,
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

// Describe describes all the registered stats metrics from the OpenEBS volumes.
func (v *VolumeStatsExporter) Describe(ch chan<- *prometheus.Desc) {
	for _, gauge := range v.gaugesList() {
		gauge.Describe(ch)
	}

	for _, counter := range v.countersList() {
		counter.Describe(ch)
	}
}

// Collect is called by the Prometheus registry when collecting
// metrics. The implementation sends each collected metric via the
// provided channel and returns once the last metric has been sent. The
// descriptor of each sent metric is one of those returned by
// Describe. Returned metrics that share the same descriptor must differ
// in their variable label values. This method may be called
// way. Bloc	king occurs at the expense of total performance of rendering
// concurrently and must therefore be implemented in a concurrency safe
// all registered metrics. Ideally, Collector implementations support
// concurrent readers.

// Collect collects all the registered stats metrics from the OpenEBS volumes.
// It tries to reconnect with the volume if there is any error via a goroutine.
func (v *VolumeStatsExporter) Collect(ch chan<- prometheus.Metric) {
	// no need to catch the error as exporter should work even if
	// there are failures in collecting the metrics due to connection
	// issues or anything else.
	switch v.CASType {
	case "cstor":
		_ = v.Cstor.collector(&v.Metrics)
	case "jiva":
		_ = v.Jiva.collector(&v.Metrics)
	}

	// collect the metrics extracted by collect method
	for _, gauge := range v.gaugesList() {
		gauge.Collect(ch)
	}
	for _, counter := range v.countersList() {
		counter.Collect(ch)
	}
}
