// Package command starts an http server and displays the collected
// metrics at "/metrics" endpoint. It collects metrics from collector.go
// You have to instantiate NewExporter by calling collector.NewExporter
// method and pass the Jiva volume controller IP.
package command

import (
	"encoding/json"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/golang/glog"
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
	glog.Info("Starting http server....")
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
		glog.Error(err)
	}
	return err
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.RawQuery == "format=json" {
		jsonHandler().ServeHTTP(w, r)
	} else {
		prometheus.Handler().ServeHTTP(w, r)
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
