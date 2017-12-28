package command

import (
	"errors"

	goflag "flag"

	"github.com/openebs/maya/cmd/maya-nodebot/mcrd"
	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"
)

type CmdStartOptions struct {
	kubeconfig string
}

//NewSubCmdIscsiLogin logs in to particular portal or all discovered portals
func NewCmdStart() *cobra.Command {
	options := CmdStartOptions{}
	//var target string
	getCmd := &cobra.Command{
		Use:   "start",
		Short: "crd watcher",
		Long: ` StoragePoolClaim custom resouce will be watched for added, updated, deleted
		events `,
		Run: func(cmd *cobra.Command, args []string) {

			util.CheckErr(options.Validate(), util.Fatal)
			mcrd.Checkcrd(options.kubeconfig)

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
	if c.kubeconfig == "" {
		return errors.New("--kubeconfig is missing. Please specify kubeconfig")
	}
	return nil
}
