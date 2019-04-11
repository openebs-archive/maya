package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Sirupsen/logrus"
	driver "github.com/openebs/maya/pkg/csidriver"
	"github.com/openebs/maya/pkg/version"
	"github.com/spf13/cobra"
)

func main() {
	_ = flag.CommandLine.Parse([]string{})
	var config = driver.NewConfig()

	cmd := &cobra.Command{
		Use:   "openebs-csi-driver",
		Short: "openebs-csi-driver",
		Run: func(cmd *cobra.Command, args []string) {
			handle(config)
		},
	}

	cmd.Flags().AddGoFlagSet(flag.CommandLine)
	cmd.PersistentFlags().StringVar(&config.Token, "token", "", "token")
	cmd.PersistentFlags().StringVar(&config.RestURL, "url", "", "url")
	cmd.PersistentFlags().StringVar(&config.NodeID, "nodeid", "node1", "node id")
	cmd.PersistentFlags().StringVar(&config.Version, "version", "", "Print the version and exit")
	cmd.PersistentFlags().StringVar(&config.Endpoint, "endpoint", "unix://csi/csi.sock", "CSI endpoint")
	cmd.PersistentFlags().StringVar(&config.DriverName, "name",
		"openebs-csi.openebs.io", "name of the driver")
	cmd.PersistentFlags().StringVar(&config.PluginType,
		"plugin", "csi-plugin", "Plugin type controller/node")

	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
}

func handle(config *driver.Config) {

	if config.Version == "" {
		config.Version = version.GetVersion()
	}

	logrus.Infof("%s - %s\n", version.GetVersion(),
		version.GetGitCommit())

	logrus.Infof("DriverName: %v Plugin: %v\nEndPoint: %v URL: %v \nNodeID: %v",
		config.DriverName, config.PluginType, config.Endpoint, config.RestURL, config.NodeID)
	drvr := driver.New(config)

	if err := drvr.Run(); err != nil {
		log.Fatalln(err)
	}
	os.Exit(0)

}
