package command

import (
	"testing"

	"github.com/openebs/maya/pkg/client/http/v1alpha1"
)

func TestInitialize(t *testing.T) {
	cases := map[string]struct {
		cmdOptions *VolumeExporterOptions
		output     string
	}{
		"Storage engine is cstor": {
			cmdOptions: &VolumeExporterOptions{
				CASType: "cstor",
			},
			output: "cstor",
		},
		"storage engine is jiva": {
			cmdOptions: &VolumeExporterOptions{
				CASType: "jiva",
			},
			output: "jiva",
		},
		"storage engine is other": {
			cmdOptions: &VolumeExporterOptions{
				CASType: "other",
			},
			output: "",
		},
	}
	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			got := Initialize(tt.cmdOptions)
			if got != tt.output {
				t.Fatalf("Initialize() => %v, want %v", got, tt.output)
			}
		})
	}
}
func TestStartMayaExporter(t *testing.T) {
	cases := map[string]struct {
		cmdOptions *VolumeExporterOptions
		err        error
		targetURL  string
	}{
		"Check for metrics": {
			err:       nil,
			targetURL: "http://localhost:9500/metrics",
		},
		"Check for json": {
			err:       nil,
			targetURL: "http://localhost:9500/metrics/?type=json",
		},
	}

	options := &VolumeExporterOptions{
		ControllerAddress: "localhost:9501",
		MetricsPath:       "/metrics",
		ListenAddress:     ":9500",
	}

	go func() {
		err := options.StartMayaExporter()
		if err != nil {
			t.Logf("Unable to start server: %v", err)
		}

	}()

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := v1alpha1.URL("", tt.targetURL)
			if (err == nil && tt.err != nil) || (err != nil && tt.err == nil) || (err != nil && tt.err != nil && err.Error() != tt.err.Error()) {
				t.Logf("first:	%v", (err != nil && tt.err != nil && err.Error() != tt.err.Error()))
				t.Fatalf("Test Name: %v Wanted: %v Got: %v", name, tt.err, tt.err)

			}

		})
	}
}
