package command

import (
	"errors"
	goflag "flag"
	"log"
	"net"
	"net/url"
	"time"

	"github.com/golang/glog"
	"github.com/openebs/maya/cmd/maya-exporter/app/collector"
	"github.com/openebs/maya/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
)

// Constants defined here are the dafault value of the flags. Which can be
// changed while running the binary.
const (
	// listenAddress is the address where exporter listens for the rest api
	// calls.
	listenAddress = ":9500"
	// metricsPath is the endpoint of exporter.
	metricsPath = "/metrics"
	// controllerAddress is the address where jiva controller listens.
	controllerAddress = "http://localhost:9501"
	// casType is the type of container attached storage (CAS) from which
	// the metrics need to be exported. Default is Jiva"
	casType = "jiva"
	// socketPath is the path where the unix domain sockets has been created
	// for the communication.
	SocketPath = "/var/run/istgt_ctl_sock"
)

// VolumeExporterOptions is used to create flags for the monitoring command
type VolumeExporterOptions struct {
	ListenAddress     string
	MetricsPath       string
	ControllerAddress string
	CASType           string
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
		"IP address from where metrics to be exported")
}

// AddCASTypeFlag is used to create flag to pass the storage engine name
func AddCASTypeFlag(cmd *cobra.Command, value *string) {
	cmd.Flags().StringVarP(value, "cas.type", "e", *value,
		"Type of container attached storage engine")
}

// NewCmdVolumeExporter is used to create command monitoring and it initialize
// monitoring flags also.
func NewCmdVolumeExporter() (*cobra.Command, error) {
	// create an instance of VolumeExporterOptions to initialize with default
	// values for the flags.
	options := VolumeExporterOptions{}
	options.ControllerAddress = controllerAddress
	options.ListenAddress = listenAddress
	options.MetricsPath = metricsPath
	options.CASType = casType
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
	AddCASTypeFlag(cmd, &options.CASType)
	return cmd, nil
}

// Run used to process commands,args and call openebs exporter and it returns
// nil on successful execution.
func Run(cmd *cobra.Command, options *VolumeExporterOptions) error {
	glog.Infof("Starting maya-exporter ...")
	option := Initialize(options)
	if len(option) == 0 {
		glog.Fatal("maya-exporter only supports jiva and cstor as storage engine")
		return nil
	}
	if option == "cstor" {
		glog.Infof("initialising maya-exporter for the cstor")
		if err := options.RegisterCstorStatsExporter(SocketPath); err != nil {
			glog.Fatal(err)
			return nil
		}
	}
	if option == "jiva" {
		log.Println("Initialising maya-exporter for the jiva")
		if err := options.RegisterJivaStatsExporter(); err != nil {
			glog.Fatal(err)
			return nil
		}
	}
	options.StartMayaExporter()
	return nil
}

// RegisterJivaStatsExporter parses the jiva controller URL and initialises an instance of
// VolumeExporter.
func (o *VolumeExporterOptions) RegisterJivaStatsExporter() error {
	controllerURL, err := url.ParseRequestURI(o.ControllerAddress)
	if err != nil {
		glog.Info(err)
		return errors.New("Error in parsing the URI")
	}
	exporter := collector.NewExporter(controllerURL)
	prometheus.MustRegister(exporter)
	return nil
}

// RegisterCstorStatsExporter creates a connection with the unix domain socket
// on the socketPath and the register the exporter for the collection of metrics.
// Fetching the large size images from the dockerhub takes time, so this retries
// upto 50 seconds if the istgt target has not started yet, else exits.
func (o *VolumeExporterOptions) RegisterCstorStatsExporter(sockPath string) error {
	var i int
retry:
	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		glog.Errorln("Dial error :", err)
		glog.Info("Sleep for 5 second and then retry initiating connection")
		time.Sleep(5 * time.Second)
		for {
			i++
			glog.Info("Retrying to connect to the server, retry count:", i)
			goto retry
		}
	}
	exporter := collector.NewStatsExporter(conn)
	prometheus.MustRegister(exporter)
	return nil
}
