package command

import (
	"errors"
	"log"
	"net"
	"net/http"
	"reflect"
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
			err: errors.New("listen tcp :9500: bind: address already in use"),
		},
	}
	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			startTestServer(t, tt.cmdOptions, ErrorMessage)
			msg := <-ErrorMessage
			if !reflect.DeepEqual(msg.Error(), tt.err.Error()) {
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

func TestJsonHandler(t *testing.T) {
	cases := map[string]struct {
		targetURL string
		err       error
	}{
		"When URL is correct": {
			targetURL: "http://localhost:9500" + metricsPath + "json/",
			err:       nil,
		},
	}

	srv := &http.Server{Addr: ":9500"}
	http.HandleFunc(metricsPath+"json/", jsonHandler)
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			// cannot panic, because this probably is an intentional close
			log.Printf("Httpserver: ListenAndServe() error: %s", err)
		}
	}()

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := v1alpha1.URL("", tt.targetURL)
			if !reflect.DeepEqual(err, tt.err) {
				t.Fatalf("TestName: %v jsonHandler() : expected %v, got %v", name, tt.err, err)
			}
		})
	}

	if err := srv.Shutdown(nil); err != nil {
		t.Fatalf("Shutting down server failded: %v", err)
	}

}
