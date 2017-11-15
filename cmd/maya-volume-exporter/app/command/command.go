package command

import (
	goflag "flag"

	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"
)

// VolumeExporterOptions is used to create flags for the monitoring command
type VolumeExporterOptions struct {
	ListenAddress     string
	MetricsPath       string
	ControllerAddress string
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

// NewCmdVolumeExporter is used to create command monitoring and it initialize
// monitoring flags also.
func NewCmdVolumeExporter() (*cobra.Command, error) {
	// create an instance of VolumeExporterOptions to initialize with default
	// values for the flags.
	options := VolumeExporterOptions{}
	options.ControllerAddress = "http://localhost:9501"
	options.ListenAddress = ":9500"
	options.MetricsPath = "/metrics"
	cmd := &cobra.Command{
		Short: "Collect metrics from OpenEBS volumes",
		Long: `  maya-volume-exporter monitors openebs volumes and exporter the metrics.
  It starts collecting metrics from the jiva volume controller at the endpoint
  "/v1/stats" Prometheus Server can collect the metrics from maya-volume-exporter.`,
		Example: `  maya-volume-exporter -a=http://localhost:8001 -c=:9500 -m=/metrics`,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(Run(cmd, &options), util.Fatal)
		},
	}

	cmd.Flags().AddGoFlagSet(goflag.CommandLine)
	goflag.CommandLine.Parse([]string{})
	AddControllerAddressFlag(cmd, &options.ControllerAddress)
	AddListenAddressFlag(cmd, &options.ListenAddress)
	AddMetricsPathFlag(cmd, &options.MetricsPath)

	return cmd, nil
}

// Run used to process commands,args and call openebs exporter and it returns
// nil on successful execution.
func Run(cmd *cobra.Command, options *VolumeExporterOptions) error {
	glog.Infof("Starting maya-volume-exporter ...")
	Entrypoint(options)
	return nil
}
