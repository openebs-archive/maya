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

package pool

import (
	"strconv"

	zpool "github.com/openebs/maya/pkg/zpool/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
)

type metrics struct {
	size                prometheus.Gauge
	usedCapacity        prometheus.Gauge
	freeCapacity        prometheus.Gauge
	usedCapacityPercent prometheus.Gauge
	status              *prometheus.GaugeVec

	zpoolCommandErrorCounter            prometheus.Gauge
	zpoolRejectRequestCounter           prometheus.Gauge
	zpoolListParseErrorCounter          prometheus.Gauge
	noPoolAvailableErrorCounter         prometheus.Gauge
	inCompleteOutputErrorCounter        prometheus.Gauge
	initializeLibUZFSClientErrorCounter prometheus.Gauge
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
		s.used,
		s.free,
		s.usedCapacityPercent,
	}
}

func parseFloat64(e string, m *metrics) float64 {
	num, err := strconv.ParseFloat(e, 64)
	if err != nil {
		m.zpoolListParseErrorCounter.Inc()
	}
	return num
}

func (s *statsFloat64) parse(stats zpool.Stats, p *pool) {
	s.size = parseFloat64(stats.Size, p.metrics)
	s.used = parseFloat64(stats.Used, p.metrics)
	s.free = parseFloat64(stats.Free, p.metrics)
	s.status = zpool.Status[stats.Status]
	s.usedCapacityPercent = parseFloat64(stats.UsedCapacityPercent, p.metrics)
}

func (m *metrics) withSize() *metrics {
	m.size = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "pool_size",
			Help:      "Size of pool",
		},
	)
	return m
}

func (m *metrics) withStatus() *metrics {
	m.status = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "pool_status",
			Help:      `Status of pool (0, 1, 2, 3, 4, 5, 6)= {"Offline", "Online", "Degraded", "Faulted", "Removed", "Unavail", "NoPoolsAvailable"}`,
		},
		[]string{"pool"},
	)
	return m
}
func (m *metrics) withUsedCapacity() *metrics {
	m.usedCapacity = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "used_pool_capacity",
			Help:      "Capacity used by pool",
		},
	)
	return m
}

func (m *metrics) withFreeCapacity() *metrics {
	m.freeCapacity = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "free_pool_capacity",
			Help:      "Free capacity in pool",
		},
	)
	return m
}

func (m *metrics) withUsedCapacityPercent() *metrics {
	m.usedCapacityPercent = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "used_pool_capacity_percent",
			Help:      "Capacity used by pool in percent",
		},
	)
	return m
}

func (m *metrics) withParseErrorCounter() *metrics {
	m.zpoolListParseErrorCounter = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "zpool_list_parse_error_count",
			Help:      "Total no of parsing errors",
		},
	)
	return m
}

func (m *metrics) withRejectRequestCounter() *metrics {
	m.zpoolRejectRequestCounter = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "zpool_list_reject_request_count",
			Help:      "Total no of rejected requests of zpool command",
		},
	)
	return m
}

func (m *metrics) withCommandErrorCounter() *metrics {
	m.zpoolCommandErrorCounter = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "zpool_list_command_error",
			Help:      "Total no of zpool command errors",
		},
	)
	return m
}

func (m *metrics) withNoPoolAvailableErrorCounter() *metrics {
	m.noPoolAvailableErrorCounter = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "zpool_list_no_pool_available_error",
			Help:      "Total no of no pool available errors",
		},
	)
	return m
}

func (m *metrics) withIncompleteOutputErrorCounter() *metrics {
	m.inCompleteOutputErrorCounter = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "zpool_list_incomplete_stdout_error",
			Help:      "Total no of incomplete stdout errors",
		},
	)
	return m
}

func (m *metrics) withInitializeLibuzfsClientErrorCounter() *metrics {
	m.initializeLibUZFSClientErrorCounter = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "openebs",
			Name:      "zpool_list_failed_to_initialize_libuzfs_client_error_counter",
			Help:      "Total no of initialize libuzfs client error",
		},
	)
	return m
}

func newMetrics() *metrics {
	return new(metrics)
}
