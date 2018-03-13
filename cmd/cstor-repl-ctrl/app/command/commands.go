package command

import (
	goflag "flag"

	"fmt"

	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"
)

var (
	cmdName = "cstorReplCtrl"
	usage   = fmt.Sprintf("%s", cmdName)
)

// CstorReplCtrlOptions defines a type for the options of cstorReplCtrl.
type CstorReplCtrlOptions struct {
	KubeConfig string
	Namespace  string
}

// AddKubeConfigFlag is used to add a config flag.
func AddKubeConfigFlag(cmd *cobra.Command, value *string) {
	cmd.Flags().StringVarP(value, "kubeconfig", "", *value,
		"Path to a kube config. Only required if out-of-cluster.")
}

// AddNamespaceFlag is used to add a namespace flag.
func AddNamespaceFlag(cmd *cobra.Command, value *string) {
	cmd.Flags().StringVarP(value, "namespace", "n", *value,
		"Namespace to deploy in. If no namespace is provided, POD_NAMESPACE env.var is used. Lastly, the 'default' namespace will be used as a last option.")
}

// NewCmdOptions creates an options Cobra command to return usage.
func NewCmdOptions() *cobra.Command {
	cmd := &cobra.Command{
		Use: "options",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}

	return cmd
}

// NewCstorReplCtrl creates a new CstorReplCtrl. This cmd includes logging,
// cmd option parsing from flags.
func NewCstorReplCtrl() (*cobra.Command, error) {
	// Define the options for CstorReplCtrl.
	options := CstorReplCtrlOptions{}

	// Create a new command.
	cmd := &cobra.Command{
		Use:   usage,
		Short: "",
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(Run(cmd, &options), util.Fatal)
		},
	}

	// Bind & parse flags defined by external projects.
	// e.g. This imports the golang/glog pkg flags into the cmd flagset.
	cmd.Flags().AddGoFlagSet(goflag.CommandLine)
	goflag.CommandLine.Parse([]string{})
	cmd.AddCommand(
		NewCmdStart(),
	)
	// Define the flags allowed in this command & store each option provided
	// as a flag, into the CstorReplCtrlOptions.
	AddKubeConfigFlag(cmd, &options.KubeConfig)
	AddNamespaceFlag(cmd, &options.Namespace)

	return cmd, nil
}

// Run CstorReplCtrl.
func Run(cmd *cobra.Command, options *CstorReplCtrlOptions) error {
	glog.Infof("Starting CstorReplCtrl...")

	return nil
}
