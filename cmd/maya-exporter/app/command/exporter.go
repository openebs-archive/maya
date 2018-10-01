// Package command starts an http server and displays the collected
// metrics at "/metrics" endpoint. It collects metrics from collector.go
// You have to instantiate NewExporter by calling collector.NewExporter
// method and pass the Jiva volume controller IP.
package command

import (
	"net/http"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Initialize returns the valid flags such as jiva and cstor and returns
// null string otherwise.
func Initialize(options *VolumeExporterOptions) string {
	switch option := options.CASType; option {
	case "jiva":
		return "jiva"
	case "cstor":
		return "cstor"
	default:
		return ""
	}
}

// We need to run several instances of Exporter for each volume just like node
// exporter on every node. At a time one instance can gather only the metrics
// from the requested volume. You need to pass the controller IP using flag -c
// at runtime as a command line argument. Type maya-exporter -h for more
// info.

// StartMayaExporter starts an HTTP server that exposes the metrics on
// "/metrics" endpoint.
func (options *VolumeExporterOptions) StartMayaExporter() error {
	glog.Info("Starting http server....")
	http.Handle(options.MetricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		homepage := `<html>
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
