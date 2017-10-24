// Package exporter start an http server and display the collected
// metrics at "/metrics" endpoint. It collect metrics from collector.go
// You have to instantiates NewExporter by calling collector.NewExporter
// method and pass the Jiva volume controller IP.
package exporter

import (
	"log"
	"net/http"
	"net/url"

	"github.com/openebs/maya/cmd/maya-agent/app/exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Entrypoint is used to monitor OpenEBS volumes. It's used by maya-agent to
// start an instance of openebs volume exporter.

// We need to run several instances of Exporter for each volume just like node
// exporter on every node. At a time one instance can gather only the metrics
// from the requested volume. You need to pass the controller IP using flag -c
// at runtime as a command line argument. Type maya-agent monitor -h for more info.
func Entrypoint(options *VolumeExporterOptions) {
	controllerURL, err := url.Parse(options.ControllerAddress)
	if err != nil {
		log.Fatal(err)
	}

	exporter := collector.NewExporter(controllerURL)
	prometheus.MustRegister(exporter)

	log.Printf("Starting Server: %s", options.ListenAddress)
	if options.MetricsPath == "" || options.MetricsPath == "/" {

		http.Handle(options.MetricsPath, promhttp.Handler())

	} else {

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

	}

	err = http.ListenAndServe(options.ListenAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
}
