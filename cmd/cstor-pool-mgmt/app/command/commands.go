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
	"fmt"

	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

var (
	cmdName = "cstor-pool-mgmt"
	usage   = fmt.Sprintf("%s", cmdName)
)

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

// NewCStorPoolMgmt creates a new CStorPoolMgmt. This cmd includes logging,
// cmd option parsing from flags.
func NewCStorPoolMgmt() (*cobra.Command, error) {
	// Create a new command.
	cmd := &cobra.Command{
		Use:   usage,
		Short: "CStor Pool Management",
		Long: `interfaces between observing the CStorPool, CStorVolumeReplica
		 objects and issues pool-volume creation and deletion`,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(Run(cmd), util.Fatal)
		},
	}
	cmd.AddCommand(
		NewCmdStart(),
	)
	return cmd, nil
}

// Run is to CStorPoolMgmt.
func Run(cmd *cobra.Command) error {
	klog.Infof("cstor-pool-mgmt watcher for CStorPool and CStorVolumeReplica objects")
	return nil
}
