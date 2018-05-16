package command

import (
	goflag "flag"
	"log"
	"net/url"

	"github.com/golang/glog"
	"github.com/openebs/maya/cmd/maya-exporter/app/collector"
	"github.com/openebs/maya/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
)

const (
	listenAddress     = ":9500"
	metricsPath       = "/metrics"
	controllerAddress = "http://localhost:9501"
	storageEngine     = "jiva"
)

// VolumeExporterOptions is used to create flags for the monitoring command
type VolumeExporterOptions struct {
	ListenAddress     string
	MetricsPath       string
	ControllerAddress string
	StorageEngine     string
	Exporter          *collector.VolumeExporter
}

// AddListenAddressFlag is used to create flag to pass the listen address of exporter.
func AddListenAddressFlag(cmd *cobra.Command, value *string) {
	cmd.Flags().StringVarP(value, "listen.addr", "a", *value,
		"Address on which to expose metrics and web interface.)")
}

// AddMetricsPathFlag is used to create flag to pass the listen path where volume
// metrics are exposed.
func AddMetricsPathFlag(cmd *cobra.Command, value *string) {
	cmd.Flags().StringVarP(value, "listen.path", "m", *value,
		"Path under which to expose metrics.")
}

// AddControllerAddressFlag is used to create flag to pass the Jiva volume
// controllers IP.
func AddControllerAddressFlag(cmd *cobra.Command, value *string) {
	cmd.Flags().StringVarP(value, "controller.addr", "c", *value,
		"Address of the Jiva volume controller.")
}

// AddStorageEngineFlag is used to create flag to pass the storage engine name
func AddStorageEngineFlag(cmd *cobra.Command, value *string) {
	cmd.Flags().StringVarP(value, "storage.engine", "e", *value,
		"Name of storage engine")
}

// NewCmdVolumeExporter is used to create command monitoring and it initialize
// monitoring flags also.
func NewCmdVolumeExporter() (*cobra.Command, error) {
	// create an instance of VolumeExporterOptions to initialize with default
	// values for the flags.
	options := VolumeExporterOptions{}
	options.ControllerAddress = "http://localhost:9501"
	options.ListenAddress = ":9500"
	options.MetricsPath = "/metrics"
	options.StorageEngine = "jiva"
	cmd := &cobra.Command{
		Short: "Collect metrics from OpenEBS volumes",
		Long: `maya-exporter can be used to monitor openebs volumes and pools.
It can be deployed alongside the openebs volume or pool containers as sidecars.`,
		Example: `maya-exporter -a=http://localhost:8001 -c=:9500 -m=/metrics`,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(Run(cmd, &options), util.Fatal)
		},
	}

	cmd.Flags().AddGoFlagSet(goflag.CommandLine)
	goflag.CommandLine.Parse([]string{})
	AddControllerAddressFlag(cmd, &options.ControllerAddress)
	AddListenAddressFlag(cmd, &options.ListenAddress)
	AddMetricsPathFlag(cmd, &options.MetricsPath)
	AddStorageEngineFlag(cmd, &options.StorageEngine)

	return cmd, nil
}

// Run used to process commands,args and call openebs exporter and it returns
// nil on successful execution.
func Run(cmd *cobra.Command, options *VolumeExporterOptions) error {
	glog.Infof("Starting maya-exporter ...")
	option := Initialize(options)
	if option == "" {
		log.Println("maya-exporter only supports jiva and cstor as storage engine")
		return nil
	}
	if option == "cstor" {
		log.Println("maya-exporter does not support cstor yet")
		return nil
	}
	if option == "jiva" {
		log.Println("Initialising maya-exporter for the jiva")
		options.RegisterJivaStatsExporter()
	}
	options.StartMayaExporter()
	return nil
}

// RegisterJivaStatsExporter parses the jiva controller URL and initialises an instance of
// VolumeExporter.
func (o *VolumeExporterOptions) RegisterJivaStatsExporter() error {
	controllerURL, err := url.ParseRequestURI(o.ControllerAddress)
	if err != nil {
		log.Println(err)
		return nil
	}
	o.Exporter = collector.NewExporter(controllerURL)
	prometheus.MustRegister(o.Exporter)
	return nil
}
