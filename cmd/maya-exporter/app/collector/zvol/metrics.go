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

package zvol

import (
	"os"
	"strconv"
	"strings"

	zpool "github.com/openebs/maya/pkg/zpool/v1alpha1"
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

	zfsCommandErrorCounter                      prometheus.Gauge
	zfsStatsParseErrorCounter                   prometheus.Gauge
	zfsStatsRejectRequestCounter                prometheus.Gauge
	zfsStatsNoDataSetAvailableErrorCounter      prometheus.Gauge
	zfsStatsInitializeLibuzfsClientErrorCounter prometheus.Gauge
}

type listMetrics struct {
	used      *prometheus.GaugeVec
	available *prometheus.GaugeVec

	zfsListParseErrorCounter                   prometheus.Gauge
	zfsListCommandErrorCounter                 prometheus.Gauge
	zfsListRequestRejectCounter                prometheus.Gauge
	zfsListNoDataSetAvailableErrorCounter      prometheus.Gauge
	zfsListInitializeLibuzfsClientErrorCounter prometheus.Gauge
}

type poolSyncMetrics struct {
	zpoolLastSyncTime             *prometheus.GaugeVec
	zpoolStateUnknown             *prometheus.GaugeVec
	zpoolLastSyncTimeCommandError *prometheus.GaugeVec
	cspiRequestRejectCounter      prometheus.Counter
}

// poolfields struct is for pool last sync time metric
type poolfields struct {
	name                          string
	zpoolLastSyncTime             float64
	zpoolStateUnknown             float64
	zpoolLastSyncTimeCommandError float64
}

type fields struct {
	name      string
	used      float64
	available float64
}

func newPoolMetrics() *poolSyncMetrics {
	return new(poolSyncMetrics)
}

// newMetrics initializes fields of the metrics and returns its instance
func newListMetrics() *listMetrics {
	return new(listMetrics)
}

func (l *listMetrics) withUsedSize() *listMetrics {
	l.used = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "volume_replica_used_size",
			Help:      "Used size of volume replica on a pool",
		},
		[]string{"name"},
	)
	return l
}
func (l *listMetrics) withAvailableSize() *listMetrics {
	l.available = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "volume_replica_available_size",
			Help:      "Available size of volume replica on a pool",
		},
		[]string{"name"},
	)
	return l
}

func (l *listMetrics) withParseErrorCounter() *listMetrics {
	l.zfsListParseErrorCounter = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "zfs_list_parse_error",
			Help:      "Total no of zfs list parse errors",
		},
	)
	return l
}

func (l *listMetrics) withCommandErrorCounter() *listMetrics {
	l.zfsListCommandErrorCounter = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "zfs_list_command_error",
			Help:      "Total no of zfs command errors",
		},
	)
	return l
}

func (l *listMetrics) withRequestRejectCounter() *listMetrics {
	l.zfsListRequestRejectCounter = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "zfs_list_request_reject_count",
			Help:      "Total no of rejected requests of zfs list",
		},
	)
	return l
}

func (l *listMetrics) withNoDatasetAvailableErrorCounter() *listMetrics {
	l.zfsListNoDataSetAvailableErrorCounter = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "zfs_list_no_dataset_available_error_counter",
			Help:      "Total no of no datasets error in zfs list command",
		},
	)
	return l
}

func (l *listMetrics) withInitializeLibuzfsClientErrorCounter() *listMetrics {
	l.zfsListInitializeLibuzfsClientErrorCounter = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "zfs_list_failed_to_initialize_libuzfs_client_error_counter",
			Help:      "Total no of failed to initialize libuzfs client error in zfs list command",
		},
	)
	return l
}

// newMetrics initializes fields of the metrics and returns its instance
func newMetrics() *metrics {
	return new(metrics)
}

func (m *metrics) withReadBytes() *metrics {
	m.readBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "total_read_bytes",
			Help:      "Total read in bytes of volume replica",
		},
		[]string{"vol", "pool"},
	)
	return m
}

