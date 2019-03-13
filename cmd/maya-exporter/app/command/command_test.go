package command

import (
	"errors"
	"reflect"
	"testing"

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

func TestRegisterJiva(t *testing.T) {
	cases := map[string]struct {
		option *VolumeExporterOptions
		output error
	}{
		"ValidURL": {
			option: &VolumeExporterOptions{
				ControllerAddress: "http://localhost:9501",
			},
			output: nil,
		},
		"InvalidURL": {
			option: &VolumeExporterOptions{
				ControllerAddress: "localhost",
			},
			output: errors.New("Error in parsing the URI"),
		},
		"EmptyURL": {
			option: &VolumeExporterOptions{
				ControllerAddress: "",
			},
			output: errors.New("Error in parsing the URI"),
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			got := tt.option.RegisterJiva()
			if !reflect.DeepEqual(got, tt.output) {
				t.Fatalf("RegisterJivaStatsExporter() => [%v], want [%v]", got, tt.output)
			}
		})
	}
}
