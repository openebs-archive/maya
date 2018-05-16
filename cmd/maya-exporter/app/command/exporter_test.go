package command

import (
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
				StorageEngine: "cstor",
			},
			output: "cstor",
		},
		"storage engine is jiva": {
			cmdOptions: &VolumeExporterOptions{
				StorageEngine: "jiva",
			},
			output: "jiva",
		},
		"storage engine is other": {
			cmdOptions: &VolumeExporterOptions{
				StorageEngine: "other",
			},
			output: "",
		},
	}
	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			got := Initialize(tt.cmdOptions)
			if !reflect.DeepEqual(got, tt.output) {
				t.Fatalf("Initialize() => %v, want %v", got, tt.output)
			}
		})
	}
}

func TestStartMayaExporter(t *testing.T) {
	ErrorMessage := make(chan error)
	cmdOptions := &VolumeExporterOptions{
		ControllerAddress: "localhost:9501",
		MetricsPath:       "/metrics",
		ListenAddress:     ":9500",
	}
	go func() {
		//Block port 9090 and attempt to start http server at 9090.
		p1, err := net.Listen("tcp", "localhost:9500")
		defer p1.Close()
		if err != nil {
			t.Log(err)
		}
		ErrorMessage <- cmdOptions.StartMayaExporter()
	}()
	msg := <-ErrorMessage
	if msg != nil {
		t.Log("Try to start http server in a port which is busy.")
		t.Log(msg)
	}
}