func (m *metrics) withWriteBytes() *metrics {
	m.writeBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "total_write_bytes",
			Help:      "Total write in bytes of volume replica",
		},
		[]string{"vol", "pool"},
	)
	return m
}

func (m *metrics) withReadCount() *metrics {
	m.readCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "total_read_count",
			Help:      "Total read io count of volume replica",
		},
		[]string{"vol", "pool"},
	)
	return m
}

func (m *metrics) withWriteCount() *metrics {
	m.writeCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "total_write_count",
			Help:      "Total write io count of volume replica",
		},
		[]string{"vol", "pool"},
	)
	return m
}

func (m *metrics) withSyncCount() *metrics {
	m.syncCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "sync_count",
			Help:      "Total sync io count of volume replica",
		},
		[]string{"vol", "pool"},
	)
	return m
}

func (m *metrics) withSyncLatency() *metrics {
	m.syncLatency = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "sync_latency",
			Help:      "Sync latency on volume replica",
		},
		[]string{"vol", "pool"},
	)
	return m
}

func (m *metrics) withReadLatency() *metrics {
	m.readLatency = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "read_latency",
			Help:      "Read latency on volume replica",
		},
		[]string{"vol", "pool"},
	)
	return m
}

func (m *metrics) withWriteLatency() *metrics {
	m.writeLatency = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "write_latency",
			Help:      "Write latency on volume replica",
		},
		[]string{"vol", "pool"},
	)
	return m
}

func (m *metrics) withReplicaStatus() *metrics {
	m.replicaStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "replica_status",
			Help:      `Status of volume replica (0, 1, 2, 3) = {"Offline", "Healthy", "Degraded", "Rebuilding"}`,
		},
		[]string{"vol", "pool"},
	)
	return m
}

func (m *metrics) withinflightIOCount() *metrics {
	m.inflightIOCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "inflight_io_count",
			Help:      "Inflight IO's count of volume replica",
		},
		[]string{"vol", "pool"},
	)
	return m
}

func (m *metrics) withDispatchedIOCount() *metrics {
	m.dispatchedIOCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "dispatched_io_count",
			Help:      "Dispatched IO's count of volume replica",
		},
		[]string{"vol", "pool"},
	)
	return m
}

func (m *metrics) withRebuildCount() *metrics {
	m.rebuildCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "rebuild_count",
			Help:      "Rebuild count of volume replica",
		},
		[]string{"vol", "pool"},
	)
	return m
}

func (m *metrics) withRebuildBytes() *metrics {
	m.rebuildBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "rebuild_bytes",
			Help:      "Rebuild bytes of volume replica",
		},
		[]string{"vol", "pool"},
	)
	return m
}

func (m *metrics) withRebuildStatus() *metrics {
	m.rebuildStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "rebuild_status",
			Help:      `Status of rebuild on volume replica (0, 1, 2, 3, 4, 5, 6)= {"INIT", "DONE", "SNAP REBUILD INPROGRESS", "ACTIVE DATASET REBUILD INPROGRESS", "ERRORED", "FAILED", "UNKNOWN"}`,
		},
		[]string{"vol", "pool"},
	)
	return m
}

func (m *metrics) withRebuildDone() *metrics {
	m.rebuildDoneCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "total_rebuild_done",
			Help:      "Total no of rebuild done on volume replica",
		},
		[]string{"vol", "pool"},
	)
	return m
}

func (m *metrics) withFailedRebuild() *metrics {
	m.rebuildFailedCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "total_failed_rebuild",
			Help:      "Total no of failed rebuilds on volume replica",
		},
		[]string{"vol", "pool"},
	)
	return m
}

func (m *metrics) withCommandErrorCounter() *metrics {
	m.zfsCommandErrorCounter = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "zfs_stats_command_error",
			Help:      "Total no of zfs command errors",
		},
	)
	return m
}

func (m *metrics) withParseErrorCounter() *metrics {
	m.zfsStatsParseErrorCounter = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "zfs_stats_parse_error_counter",
			Help:      "Total no of zfs stats parse errors",
		},
	)
	return m
}

