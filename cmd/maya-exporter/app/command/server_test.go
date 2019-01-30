package command

import (
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

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
		httpErr int
	}{
		"When URL is correct": {
			httpErr: http.StatusOK,
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/json", nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(jsonHandleFunc)

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.httpErr {
				t.Fatalf("handler returned wrong status code: got %v want %v",
					status, http.StatusOK)
			}
		})
	}
}

func TestMetricHandler(t *testing.T) {
	cases := map[string]struct {
		targetURL string
		httpErr   int
	}{
		"When metrics is requested protobuf format": {
			targetURL: "/metrics/",
			httpErr:   http.StatusOK,
		},
		"When metrics is requested in json format": {
			targetURL: "/metrics/?format=json",
			httpErr:   http.StatusOK,
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.targetURL, nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(metricsHandler)

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.httpErr {
				t.Fatalf("handler returned wrong status code: got %v want %v",
					status, tt.httpErr)
			}
		})
	}
}
