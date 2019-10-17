package zvol

import (
	"os"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type poolSyncMetrics struct {
	zpoolLastSyncTime             *prometheus.GaugeVec
	zpoolStateUnknown             *prometheus.GaugeVec
	zpoolLastSyncTimeCommandError *prometheus.GaugeVec
}

type poolfields struct {
	name                          string
	zpoolLastSyncTime             float64
	zpoolStateUnknown             float64
	zpoolLastSyncTimeCommandError float64
}

func newPoolMetrics() *poolSyncMetrics {
	return new(poolSyncMetrics)
}

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

func poolMetricParser(stdout []byte, p *poolSyncMetrics) *poolfields {
	if len(string(stdout)) == 0 {
		pool := poolfields{
			name:                          os.Getenv("HOSTNAME"),
			zpoolLastSyncTime:             0,
			zpoolLastSyncTimeCommandError: 0,
			zpoolStateUnknown:             1,
		}
		return &pool
	}

	pools := strings.Split(string(stdout), "\n")
	f := strings.Fields(pools[0])
	if len(f) < 2 {
		return nil
	}

	pool := poolfields{
		name:                          f[0],
		zpoolLastSyncTime:             poolSyncTimeParseFloat64(f[2]),
		zpoolStateUnknown:             0,
		zpoolLastSyncTimeCommandError: 0,
	}

	return &pool
}

func poolSyncTimeParseFloat64(e string) float64 {
	num, err := strconv.ParseFloat(e, 64)
	if err != nil {
		return 0

	}
	return num
}
