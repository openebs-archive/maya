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
