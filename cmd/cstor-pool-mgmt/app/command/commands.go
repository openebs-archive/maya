/*
Copyright 2018 The OpenEBS Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package command

import (
	goflag "flag"

	"fmt"

	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"
)

var (
	cmdName = "cstor-pool-mgmt"
	usage   = fmt.Sprintf("%s", cmdName)
)

// CStorPoolMgmtOptions defines a type for the options of CStorPoolMgmt.
type CStorPoolMgmtOptions struct {
	KubeConfig string
	Namespace  string
}

// AddKubeConfigFlag is used to add a config flag.
func AddKubeConfigFlag(cmd *cobra.Command, value *string) {
	cmd.Flags().StringVarP(value, "kubeconfig", "", *value,
		"Path to a kube config. Only required if out-of-cluster.")
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

// NewCstorPoolMgmt creates a new CstorPoolMgmt. This cmd includes logging,
// cmd option parsing from flags.
func NewCstorPoolMgmt() (*cobra.Command, error) {
	// Define the options for CstorPoolMgmt.
	options := CStorPoolMgmtOptions{}

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
	// as a flag, into the CStorPoolMgmtOptions.
	AddKubeConfigFlag(cmd, &options.KubeConfig)

	return cmd, nil
}

// Run is to CStorPoolMgmt.
func Run(cmd *cobra.Command, options *CStorPoolMgmtOptions) error {
	glog.Infof("Starting cstor-pool-mgmt...")

	return nil
}
