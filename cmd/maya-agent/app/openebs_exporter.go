package app

import (
	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

// OpenEBSExporterOptions is used to create flags for the monitoring command
type OpenEBSExporterOptions struct {
	ListenAddress     string
	MetricsPath       string
	ControllerAddress string
}

func AddListenAddressFlag(cmd *cobra.Command, value *string) {
	cmd.Flags().StringVarP(value, "listen.addr", "a", *value,
		"Address on which to expose metrics and web interface.)")
}

func AddMetricsPathFlag(cmd *cobra.Command, value *string) {
	cmd.Flags().StringVarP(value, "listen.path", "m", *value,
		"Path under which to expose metrics.")
}

func AddControllerAddressFlag(cmd *cobra.Command, value *string) {
	cmd.Flags().StringVarP(value, "controller.addr", "c", *value,
		"Address of the OpenEBS controller monitoring.")
}

// NewCmdOpenEBSExporter is used to create command monitoring and it initialize
// monitoring flags also.
func NewCmdOpenEBSExporter() *cobra.Command {
	// create an instance of OpenEBSExporterOptions to initialize with default
	// values for the flags.
	options := OpenEBSExporterOptions{}
	options.ControllerAddress = "http://localhost:9501"
	options.ListenAddress = ":9500"
	options.MetricsPath = "/metrics"
	cmd := &cobra.Command{
		Use:   "monitor",
		Short: "Collect metrics from OpenEBS volumes",
		Long: `  monitor command is used to start monitoring openebs volumes. It
  start collecting metrics from the jiva controller at the endpoint
  "/v1/stats" and push it to Prometheus Server`,
		Example: `  maya-agent monitor -a=http://localhost:8001 -c=:9500 -m=/metrics`,
		Run: func(cmd *cobra.Command, args []string) {
			checkErr(run(cmd, &options), fatal)
		},
	}

	AddControllerAddressFlag(cmd, &options.ControllerAddress)
	AddListenAddressFlag(cmd, &options.ListenAddress)
	AddMetricsPathFlag(cmd, &options.MetricsPath)

	return cmd
}

// run used to process commands,args and call openebs exporter and it returns
// nil on successful execution.
func run(cmd *cobra.Command, options *OpenEBSExporterOptions) error {
	glog.Infof("Starting openebs-exporter ...")
	Entrypoint(options)
	return nil
}
