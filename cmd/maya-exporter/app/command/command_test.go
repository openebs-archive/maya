package command

import (
	"reflect"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
)

var (
	cmd = &cobra.Command{
		Short: "Collect metrics from OpenEBS volumes",
		Long: `maya-exporter can be used to monitor openebs volumes and pools.
It can be deployed alongside the openebs volume or pool containers as sidecars.`,
		Example: `maya-exporter -a=http://localhost:8001 -c=:9500 -m=/metrics`,
	}
)

func TestAddListenAddressFlag(t *testing.T) {
	cases := map[string]struct {
		input *cobra.Command
		value string
	}{
		"Valid Command": {
			input: cmd,
			value: ":9500",
		},
	}
	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			AddListenAddressFlag(tt.input, &tt.value)
		})
	}
}

func TestAddMetricsPathFlag(t *testing.T) {
	cases := map[string]struct {
		input *cobra.Command
		value string
	}{
		"Valid Command": {
			input: cmd,
			value: ":9500",
		},
	}
	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			AddMetricsPathFlag(tt.input, &tt.value)
		})
	}
}

func TestAddControllerAddressFlag(t *testing.T) {
	cases := map[string]struct {
		input *cobra.Command
		value string
	}{
		"Valid Command": {
			input: cmd,
			value: ":9500",
		},
	}
	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			AddControllerAddressFlag(tt.input, &tt.value)
		})
	}
}

func TestAddStorageEngineFlag(t *testing.T) {
	cases := map[string]struct {
		input *cobra.Command
		value string
	}{
		"Valid Command": {
			input: cmd,
			value: ":9500",
		},
	}
	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			AddStorageEngineFlag(tt.input, &tt.value)
		})
	}
}

func TestRegisterJivaStatsExporter(t *testing.T) {

	cases := map[string]struct {
		input  *VolumeExporterOptions
		output error
	}{
		"ValidURL": {
			input: &VolumeExporterOptions{
				ControllerAddress: "http://localhost:9501",
			},
			output: nil,
		},
		"InvalidURL": {
			input: &VolumeExporterOptions{
				ControllerAddress: "localhost",
			},
			output: nil,
		},
	}

	for name, tt := range cases {
		prometheus.Unregister(tt.input.Exporter)
		t.Run(name, func(t *testing.T) {
			got := tt.input.RegisterJivaStatsExporter()
			if !reflect.DeepEqual(got, tt.output) {
				t.Fatalf("RegisterJivaStatsExporter() => [%v], want [%v]", got, tt.output)
			}
		})
	}

}
