package app

import (
	"fmt"
	"os"
	"strings"

	goflag "flag"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

var (
	cmdName = "maya-agent"
	usage   = fmt.Sprintf("%s", cmdName)
)

// Define a type for the options of MayaAgent
type MayaAgentOptions struct {
	KubeConfig string
	Namespace  string
}

func AddKubeConfigFlag(cmd *cobra.Command, value *string) {
	cmd.Flags().StringVarP(value, "kubeconfig", "", *value, "Path to a kube config. Only required if out-of-cluster.")
}

func AddNamespaceFlag(cmd *cobra.Command, value *string) {
	cmd.Flags().StringVarP(value, "namespace", "n", *value, "Namespace to deploy in. If no namespace is provided, POD_NAMESPACE env. var is used. Lastly, the 'default' namespace will be used as a last option.")
}

// Fatal prints the message (if provided) and then exits. If V(2) or greater,
// glog.Fatal is invoked for extended information.
func fatal(msg string) {
	if glog.V(2) {
		glog.FatalDepth(2, msg)
	}
	if len(msg) > 0 {
		// add newline if needed
		if !strings.HasSuffix(msg, "\n") {
			msg += "\n"
		}
		fmt.Fprint(os.Stderr, msg)
	}
	os.Exit(1)
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

// Create a new maya-agent. This cmd includes logging,
// cmd option parsing from flags
func NewMayaAgent() (*cobra.Command, error) {
	// Define the options for MayaAgent
	options := MayaAgentOptions{}

	// Create a new command
	cmd := &cobra.Command{
		Use:   usage,
		Short: "",
		Run: func(cmd *cobra.Command, args []string) {
			checkErr(Run(cmd, &options), fatal)
		},
	}

	// Bind & parse flags defined by external projects.
	// e.g. This imports the golang/glog pkg flags into the cmd flagset
	cmd.Flags().AddGoFlagSet(goflag.CommandLine)
	goflag.CommandLine.Parse([]string{})
	cmd.AddCommand(
		NewCmdOpenEBSExporter(),
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

func checkErr(err error, handleErr func(string)) {
	if err == nil {
		return
	}
	handleErr(err.Error())
}
