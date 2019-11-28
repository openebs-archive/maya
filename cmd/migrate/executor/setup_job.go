/*
Copyright 2019 The OpenEBS Authors.

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

package executor

import (
	"flag"
	//"fmt"
	//"os"
	"strings"

	//"k8s.io/klog"

	"github.com/spf13/cobra"
)

var spcName, openebsNamespace string

// NewJob will setup a new migrate job
func NewJob() *cobra.Command {
	// Create a new command.
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "OpenEBS Migrate Utility",
		Long: `An utility to migrate OpenEBS SPC Pools to CSPC and Non-CSI Volumes to CSI Volumes,
			run as a Kubernetes Job`,
		PersistentPreRun: PreRun,
	}

	cmd.AddCommand(
		NewMigratePoolJob(),
	)

	cmd.PersistentFlags().StringVarP(&options.openebsNamespace,
		"openebs-namespace", "",
		options.openebsNamespace,
		"namespace where openebs components are installed.")

	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	// Hack: Without the following line, the logs will be prefixed with Error
	_ = flag.CommandLine.Parse([]string{})

	return cmd
}

// PreRun will check for environement variables to be read and intialized.
func PreRun(cmd *cobra.Command, args []string) {
	namespace := getOpenEBSNamespace()
	if len(strings.TrimSpace(namespace)) != 0 {
		openebsNamespace = namespace
	}
}
