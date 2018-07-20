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

	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"
)

var (
	cmdName = "cstor-volume-mgmt"
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

// NewCStorVolumeMgmt creates a new CStorVolume CRD watcher command.
func NewCStorVolumeMgmt() (*cobra.Command, error) {
	// Create a new command.
	cmd := &cobra.Command{
		Use:   usage,
		Short: "CStor Volume Management",
		Long: `interfaces between observing the CStorVolume
		 objects and volume controller creation`,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(Run(cmd), util.Fatal)
		},
	}
	cmd.AddCommand(
		NewCmdStart(),
	)
	return cmd, nil
}

// Run is to run cstor-volume-mgmt command without any arguments
func Run(cmd *cobra.Command) error {
	glog.Infof("cstor-volume-mgmt watcher for CStorVolume objects")
	return nil
}
