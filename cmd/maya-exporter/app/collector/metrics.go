// Copyright © 2017-2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

import (
	"encoding/json"

	v1 "github.com/openebs/maya/pkg/stats/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/klog"
)

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

// metrics keeps all the volume related stats values into the respective fields.
type metrics struct {
	actualUsed                 prometheus.Gauge
	logicalSize                prometheus.Gauge
	sectorSize                 prometheus.Gauge
	reads                      prometheus.Gauge
	totalReadTime              prometheus.Gauge
	totalReadBlockCount        prometheus.Gauge
	totalReadBytes             prometheus.Gauge
	writes                     prometheus.Gauge
	totalWriteTime             prometheus.Gauge
	totalWriteBlockCount       prometheus.Gauge
	totalWriteBytes            prometheus.Gauge
	sizeOfVolume               prometheus.Gauge
	volumeStatus               prometheus.Gauge
	parseErrorCounter          prometheus.Gauge
	connectionRetryCounter     prometheus.Gauge
	connectionErrorCounter     prometheus.Gauge
	healthyReplicaCounter      prometheus.Gauge
	degradedReplicaCounter     prometheus.Gauge
	totalReplicaCounter        prometheus.Gauge
	isClientConnected          prometheus.Gauge
	volumeUpTime               *prometheus.GaugeVec
	targetRejectRequestCounter prometheus.Counter
}

// stats keep the values of read/write I/O's and
// other volume statistics per second.
type stats struct {
	got                  bool
	casType, iqn         string
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
	totalReplicaCount    float64
	healthyReplicaCount  float64
	degradedReplicaCount float64
	offlineReplicaCount  float64
	isClientConnected    float64
	name                 string
	replicas             []v1.Replica
	status               v1.TargetMode
	address              string
}

// MetricsInitializer returns the Metrics instance used for registration
// of exporter while instantiating JivaStatsExporter and
// CstorStatsExporter.
func Metrics(cas string) metrics {
	return metrics{
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

		volumeStatus: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "volume_status",
				Help:      "Status of volume: (1, 2, 3, 4) = {Offline, Degraded, Healthy, Unknown}",
			}),

		volumeUpTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "volume_uptime",
				Help:      "Time since volume has registered",
			},
			[]string{"volName", "castype"},
		),

		connectionRetryCounter: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "connection_retry_total",
				Help:      "Total no of connection retry requests",
			},
		),

		connectionErrorCounter: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "connection_error_total",
				Help:      "Total no of connection errors",
			},
		),

		parseErrorCounter: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "parse_error_total",
				Help:      "Total no of parsing errors",
			},
		),

		healthyReplicaCounter: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "healthy_replica_count",
				Help:      "Total no of healthy replicas",
			},
		),

		degradedReplicaCounter: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "degraded_replica_count",
				Help:      "Total no of degraded/ro replicas",
			},
		),

		totalReplicaCounter: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "total_replica_count",
				Help:      "Total no of replicas connected to cas",
			},
		),

		isClientConnected: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "iscsi_initiator_login_status",
				Help:      "iSCSI Initiator to jiva target login status: (0, 1): {Not Logged In, Logged In}",
			},
		),

		targetRejectRequestCounter: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: "openebs",
				Name:      "target_reject_request_counter",
				Help:      "Total no of rejected GET http requests if a request is already in progress",
			},
		),
	}
}

func (v *stats) getReplicaCount() {
	var (
		ro, rw float64
	)
	for _, rep := range v.replicas {
		switch rep.Mode {
		case readOnly, writeOnly, errored, degraded:
			ro++
		case readWrite, healthy:
			rw++
		default:
			klog.Error("Unknown replica mode: ", rep.Mode)
		}
	}
	v.degradedReplicaCount = ro
	v.healthyReplicaCount = rw
}

func (v *stats) getVolumeStatus() volumeStatus {
	switch v.status {
	case targetReadOnly, targetOffline:
		return Offline
	case targetReadWrite, targetHealthy:
		return Healthy
	case targetDegraded:
		return Degraded
	default:
		return Unknown
	}
}

func parseFloat64(entity json.Number, metrics *metrics) float64 {
	num, err := entity.Float64()
	if err != nil {
		klog.Error("failed to parse, err: ", err)
		metrics.parseErrorCounter.Inc()
	}
	return num
}
