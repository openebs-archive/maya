package command

import (
	"errors"
	"net"
	"reflect"
	"testing"
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
	ErrorMessage := make(chan error)
	cases := map[string]struct {
		cmdOptions *VolumeExporterOptions
		err        error
	}{
		"If port is busy and path is `/metrics`": {
			cmdOptions: &VolumeExporterOptions{
				ControllerAddress: "localhost:9501",
				MetricsPath:       "/metrics",
				ListenAddress:     ":9500",
			},
			err: errors.New("bind address already in use, please use another address"),
		},
	}
	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			startTestServer(t, tt.cmdOptions, ErrorMessage)
			msg := <-ErrorMessage
			if !reflect.DeepEqual(msg, tt.err) {
				t.Fatalf("StartMayaExporter() : expected %v, got %v", tt.err, msg)
			}
		})
	}
}

func startTestServer(t *testing.T, options *VolumeExporterOptions, errMsg chan error) {
	go func() {
		//Block port 9500 and attempt to start http server at 9500.
		listener, err := net.Listen("tcp", "localhost:9500")
		defer listener.Close()
		if err != nil {
			t.Log(err)
		}
		errMsg <- options.StartMayaExporter()
	}()
}
