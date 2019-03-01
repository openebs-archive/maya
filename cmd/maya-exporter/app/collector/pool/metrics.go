package pool

import (
	"strconv"

	"github.com/golang/glog"
	zpool "github.com/openebs/maya/pkg/zpool/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
)

type metrics struct {
	size                prometheus.Gauge
	status              prometheus.Gauge
	usedCapacity        prometheus.Gauge
	freeCapacity        prometheus.Gauge
	usedCapacityPercent prometheus.Gauge

	parseErrorCounter        prometheus.Gauge
	zpoolCommandErrorCounter prometheus.Gauge
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

// newMetrics initializes fields of the metrics and returns its instance
func newMetrics() metrics {
	return metrics{
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

		zpoolCommandErrorCounter: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "openebs",
				Name:      "zpool_command_error",
				Help:      "zpool command error counter",
			},
		),
	}
}
