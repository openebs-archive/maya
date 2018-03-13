package command

import (
	goflag "flag"

	"github.com/openebs/maya/cmd/cstor-repl-ctrl/crdops"
	"github.com/spf13/cobra"
)

type CmdStartOptions struct {
	kubeconfig string
}

// NewCmdStart starts watching for Cstor-CRD events.
func NewCmdStart() *cobra.Command {
	options := CmdStartOptions{}
	getCmd := &cobra.Command{
		Use:   "start",
		Short: "starts cstorPool and cstorReplica watcher",
		Long: ` cstorPool and cstorReplica crds will be watched for added, updated, deleted
		events `,
		Run: func(cmd *cobra.Command, args []string) {
			crdops.CrdOperations(options.kubeconfig)
		},
	}

	// Bind & parse flags defined by external projects.
	// e.g. This imports the golang/glog pkg flags into the cmd flagset.
	getCmd.Flags().AddGoFlagSet(goflag.CommandLine)
	goflag.CommandLine.Parse([]string{})

	getCmd.Flags().StringVar(&options.kubeconfig, "kubeconfig", "",
		`kubeconfig needs to be specified if out of cluster`)
	return getCmd
}
