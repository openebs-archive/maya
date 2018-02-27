package command

import (


	goflag "flag"

	"github.com/openebs/maya/cmd/cstor-sidecar/ccrd"
	"github.com/spf13/cobra"
)

type CmdStartOptions struct {
	kubeconfig string
}

//NewCmdStart starts watching for Cstor-CRD events
func NewCmdStart() *cobra.Command {
	options := CmdStartOptions{}
	//var target string
	getCmd := &cobra.Command{
		Use:   "start",
		Short: "crd watcher",
		Long: ` Cstor-CRD will be watched for added, updated, deleted
		events `,
		Run: func(cmd *cobra.Command, args []string) {

			//util.CheckErr(options.Validate(), util.Fatal)
			ccrd.Checkcrd(options.kubeconfig)

		},
	}

	// Bind & parse flags defined by external projects.
	// e.g. This imports the golang/glog pkg flags into the cmd flagset
	getCmd.Flags().AddGoFlagSet(goflag.CommandLine)
	goflag.CommandLine.Parse([]string{})

	getCmd.Flags().StringVar(&options.kubeconfig, "kubeconfig", "",
		`kubeconfig needs to be specified if out of cluster`)
	return getCmd
}

func (c *CmdStartOptions) Validate() error {

	return nil
}
