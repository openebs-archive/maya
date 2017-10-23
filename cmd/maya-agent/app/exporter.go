package app

import (
	"log"
	"net/http"
	"net/url"

	"github.com/openebs/maya/cmd/maya-agent/app/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Entrypoint(options *OpenEBSExporterOptions) {
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
