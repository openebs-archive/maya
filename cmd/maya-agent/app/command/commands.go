package command

import (
	goflag "flag"

	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"
)

var (
	cmdName = "maya-agent"
	usage   = cmdName
)

// MayaAgentOptions defines a type for the options of MayaAgent
type MayaAgentOptions struct {
	KubeConfig string
	Namespace  string
}

//AddKubeConfigFlag is used to add a config flag
func AddKubeConfigFlag(cmd *cobra.Command, value *string) {
	cmd.Flags().StringVarP(value, "kubeconfig", "", *value,
		"Path to a kube config. Only required if out-of-cluster.")
}

//AddNamespaceFlag is used to add a namespace flag
func AddNamespaceFlag(cmd *cobra.Command, value *string) {
	cmd.Flags().StringVarP(value, "namespace", "n", *value,
		"Namespace to deploy in. If no namespace is provided, POD_NAMESPACE env.var is used. Lastly, the 'default' namespace will be used as a last option.")
}

// NewCmdOptions creates an options Cobra command to return usage
func NewCmdOptions() *cobra.Command {
	cmd := &cobra.Command{
		Use: "options",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}

	return cmd
}

// NewMayaAgent creates a new maya-agent. This cmd includes logging,
// cmd option parsing from flags
func NewMayaAgent() (*cobra.Command, error) {
	// Define the options for MayaAgent
	options := MayaAgentOptions{}

	// Create a new command
	cmd := &cobra.Command{
		Use:   usage,
		Short: "",
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(Run(cmd, &options), util.Fatal)
		},
	}

	// Bind & parse flags defined by external projects.
	// e.g. This imports the golang/glog pkg flags into the cmd flagset
	cmd.Flags().AddGoFlagSet(goflag.CommandLine)
	goflag.CommandLine.Parse([]string{})
	cmd.AddCommand(
		NewCmdBlockDevice(), //Add new command on block device
		NewCmdIscsi(),       //Add new command for iscsi operations
	)
	// Define the flags allowed in this command & store each option provided
	// as a flag, into the MayaAgentOptions
	AddKubeConfigFlag(cmd, &options.KubeConfig)
	AddNamespaceFlag(cmd, &options.Namespace)

	return cmd, nil
}

// Run maya-agent
func Run(cmd *cobra.Command, options *MayaAgentOptions) error {
	glog.Infof("Starting maya-agent...")

	return nil
}
