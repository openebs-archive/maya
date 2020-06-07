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

// Package command starts an http server and displays the collected
// metrics at "/metrics" endpoint. It collects metrics from collector.go
// You have to instantiate NewExporter by calling collector.NewExporter
// method and pass the Jiva volume controller IP.
package command

import (
	"encoding/json"
	"net/http"
	_ "net/http/pprof"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/klog"
)

// We need to run several instances of Exporter for each volume just like node
// exporter on every node. At a time one instance can gather only the metrics
// from the requested volume. You need to pass the controller IP using flag -c
// at runtime as a command line argument. Type maya-exporter -h for more
// info.

// StartMayaExporter starts an HTTP server that exposes the metrics on
// <clusterIP>:9500/metrics endpoint and for getting the metrics in json
// the "type=json" action can be used. e.g <clusterIP>:9500/metrics/?type=json
func (options *VolumeExporterOptions) StartMayaExporter() error {
	klog.Info("Starting http server....")
	http.HandleFunc(options.MetricsPath, metricsHandler) // For backward compatibility
	http.HandleFunc(options.MetricsPath+"/", metricsHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		homepage := `
<html>
	<head><title>OpenEBS Exporter</title></head>
<body>
	<h1>OpenEBS Exporter</h1>
	<p><a href="` + options.MetricsPath + `">Metrics</a></p>
</body>
</html>
`
		w.Write([]byte(homepage))
	})
	err := http.ListenAndServe(options.ListenAddress, nil)
	if err != nil {
		klog.Error(err)
	}
	return err
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.RawQuery == "format=json" {
		jsonHandler().ServeHTTP(w, r)
	} else {
		promhttp.Handler().ServeHTTP(w, r)
	}
}

func jsonHandler() http.Handler {
	return http.HandlerFunc(jsonHandleFunc)
}

func jsonHandleFunc(w http.ResponseWriter, r *http.Request) {
	metricsFamily, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		http.Error(w, "Error fetching metrics : "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(metricsFamily)
	if err != nil {
		http.Error(w, "Error encoding metric family: "+err.Error(), http.StatusInternalServerError)
	}
}
