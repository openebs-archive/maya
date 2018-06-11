package collector

import (
	"net"

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
)

// Exporter interface defines the methods that to be implemented by
// the CstorStatsExporter and JivaStatsExporter
type Exporter interface {
	collector()
	parser()
}

// Collector is the interface implemented by struct that can be used by
// Prometheus to collect metrics. A Collector has to be registered for
// collection of  metrics. Basically it has two methods Describe and Collect.

// CstorStatsExporter implements the prometheus.Collector interface. It exposes
// the metrics of a OpenEBS (cstor) volume.
type CstorStatsExporter struct {
	Conn    net.Conn
	Metrics Metrics
}

// JivaStatsExporter implements the prometheus.Collector interface. It exposes
// the metrics of a OpenEBS (Jiva) volume.
type JivaStatsExporter struct {
	VolumeControllerURL string
	Metrics             Metrics
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
	readIOPS               prometheus.Gauge
	readTimePS             prometheus.Gauge
	readBlockCountPS       prometheus.Gauge
	writeIOPS              prometheus.Gauge
	writeTimePS            prometheus.Gauge
	writeBlockCountPS      prometheus.Gauge
	readLatency            prometheus.Gauge
	writeLatency           prometheus.Gauge
	avgReadBlockCountPS    prometheus.Gauge
	avgWriteBlockCountPS   prometheus.Gauge
	sizeOfVolume           prometheus.Gauge
	volumeUpTime           *prometheus.GaugeVec
	connectionRetryCounter *prometheus.CounterVec
	connectionErrorCounter *prometheus.CounterVec
}

// MetricsDiff keep the difference of the read/write I/O's and
// other volume statistics per second.
type MetricsDiff struct {
	readIOPS             float64
	writeIOPS            float64
	readBlockCountPS     float64
	writeBlockCountPS    float64
	readTimePS           float64
	writeTimePS          float64
	readLatency          float64
	writeLatency         float64
	avgReadBlockCountPS  float64
	avgWriteBlockCountPS float64
	size                 float64
	sectorSize           float64
	logicalSize          float64
	actualSize           float64
	uptime               float64
}

var (
	// Declare the Metrics instance to reuse it in registration
	// of exporter while instantiating JivaStatsExporter and
	// CstorStatsExporter.
	metrics = Metrics{
		actualUsed: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "OpenEBS",
			Name:      "actual_used",
			Help:      "Actual volume size used",
		}),

		logicalSize: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "OpenEBS",
			Name:      "logical_size",
			Help:      "Logical size of volume",
		}),

		sectorSize: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "OpenEBS",
			Name:      "sector_size",
			Help:      "sector size of volume",
		}),

		readIOPS: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "OpenEBS",
			Name:      "read_iops",
			Help:      "Read Input/Outputs on Volume",
		}),

		readTimePS: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "OpenEBS",
			Name:      "read_time_per_second",
			Help:      "Read time on volume per second",
		}),

		readBlockCountPS: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "OpenEBS",
			Name:      "read_block_count_per_second",
			Help:      "Read Block count of volume per second",
		}),

		writeIOPS: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "OpenEBS",
			Name:      "write_iops",
			Help:      "Write Input/Outputs on Volume per second",
		}),

		writeTimePS: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "OpenEBS",
			Name:      "write_time_per_second",
			Help:      "Write time on volume per second",
		}),

		writeBlockCountPS: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "OpenEBS",
			Name:      "write_block_count_per_second",
			Help:      "Write Block count of volume per second",
		}),

		readLatency: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "OpenEBS",
			Name:      "read_latency",
			Help:      "Read Latency count of volume",
		}),

		writeLatency: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "OpenEBS",
			Name:      "write_latency",
			Help:      "Write Latency count of volume",
		}),

		avgReadBlockCountPS: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "OpenEBS",
			Name:      "avg_read_block_count_per_second",
			Help:      "Average Read Block count of volume per second",
		}),

		avgWriteBlockCountPS: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "OpenEBS",
			Name:      "avg_write_block_count_per_second",
			Help:      "Average Write Block count of volume per second",
		}),

		sizeOfVolume: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "OpenEBS",
			Name:      "size_of_volume",
			Help:      "Size of the volume requested",
		}),

		volumeUpTime: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "OpenEBS",
			Name:      "volume_uptime",
			Help:      "Time since volume has registered",
		},
			[]string{"volName", "iqn", "portal"},
		),
		connectionRetryCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "connection_retry_total",
				Help: "Total no of connection retry requests",
			},
			[]string{"err"},
		),
		connectionErrorCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "connection_error_total",
				Help: "Total no of connection errors",
			},
			[]string{"err"},
		),
	}
)
