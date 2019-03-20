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
	goflag "flag"
	"net/url"
	"time"

	"github.com/golang/glog"
	"github.com/openebs/maya/cmd/maya-exporter/app/collector"
	"github.com/openebs/maya/cmd/maya-exporter/app/collector/pool"
	"github.com/openebs/maya/cmd/maya-exporter/app/collector/zvol"
	types "github.com/openebs/maya/pkg/exec"
	exec "github.com/openebs/maya/pkg/exec/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
)

// Constants defined here are the default value of the flags. Which can be
// changed while running the binary.
const (
	// listenAddress is the address where exporter listens for the rest api
	// calls.
	listenAddress = ":9500"
	// metricsPath is the endpoint of exporter.
	metricsPath = "/metrics"
	// socketPath where istgt is listening
	socketPath = "/var/run/istgt_ctl_sock"
	// controllerAddress is the address where jiva controller listens.
	controllerAddress = "http://localhost:9501"
	// casType is the type of container attached storage (CAS) from which
	// the metrics need to be exported. Default is Jiva"
	casType = "jiva"
	// timeout is the timeout for executing a command
	timeout = 30 * time.Second
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
	switch options.CASType {
	case "cstor":
		glog.Infof("Initialising maya-exporter for the cstor")
		options.RegisterCstor()
	case "jiva":
		glog.Infof("Initialising maya-exporter for the jiva")
		if err := options.RegisterJiva(); err != nil {
			glog.Fatal(err)
			return nil
		}
	case "pool":
		glog.Infof("Initialising maya-exporter for the cstor pool")
		options.RegisterPool()
	default:
		return errors.New("unsupported CAS")
	}
	options.StartMayaExporter()
	return nil
}

// RegisterJiva parses the jiva controller URL and
// initialises an instance of Jiva.This returns err
// if the URL is not correct.
func (o *VolumeExporterOptions) RegisterJiva() error {
	url, err := url.ParseRequestURI(o.ControllerAddress)
	if err != nil {
		glog.Error(err)
		return errors.New("Error in parsing the URI")
	}
	jiva := collector.Jiva(url)
	exporter := collector.New(jiva)
	prometheus.MustRegister(exporter)
	glog.Info("Registered maya exporter for jiva")
	return nil
}

// RegisterCstor initiates the connection with the cstor and register
// the exporter with Prometheus for collecting the metrics.This doesn't returns
// error because that case is handled in InitiateConnection().
func (o *VolumeExporterOptions) RegisterCstor() {
	cstor := collector.Cstor(socketPath)
	exporter := collector.New(cstor)
	prometheus.MustRegister(exporter)
	glog.Info("Registered maya exporter for cstor")
	return
}

// RegisterPool registers pool collector which collects
// pool level metrics
func (o *VolumeExporterOptions) RegisterPool() {
	p := pool.New(buildRunner(timeout, "zpool", "list", "-Hp"))
	z := zvol.New(buildRunner(timeout, "zfs", "stats"))
	l := zvol.NewVolumeList(buildRunner(timeout, "zfs", "list", "-Hp"))
	prometheus.MustRegister(p, z, l)
	glog.Info("Registered maya exporter for cstor pool")
	return
}

func buildRunner(timeout time.Duration, cmd string, args ...string) types.Runner {
	return exec.StdoutBuilder().
		WithTimeout(timeout).
		WithCommand(cmd).
		WithArgs(args...).
		Build()
}