func (m *metrics) withRequestRejectCounter() *metrics {
	m.zfsStatsRejectRequestCounter = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "zfs_stats_reject_request_count",
			Help:      "Total no of rejected requests of zfs stats",
		},
	)
	return m
}

func (m *metrics) withNoDatasetAvailableErrorCounter() *metrics {
	m.zfsStatsNoDataSetAvailableErrorCounter = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "zfs_stats_no_dataset_available_error_counter",
			Help:      "Total no of no datasets error in zfs stats command",
		},
	)
	return m
}

func (m *metrics) withInitializeLibuzfsClientErrorCounter() *metrics {
	m.zfsStatsInitializeLibuzfsClientErrorCounter = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "zfs_stats_failed_to_initialize_libuzfs_client_error_counter",
			Help:      "Total no of failed to initialize libuzfs client error in zfs stats command",
		},
	)
	return m
}

// All new metrics related to pool last sync time
func (p *poolSyncMetrics) withZpoolStateUnknown() *poolSyncMetrics {

	p.zpoolStateUnknown = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "zpool_state_unknown",
			Help:      "zpool state unknown",
		},
		[]string{"pool"},
	)
	return p
}

func (p *poolSyncMetrics) withRequestRejectCounter() *poolSyncMetrics {
	p.cspiRequestRejectCounter = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "zfs_get_livenesstimestamp_request_reject_count",
			Help:      "Total no of rejected requests for pool liveness",
		},
	)
	return p
}

func (p *poolSyncMetrics) withzpoolLastSyncTimeCommandError() *poolSyncMetrics {

	p.zpoolLastSyncTimeCommandError = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "zpool_sync_time_command_error",
			Help:      "Zpool sync time command error",
		},
		[]string{"pool"},
	)
	return p
}

func (p *poolSyncMetrics) withZpoolLastSyncTime() *poolSyncMetrics {

	p.zpoolLastSyncTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "zpool_last_sync_time",
			Help:      "Last sync time of pool",
		},
		[]string{"pool"},
	)
	return p
}

func parseFloat64(e string, m *listMetrics) float64 {
	num, err := strconv.ParseFloat(e, 64)
	if err != nil {
		m.zfsListParseErrorCounter.Inc()
	}
	return num
}

func listParser(stdout []byte, m *listMetrics) []fields {
	if len(string(stdout)) == 0 {
		m.zfsListParseErrorCounter.Inc()
		return nil
	}
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
	return list
}

// poolMetricParser is used to parse output from zfs get io.openebs:livenesstimestamp
func poolMetricParser(stdout []byte) *poolfields {
	if len(string(stdout)) == 0 {
		pool := poolfields{
			name:                          os.Getenv("HOSTNAME"),
			zpoolLastSyncTime:             zpool.ZpoolLastSyncCommandErrorOrUnknownUnset,
			zpoolLastSyncTimeCommandError: zpool.ZpoolLastSyncCommandErrorOrUnknownUnset,
			zpoolStateUnknown:             zpool.ZpoolLastSyncCommandErrorOrUnknownSet,
		}
		return &pool
	}

	pools := strings.Split(string(stdout), "\n")
	f := strings.Fields(pools[0])
	if len(f) < 2 {
		return nil
	}

	pool := poolfields{
		name:                          os.Getenv("HOSTNAME"),
		zpoolLastSyncTime:             poolSyncTimeParseFloat64(f[2]),
		zpoolStateUnknown:             zpool.ZpoolLastSyncCommandErrorOrUnknownUnset,
		zpoolLastSyncTimeCommandError: zpool.ZpoolLastSyncCommandErrorOrUnknownUnset,
	}

	return &pool
}

// poolSyncTimeParseFloat64 is used to convert epoch timestamp in string to float64
func poolSyncTimeParseFloat64(e string) float64 {
	num, err := strconv.ParseFloat(e, 64)
	if err != nil {
		return 0

	}
	return num
}
